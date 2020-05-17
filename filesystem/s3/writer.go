package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type writer struct {
	pw     *io.PipeWriter
	synced bool
}

func newWriter(ctx context.Context, u *s3manager.Uploader, bucket, name string) *writer {
	pr, pw := io.Pipe()
	go func() {
		defer pr.Close()
		if _, err := u.UploadWithContext(ctx, &s3manager.UploadInput{
			Body:   pr,
			Bucket: &bucket,
			Key:    &name,
		}); err != nil {
			pr.CloseWithError(fmt.Errorf("upload: %w", err))
		}
	}()
	return &writer{pw: pw}
}

func (w *writer) Write(p []byte) (n int, err error) { return w.pw.Write(p) }

// Sync closes pw, which in turn makes all subsequent calls from Read
// return io.EOF, indicating a successful upload. This then persists the file
// to s3.
func (w *writer) Sync() error {
	if err := w.pw.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	w.synced = true
	return nil
}

// Close closes pw, if Sync has not been called, it will close with
// io.UnexpectedEOF, causing the upload to fail thus not persisting the data
// to the bucket.
func (w *writer) Close() error {
	if !w.synced {
		return w.pw.CloseWithError(io.ErrUnexpectedEOF)
	}
	return w.pw.Close()
}
