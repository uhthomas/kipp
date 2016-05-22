package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"io/ioutil"
	"testing"
)

type decryptTest struct {
	key, iv, data []byte
}

var decrypt decryptTest

func init() {
	b, err := Random((50 << 20) + 48)
	if err != nil {
		panic(err)
	}
	decrypt = decryptTest{b[:32], b[32:48], b[48:]}
}

func BenchmarkDecrypter(b *testing.B) {
	var buf bytes.Buffer
	e, err := NewEncrypter(&buf, decrypt.key, decrypt.iv)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := io.Copy(e, bytes.NewReader(decrypt.data)); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(5 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d, err := NewDecrypter(&buf, decrypt.key, decrypt.iv)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(ioutil.Discard, d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdDecrypter(b *testing.B) {
	var buf bytes.Buffer
	block, err := aes.NewCipher(decrypt.key)
	if err != nil {
		b.Fatal(err)
	}
	s := &cipher.StreamWriter{S: cipher.NewCTR(block, decrypt.iv), W: &buf}
	if _, err := io.Copy(s, bytes.NewReader(decrypt.data)); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(5 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block, err := aes.NewCipher(decrypt.key)
		if err != nil {
			b.Fatal(err)
		}
		r := &cipher.StreamReader{S: cipher.NewCTR(block, decrypt.iv), R: &buf}
		if _, err := io.Copy(ioutil.Discard, r); err != nil {
			b.Fatal(err)
		}
	}
}
