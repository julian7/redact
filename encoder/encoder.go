package encoder

import (
	"github.com/pkg/errors"
)

const (
	// TypeAES256GCM96 is AES256-GCM96 encoding type
	TypeAES256GCM96 = iota
)

// EncoderFactory generates a new encoder
type EncoderFactory func(aes, hmac []byte) (Encoder, error)

var (
	encoders = map[int]EncoderFactory{
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

// RegisterEncoder registers a new encoder
func RegisterEncoder(encType int, factory EncoderFactory) error {
	if _, ok := encoders[encType]; ok {
		return errors.Errorf("encoder type %d already exists", encType)
	}
	encoders[encType] = factory
	return nil
}

// UnregisterEncoder removes an encoder
func UnregisterEncoder(encType int) error {
	if _, ok := encoders[encType]; !ok {
		return errors.Errorf("encoder type %d doesn't exist", encType)
	}
	delete(encoders, encType)
	return nil
}
