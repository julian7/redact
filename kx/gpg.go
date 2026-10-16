package kx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-billy/v5/util"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
)

// SecretKeyFromExchange loads data from an encrypted secret key into provided
// object, and then saves it.
func SecretKeyFromExchange(redactRepo *repo.Repo, fingerprint []byte) (io.ReadCloser, error) {
	stub, err := redactRepo.GetExchangeFilenameStubFor(fingerprint, nil)
	if err != nil {
		return nil, fmt.Errorf("finding key in exchange dir: %w", err)
	}

	ciphertext, err := redactRepo.Workdir.Open(repo.ExchangeSecretKeyFile(stub))
	if err != nil {
		return nil, fmt.Errorf("opening exchange secret key: %w", err)
	}
	defer ciphertext.Close()

	reader, err := gpgutil.DecryptWithKey(ciphertext, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key from exchange dir: %w", err)
	}

	return reader, nil
}

// SaveGPGKeyToKX saves secret key into key exchange, encrypted with OpenPGP key
func SaveGPGKeyToKX(redactRepo *repo.Repo, key *openpgp.Entity, writerCallback func(io.Writer) error) error {
	kxstub, err := redactRepo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint, nil)
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

// LoadGPGPubkeysFromKX loads a public key from key exchange
func LoadGPGPubkeysFromKX(redactRepo *repo.Repo, fingerprint []byte) (openpgp.EntityList, error) {
	stub, err := redactRepo.GetExchangeFilenameStubFor(fingerprint, nil)
	if err != nil {
		return nil, fmt.Errorf("finding exchange public key: %w", err)
	}

	file, err := redactRepo.Workdir.Open(repo.ExchangePubKeyFile(stub))
	if err != nil {
		return nil, fmt.Errorf("opening public key for %x: %w", fingerprint, err)
	}
	defer file.Close()

	pubkey, err := gpgutil.LoadPubKey(file, true)
	if err != nil {
		return nil, fmt.Errorf("loading public key for %x: %w", fingerprint, err)
	}

	return pubkey, nil
}

// SaveGPGPubkeyToKX saves public OpenPGP key into key exchange
func SaveGPGPubkeyToKX(redactRepo *repo.Repo, key *openpgp.Entity) error {
	kxstub, err := redactRepo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint, nil)
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

// UpdateGPGKeysInKX updates all key exchange secret keys with new data
func UpdateGPGKeysInKX(redactRepo *repo.Repo, writerCallback func(io.Writer) error) (int, error) {
	kxdir := repo.DefaultKeyExchangeDir
	updated := 0

	err := util.Walk(redactRepo.Workdir, kxdir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}

		if !strings.HasSuffix(path, repo.ExtKeyArmor) {
			return nil
		}

		fingerprintText := strings.TrimSuffix(filepath.Base(path), repo.ExtKeyArmor)

		fingerprint, err := hex.DecodeString(fingerprintText)
		if err != nil {
			return nil // nolint:nilerr
		}

		keys, err := LoadGPGPubkeysFromKX(redactRepo, fingerprint)
		if err != nil {
			return fmt.Errorf("loading public key for %s: %w", fingerprintText, err)
		}

		if len(keys) != 1 {
			return fmt.Errorf("key %s has %d public keys", fingerprintText, len(keys))
		}

		updated++

		err = SaveGPGKeyToKX(redactRepo, keys[0], writerCallback)
		if err != nil {
			return fmt.Errorf(
				"saving secret key encrypted with key %s: %w",
				fingerprintText,
				err,
			)
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("updating secret key in exchange dir: %w", err)
	}

	return updated, nil
}
