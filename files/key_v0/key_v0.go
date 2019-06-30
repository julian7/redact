package keyv0

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"

	"github.com/pkg/errors"
)

const (
	// AESKeySize key size of AES key (32 bytes for AEC SHA-256)
	AESKeySize = 32
	// HMACKeySize key size of HMAC key (64 bytes for HMAC SHA-256)
	HMACKeySize = 64
)

// KeyV0 stores an AES-256 and a HMAC SHA-256 key
type KeyV0 struct {
	Epoch    uint32
	AESData  [AESKeySize]byte
	HMACData [HMACKeySize]byte
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

// AES returns AES key
func (k *KeyV0) AES() []byte {
	return k.AESData[:]
}

// HMAC returns HMAC key
func (k *KeyV0) HMAC() []byte {
	return k.HMACData[:]
}

// Generate generates new keys to master key
func (k *KeyV0) Generate() error {
	_, err := rand.Read(k.AESData[:])
	if err != nil {
		return errors.Wrap(err, "generating AES key")
	}
	_, err = rand.Read(k.HMACData[:])
	if err != nil {
		return errors.Wrap(err, "generating HMAC key")
	}
	return nil
}

func (k *KeyV0) String() string {
	return fmt.Sprintf("#%d %s", k.Epoch, k.hash()[:8])
}

func (k *KeyV0) hash() string {
	return fmt.Sprintf(
		"%x",
		sha1.Sum(
			[]byte(
				fmt.Sprintf("%s|%s", string(k.AES()), string(k.HMAC())),
			),
		),
	)
}
