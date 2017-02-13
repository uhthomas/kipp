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

type encryptTest struct {
	key, iv []byte
	r       io.ReadSeeker
}

var encrypt encryptTest

func init() {
	b, err := Random((5 << 20) + 48)
	if err != nil {
		panic(err)
	}
	encrypt = encryptTest{b[:32], b[32:48], bytes.NewReader(b[48:])}
}

func BenchmarkEncrypter(b *testing.B) {
	b.SetBytes(5 << 20)
	for i := 0; i < b.N; i++ {
		encrypt.r.Seek(0, os.SEEK_SET)
		e, err := NewEncrypter(ioutil.Discard, encrypt.key, encrypt.iv)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(e, encrypt.r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdEncrypter(b *testing.B) {
	b.SetBytes(5 << 20)
	for i := 0; i < b.N; i++ {
		encrypt.r.Seek(0, os.SEEK_SET)
		block, err := aes.NewCipher(encrypt.key)
		if err != nil {
			b.Fatal(err)
		}
		s := &cipher.StreamWriter{S: cipher.NewCTR(block, encrypt.iv), W: ioutil.Discard}
		if _, err := io.Copy(s, encrypt.r); err != nil {
			b.Fatal(err)
		}
	}
}
