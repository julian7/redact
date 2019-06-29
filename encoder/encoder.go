package encoder

import (
	"github.com/pkg/errors"
)

const (
	// TypeAES256GCM96 is AES256-GCM96 encoding type
	TypeAES256GCM96 = iota
)

var (
	encoders = map[int]func([]byte, []byte) (Encoder, error){
		TypeAES256GCM96: NewAES256GCM96,
	}
)

// Encoder can encode and decode data
type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

// NewEncoder returns a new Encoder initialized with a key handler
func NewEncoder(encType int, aes, hmac []byte) (Encoder, error) {
	encoder, ok := encoders[encType]
	if !ok {
		return nil, errors.Errorf("invalid encoding type %d", encType)
	}
	return encoder(aes, hmac)
}
