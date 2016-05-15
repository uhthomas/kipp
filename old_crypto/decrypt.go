package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

// type Decrypter struct {
// 	w  io.Writer
// 	sr cipher.StreamReader
// }

// func NewDecrypter(key []byte, w io.Writer) *Decrypter {
// 	b, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	iv := make([]byte, b.BlockSize())
// 	s := cipher.NewCFBDecrypter(b, iv[:])
// 	return &Decrypter{w: w, sr: cipher.StreamReader{S: s}}
// }

// func (d *Decrypter) Decrypt(r io.Reader) (n int64, err error) {
// 	d.sr.R = r
// 	return io.Copy(d.w, d.sr)
// }

// type Decrypter struct {
// 	S      *CTR
// 	R      io.ReadSeeker
// 	offset int64
// }

// func NewDecrypter(key []byte, r io.ReadSeeker) (*Decrypter, error) {
// 	b, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	iv := make([]byte, b.BlockSize())
// 	return &Decrypter{NewCTR(b, iv), r, 0}, nil
// }

// func (d *Decrypter) Read(b []byte) (n int, err error) {
// 	n, err = d.R.Read(b)
// 	d.S.Set(d.R, int(d.offset))
// 	d.S.XORKeyStream(b[:n], b[:n])
// 	return
// }

// func (d *Decrypter) Seek(offset int64, whence int) (ret int64, err error) {
// 	d.offset = offset
// 	fmt.Println(offset)
// 	return d.R.Seek(offset, whence)
// }

type Decrypter struct {
	B    cipher.Block
	IV   []byte
	R    io.ReadSeeker
	ctr  []byte
	read int64
}

func NewDecrypter(key []byte, r io.ReadSeeker) (*Decrypter, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, b.BlockSize())
	return &Decrypter{
		B:   b,
		IV:  Duplicate(iv),
		R:   r,
		ctr: Duplicate(iv),
	}, nil
}

func (d *Decrypter) XORKeyStream(dst, src []byte) {
	var read int
	for len(src) > 0 {
		block := d.block()
		blockBytes := d.read % int64(d.B.BlockSize())
		n := XORBytes(dst, src, block[blockBytes:])
		d.read += int64(n)
		read += n
		dst = dst[n:]
		src = src[n:]
	}
}

func (d *Decrypter) Read(p []byte) (n int, err error) {
	n, err = d.R.Read(p)
	if err != nil {
		return
	}
	d.XORKeyStream(p[:n], p[:n])
	return
}

func (d *Decrypter) Seek(offset int64, whence int) (ret int64, err error) {
	d.read = offset
	return d.R.Seek(offset, whence)
}

func (d *Decrypter) block() (b []byte) {
	total := d.read / int64(d.B.BlockSize())
	copy(d.ctr, d.IV)
	for ; total > 0; total-- {
		for i := len(d.ctr) - 1; i >= 0; i-- {
			d.ctr[i]++
			if d.ctr[i] != 0 {
				break
			}
		}
	}
	b = make([]byte, d.B.BlockSize())
	d.B.Encrypt(b, d.ctr)
	return
}
