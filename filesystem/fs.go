package filesystem

import (
	"context"
	"io"
)

// A FileSystem is a persistent store of objects uniquely identified by name.
type FileSystem interface {
	Create(ctx context.Context, name string) (Writer, error)
	Open(ctx context.Context, name string) (Reader, error)
	Remove(ctx context.Context, name string) error
}

type Writer interface {
	io.WriteCloser
	// Sync flushes the data to persistent storage. Sync must be called
	// be called before Close, otherwise the implementation should abort
	// the write.
	Sync() error
}

type Reader interface {
	io.ReadSeeker
	io.Closer
}
