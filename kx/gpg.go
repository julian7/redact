package kx

import (
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

	reader, err := gpgutil.DecryptWithKey(
		repo.ExchangeSecretKeyFile(stub),
		fingerprint,
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key from exchange dir: %w", err)
	}

	return reader, nil
}

// SaveGPGKeyToKX saves secret key into key exchange, encrypted with OpenPGP key
func SaveGPGKeyToKX(redactRepo *repo.Repo, key *openpgp.Entity, writerCallback func(io.Writer)) error {
	kxstub, err := redactRepo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint, nil)
	if err != nil {
		return err
	}

	secretName := repo.ExchangeSecretKeyFile(kxstub)

	secretWriter, err := os.OpenFile(secretName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exchange secret key: %w", err)
	}
	defer secretWriter.Close()

	r, w := io.Pipe()

	go func(writer io.WriteCloser) {
		writerCallback(writer)
		writer.Close()
	}(w)

	return gpgutil.Encrypt(r, secretWriter, key)
}

// LoadGPGPubkeysFromKX loads a public key from key exchange
func LoadGPGPubkeysFromKX(redactRepo *repo.Repo, fingerprint []byte) (openpgp.EntityList, error) {
	stub, err := redactRepo.GetExchangeFilenameStubFor(fingerprint, nil)
	if err != nil {
		return nil, fmt.Errorf("finding exchange public key: %w", err)
	}

	pubkey, err := gpgutil.LoadPubKeyFromFile(repo.ExchangePubKeyFile(stub), true)
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

	pubkeyName := repo.ExchangePubKeyFile(kxstub)

	pubkeyWriter, err := redactRepo.Workdir.OpenFile(pubkeyName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exchange pubkey file: %w", err)
	}

	defer pubkeyWriter.Close()

	if err := gpgutil.SavePubKey(pubkeyWriter, key, true); err != nil {
		return fmt.Errorf("serializing public key to exchange store: %w", err)
	}

	return nil
}

// UpdateGPGKeysInKX updates all key exchange secret keys with new data
func UpdateGPGKeysInKX(redactRepo *repo.Repo, writerCallback func(io.Writer)) (int, error) {
	kxdir := redactRepo.ExchangeDir()
	updated := 0

	err := util.Walk(redactRepo.Workdir, kxdir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}

		if !strings.HasSuffix(path, repo.ExtKeyArmor) {
			return nil
		}

		fingerprintText := strings.TrimRight(filepath.Base(path), repo.ExtKeyArmor)

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
