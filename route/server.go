package route

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

type server struct {
	UploadSize int64
}

func Listen() error {
	s, r := &server{150 << 20}, mux.NewRouter()
	r.HandleFunc("/c{slug}", s.View)
	r.HandleFunc("/_/upload", s.Upload).Methods("POST")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(filepath.Join("_", "public")))))
	return http.ListenAndServe(":1337", r)
}
