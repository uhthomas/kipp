package filesystemutil

import (
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/uhthomas/kipp/filesystem"
	"github.com/uhthomas/kipp/filesystem/local"
	"github.com/uhthomas/kipp/filesystem/s3"
)

// Parse parses s, and will create the appropriate filesystem for the scheme.
func Parse(s string) (filesystem.FileSystem, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "":
		return local.New(u.Path, os.TempDir())
	// s3 follows the form:
	//      s3://key:token@region/bucket?endpoint=optional
	case "s3":
		c := &aws.Config{Region: &u.Host}
		if u.User != nil {
			p, _ := u.User.Password()
			c.Credentials = credentials.NewStaticCredentials(u.User.Username(), p, "")
		}
		if e := u.Query().Get("endpoint"); e != "" {
			c.Endpoint = &e
		}
		return s3.New(u.Path, c)
	}
	return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
}
