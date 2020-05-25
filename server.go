package kipp

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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

	"github.com/uhthomas/kipp/database"
	"github.com/uhthomas/kipp/filesystem"
	"github.com/zeebo/blake3"
)

// Server acts as the HTTP server and configuration.
type Server struct {
	Database   database.Database
	FileSystem filesystem.FileSystem

	Lifetime   time.Duration
	Limit      int64
	PublicPath string
}

// ServeHTTP will serve HTTP requests. It first tries to determine if the
// request is for uploading, it then tries to serve static files and then will
// try to serve public files.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	case http.MethodPost:
		if r.URL.Path == "/" {
			s.UploadHandler(w, r)
			return
		}
		fallthrough
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
		if f, err := http.Dir(s.PublicPath).Open(name); !os.IsNotExist(err) {
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

		// trim anything after the first "."
		if i := strings.Index(name, "."); i > -1 {
			name = name[:i]
		}

		e, err := s.Database.Lookup(r.Context(), name)
		if err != nil {
			if errors.Is(err, database.ErrNoResults) {
				return nil, os.ErrNotExist
			}
			return nil, err
		}

		cache := "max-age=31536000" // ~ 1 year
		if e.Lifetime != nil {
			now := time.Now()
			if e.Lifetime.Before(now) {
				return nil, os.ErrNotExist
			}
			cache = fmt.Sprintf(
				"public, must-revalidate, max-age=%d",
				int(e.Lifetime.Sub(now).Seconds()),
			)
		}

		f, err := s.FileSystem.Open(r.Context(), e.Slug)
		if err != nil {
			return nil, err
		}

		// Detect content type before serving content to filter html files
		ctype := mime.TypeByExtension(filepath.Ext(e.Name))
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
			url.PathEscape(e.Name),
		))
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Etag", strconv.Quote(e.Sum))
		if e.Lifetime != nil {
			w.Header().Set("Expires", e.Lifetime.Format(http.TimeFormat))
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		return &file{Reader: f, entry: e}, nil
	})).ServeHTTP(w, r)
}

// UploadHandler write the contents of the "file" part to a filesystem.Reader,
// persists the entry to the database and writes the location of the file
// to the response.
func (s Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Due to the overhead of multipart bodies, the actual limit for files
	// is smaller than it should be. It's not really feasible to calculate
	// the overhead so this is *good enough* for the time being.
	//
	// TODO(thomas): is there a better way to limit the size for the
	//      part, rather than the whole body?
	if r.ContentLength > s.Limit {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.Limit)

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var p *multipart.Part
	for {
		if p, err = mr.NextPart(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

	// 9 bytes as base64 is most efficient when aligned to len(b) % 3
	var b [9]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slug := base64.RawURLEncoding.EncodeToString(b[:])

	f, err := s.FileSystem.Create(r.Context(), slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	h := blake3.New()
	n, err := io.Copy(io.MultiWriter(f, h), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := database.Entry{
		Slug:      slug,
		Name:      name,
		Sum:       base64.RawURLEncoding.EncodeToString(h.Sum(nil)),
		Size:      n,
		Timestamp: time.Now(),
	}

	if s.Lifetime > 0 {
		l := e.Timestamp.Add(s.Lifetime)
		e.Lifetime = &l
	}

	if err := s.Database.Create(r.Context(), e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := f.Sync(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(name)

	var buf strings.Builder
	buf.Grow(len(slug) + len(ext) + 2)
	buf.WriteRune('/')
	buf.WriteString(slug)
	buf.WriteString(ext)

	http.Redirect(w, r, buf.String(), http.StatusSeeOther)

	buf.WriteRune('\n')
	_, _ = w.Write([]byte(buf.String()))
}
