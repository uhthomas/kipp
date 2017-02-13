package crypto

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
