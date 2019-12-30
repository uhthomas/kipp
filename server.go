package kipp

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
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

	"github.com/boltdb/bolt"
	"github.com/minio/blake2b-simd"
)

// Server acts as the HTTP server, configuration and provides essential core
// functions such as Cleanup.
type Server struct {
	DB         *bolt.DB
	Expiration time.Duration
	Max        int64
	FilePath   string
	TempPath   string
	PublicPath string
}

// Cleanup will delete expired files and remove files associated with it as
// long as it is not used by any other files.
func (s Server) Cleanup() (err error) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(time.Now().Unix()))

	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer func() {
		if e := tx.Commit(); err == nil {
			err = e
		}
	}()

	c := tx.Bucket([]byte("ttl")).Cursor()
	for k, v := c.First(); k != nil && bytes.Compare(k, b[:]) <= 0; k, v = c.Next() {
		if err := c.Delete(); err != nil {
			return err
		}
		sum := base64.RawURLEncoding.EncodeToString(v)
		if err := os.Remove(filepath.Join(s.FilePath, sum)); err != nil {
			return err
		}
	}
	return nil
}

// ServeHTTP will serve HTTP requests. It first tries to determine if the
// request is for uploading, it then tries to serve static files and then will
// try to serve public files.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		f, err := http.Dir(s.PublicPath).Open(name)
		if !os.IsNotExist(err) {
			d, err := f.Stat()
			if err != nil {
				return nil, err
			}
			if !d.IsDir() {
				w.Header().Set("Cache-Control", "max-age=31536000")
				// nginx style weak Etag
				w.Header().Set("Etag", fmt.Sprintf(
					`W/"%x-%x"`,
					d.ModTime().Unix(), d.Size(),
				))
			}
			return f, nil
		}
		dir, name := path.Split(name)
		if dir != "/" {
			return nil, os.ErrNotExist
		}
		// trim anything after "."
		if i := strings.Index(name, "."); i > -1 {
			name = name[:i]
		}
		var out struct {
			Checksum  string
			CreatedAt time.Time
			Expires   *time.Time
			Name      string
			Size      uint64
		}
		if err := s.DB.View(func(tx *bolt.Tx) error {
			return gob.NewDecoder(bytes.NewReader(tx.Bucket([]byte("files")).Get([]byte(name)))).Decode(&out)
		}); err == io.EOF {
			return nil, os.ErrNotExist
		} else if err != nil {
			return nil, err
		}

		// 1 year
		cache := "max-age=31536000"
		if out.Expires != nil {
			// duration in seconds until expiration
			d := int(time.Until(*out.Expires).Seconds())
			if d <= 0 {
				// catch expired files. the cleanup worker should delete the
				// file on its own at some point
				return nil, os.ErrNotExist
			}
			cache = fmt.Sprintf("public, must-revalidate, max-age=%d", d)
		}
		f, err = os.Open(filepath.Join(s.FilePath, out.Checksum))
		if err != nil {
			// // looks weird, but we don't want the file server to serve 404.
			// // this is a fatal error and should never happen
			// if os.IsNotExist(err) {
			// 	err = errors.New(err.Error())
			// }
			return nil, err
		}
		// Detect content type before serving content to filter html files
		ctype := mime.TypeByExtension(filepath.Ext(out.Name))
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
			url.PathEscape(out.Name),
		))
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Etag", strconv.Quote(out.Checksum))
		if out.Expires != nil {
			w.Header().Set("Expires", out.Expires.Format(http.TimeFormat))
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		return file{f, out.CreatedAt}, nil
	})).ServeHTTP(w, r)
}

// UploadHandler will read the request body and write it to the disk whilst also
// calculating a blake2b checksum. It will then insert the file information
// into the database and if the file doesn't already exist, it will be moved
// into the FilePath. It will then return StatusSeeOther with the location
// of the file.
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
	sum := h.Sum(nil)
	// 9 byte ID as base64 is most efficient when it aligns to len(b) % 3
	var b [9]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := base64.RawURLEncoding.EncodeToString(b[:])
	now := time.Now().UTC()
	data := struct {
		Checksum  string
		CreatedAt time.Time
		Expires   *time.Time
		Name      string
		Size      uint64
	}{base64.RawURLEncoding.EncodeToString(sum), now, nil, name, uint64(n)}
	if s.Expiration > 0 {
		e := time.Now().Add(s.Expiration)
		data.Expires = &e
	}
	if err := s.DB.Update(func(tx *bolt.Tx) error {
		p := filepath.Join(s.FilePath, data.Checksum)
		d, err := os.Stat(p)
		// switch {
		// // If the file exists then we should delete the current ttl (if any)
		// case err == nil:
		// 	var b [8]byte
		// 	binary.BigEndian.PutUint64(b[:], uint64(d.ModTime().Unix()))
		// 	if err := tx.Bucket([]byte("ttl")).Delete(append(b[:], sum...)); err != nil {
		// 		return err
		// 	}
		// // if the file doesn't exist then it should be created
		// case os.IsNotExist(err):
		// 	if err := os.Rename(tf.Name(), p); err != nil {
		// 		return err
		// 	}
		// default:
		// 	return err
		// }

		if os.IsNotExist(err) {
			if err := os.Rename(tf.Name(), p); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		var ttl time.Time
		if err == nil {
			ttl = d.ModTime()
		}
		// if the ttl is valid
		if ttl.IsZero() || ttl.Unix() != 0 {
			if data.Expires == nil {
				// and the new ttl is not then invalidate it
				ttl = time.Unix(0, 0)
			} else if e := *data.Expires; ttl.Before(e) {
				// and the current ttl is before the new ttl
				ttl = e
			}
		}

		// if the file exists, the ttl is valid and the old ttl is before the new ttl
		if err == nil && d.ModTime().Unix() != 0 && d.ModTime().Before(ttl) {
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(d.ModTime().Unix()))
			if err := tx.Bucket([]byte("ttl")).Delete(append(b[:], sum...)); err != nil {
				return err
			}
		}

		// if the file doesn't exist or the new ttl is not equal to the current ttl then update it
		if os.IsNotExist(err) || !d.ModTime().Equal(ttl) {
			if err := os.Chtimes(p, now, ttl); err != nil {
				return err
			}

			// if the new ttl is valid then add it to the bucket
			if ttl.Unix() != 0 {
				var b [8]byte
				binary.BigEndian.PutUint64(b[:], uint64(ttl.Unix()))
				if err := tx.Bucket([]byte("ttl")).Put(append(b[:], sum...), sum); err != nil {
					return err
				}
			}
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			return err
		}
		return tx.Bucket([]byte("files")).Put([]byte(id), buf.Bytes())
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(name)

	var buf strings.Builder
	buf.Grow(len(id) + len(ext) + 2)
	buf.WriteRune('/')
	buf.WriteString(id)
	buf.WriteString(ext)

	http.Redirect(w, r, buf.String(), http.StatusSeeOther)

	buf.WriteRune('\n')
	_, _ = w.Write([]byte(buf.String()))
}
