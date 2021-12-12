package keyv0

import (
	"crypto/rand"
	"crypto/sha1" //nolint:gosec
	"fmt"
)

const (
	// SecretSize is the size of a 32-byte AES and 64-byte HMAC key (for SHA-256 and HMAC SHA-256)
	SecretSize = 96
)

// KeyV0 stores an AES-256 and a HMAC SHA-256 key
type KeyV0 struct {
	Epoch      uint32
	SecretData [SecretSize]byte
}

// NewKey creates a new Key struct based on parameter input
func NewKey(epoch uint32) *KeyV0 {
	key := &KeyV0{Epoch: epoch}

	return key
}

// Version returns key epoch version
func (k *KeyV0) Version() uint32 {
	return k.Epoch
}

// Type returns key format type
func (k *KeyV0) Type() uint32 {
	return 0
}

// Secret returns Secret key
func (k *KeyV0) Secret() []byte {
	return k.SecretData[:]
}

// Generate generates new keys to secret key
func (k *KeyV0) Generate() error {
	_, err := rand.Read(k.SecretData[:])
	if err != nil {
		return fmt.Errorf("generating Secret key: %w", err)
	}

	return nil
}

func (k *KeyV0) String() string {
	return fmt.Sprintf("#%d %s", k.Epoch, fmt.Sprintf("%x", sha1.Sum(k.Secret()))[:8]) //nolint:gosec
}
