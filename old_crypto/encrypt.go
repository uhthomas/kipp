package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

type Encrypter struct {
	S cipher.Stream
	W io.Writer
}

func NewEncrypter(key []byte, w io.Writer) (*Encrypter, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, b.BlockSize())
	return &Encrypter{NewCTR(b, iv), w}, nil
}

func (e *Encrypter) Write(b []byte) (n int, err error) {
	c := make([]byte, len(b))
	e.S.XORKeyStream(c, b)
	return e.W.Write(c)
}
