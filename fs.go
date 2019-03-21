package kipp

import (
	"net/http"
	"os"
	"time"
)

// fileSystemFunc implements http.FileSystem.
type fileSystemFunc func(string) (http.File, error)

func (f fileSystemFunc) Open(name string) (http.File, error) { return f(name) }

// file wraps http.File to provide correct Last-Modified times.
type file struct {
	http.File
	modTime time.Time
}

func (f file) Stat() (os.FileInfo, error) {
	d, err := f.File.Stat()
	if err == nil {
		d = fileInfo{d, f.modTime}
	}
	return d, err
}

type fileInfo struct {
	os.FileInfo
	modTime time.Time
}

func (d fileInfo) ModTime() time.Time {
	return d.modTime
}
