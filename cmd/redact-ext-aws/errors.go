package main

import "errors"

var (
	ErrAlreadyWritten   = errors.New("secret already written to param store")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrMissingKeyID     = errors.New("missing keyid (KMS key ID)")
	ErrMissingParamPath = errors.New("missing param (Parameter Store Name)")
)
