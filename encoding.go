package conf

// Encoding allows the encoding of conf to easily be changed to encodings such
// as base32, base64 or hex.
type Encoding interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}
