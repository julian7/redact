package main

import "errors"

var (
	ErrAlreadyWritten  = errors.New("secret already written to key vault")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingKeyvault = errors.New("missing keyvault setting")
	ErrMissingSecret   = errors.New("missing secret setting")
)
