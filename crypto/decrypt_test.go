package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type decryptTest struct {
	key, iv []byte
	r       io.ReadSeeker
}

var decrypt decryptTest

func init() {
	b, err := Random((5 << 20) + 48)
	if err != nil {
		panic(err)
	}
	key, iv := b[:32], b[32:48]
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	newCTR(block, iv).XORKeyStream(b[48:], b[48:])
	decrypt = decryptTest{key, iv, bytes.NewReader(b[48:])}
}

func BenchmarkDecrypter(b *testing.B) {
	b.SetBytes(5 << 20)
	for i := 0; i < b.N; i++ {
		decrypt.r.Seek(0, os.SEEK_SET)
		d, err := NewDecrypter(decrypt.r, decrypt.key, decrypt.iv)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(ioutil.Discard, d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdDecrypter(b *testing.B) {
	b.SetBytes(5 << 20)
	for i := 0; i < b.N; i++ {
		decrypt.r.Seek(0, os.SEEK_SET)
		block, err := aes.NewCipher(decrypt.key)
		if err != nil {
			b.Fatal(err)
		}
		r := &cipher.StreamReader{S: cipher.NewCTR(block, decrypt.iv), R: decrypt.r}
		if _, err := io.Copy(ioutil.Discard, r); err != nil {
			b.Fatal(err)
		}
	}
}
