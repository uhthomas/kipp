package filesystemutil

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
		return local.New(u.Path)
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
