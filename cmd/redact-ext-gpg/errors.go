package main

import "errors"

var (
	ErrGPGExtensionFlagUnset        = errors.New("GPG extension is skipped because FF_GIT_EXT is unset")
	ErrNoLocalKeyExchangeCopy       = errors.New("no local GnuPG identity has an encrypted Key Exchange copy")
	ErrIdentityHasNoKeyExchangeCopy = errors.New("GnuPG identity has no encrypted Key Exchange copy")
	ErrGPGIdentitySelectionRequired = errors.New("GnuPG identity selection is required")
	ErrGrantRecipientInputRequired  = errors.New("grant requires recipient=, file=, or armor=; see grant --help")
	ErrRevokeRecipientInputRequired = errors.New("revoke requires recipient, file, or armor settings")
	ErrNoRecipientPublicKeys        = errors.New("no recipient public keys were found")
	ErrMissingKeyExchangeCollection = errors.New("missing Key Exchange collection on stdin")
)
