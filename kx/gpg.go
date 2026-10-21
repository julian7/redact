package kx

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
)

// SaveGPGKeyToKX saves secret key into key exchange, encrypted with OpenPGP key
func SaveGPGKeyToKX(redactRepo *repo.Repo, key *openpgp.Entity, writerCallback func(io.Writer) error) error {
	kxstub, err := redactRepo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	plaintext := bytes.Buffer{}
	if err = writerCallback(&plaintext); err != nil {
		return fmt.Errorf("reading key: %w", err)
	}

	cyphertext := bytes.Buffer{}
	if err = gpgutil.Encrypt(&plaintext, &cyphertext, key); err != nil {
		return fmt.Errorf("encrypting key: %w", err)
	}

	tmpfile, err := redactRepo.Workdir.TempFile(repo.DefaultKeyExchangeDir, "kx-")
	if err != nil {
		return fmt.Errorf("creating temp file for encrypted key: %w", err)
	}

	defer redactRepo.Workdir.Remove(tmpfile.Name()) // nolint:errcheck

	if _, err = io.Copy(tmpfile, &cyphertext); err != nil {
		_ = tmpfile.Close()

		return fmt.Errorf("writing encrypted key: %w", err)
	}

	if err = tmpfile.Close(); err != nil {
		return fmt.Errorf("closing encrypted key: %w", err)
	}

	if err = redactRepo.Workdir.Rename(tmpfile.Name(), repo.ExchangeSecretKeyFile(kxstub)); err != nil {
		return fmt.Errorf("updating encrypted key: %w", err)
	}

	return nil
}

// SaveGPGPubkeyToKX saves public OpenPGP key into key exchange
func SaveGPGPubkeyToKX(redactRepo *repo.Repo, key *openpgp.Entity) error {
	kxstub, err := redactRepo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	pubkeyWriter, err := redactRepo.Workdir.TempFile(repo.DefaultKeyExchangeDir, "kx-")
	if err != nil {
		return fmt.Errorf("creating temp file for public key: %w", err)
	}

	defer redactRepo.Workdir.Remove(pubkeyWriter.Name()) // nolint:errcheck

	if err := gpgutil.SavePubKey(pubkeyWriter, key, true); err != nil {
		_ = pubkeyWriter.Close()

		return fmt.Errorf("serializing public key to exchange store: %w", err)
	}

	if err := pubkeyWriter.Close(); err != nil {
		return fmt.Errorf("closing public key: %w", err)
	}

	if err := redactRepo.Workdir.Rename(pubkeyWriter.Name(), repo.ExchangePubKeyFile(kxstub)); err != nil {
		return fmt.Errorf("updating public key: %w", err)
	}

	return nil
}
