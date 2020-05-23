package s3

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type writer struct {
	pw     *io.PipeWriter
	synced bool
	c      <-chan error
}

func newWriter(ctx context.Context, u *s3manager.Uploader, bucket, name string) *writer {
	pr, pw := io.Pipe()
	c := make(chan error)
	go func() {
		defer close(c)
		defer pr.Close()
		_, err := u.UploadWithContext(ctx, &s3manager.UploadInput{
			Body:   pr,
			Bucket: &bucket,
			Key:    &name,
		})
		if err != nil {
			log.Println(err)
			pr.CloseWithError(fmt.Errorf("upload: %w", err))
		}
		c <- err
	}()
	return &writer{pw: pw, c: c}
}

func (w *writer) Write(p []byte) (n int, err error) { return w.pw.Write(p) }

// Sync closes pw, which in turn makes all subsequent calls from Read
// return io.EOF, indicating a successful upload. The file is then persisted
// to s3.
func (w *writer) Sync() error {
	if err := w.pw.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	w.synced = true
	// wait for the upload to finish
	return <-w.c
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
