package main

import "errors"

var (
	ErrGPGKeyNotFound    = errors.New("nobody to grant access to")
	ErrExtensionNotFound = errors.New("extension not added")
	ErrOptions           = errors.New("invalid command line options")
	ErrSeek              = errors.New("cannot return to start of file")
	ErrNoSuitableKey     = errors.New("no suitable key found")
	ErrKeyAlreadyExists  = errors.New("secret key already exists")
)
