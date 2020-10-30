package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type reader struct {
	ctx          context.Context
	client       *s3.S3
	bucket, name string
	obj          *s3.GetObjectOutput
	offset, size int64
}

func newReader(ctx context.Context, client *s3.S3, bucket, name string) *reader {
	return &reader{
		ctx:    ctx,
		client: client,
		bucket: bucket,
		name:   name,
	}
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.obj != nil {
		return r.obj.Body.Read(p)
	}
	if err := r.reset(); err != nil {
		return 0, fmt.Errorf("reset: %w", err)
	}
	return r.Read(p)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.offset
	case io.SeekEnd:
		offset = r.size - offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	if offset < 0 {
		return 0, errors.New("invalid offset")
	}
	r.Close()
	r.obj, r.offset = nil, offset
	return offset, nil
}

func (r *reader) Close() error {
	if r.obj == nil {
		return nil
	}
	return r.obj.Body.Close()
}

func (r *reader) reset() error {
	r.Close()
	in := &s3.GetObjectInput{Bucket: &r.bucket, Key: &r.name}
	if r.offset > 0 {
		in.Range = aws.String(fmt.Sprintf("bytes=%d-", r.offset))
	}
	obj, err := r.client.GetObjectWithContext(r.ctx, in)
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	r.obj = obj
	if r.size == 0 {
		r.size = *obj.ContentLength
	}
	return nil
}
