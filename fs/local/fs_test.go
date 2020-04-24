package local_test

import (
	"testing"

	"github.com/uhthomas/kipp/fs"
	"github.com/uhthomas/kipp/fs/local"
)

func TestFileSystem(t *testing.T) {
	var i interface{} = (*local.FileSystem)(nil)
	if _, ok := i.(fs.FileSystem); !ok {
		t.Fatal("local.FileSystem does not implement fs.FileSystem")
	}
}
