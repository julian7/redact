package encoder

import (
	"crypto/aes"
	"crypto/cipher"
)

const (
	// AES256KeySize is the standard key size for AES-256 encoding
	AES256KeySize = 32
)

// AES256GCM96 can take a KeyHandler, and stores encryption and HMAC keys
type AES256GCM96 []byte

// NewAES256GCM96 returns a new Encoder initialized with a key handler
func NewAES256GCM96(key []byte) (AEAD, error) {
	if len(key) < AES256KeySize {
		return nil, ErrKeyTooSmall
	}

	return AES256GCM96(key[:AES256KeySize]), nil
}

func (enc AES256GCM96) KeySize() int { return AES256KeySize }

func (enc AES256GCM96) String() string { return "AES256-GCM96" }

func (enc AES256GCM96) AEAD() (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(enc)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}
