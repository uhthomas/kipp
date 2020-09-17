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

// A FileSystem contains information about the local filesystem.
type FileSystem struct{ dir, tmp string }

// New creates a new FileSystem, and makes the relevant directories for
// dir and tmp.
func New(dir string) (*FileSystem, error) {
	tmp := filepath.Join(dir, "tmp")
	if err := os.MkdirAll(tmp, 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &FileSystem{dir: dir, tmp: tmp}, nil
}

// Create writer r to a temporary file, and links it to a permanent location
// upon success.
func (fs FileSystem) Create(_ context.Context, name string, r io.Reader) error {
	f, err := ioutil.TempFile(fs.tmp, "kipp")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	if err := os.Link(f.Name(), filepath.Join(fs.dir, name)); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	return nil
}

// Open opens the named file.
func (fs FileSystem) Open(_ context.Context, name string) (filesystem.Reader, error) {
	return os.Open(filepath.Join(fs.dir, name))
}

// Remove removes the named file.
func (fs FileSystem) Remove(_ context.Context, name string) error {
	return os.Remove(name)
}
