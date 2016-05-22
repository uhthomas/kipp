package crypto

import (
	"crypto/cipher"
	"math/big"
)

const streamBufferSize = 512

type ctr struct {
	b       cipher.Block
	ctr     []byte
	iv      []byte
	out     []byte
	outUsed int
}

func newCTR(block cipher.Block, iv []byte) *ctr {
	if len(iv) != block.BlockSize() {
		panic("IV length must equal block size")
	}
	bufSize := streamBufferSize
	if bs := block.BlockSize(); bufSize < bs {
		bufSize = bs
	}
	return &ctr{block, Duplicate(iv), Duplicate(iv), make([]byte, 0, bufSize), 0}
}

func (x *ctr) XORKeyStream(dst, src []byte) {
	for len(src) > 0 {
		if x.outUsed >= len(x.out)-x.b.BlockSize() {
			x.refill()
		}
		n := XORBytes(dst, src, x.out[x.outUsed:])
		dst, src = dst[n:], src[n:]
		x.outUsed += n
	}
}

func (x *ctr) seek(offset int64) {
	b := &big.Int{}
	x.ctr = b.SetBytes(x.iv).
		Add(b, big.NewInt(offset/int64(x.b.BlockSize()))).
		Bytes()
	x.ctr = append(make([]byte, x.b.BlockSize()-len(x.ctr)), x.ctr...)
	x.outUsed = len(x.out)
	x.refill()
	x.outUsed = int(offset) % x.b.BlockSize()
}

func (x *ctr) refill() {
	remain := len(x.out) - x.outUsed
	copy(x.out, x.out[x.outUsed:])
	x.out = x.out[:cap(x.out)]
	bs := x.b.BlockSize()
	for remain <= len(x.out)-bs {
		x.b.Encrypt(x.out[remain:], x.ctr)
		remain += bs
		for i := len(x.ctr) - 1; i >= 0; i-- {
			x.ctr[i]++
			if x.ctr[i] != 0 {
				break
			}
		}
	}
	x.out = x.out[:remain]
	x.outUsed = 0
}
