package kx

import (
	"fmt"
	"io"

	"github.com/julian7/redact/files"
)

// LoadSecretKeyFromReader reads secret key from an io.Reader, and saves it into its
// place.
func LoadSecretKeyFromReader(secretkey *files.SecretKey, reader io.Reader) error {
	if err := secretkey.Read(reader); err != nil {
		return fmt.Errorf("reading unencrypted secret key: %w", err)
	}

	if err := secretkey.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	return nil
}
