package main

import "errors"

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrMissingKeyvault    = errors.New("missing keyvault setting")
	ErrMissingSecret      = errors.New("missing secret setting")
	ErrNoExportedKey      = errors.New("secret contains no exported key")
	ErrMissingExportedKey = errors.New("missing exported key on stdin")
	ErrNotApplicable      = errors.New("not applicable")
	ErrNotSupported       = errors.New("unsupported extension command")
)
