package encoder

import (
	"fmt"
)

const (
	// TypeAES256GCM96 is AES256-GCM96 encoding type
	TypeAES256GCM96 = iota
)

// Factory generates a new encoder
type Factory func(key []byte) (Encoder, error)

var (
	encoders = map[int]Factory{
		TypeAES256GCM96: NewAES256GCM96,
	}
)

// Encoder can encode and decode data
type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

// NewEncoder returns a new Encoder initialized with a key handler
func NewEncoder(encType int, key []byte) (Encoder, error) {
	encoder, ok := encoders[encType]
	if !ok {
		return nil, fmt.Errorf("invalid encoding type %d", encType)
	}

	return encoder(key)
}

// RegisterEncoder registers a new encoder
func RegisterEncoder(encType int, factory Factory) error {
	if _, ok := encoders[encType]; ok {
		return fmt.Errorf("encoder type %d already exists", encType)
	}

	encoders[encType] = factory

	return nil
}

// UnregisterEncoder removes an encoder
func UnregisterEncoder(encType int) error {
	if _, ok := encoders[encType]; !ok {
		return fmt.Errorf("encoder type %d doesn't exist", encType)
	}

	delete(encoders, encType)

	return nil
}
