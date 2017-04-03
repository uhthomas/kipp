package ctr

import (
	"crypto/aes"
	"io"
)

type Reader struct {
	*ctr
	R io.Reader
}

func NewReader(r io.Reader, key, iv []byte) (*Reader, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Reader{newCTR(b, iv), r}, nil
}

func (d *Reader) Read(p []byte) (n int, err error) {
	n, err = d.R.Read(p)
	d.XORKeyStream(p[:n], p[:n])
	return
}

func (d *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := d.R.(io.Seeker); ok {
		ret, err = s.Seek(offset, whence)
		d.seek(ret)
	}
	return
}

func (d *Reader) Close() (err error) {
	if c, ok := d.R.(io.Closer); ok {
		err = c.Close()
	}
	return
}

type Writer struct {
	*ctr
	W io.Writer
}

func NewWriter(w io.Writer, key, iv []byte) (*Writer, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Writer{newCTR(b, iv), w}, nil
}

func (e *Writer) Write(b []byte) (n int, err error) {
	c := make([]byte, len(b))
	e.XORKeyStream(c, b)
	return e.W.Write(c)
}

func (e *Writer) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := e.W.(io.Seeker); ok {
		ret, err = s.Seek(offset, whence)
		e.seek(ret)
	}
	return
}

func (e *Writer) Close() (err error) {
	if c, ok := e.W.(io.Closer); ok {
		err = c.Close()
	}
	return
}
