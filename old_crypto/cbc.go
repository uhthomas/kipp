package crypto

import "crypto/cipher"

// import "crypto/cipher"

// type CTR struct {
// }

// func NewCTR(block cipher.Block, iv []byte) cipher.Stream {

// }

// func (c *CTR) XORKeyStream(dst, src []byte) {
//   for len(src) > 0 {

//   }
// }

type CBC struct {
	b         cipher.Block
	blockSize int
	iv        []byte
	oiv       []byte
	tmp       []byte
}

func NewCBCDecrypter(b cipher.Block, iv []byte) *CBC {
	if len(iv) != b.BlockSize() {
		panic("IV length must equal block size")
	}
	return &CBC{b, b.BlockSize(), Duplicate(iv), Duplicate(iv), make([]byte, b.BlockSize())}
}

func (x *CBC) BlockSize() int {
	return x.blockSize
}

func (x *CBC) CryptBlocks(dst, src []byte) {
	if len(src)%x.BlockSize() != 0 {
		panic("input not full blocks")
	}
	if len(dst) < len(src) {
		panic("output smaller than input")
	}
	if len(src) == 0 {
		return
	}
	for len(src) > 0 {
		block := src[:x.BlockSize()]
		output := make([]byte, x.BlockSize())
		x.b.Decrypt(output, block)
		XORBytes(dst, output, x.iv)
		dst = dst[x.BlockSize():]
		src = src[x.BlockSize():]
		copy(x.iv, block)
	}
}

func (x *CBC) IV(iv []byte) {
	copy(x.iv, iv)
}
