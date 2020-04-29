package local

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/uhthomas/kipp/filesystem"
)

// A FileSystem contains information about the local filesystem.
type FileSystem struct{ dir, tmp string }

// New creates a new FileSystem, and makes the relevant directories for
// dir and tmp.
func New(dir, tmp string) (*FileSystem, error) {
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(tmp, 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &FileSystem{dir: dir, tmp: tmp}, nil
}

// Create creates a temporary writer, and wraps it so when the writer is closed,
// it is moved to name.
func (fs FileSystem) Create(_ context.Context, name string) (filesystem.Writer, error) {
	f, err := ioutil.TempFile(fs.tmp, "kipp")
	if err != nil {
		return nil, fmt.Errorf("temp writer: %w", err)
	}
	return &writer{f: f, name: filepath.Join(fs.dir, name)}, nil
}

// Open opens the named writer.
func (fs FileSystem) Open(_ context.Context, name string) (filesystem.Reader, error) {
	return os.Open(filepath.Join(fs.dir, name))
}

// Remove removes the named writer.
func (fs FileSystem) Remove(_ context.Context, name string) error {
	return os.Remove(name)
}
