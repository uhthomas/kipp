package local_test

import (
	"testing"

	"github.com/uhthomas/kipp/filesystem"
	"github.com/uhthomas/kipp/filesystem/local"
)

func TestFileSystem(t *testing.T) {
	var i interface{} = (*local.FileSystem)(nil)
	if _, ok := i.(filesystem.FileSystem); !ok {
		t.Fatal("local.FileSystem does not implement fs.FileSystem")
	}
}
