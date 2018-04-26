package kipp

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/minio/blake2b-simd"
)

// Server acts as the HTTP server, configuration and provides essential core
// functions such as Cleanup.
type Server struct {
	DB          *sql.DB
	Encoding    Encoding
	Expiration  time.Duration
	Max         int64
	FilePath    string
	TempPath    string
	PublicPath  string
	ProxyHeader string
}

// Cleanup will delete expired content and remove files associated with it as
// long as it is not used by any other content.
func (s Server) Cleanup() error {
	if _, err := s.DB.Exec("DELETE FROM files WHERE expires < ?", time.Now()); err != nil {
		return err
	}
	return filepath.Walk(s.FilePath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() || filepath.Dir(path) != filepath.Clean(s.FilePath) {
			return nil
		}
		var exists bool
		if s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM files WHERE checksum = ?)", f.Name()).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		return os.Remove(path)
	})
}

// ServeHTTP will serve HTTP requests. It first tries to determine if the
// request is for uploading, it then tried to serve static files and then will
// try to serve content.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: version?
	w.Header().Set("Server", "conf")
	if r.URL.Path == "/" && r.Method == http.MethodPost {
		s.UploadHandler(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		allow := "GET, HEAD, OPTIONS"
		if r.URL.Path == "/" {
			allow = "GET, HEAD, OPTIONS, POST"
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", allow)
		} else {
			w.Header().Set("Allow", allow)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}
	http.FileServer(fileSystemFunc(func(name string) (http.File, error) {
		fmt.Println(name)
		f, err := http.Dir(s.PublicPath).Open(name)
		if !os.IsNotExist(err) {
			f = &file{File: f}
			fi, err := f.Stat()
			if err != nil {
				return nil, err
			}
			if !fi.IsDir() {
				w.Header().Set("Cache-Control", "max-age=31536000")
				// nginx style weak Etag
				w.Header().Set("Etag", fmt.Sprintf(
					`W/"%x-%x"`,
					fi.ModTime().Unix(),
					fi.Size(),
				))
			}
			return f, err
		}
		// we don't need to do do path.Clean as htto.FileServer.ServeHTTP does
		// this for us. It also ensures it begins with a / so we can use that to
		// determine which directory we're in
		dir, name := path.Split(name)
		if dir != "/" {
			return nil, os.ErrNotExist
		}
		if i := strings.Index(name, "."); i > -1 {
			name = name[:i]
		}
		var (
			checksum  string
			createdAt time.Time
			expires   *time.Time
		)
		if err := s.DB.QueryRow(
			"SELECT checksum, created_at, expires, name FROM files WHERE id = ?",
			name,
		).Scan(&checksum, &createdAt, &expires, &name); err != nil {
			if err == sql.ErrNoRows {
				return nil, os.ErrNotExist
			}
			return nil, err
		}
		// 1 year
		cache := "max-age=31536000"
		if expires != nil {
			// duration in seconds until expiration
			d := int(time.Until(*expires).Seconds())
			if d > 0 {
				cache = fmt.Sprintf("public, must-revalidate, max-age=%d", d)
			} else {
				// catch expired files. the cleanup worker should delete the
				// file on its own at some point
				return nil, os.ErrNotExist
			}
		}
		f, err = os.Open(filepath.Join(s.FilePath, checksum))
		if err != nil {
			// looks weird, but we don't want the file server to serve 404.
			// this is a fatal error and should never happen
			if os.IsNotExist(err) {
				err = errors.New(err.Error())
			}
			return nil, err
		}
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		f = &file{f, fileInfo{fi, createdAt}}
		// Detect content type before serving content to filter html files
		ctype := mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			var b [512]byte
			n, _ := io.ReadFull(f, b[:])
			ctype = http.DetectContentType(b[:n])
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, errors.New("seeker can't seek")
			}
		}
		// catches text/html and text/html; charset=utf-8
		const prefix = "text/html"
		if strings.HasPrefix(ctype, prefix) {
			ctype = "text/plain" + ctype[len(prefix):]
		}
		w.Header().Set("Cache-Control", cache)
		w.Header().Set("Content-Disposition", fmt.Sprintf(
			"filename=%q; filename*=UTF-8''%[1]s",
			url.PathEscape(name),
		))
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Etag", strconv.Quote(checksum))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		return f, nil
	})).ServeHTTP(w, r)
	//
	f, err := http.Dir(s.PublicPath).Open(name)
	if !os.IsNotExist(err) {
		http.FileServer(fileSystemFunc(func(name string) (http.File, error) {
			return f, nil
		})).ServeHTTP(w, r)
		return
	}
	// name, etag, expires, file
}

// UploadHandler will read the request body and write it to the disk whilst also
// calculating a blake2b checksum. It will then insert the content information
// into the database and if the file doesn't already exist, it will be moved
// into the FilePath. It will then return the expiration date and URL to the
// client.
func (s Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > s.Max {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	// Find the multipart body to read from.
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var p *multipart.Part
	for {
		p, err = mr.NextPart()
		if err == io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p.FormName() == "file" {
			break
		}
	}
	defer p.Close()
	name := p.FileName()
	if len(name) > 255 {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	// Create temporary file to be used for storing uploads.
	tf, err := ioutil.TempFile(s.TempPath, "kipp-upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If the file is uploaded successfully and renamed this operation will fail deliberately.
	defer os.Remove(tf.Name())
	defer tf.Close()
	// Hash and save the file.
	h := blake2b.New512()
	n, err := io.Copy(io.MultiWriter(tf, h), http.MaxBytesReader(w, p, s.Max))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// base64 for checksum encoding since it's slightly more compact than base32 and
	// is unlikely to be read by humans
	sum := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	var expires *time.Time
	if s.Expiration > 0 {
		e := time.Now().Add(s.Expiration)
		expires = &e
	}
	// 9 byte ID as base64 is most efficient when it aligns to len(b) % 3
	b := make([]byte, 9)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := base64.RawURLEncoding.EncodeToString(b)
	tx := s.DB.Begin()
	if _, err := tx.Exec("INSERT INTO files (checksum, expires, id, name, size) VALUES (?, ?, ?, ?, ?) ", sum, expires, id, name, n); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var exists bool
	if err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM files WHERE checksum = ?)", sum).Scan(&exists); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		p := filepath.Join(s.FilePath, sum)
		if err := os.Rename(tf.Name(), p); err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(struct {
		Expires *time.Time `json:"expires,omitempty"`
		Path    string     `json:"path"`
	}{expires, id + filepath.Ext(name)})
}
