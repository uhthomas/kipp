package crypto

import (
	"crypto/aes"
	"io"
)

type Encrypter struct {
	*ctr
	W io.Writer
}

func NewEncrypter(w io.Writer, key, iv []byte) (*Encrypter, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Encrypter{newCTR(b, iv), w}, nil
}

func (e *Encrypter) Write(b []byte) (n int, err error) {
	c := make([]byte, len(b))
	e.XORKeyStream(c, b)
	return e.W.Write(c)
}

func (e *Encrypter) Seek(offset int64, whence int) (ret int64, err error) {
	if w, ok := e.W.(io.WriteSeeker); ok {
		ret, err = w.Seek(offset, whence)
		e.p = ret
		e.refill()
	}
	return
}

func (e *Encrypter) Close() (err error) {
	if c, ok := e.W.(io.Closer); ok {
		err = c.Close()
	}
	return
}
