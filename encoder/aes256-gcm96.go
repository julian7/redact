package encoder

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	// AES256KeySize is the standard key size for AES-256 encoding
	AES256KeySize = 32
	// HMAC256KeySize is the standard key size for HMAC AES-256 hashing
	HMAC256KeySize = 64
)

// AES256GCM96 can take a KeyHandler, and stores encryption and HMAC keys
type AES256GCM96 struct {
	key  []byte
	hmac []byte
}

// NewAES256GCM96 returns a new Encoder initialized with a key handler
func NewAES256GCM96(key []byte) (Encoder, error) {
	if len(key) < AES256KeySize+HMAC256KeySize {
		return nil, errors.New("key too small")
	}

	return &AES256GCM96{key: key[:AES256KeySize], hmac: key[AES256KeySize : AES256KeySize+HMAC256KeySize]}, nil
}

// Encode takes a value, and encrypts it in a convergent way, making sure the same
// input provides the same output every time, while not leaking secret information
// about its encryption keys
func (enc AES256GCM96) Encode(value []byte) ([]byte, error) {
	gcm, err := getGCM(enc.key)
	if err != nil {
		return nil, err
	}

	nonce, err := calculateHMAC(enc.hmac, value)
	if err != nil {
		return nil, err
	}

	nonce = nonce[:gcm.NonceSize()]
	ciphertext := gcm.Seal(nil, nonce, value, nil)
	ciphertext = append(nonce, ciphertext...)

	return ciphertext, nil
}

// Decode takes a byte stream, and decrypts its contents
func (enc AES256GCM96) Decode(ciphertext []byte) ([]byte, error) {
	gcm, err := getGCM(enc.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too small")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting with AES256-GCM: %w", err)
	}

	hmacSum, err := calculateHMAC(enc.hmac, plaintext)

	if err != nil {
		return nil, fmt.Errorf("calculating HMAC: %w", err)
	}

	if !bytes.Equal(hmacSum[:gcm.NonceSize()], nonce) {
		return nil, errors.New("HMAC checksum invalid")
	}

	return plaintext, nil
}

func getGCM(key []byte) (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func calculateHMAC(key, value []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("HMAC key is empty")
	}

	nonceHMAC := hmac.New(sha256.New, key)

	if _, err := nonceHMAC.Write(value); err != nil {
		return nil, err
	}

	return nonceHMAC.Sum(nil), nil
}
