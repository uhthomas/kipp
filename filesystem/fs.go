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

// A Writer is a writable stream for persisting files. Calls to Close should
// cleanup the file if Sync has not been called.
type Writer interface {
	io.WriteCloser
	// Sync flushes the data to persistent storage. Sync must be called
	// be called before Close, otherwise the implementation should abort
	// the write.
	Sync() error
}

// A Reader is a readable, seekable and closable file stream.
type Reader interface {
	io.ReadSeeker
	io.Closer
}

// The Locator interface is implemented by Readers that support resolving
// a permanent URL location.
//
// The location will be served in the Content-Location header for the upstream
// response.
type Locator interface {
	// Locate resolves a permanent, publicly accessible URL for the
	// underlying content.
	Locate(ctx context.Context) (string, error)
}
