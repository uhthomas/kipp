package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"testing"
)

var (
	key, iv, by []byte
	d           *bytes.Reader
)

func init() {
	b, err := Random((50 << 20) + 48)
	if err != nil {
		panic(err)
	}
	key, iv, by, d = b[:32], b[32:48], b[48:], bytes.NewReader(b[48:])
}

func BenchmarkEncrypter(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		e, err := NewEncrypter(&buf, key, iv)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(e, d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdEncrypter(b *testing.B) {
	discard := make([]byte, len(by))
	for i := 0; i < b.N; i++ {
		block, err := aes.NewCipher(key)
		if err != nil {
			b.Fatal(err)
		}
		s := cipher.NewCTR(block, iv)
		s.XORKeyStream(discard, by)
	}
}

func BenchmarkStdStreamEncrypter(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		block, err := aes.NewCipher(key)
		if err != nil {
			b.Fatal(err)
		}
		s := &cipher.StreamWriter{S: cipher.NewCTR(block, iv), W: &buf}
		if _, err := io.Copy(s, d); err != nil {
			b.Fatal(err)
		}
	}
}
