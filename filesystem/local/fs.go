package local

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/uhthomas/kipp/filesystem"
)

type FileSystem struct{ path, tempPath string }

func New(path, tempPath string) *FileSystem {
	return &FileSystem{path: path, tempPath: tempPath}
}

// Create creates a temporary file, and wraps it so when the file is closed,
// it is moved to name.
//
// there's some oddities we have to consider:
// the way kipp works at the moment, we upload to a temporary file and then call
// os.Link, renaming the temporary file to that of the sum of the content. This
// is great for deduplication, but is it necessary?
func (fs FileSystem) Create(_ context.Context, name string) (io.WriteCloser, error) {
	f, err := ioutil.TempFile(fs.tempPath, "kipp")
	if err != nil {
		return nil, fmt.Errorf("temp file: %w", err)
	}
	return &tempFile{File: f, name: filepath.Join(fs.path, name)}, nil
}

// Open opens the named file.
func (fs FileSystem) Open(_ context.Context, name string) (filesystem.ReadSeekCloser, error) {
	return os.Open(filepath.Join(fs.path, name))
}

// Remove removes the named file.
func (fs FileSystem) Remove(_ context.Context, name string) error {
	return os.Remove(name)
}
