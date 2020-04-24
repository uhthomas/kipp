package fs

import (
	"context"
	"io"
)

// A FileSystem is a persistent store of objects uniquely identified by name.
type FileSystem interface {
	Create(ctx context.Context, name string) (io.WriteCloser, error)
	Open(ctx context.Context, name string) (ReadSeekCloser, error)
	Remove(ctx context.Context, name string) error
}

type ReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
}
