package encoder

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"fmt"
	"strings"
)

const (
	// TypeAES256GCM96 is AES256-GCM96 encoding type
	TypeAES256GCM96 = uint32(0)
	// TypeChaCha20Poly1305 is ChaCha20-Poly1305 encoding type
	TypeChaCha20Poly1305 = uint32(1)
)

// Encoder can encode and decode data
type AEAD interface {
	AEAD() (cipher.AEAD, error)
	KeySize() int
}

type Encoder struct {
	enc  AEAD
	hmac HMAC256
	name string
}

type Entry struct {
	factory Factory
	name    string
}
type Factory func([]byte) (AEAD, error)

var (
	ErrCyphertextSmall     = errors.New("ciphertext too small")
	ErrEncoderNotFound     = errors.New("encoder not found")
	ErrEncoderTypeExists   = errors.New("encoder type already exists")
	ErrEncoderTypeInvalid  = errors.New("invalid encoding type")
	ErrEncoderTypeNotFound = errors.New("encoder type not found")
	ErrInvalidChecksum     = errors.New("invalid HMAC checksum")
	ErrKeyTooSmall         = errors.New("key too small")

	encoders = map[uint32]Entry{
		TypeAES256GCM96:      {factory: NewAES256GCM96, name: "AES256-GCM96"},
		TypeChaCha20Poly1305: {factory: NewChaCha20Poly1305, name: "ChaCha20-Poly1305"},
	}
)

func FindEncoder(name string) (uint32, error) {
	name = strings.ToLower(name)
	for typ, entry := range encoders {
		if name == strings.ToLower(entry.name) {
			return typ, nil
		}
	}

	return 0, ErrEncoderNotFound
}

func Name(typ uint32) string {
	enc, ok := encoders[typ]
	if !ok {
		return "unknown type"
	}

	return enc.name
}

// NewEncoder returns a new Encoder initialized with a key handler
func NewEncoder(encType uint32, key []byte) (*Encoder, error) {
	Entry, ok := encoders[encType]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrEncoderTypeInvalid, encType)
	}

	enc, err := Entry.factory(key)
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
		name: Entry.name,
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
		return nil, ErrCyphertextSmall
	}

	nonce := ciphertext[:nonceLen]
	ciphertext = ciphertext[nonceLen:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting with %s: %w", e.name, err)
	}

	hmacSum, err := e.hmac.Sum(plaintext)

	if err != nil {
		return nil, fmt.Errorf("getting HMAC nonce: %w", err)
	}

	if !bytes.Equal(hmacSum[:len(nonce)], nonce) {
		return nil, ErrInvalidChecksum
	}

	return plaintext, nil
}

// RegisterEncoder registers a new encoder
func RegisterEncoder(encType uint32, name string, factory Factory) error {
	if _, ok := encoders[encType]; ok {
		return fmt.Errorf("%d: %w", encType, ErrEncoderTypeExists)
	}

	encoders[encType] = Entry{factory: factory, name: name}

	return nil
}

// UnregisterEncoder removes an encoder
func UnregisterEncoder(encType uint32) error {
	if _, ok := encoders[encType]; !ok {
		return fmt.Errorf("%d: %w", encType, ErrEncoderTypeNotFound)
	}

	delete(encoders, encType)

	return nil
}
