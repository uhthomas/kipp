package b2

import (
	"context"
	"fmt"
	"io"

	"github.com/kurin/blazer/b2"
)

type reader struct {
	ctx  context.Context
	obj  *b2.Object
	size int64
	i    int64
	rc   io.ReadCloser
}

// Read reads from rc, adding n to i.
func (r *reader) Read(b []byte) (n int, err error) {
	n, err = r.rc.Read(b)
	r.i += int64(n)
	return n, err
}

// Seek closes rc and calculates the new offset, replacing rc with a new
// reader for the given offset.
func (r *reader) Seek(offset int64, whence int) (n int64, err error) {
	if err := r.rc.Close(); err != nil {
		return 0, fmt.Errorf("close: %w", err)
	}
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.i
	case io.SeekEnd:
		offset += r.size
	}
	r.i = offset
	r.rc = r.obj.NewRangeReader(r.ctx, offset, -1)
	return r.i, nil
}

// Close closes rc.
func (r *reader) Close() error { return r.rc.Close() }
