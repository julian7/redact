package encoder

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/pkg/errors"
)

// AES256GCM96 can take a KeyHandler, and stores encryption and HMAC keys
type AES256GCM96 struct {
	key  []byte
	hmac []byte
}

// NewAES256GCM96 returns a new Encoder initialized with a key handler
func NewAES256GCM96(aes, hmac []byte) (Encoder, error) {
	var encoder AES256GCM96
	fullLength := 64
	length := fullLength / 2
	key, err := DeriveKey(aes, hmac, fullLength)
	if err != nil {
		return nil, err
	}
	if len(key) < fullLength {
		return nil, errors.New("could not derive key; input is too small")
	}
	encoder.key = key[:length]
	if len(encoder.key) != length {
		return nil, errors.New("could not derive encryption key; length doesn't match")
	}
	encoder.hmac = key[length:]
	if len(encoder.hmac) != length {
		return nil, errors.New("could not derive hmac key; length doesn't match")
	}
	return encoder, nil
}

// Encode takes a value, and encrypts it in a convergent way, making sure the same
// input provides the same output every time, while not leaking secret information
// about its encryption keys
func (enc *AES256GCM96) Encode(value []byte) ([]byte, error) {
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
func (enc *AES256GCM96) Decode(ciphertext []byte) ([]byte, error) {
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
		return nil, errors.Wrap(err, "decrypting with AES256-GCM")
	}

	hmacSum, err := calculateHMAC(enc.hmac, plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "calculating HMAC")
	}
	if bytes.Compare(hmacSum[:gcm.NonceSize()], nonce) != 0 {
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
	nonceHMAC.Write(value)
	return nonceHMAC.Sum(nil), nil
}
