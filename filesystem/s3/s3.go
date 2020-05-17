package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/uhthomas/kipp/filesystem"
)

// FileSystem is an abstraction over an s3 bucket which allows for the creation,
// opening and removal of objects.
type FileSystem struct {
	client   *s3.S3
	uploader *s3manager.Uploader
	bucket   string
}

// New creates a new aws session and s3 client.
func New(bucket string, config *aws.Config) (*FileSystem, error) {
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}
	c := s3.New(sess)
	return &FileSystem{
		client:   c,
		uploader: s3manager.NewUploaderWithClient(c),
		bucket:   bucket,
	}, nil
}

// Create creates a new writer which uploads the s3 object, name, to the bucket.
func (fs *FileSystem) Create(ctx context.Context, name string) (filesystem.Writer, error) {
	return newWriter(ctx, fs.uploader, fs.bucket, name), nil
}

// Open gets the object with the specified key, name.
func (fs *FileSystem) Open(ctx context.Context, name string) (filesystem.Reader, error) {
	r := &reader{
		ctx:    ctx,
		client: fs.client,
		bucket: fs.bucket,
		name:   name,
	}
	if err := r.reset(); err != nil {
		return nil, err
	}
	return r, nil
}

// Remove removes the s3 object specified with key, name, from the bucket.
func (fs *FileSystem) Remove(ctx context.Context, name string) error {
	if _, err := fs.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &fs.bucket,
		Key:    &name,
	}); err != nil {
		return fmt.Errorf("delete object %s/%s: %w", fs.bucket, name, err)
	}
	return nil
}
