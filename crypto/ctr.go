package crypto

import (
	"crypto/cipher"
	"math/big"
)

const streamBufferSize = 512

type ctr struct {
	c   int
	b   cipher.Block
	buf []byte
	ctr []byte
	iv  []byte
	p   int64
}

func newCTR(block cipher.Block, iv []byte) *ctr {
	if len(iv) != block.BlockSize() {
		panic("IV length must equal block size")
	}
	buffer := streamBufferSize
	if bs := block.BlockSize(); bs > buffer {
		buffer = bs
	}
	return &ctr{0, block, make([]byte, 0, buffer), Duplicate(iv), Duplicate(iv), 0}
}

func (x *ctr) XORKeyStream(dst, src []byte) {
	for len(src) > 0 {
		if x.c >= len(x.buf)-x.b.BlockSize() {
			x.refill()
		}
		n := XORBytes(dst, src, x.buf[x.c:])
		x.p += int64(n)
		x.c += n
		dst, src = dst[n:], src[n:]
	}
}

func (x *ctr) refill() {
	b := &big.Int{}
	x.ctr = b.
		SetBytes(x.iv).
		Add(b, big.NewInt(x.p/int64(x.b.BlockSize()))).
		Bytes()
	x.ctr = append(make([]byte, x.b.BlockSize()-len(x.ctr)), x.ctr...)
	x.buf = x.buf[:cap(x.buf)]
	for i := 0; i < streamBufferSize; i += x.b.BlockSize() {
		x.b.Encrypt(x.buf[i:], x.ctr)
		x.increment()
	}
	x.c = int(x.p) % x.b.BlockSize()
	return
}

func (x *ctr) increment() {
	for i := len(x.ctr) - 1; i >= 0; i-- {
		x.ctr[i]++
		if x.ctr[i] != 0 {
			break
		}
	}
}
