package route

import (
	"conf/crypto"
	"conf/model"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

func (s *server) View(w http.ResponseWriter, r *http.Request) {
	slug := strings.Split(mux.Vars(r)["slug"], ".")[0]
	var c model.Content
	if model.DB.First(&c, "slug = ?", slug).RecordNotFound() {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	f, err := os.Open(filepath.Join("_", "files", c.Hash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d, err := crypto.NewDecrypter(f, c.Key, c.IV)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%q", c.Name))
	http.ServeContent(w, r, c.Name, c.CreatedAt, d)
}
