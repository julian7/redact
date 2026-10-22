// Package ageutil wraps the filippo.io/age library with the small set of
// helpers the age Key Exchange extension needs: stream encryption and
// decryption, parsing of single recipient and identity strings, reading of
// identities and recipients files, including the passphrase-protected identity
// files that hold their keys as age ciphertext, and discovery of local
// identities.
package ageutil

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

func Encrypt(reader io.Reader, writer io.Writer, recipients ...age.Recipient) error {
	encrypter, err := age.Encrypt(writer, recipients...)
	if err != nil {
		return fmt.Errorf("starting age encryption: %w", err)
	}

	if _, err := io.Copy(encrypter, reader); err != nil {
		_ = encrypter.Close()

		return fmt.Errorf("writing age plaintext: %w", err)
	}

	if err := encrypter.Close(); err != nil {
		return fmt.Errorf("closing age encryption: %w", err)
	}

	return nil
}

func Decrypt(reader io.Reader, writer io.Writer, identities ...age.Identity) error {
	decrypter, err := age.Decrypt(reader, identities...)
	if err != nil {
		return fmt.Errorf("decrypting age ciphertext: %w", err)
	}

	if _, err := io.Copy(writer, decrypter); err != nil {
		return fmt.Errorf("reading age plaintext: %w", err)
	}

	return nil
}

// The age library exposes generic parsing only for readers, so dispatch by type for a single string.
func ParseRecipient(text string) (age.Recipient, error) {
	switch {
	case strings.HasPrefix(text, "age1pq1"):
		return age.ParseHybridRecipient(text)
	case strings.HasPrefix(text, "age1"):
		return age.ParseX25519Recipient(text)
	default:
		// The value may be a private key handed to a recipient argument, so do
		// not echo it back.
		return nil, errors.New("unknown age recipient type")
	}
}

func ParseIdentity(text string) (age.Identity, error) {
	switch {
	case strings.HasPrefix(text, "AGE-SECRET-KEY-1"):
		return age.ParseX25519Identity(text)
	case strings.HasPrefix(text, "AGE-SECRET-KEY-PQ-1"):
		return age.ParseHybridIdentity(text)
	default:
		return nil, errors.New("unknown age identity type")
	}
}

// Treat missing files as unavailable identities so discovery can continue.
func ParseIdentitiesFile(filename string, passphrase PassphraseFunc) ([]age.Identity, error) {
	file, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("opening age identities file: %w", err)
	}
	defer file.Close()

	return readIdentities(filename, bufio.NewReader(file), passphrase)
}

func ParseRecipientsFile(filename string) ([]age.Recipient, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening age recipients file: %w", err)
	}
	defer file.Close()

	recipients, err := age.ParseRecipients(file)
	if err != nil {
		if err.Error() == "no recipients found" {
			return nil, nil
		}

		return nil, fmt.Errorf("reading age recipients file: %w", err)
	}

	return recipients, nil
}

// The Recipient interface lacks String, so dispatch by concrete type.
func RecipientText(recipient age.Recipient) (string, error) {
	switch rec := recipient.(type) {
	case *age.X25519Recipient:
		return rec.String(), nil
	case *age.HybridRecipient:
		return rec.String(), nil
	default:
		return "", fmt.Errorf("unsupported age recipient type %T", recipient)
	}
}

func RecipientString(identity age.Identity) (string, error) {
	switch id := identity.(type) {
	case *age.X25519Identity:
		return id.Recipient().String(), nil
	case *age.HybridIdentity:
		return id.Recipient().String(), nil
	default:
		return "", fmt.Errorf("unsupported age identity type %T", identity)
	}
}

func PrintRecipientTo(writer io.Writer, recipient string) error {
	// Keep X25519 recipients whole while bounding hybrid recipient output.
	const previewLength = 62

	preview := recipient
	if len(preview) > previewLength {
		preview = preview[:previewLength] + "..."
	}

	if _, err := fmt.Fprintf(writer, "age recipient: %s (%s)\n", preview, RecipientID(recipient)); err != nil {
		return fmt.Errorf("printing age recipient: %w", err)
	}

	return nil
}

// Use 160 bits of SHA-256 so identifiers stay compact even for hybrid recipients.
func RecipientID(recipient string) string {
	digest := sha256.Sum256([]byte(recipient))

	return hex.EncodeToString(digest[:20])
}

func DefaultIdentityFile() string {
	if env := os.Getenv("AGE_IDENTITIES_FILE"); env != "" {
		return env
	}

	config, err := userConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(config, "age", "identities")
}

func ReadIdentities(reader io.Reader, name string, passphrase PassphraseFunc) ([]age.Identity, error) {
	return readIdentities(name, bufio.NewReader(reader), passphrase)
}
