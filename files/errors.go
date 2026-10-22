package files

import "errors"

var (
	ErrAlreadyEncoded     = errors.New("already encoded")
	ErrExcessiveRights    = errors.New("excessive rights")
	ErrInsufficientRights = errors.New("insufficient rights")
	ErrInvalidPreamble    = errors.New("invalid file preamble")
	ErrInvalidKeyPreamble = errors.New("invalid key file preamble")
	ErrInvalidKeyType     = errors.New("invalid key type")
	ErrInvalidKey         = errors.New("invalid key")
	ErrNoPEMData          = errors.New("no data found in PEM")
	ErrPEMTypeMismatch    = errors.New("PEM type mismatch")
	ErrNoKeysLoaded       = errors.New("no keys loaded")
	ErrKeydirNotDirectory = errors.New("keydir is not a directory")
)
