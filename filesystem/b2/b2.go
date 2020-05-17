package b2

import (
	"context"
	"fmt"
	"io"

	"github.com/kurin/blazer/b2"
	"github.com/uhthomas/kipp/filesystem"
)

type FileSystem struct {
}

func New(ctx context.Context, account, key string) (*FileSystem, error) {
	b2.NewClient(ctx, account, key)
	return nil, nil
}

func (fs *FileSystem) Create(ctx context.Context, name string) (filesystem.Writer, error) {
	return nil, nil
}

func (fs *FileSystem) Open(ctx context.Context, name string) (filesystem.Reader, error) {
	r := &reader{
		ctx:  ctx,
		size: 0, // ??? :(
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}
	return r, nil
}

func (fs *FileSystem) Remove(ctx context.Context, name string) error {
	return nil
}
