package local

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/uhthomas/kipp/fs"
)

type FileSystem string

func New(path string) FileSystem {
	return FileSystem(path)
}

// there's some oddities we have to consider:
// the way kipp works at the moment, we upload to a temporary file and then call
// os.Link, renaming the temporary file to that of the sum of the content. This
// is great for deduplication, but is it necessary?
func (fs FileSystem) Create(_ context.Context, name string) (io.WriteCloser, error) {
	return os.Create(filepath.Join(string(fs), name))
}

func (fs FileSystem) Open(_ context.Context, name string) (fs.ReadSeekCloser, error) {
	return os.Open(filepath.Join(string(fs), name))
}
