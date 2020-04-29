package kipp

import (
	"fmt"
	"io"
	"os"
	"testing"
	"testing/quick"
	"time"

	"github.com/uhthomas/kipp/database"
)

type fakeFileSystemReader struct{ limit, off int64 }

func (r fakeFileSystemReader) Read(b []byte) (n int, err error) {
	if r.limit == r.off {
		return 0, io.EOF
	}
	if l := r.limit - r.off; int64(len(b)) >= l {
		b = b[:l]
	}
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
	r.off += int64(len(b))
	return len(b), nil
}

func (r fakeFileSystemReader) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.off
	case io.SeekEnd:
		offset += r.limit
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	r.off = offset
	return offset, nil
}

func (fakeFileSystemReader) Close() error { return nil }

func TestFileReaddir(t *testing.T) {
	f := &file{
		Reader: fakeFileSystemReader{limit: 10},
		entry:  database.Entry{},
	}
	files, err := (&file{}).Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(files); l > 0 {
		t.Fatalf("unexpected number of files; got %d, want 0", l)
	}
}

func TestFileStat(t *testing.T) {
	e := database.Entry{
		Slug:      "some slug",
		Name:      "some name",
		Sum:       "some sum",
		Size:      123456,
		Lifetime:  nil,
		Timestamp: time.Unix(1234, 5678),
	}
	fi, err := (&file{entry: e}).Stat()
	if err != nil {
		t.Fatal(err)
	}
	if err := quick.CheckEqual(e, fi.(*fileInfo).entry, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFileInfo(t *testing.T) {
	e := database.Entry{
		Name:      "some name",
		Size:      123456,
		Timestamp: time.Unix(1234, 5678),
	}
	fi := &fileInfo{entry: e}
	if got, want := fi.Name(), e.Name; got != want {
		t.Fatalf("unexpected name; got %s, want %s", got, want)
	}
	if got, want := fi.Size(), e.Size; got != want {
		t.Fatalf("unexpected size; got %d, want %d", got, want)
	}
	if got, want := fi.Mode(), os.FileMode(0600); got != want {
		t.Fatalf("unexpected mode; got %d, want %d", got, want)
	}
	if fi.IsDir() {
		t.Fatalf("file info reports it's a directory when it shouldn't")
	}
	if fi.Sys() != nil {
		t.Fatal("sys is not nil")
	}
	if got, want := fi.ModTime(), e.Timestamp; got != want {
		t.Fatalf("unexpected mod time; got %s, want %s", got, want)
	}
}
