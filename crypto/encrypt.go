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

func (e *Encrypter) ReadFrom(r io.Reader) (n int64, err error) {
	b := make([]byte, 0, 32<<10)
	for {
		nr, err := r.Read(b[:cap(b)])
		n += int64(nr)
		switch {
		case err == io.EOF:
			return n, nil
		case err != nil:
			return n, err
		}
		if _, err := e.Write(b[:nr]); err != nil {
			return n, err
		}
	}
}

func (e *Encrypter) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := e.W.(io.Seeker); ok {
		ret, err = s.Seek(offset, whence)
		e.seek(ret)
	}
	return
}

func (e *Encrypter) Close() (err error) {
	if c, ok := e.W.(io.Closer); ok {
		err = c.Close()
	}
	return
}
