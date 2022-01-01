package encoder

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

const (
	// HMAC256KeySize is the standard key size for HMAC AES-256 hashing
	HMAC256KeySize = 64
)

// HMAC256 calculates IV for AEAD using HMAC-256
type HMAC256 []byte

func NewHMAC256(key []byte) HMAC256 {
	return HMAC256(key[:HMAC256KeySize])
}

func (key HMAC256) Sum(plaintext []byte) ([]byte, error) {
	nonce := hmac.New(sha256.New, []byte(key))
	if _, err := nonce.Write(plaintext); err != nil {
		return nil, fmt.Errorf("calculating HMAC IV: %w", err)
	}

	return nonce.Sum(nil), nil
}
