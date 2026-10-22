package main

import "errors"

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrMissingKeyID       = errors.New("missing keyid (KMS key ID)")
	ErrMissingParamPath   = errors.New("missing param (Parameter Store Name)")
	ErrNoExportedKey      = errors.New("parameter contains no exported key")
	ErrMissingExportedKey = errors.New("missing exported key on stdin")
	ErrNotApplicable      = errors.New("not applicable")
	ErrNotSupported       = errors.New("unsupported extension command")
)
