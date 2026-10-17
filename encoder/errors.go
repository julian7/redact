package encoder

import "errors"

var (
	ErrCyphertextSmall     = errors.New("ciphertext too small")
	ErrEncoderNotFound     = errors.New("encoder not found")
	ErrEncoderTypeExists   = errors.New("encoder type already exists")
	ErrEncoderTypeInvalid  = errors.New("invalid encoding type")
	ErrEncoderTypeNotFound = errors.New("encoder type not found")
	ErrInvalidChecksum     = errors.New("invalid HMAC checksum")
	ErrKeyTooSmall         = errors.New("key too small")
)
