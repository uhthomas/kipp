package conf

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/6f7262/conf/ctr"
	"github.com/jinzhu/gorm"
	blake2b "github.com/minio/blake2b-simd"
)

// Server is used to serve http requests and acts as a config.
type Server struct {
	DB         *gorm.DB
	Expiration time.Duration
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
	// TODO: version?
	w.Header().Set("Server", "conf")
	hasPrefix := func(s string) bool {
		return strings.HasPrefix(r.URL.Path, s)
	}
	switch {
	case r.URL.Path == "/upload":
		s.UploadHandler(w, r)
	case r.URL.Path == "/", hasPrefix("/css"), hasPrefix("/fonts"), hasPrefix("/js"):
		s.StaticHandler(w, r)
	default:
		s.ContentHandler(w, r)
	}
}

// StaticHandler will server static content given the url path.
func (s Server) StaticHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		return
	case r.Method != http.MethodHead && r.Method != http.MethodGet:
		w.Header().Set("Allow", "GET, HEAD, OPTIONS")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, path.Join(s.PublicPath, r.URL.Path))
}

// ContentHandler will query the database for the given slug. If the slug doesn't exist it
// will return 404 otherwise it will decrypt the file and serve it.
func (s Server) ContentHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		return
	case r.Method != http.MethodHead && r.Method != http.MethodGet:
		w.Header().Set("Allow", "GET, HEAD, OPTIONS")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// split the path to allow for extensions
	slug := strings.Split(r.URL.Path, ".")[0][1:]
	b, err := base64.RawURLEncoding.DecodeString(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// fragment:10 | sum:16 = 26
	if len(b) != 26 {
		http.NotFound(w, r)
		return
	}
	var c Content
	if s.DB.First(&c, "fragment = ?", b[:10]).RecordNotFound() {
		http.NotFound(w, r)
		return
	}
	// remove fragment:10 from b to make it sum:16
	b = b[10:]
	// verify sum is valid
	h := hmac.New(sha256.New, c.Secret)
	h.Write(b)
	if !hmac.Equal(h.Sum(nil), c.MAC) {
		http.NotFound(w, r)
		return
	}
	k := make([]byte, 16)
	for i := 0; i < 16; i++ {
		k[i] = c.Secret[i] ^ b[i]
	}
	f, err := os.Open(filepath.Join(s.FilePath, c.Hash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cr, err := ctr.NewReader(f, k, c.IV)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Detect content type before serving content to filter html files
	ctype := mime.TypeByExtension(filepath.Ext(c.Name))
	if ctype == "" {
		var b [512]byte
		n, _ := io.ReadFull(cr, b[:])
		ctype = http.DetectContentType(b[:n])
		if _, err := cr.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "seeker can't seek", http.StatusInternalServerError)
			return
		}
	}
	// catches text/html and text/html; charset=utf-8
	if strings.HasPrefix(ctype, "text/html") {
		ctype = "text/plain; charset=utf-8"
	}
	// 1 year
	cache := "31536000"
	if e := c.Expires; e != nil {
		// duration in seconds until expiration
		d := int(time.Until(*e).Seconds())
		// if expired then the request should return 404 and send the content to
		// the cleanup worker -- for now we won't do anything and we'll let the
		// worker clean up content in its own time.
		if d > 0 {
			cache = fmt.Sprintf("private, must-revalidate, max-age=%d", d)
		} else {
			cache = "no-cache"
		}
	}
	w.Header().Set("Cache-Control", cache)
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%q", c.Name))
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Etag", strconv.Quote(c.Hash))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, c.Name, c.CreatedAt, cr)
}

// UploadHandler serves as a handler for uploading files to conf. It will
// generate a random key and iv then both hash and encrypt the body. After that,
// conf generates a secret (sum[32:48] ^ key) as well as a MAC using the
// secret as the key and sum[32:48] as the body. The expiration date and path is
// then sent to the client in JSON form.
func (s Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST")
		return
	case r.Method != http.MethodPost:
		w.Header().Set("Allow", "OPTIONS, POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	case r.ContentLength > s.Max:
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
	if name == "" || len(name) > 255 {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
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
	// Generate a random key and iv. key:16 | iv:16 = 32
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	k, iv, h := k[:16], k[16:], blake2b.New512()
	cw, err := ctr.NewWriter(tf, k, iv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Hash and save the file.
	n, err := io.Copy(io.MultiWriter(cw, h), http.MaxBytesReader(w, p, s.Max))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sum := h.Sum(nil)[:48]
	// Find the content
	c := Content{Hash: base64.RawURLEncoding.EncodeToString(sum[:32])}
	// remove hash:32 from sum
	sum = sum[32:]
	if s.DB.First(&c, "hash = ?", c.Hash).RecordNotFound() {
		p := filepath.Join(s.FilePath, c.Hash)
		if err := os.Rename(tf.Name(), p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c.IV = iv
		c.Secret = make([]byte, 16)
		for i := 0; i < 16; i++ {
			c.Secret[i] = k[i] ^ sum[i]
		}
		h := hmac.New(sha256.New, c.Secret)
		h.Write(sum)
		c.MAC = h.Sum(nil)
	}
	c = Content{
		Hash:   c.Hash,
		Name:   name,
		Size:   n,
		Secret: c.Secret,
		IV:     c.IV,
		MAC:    c.MAC,
	}
	if s.Expiration > 0 {
		e := time.Now().Add(s.Expiration)
		c.Expires = &e
	}
	if err := s.DB.Create(&c).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var b bytes.Buffer
	e := base64.NewEncoder(base64.RawURLEncoding, &b)
	e.Write(c.Fragment)
	e.Write(sum)
	e.Close()
	if ext := filepath.Ext(c.Name); ext != "" {
		b.WriteString(ext)
	}
	json.NewEncoder(w).Encode(struct {
		Expires *time.Time `json:"expires,omitempty"`
		Path    string     `json:"path"`
	}{c.Expires, b.String()})
}
