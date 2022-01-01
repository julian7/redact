package encoder

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"fmt"
)

const (
	// TypeAES256GCM96 is AES256-GCM96 encoding type
	TypeAES256GCM96 = iota
	// TypeChaCha20Poly1305 is ChaCha20-Poly1305 encoding type
	TypeChaCha20Poly1305
)

// Encoder can encode and decode data
type AEAD interface {
	AEAD() (cipher.AEAD, error)
	KeySize() int
	String() string
}

type Encoder struct {
	enc  AEAD
	hmac HMAC256
}

type Factory func([]byte) (AEAD, error)

var (
	ErrKeyTooSmall = errors.New("key too small")

	encoders = map[int]Factory{
		TypeAES256GCM96:      NewAES256GCM96,
		TypeChaCha20Poly1305: NewChaCha20Poly1305,
	}
)

// NewEncoder returns a new Encoder initialized with a key handler
func NewEncoder(encType int, key []byte) (*Encoder, error) {
	encoder, ok := encoders[encType]
	if !ok {
		return nil, fmt.Errorf("invalid encoding type %d", encType)
	}

	enc, err := encoder(key)
	if err != nil {
		return nil, fmt.Errorf("%s encoding error: %w", enc, err)
	}

	keysize := enc.KeySize()
	if len(key) < keysize+HMAC256KeySize {
		return nil, ErrKeyTooSmall
	}

	return &Encoder{
		enc:  enc,
		hmac: NewHMAC256(key[keysize : keysize+HMAC256KeySize]),
	}, nil
}

// Encode takes a value, and encrypts it in a convergent way, making sure the same
// input provides the same output every time, while not leaking secret information
// about its encryption keys
func (e *Encoder) Encode(plaintext []byte) ([]byte, error) {
	nonce, err := e.hmac.Sum(plaintext)
	if err != nil {
		return nil, fmt.Errorf("getting HMAC nonce: %w", err)
	}

	aead, err := e.enc.AEAD()
	if err != nil {
		return nil, err
	}

	nonce = nonce[:aead.NonceSize()]
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	ciphertext = append(nonce, ciphertext...)

	return ciphertext, nil
}

// Decode takes a byte stream, and decrypts its contents
func (e *Encoder) Decode(ciphertext []byte) ([]byte, error) {
	aead, err := e.enc.AEAD()
	if err != nil {
		return nil, err
	}

	nonceLen := aead.NonceSize()
	if len(ciphertext) < nonceLen {
		return nil, errors.New("ciphertext too small")
	}

	nonce := ciphertext[:nonceLen]
	ciphertext = ciphertext[nonceLen:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting with %s: %w", e.enc.String(), err)
	}

	hmacSum, err := e.hmac.Sum(plaintext)

	if err != nil {
		return nil, fmt.Errorf("getting HMAC nonce: %w", err)
	}

	if !bytes.Equal(hmacSum[:len(nonce)], nonce) {
		return nil, errors.New("HMAC checksum invalid")
	}

	return plaintext, nil
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
