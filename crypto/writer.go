package crypto

import (
	"crypto/aes"
	"io"
)

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
