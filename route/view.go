package route

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/6f7262/conf/crypto"
	"github.com/6f7262/conf/model"

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
	cr, err := crypto.NewReader(f, c.Key, c.IV)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c.Expires != nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", int(math.Abs(time.Since(*c.Expires).Seconds()))))
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%q", c.Name))
	w.Header().Set("Etag", strconv.Quote(c.Hash))
	http.ServeContent(w, r, c.Name, c.CreatedAt, cr)
}
