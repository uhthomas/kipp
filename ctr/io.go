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

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.R.Read(p)
	r.XORKeyStream(p[:n], p[:n])
	return
}

func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := r.R.(io.Seeker); ok {
		ret, err = s.Seek(offset, whence)
		r.seek(ret)
	}
	return
}

func (r *Reader) Close() (err error) {
	if c, ok := r.R.(io.Closer); ok {
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

func (w *Writer) Write(b []byte) (n int, err error) {
	c := make([]byte, len(b))
	w.XORKeyStream(c, b)
	return w.W.Write(c)
}

func (w *Writer) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := w.W.(io.Seeker); ok {
		ret, err = s.Seek(offset, whence)
		w.seek(ret)
	}
	return
}

func (w *Writer) Close() (err error) {
	if c, ok := w.W.(io.Closer); ok {
		err = c.Close()
	}
	return
}
