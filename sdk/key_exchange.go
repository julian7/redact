package sdk

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-billy/v5/util"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk/git"
)

// SecretKeyFromExchange loads data from an encrypted secret key into provided
// object, and then saves it.
func SecretKeyFromExchange(repo *git.Repo, fingerprint []byte) (io.ReadCloser, error) {
	stub, err := repo.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("finding key in exchange dir: %w", err)
	}

	reader, err := gpgutil.DecryptWithKey(
		git.ExchangeSecretKeyFile(stub),
		fingerprint,
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key from exchange dir: %w", err)
	}

	return reader, nil
}

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

// SaveSecretExchange saves secret key into key exchange, encrypted with OpenPGP key
func SaveSecretExchange(repo *git.Repo, key *openpgp.Entity, writerCallback func(io.Writer)) error {
	kxstub, err := repo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	secretName := git.ExchangeSecretKeyFile(kxstub)

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

// LoadPubkeysFromExchange loads a public key from key exchange
func LoadPubkeysFromExchange(repo *git.Repo, fingerprint []byte) (openpgp.EntityList, error) {
	stub, err := repo.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("finding exchange public key: %w", err)
	}

	pubkey, err := gpgutil.LoadPubKeyFromFile(git.ExchangePubKeyFile(stub), true)
	if err != nil {
		return nil, fmt.Errorf("loading public key for %x: %w", fingerprint, err)
	}

	return pubkey, nil
}

// SavePubkeyExchange saves public OpenPGP key into key exchange
func SavePubkeyExchange(repo *git.Repo, key *openpgp.Entity) error {
	kxstub, err := repo.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	pubkeyName := git.ExchangePubKeyFile(kxstub)

	pubkeyWriter, err := repo.Filesystem.OpenFile(pubkeyName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exchange pubkey file: %w", err)
	}

	defer pubkeyWriter.Close()

	if err := gpgutil.SavePubKey(pubkeyWriter, key, true); err != nil {
		return fmt.Errorf("serializing public key to exchange store: %w", err)
	}

	return nil
}

// UpdateSecretExchangeKeys updates all key exchange secret keys with new data
func UpdateSecretExchangeKeys(repo *git.Repo, writerCallback func(io.Writer)) (int, error) {
	kxdir := repo.ExchangeDir()
	updated := 0

	err := util.Walk(repo.Filesystem, kxdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}

		if !strings.HasSuffix(path, git.ExtKeyArmor) {
			return nil
		}

		fingerprintText := strings.TrimRight(filepath.Base(path), git.ExtKeyArmor)

		fingerprint, err := hex.DecodeString(fingerprintText)
		if err != nil {
			return nil // nolint:nilerr
		}

		keys, err := LoadPubkeysFromExchange(repo, fingerprint)
		if err != nil {
			return fmt.Errorf("loading public key for %s: %w", fingerprintText, err)
		}

		if len(keys) != 1 {
			return fmt.Errorf("key %s has %d public keys", fingerprintText, len(keys))
		}

		updated++

		err = SaveSecretExchange(repo, keys[0], writerCallback)
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
