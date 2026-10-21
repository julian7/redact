package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/julian7/redact/ageutil"
	"github.com/julian7/redact/repo"
)

const (
	KeyExchangeDir     = repo.DefaultKeyExchangeDir
	ExtAgeSecret       = ".age.key"
	ExtAgePublic       = ".age.pub"
	GitAttributesFile  = repo.GitAttributesFile
	GitAttributesEntry = "*.key binary"
)

func exchangeDirectory() (string, error) {
	r := &repo.Repo{}
	if err := r.SetupRepo(); err != nil {
		return "", err
	}

	return filepath.Join(r.Workdir.Root(), KeyExchangeDir), nil
}

func ExchangeAgeSecretKeyFile(recipient string) string {
	return ageutil.RecipientID(recipient) + ExtAgeSecret
}

func exchangePath(stub string) (string, error) {
	r := &repo.Repo{}
	if err := r.SetupRepo(); err != nil {
		return "", err
	}

	filename, err := r.GetExchangeFilename(stub)
	if err != nil {
		return "", err
	}

	return filepath.Join(r.Workdir.Root(), filename), nil
}

func saveAgeKeyToKX(recipient age.Recipient, collection []byte) error {
	text, err := ageutil.RecipientText(recipient)
	if err != nil {
		return fmt.Errorf("identifying recipient: %w", err)
	}

	filename, err := exchangePath(ExchangeAgeSecretKeyFile(text))
	if err != nil {
		return err
	}

	var ciphertext bytes.Buffer

	if err := ageutil.Encrypt(bytes.NewReader(collection), &ciphertext, recipient); err != nil {
		return fmt.Errorf("encrypting Key Exchange collection: %w", err)
	}

	if err := writeExchangeFile(filename, ciphertext.Bytes()); err != nil {
		return err
	}

	publicFilename := filepath.Join(filepath.Dir(filename), ageutil.RecipientID(text)+ExtAgePublic)

	return writeExchangeFile(publicFilename, []byte(text+"\n"))
}

func writeExchangeFile(filename string, contents []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(filename), "kx-")
	if err != nil {
		return fmt.Errorf("creating encrypted Key Exchange file: %w", err)
	}
	defer os.Remove(temporary.Name()) //nolint:errcheck

	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()

		return fmt.Errorf("writing encrypted Key Exchange file: %w", err)
	}

	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing encrypted Key Exchange file: %w", err)
	}

	if err := os.Rename(temporary.Name(), filename); err != nil {
		return fmt.Errorf("replacing encrypted Key Exchange file: %w", err)
	}

	return nil
}
