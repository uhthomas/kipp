package kipp

import (
	"net/http"
	"os"
	"time"
)

// fileSystemFunc implements http.FileSystem
type fileSystemFunc func(string) (http.File, error)

func (f fileSystemFunc) Open(name string) (http.File, error) { return f(name) }

// file will cache a file's FileInfo
type file struct {
	http.File
	modTime time.Time
}

func (f file) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if err == nil {
		fi = fileInfo{fi, f.modTime}
	}
	return fi, err
}

type fileInfo struct {
	os.FileInfo
	modTime time.Time
}

func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}
