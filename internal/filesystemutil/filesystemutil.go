package filesystemutil

import (
	"fmt"
	"net/url"
	"os"

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
	case "s3":
		return s3.New("", "", "")
	}
	return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
}
