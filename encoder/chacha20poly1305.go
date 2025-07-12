package encoder

import (
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// ChaCha20Poly1305KeySize is the standard key size for ChaCha20-Poly1305 encoding
	ChaCha20Poly1305KeySize = 32
)

// ChaCha20Poly1305 can take a KeyHandler, and stores encryption and HMAC keys
type ChaCha20Poly1305 []byte

// NewChaCha20Poly1305 returns a new Encoder initialized with a key handler
func NewChaCha20Poly1305(key []byte) (AEAD, error) {
	if len(key) < ChaCha20Poly1305KeySize {
		return nil, ErrKeyTooSmall
	}

	return ChaCha20Poly1305(key[:ChaCha20Poly1305KeySize]), nil
}

func (enc ChaCha20Poly1305) KeySize() int { return ChaCha20Poly1305KeySize }

func (enc ChaCha20Poly1305) String() string { return "ChaCha20-Poly1305" }

func (enc ChaCha20Poly1305) AEAD() (cipher.AEAD, error) {
	return chacha20poly1305.New(enc)
}
