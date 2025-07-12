package repo

import "errors"

var (
	ErrExchangeIsNotDir  = errors.New("key exchange is not a directory")
	ErrRedactKeyNotFound = errors.New("redact key not found")
)
