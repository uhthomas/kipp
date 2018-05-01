package kipp

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func MakeServer() (s kipp.Server, err error) {
	s.DB, err = (kipp.Driver{
		Dialect: "sqlite3",
		Path:    ":memory:",
	}).Open()
	if err != nil {
		return s, err
	}
	s.FilePath, err = ioutil.TempDir("", "")
	if err != nil {
		return s, err
	}
	s.TempPath, err = ioutil.TempDir("", "")
	if err != nil {
		return s, err
	}
	s.Max = 150 << 20
	return s, nil
}

func TestServerUploadHandler(t *testing.T) {
	s, err := MakeServer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(s.FilePath)
	defer os.RemoveAll(s.TempPath)
	tests := []struct {
		io.Reader
		Name     string
		Length   uint
		Expected int
	}{{nil, "", 0, 0}}
}
