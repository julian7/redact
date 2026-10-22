package main

import "errors"

var (
	// ErrExchangeIsNotDir mirrors what redact reports for the same problem, so
	// that users see one wording for one problem.
	ErrExchangeIsNotDir             = errors.New("key exchange is not a directory")
	ErrNoLocalKeyExchangeCopy       = errors.New("no local age identity has an encrypted Key Exchange copy")
	ErrIdentityHasNoKeyExchangeCopy = errors.New("age identity has no encrypted Key Exchange copy")
	ErrAgeIdentitySelectionRequired = errors.New("age identity selection is required")
	ErrGrantRecipientInputRequired  = errors.New("grant requires recipient= or file=; see grant --help")
	ErrRevokeRecipientInputRequired = errors.New("revoke requires recipient or file settings")
	ErrNoRecipientPublicKeys        = errors.New("no recipient public keys were found")
	ErrMissingKeyExchangeCollection = errors.New("missing Key Exchange collection on stdin")
)
