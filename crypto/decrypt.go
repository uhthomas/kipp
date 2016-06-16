package crypto

import (
	"crypto/aes"
	"io"
)

type Decrypter struct {
	*ctr
	R io.Reader
}

func NewDecrypter(r io.Reader, key, iv []byte) (*Decrypter, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Decrypter{newCTR(b, iv), r}, nil
}

func (d *Decrypter) Read(p []byte) (n int, err error) {
	n, err = d.R.Read(p)
	d.XORKeyStream(p[:n], p[:n])
	return
}

func (d *Decrypter) WriteTo(w io.Writer) (n int64, err error) {
	b := make([]byte, 0, 32<<10)
	for {
		nr, err := d.Read(b[:cap(b)])
		n += int64(nr)
		switch {
		case err == io.EOF:
			return n, nil
		case err != nil:
			return n, err
		}
		if _, err := w.Write(b[:nr]); err != nil {
			return n, err
		}
	}
}

func (d *Decrypter) Seek(offset int64, whence int) (ret int64, err error) {
	if r, ok := d.R.(io.ReadSeeker); ok {
		ret, err = r.Seek(offset, whence)
		d.seek(ret)
	}
	return
}

func (d *Decrypter) Close() (err error) {
	if c, ok := d.R.(io.Closer); ok {
		err = c.Close()
	}
	return
}
