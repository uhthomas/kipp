package filesystem

import (
	"context"
	"io"
)

// A FileSystem is a persistent store of objects uniquely identified by name.
type FileSystem interface {
	// Create creates an object with the specified name, and will read
	// from r up to io.EOF. The reader is explicitly passed in to allow
	// implementations to cleanup, and guarantee consistency.
	Create(ctx context.Context, name string, r io.Reader) error
	Open(ctx context.Context, name string) (Reader, error)
	Remove(ctx context.Context, name string) error
}

// A Reader is a readable, seekable and closable file stream.
type Reader interface {
	io.ReadSeeker
	io.Closer
}

// PipeReader pipes r to f(w).
func PipeReader(f func(w io.Writer) error) io.Reader {
	pr, pw := io.Pipe()
	go func() { pw.CloseWithError(f(pw)) }()
	return pr
}
