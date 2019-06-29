package encoder

import (
	"bytes"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/hkdf"
)

// DeriveKey computes derived key from context and from a specific version of AES key
func DeriveKey(aes, hmac []byte, length int) ([]byte, error) {
	if len(hmac) == 0 {
		return nil, errors.New("missing context")
	}
	der := bytes.NewBuffer(nil)
	der.Grow(length)
	stream := &io.LimitedReader{
		R: hkdf.New(sha256.New, aes, nil, hmac),
		N: int64(length),
	}

	n, err := der.ReadFrom(stream)
	if err != nil {
		return nil, errors.Wrap(err, "reading derived bytes")
	}
	if n != int64(length) {
		return nil, errors.Errorf("cannot read enough derived bytes; needed %d, got %d", length, n)
	}
	return der.Bytes(), nil
}
