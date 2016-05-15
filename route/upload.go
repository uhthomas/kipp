package route

import (
	"conf/crypto"
	"conf/model"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (s *server) Upload(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > s.UploadSize {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
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
	tf, err := ioutil.TempFile(filepath.Join("_", "tmp"), "conf-upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tf.Close()
	k, err := crypto.Random(48)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	k, iv, h := k[:32], k[32:], sha256.New()
	e, err := crypto.NewEncrypter(tf, k, iv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, err := io.Copy(io.MultiWriter(e, h), http.MaxBytesReader(w, f, s.UploadSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hs := hex.EncodeToString(h.Sum(nil))
	var c model.Content
	if model.DB.First(&c, "hash = ?", hs).RecordNotFound() {
		if err := os.Rename(tf.Name(), filepath.Join("_", "files", hs)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		k, iv = c.Key, c.IV
		defer os.Remove(tf.Name())
	}
	c = model.Content{
		Name:      name,
		Extension: filepath.Ext(name),
		Hash:      hs,
		Size:      uint64(n),
		Key:       k,
		IV:        iv,
	}
	if r.URL.Query().Get("permanent") != "true" {
		e := time.Now().Add(24 * time.Hour)
		c.Expires = &e
	}
	if err := model.DB.Create(&c).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&c)
}
