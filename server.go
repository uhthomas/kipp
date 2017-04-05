package conf

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/6f7262/conf/ctr"
	blake2b "github.com/minio/blake2b-simd"
)

type Server struct {
	DB         *DB
	Max        int64
	FilePath   string
	TempPath   string
	PublicPath string
}

// Cleanup will delete expired content and remove files associated with it as
// long as it is not used by any other content.
func (s Server) Cleanup() error {
	if err := s.DB.Delete(&Content{}, "expires < ?", time.Now()).Error; err != nil {
		return err
	}
	return filepath.Walk(s.FilePath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() || filepath.Dir(path) != s.FilePath {
			return nil
		}
		if !s.DB.Where("hash = ?", f.Name()).Find(&Content{}).RecordNotFound() {
			return nil
		}
		return os.Remove(filepath.Join(s.FilePath, f.Name()))
	})
}

// ServeHTTP will serve HTTP requests. /, /css, /fonts, /js and /upload are all
// static routes and any other route is considered a request for content.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hasPrefix := func(s string) bool {
		return strings.HasPrefix(r.URL.Path, s)
	}
	switch {
	case hasPrefix("/upload"):
		s.UploadHandler(w, r)
	case r.URL.Path == "/", hasPrefix("/css"), hasPrefix("/fonts"), hasPrefix("/js"):
		http.ServeFile(w, r, path.Join(s.PublicPath, r.URL.Path))
	default:
		s.ContentHandler(w, r)
	}
}

// Content will query the database for the given slug. If the slug doesn't exist it
// will return 404 otherwise it will decrypt the file and serve it.
func (s Server) ContentHandler(w http.ResponseWriter, r *http.Request) {
	// split the path to allow for extensions
	slug := strings.Split(r.URL.Path, ".")[0][1:]
	var c Content
	if s.DB.First(&c, "slug = ?", slug).RecordNotFound() {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(filepath.Join(s.FilePath, c.Hash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cr, err := ctr.NewReader(f, c.Key, c.IV)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c.Expires != nil {
		d := int(math.Abs(time.Since(*c.Expires).Seconds()))
		w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", d))
	}
	// Detect content type before serving content to filter html files
	ctype := mime.TypeByExtension(filepath.Ext(c.Name))
	if ctype == "" {
		var b [512]byte
		n, _ := io.ReadFull(cr, b[:])
		ctype = http.DetectContentType(b[:n])
		if _, err := cr.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "seeker can't seek", http.StatusInternalServerError)
		}
	}
	// catches text/html and text/html; charset=utf-8
	if strings.HasPrefix(ctype, "text/html") {
		ctype = "text/plain; charset=utf-8"
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%q", c.Name))
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Etag", strconv.Quote(c.Hash))
	http.ServeContent(w, r, c.Name, c.CreatedAt, cr)
}

// Upload serves as a handler for uploading files to conf. It will read the body
// of a http request, generate a blake2 hash and generate a random key and iv,
// encrypting it using conf/ctr. It will then create the content model
// and insert it into conf's database before returning that model as the
// response.
func (s Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Reject invalid requests
	switch {
	case r.Method == http.MethodOptions:
		return
	case r.Method != http.MethodPost:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	case r.ContentLength > s.Max:
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	// Find the multipart body to read from.
	var (
		f    io.ReadCloser
		name string
	)
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if name = p.FileName(); name != "" && p.FormName() == "file" {
			f = p
			defer f.Close()
			break
		}
	}
	// Create temporary file to be used for storing uploads.
	tf, err := ioutil.TempFile(s.TempPath, "conf-upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If the file is uploaded successfully and renamed this operation will fail.
	defer os.Remove(tf.Name())
	defer tf.Close()
	// Generate a random key and iv. key + iv + slug = 32 + 16 + 10
	k := make([]byte, 58)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	k, iv, slug, h := k[:32], k[32:48], k[48:], blake2b.New256()
	cw, err := ctr.NewWriter(tf, k, iv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Hash the file then create and save the content model.
	n, err := io.Copy(io.MultiWriter(cw, h), http.MaxBytesReader(w, f, s.Max))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hs := hex.EncodeToString(h.Sum(nil))
	var c Content
	if s.DB.First(&c, "hash = ?", hs).RecordNotFound() {
		if err := os.Rename(tf.Name(), filepath.Join(s.FilePath, hs)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		k, iv = c.Key, c.IV
	}
	c = Content{
		Name:      name,
		Extension: filepath.Ext(name),
		Slug:      hex.EncodeToString(slug),
		Hash:      hs,
		Size:      n,
		Key:       k,
		IV:        iv,
	}
	if r.URL.Query().Get("permanent") != "true" {
		e := time.Now().Add(24 * time.Hour)
		c.Expires = &e
	}
	if err := s.DB.Create(&c).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&c)
}
