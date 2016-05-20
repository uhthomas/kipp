package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"testing"
)

type encryptTest struct {
	key, iv, data []byte
}

var encrypt encryptTest

func init() {
	b, err := Random((50 << 20) + 48)
	if err != nil {
		panic(err)
	}
	encrypt = encryptTest{b[:32], b[32:48], b[48:]}
}

func BenchmarkEncrypter(b *testing.B) {
	r := bytes.NewReader(encrypt.data)
	var buf bytes.Buffer
	b.SetBytes(5 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, err := NewEncrypter(&buf, encrypt.key, encrypt.iv)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(e, r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdEncrypter(b *testing.B) {
	discard := make([]byte, len(encrypt.data))
	b.SetBytes(5 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block, err := aes.NewCipher(encrypt.key)
		if err != nil {
			b.Fatal(err)
		}
		s := cipher.NewCTR(block, encrypt.iv)
		s.XORKeyStream(discard, encrypt.data)
	}
}

func BenchmarkStdStreamEncrypter(b *testing.B) {
	r := bytes.NewReader(encrypt.data)
	var buf bytes.Buffer
	b.SetBytes(5 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block, err := aes.NewCipher(encrypt.key)
		if err != nil {
			b.Fatal(err)
		}
		s := &cipher.StreamWriter{S: cipher.NewCTR(block, encrypt.iv), W: &buf}
		if _, err := io.Copy(s, r); err != nil {
			b.Fatal(err)
		}
	}
}
