package sdk

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/spf13/afero"
)

type ErrExchangeDir struct {
	err error
}

func (e *ErrExchangeDir) Error() string {
	return e.err.Error()
}

func (e *ErrExchangeDir) Unwrap() error {
	err, ok := e.err.(interface{ Unwrap() error })
	if ok {
		return err.Unwrap()
	}
	return nil
}

// LoadSecretKeyFromExchange loads data from an encrypted secret key into provided
// object, and then saves it.
func LoadSecretKeyFromExchange(secretkey *files.SecretKey, fingerprint []byte) error {
	stub, err := secretkey.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return fmt.Errorf("finding key in exchange dir: %w", err)
	}

	reader, err := gpgutil.DecryptWithKey(
		files.ExchangeSecretKeyFile(stub),
		fingerprint,
	)
	if err != nil {
		return fmt.Errorf("decrypt secret key from exchange dir: %w", err)
	}

	defer reader.Close()

	return LoadSecretKeyFromReader(secretkey, reader)
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
func SaveSecretExchange(secretkey *files.SecretKey, key *openpgp.Entity) error {
	kxstub, err := secretkey.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	secretName := files.ExchangeSecretKeyFile(kxstub)

	secretWriter, err := os.OpenFile(secretName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exchange secret key: %w", err)
	}
	defer secretWriter.Close()

	r, w := io.Pipe()

	go func(writer io.WriteCloser) {
		_ = secretkey.SaveTo(writer)
		writer.Close()
	}(w)

	return gpgutil.Encrypt(r, secretWriter, key)
}

// LoadPubkeysFromExchange loads a public key from key exchange
func LoadPubkeysFromExchange(secretkey *files.SecretKey, fingerprint []byte) (openpgp.EntityList, error) {
	stub, err := secretkey.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("finding exchange public key: %w", err)
	}

	pubkey, err := gpgutil.LoadPubKeyFromFile(files.ExchangePubKeyFile(stub), true)
	if err != nil {
		return nil, fmt.Errorf("loading public key for %x: %w", fingerprint, err)
	}

	return pubkey, nil
}

// SavePubkeyExchange saves public OpenPGP key into key exchange
func SavePubkeyExchange(secretkey *files.SecretKey, key *openpgp.Entity) error {
	kxstub, err := secretkey.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	pubkeyName := files.ExchangePubKeyFile(kxstub)

	pubkeyWriter, err := os.OpenFile(pubkeyName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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
func UpdateSecretExchangeKeys(secretkey *files.SecretKey) (int, error) {
	kxdir, err := secretkey.ExchangeDir()
	if err != nil {
		return 0, &ErrExchangeDir{
			err: fmt.Errorf("fetching key exchange dir: %w", err),
		}
	}

	updated := 0

	err = afero.Walk(secretkey.Fs, kxdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !strings.HasSuffix(path, files.ExtKeyArmor) {
			return nil
		}

		fingerprintText := strings.TrimRight(filepath.Base(path), files.ExtKeyArmor)

		fingerprint, err := hex.DecodeString(fingerprintText)
		if err != nil {
			return nil
		}

		keys, err := LoadPubkeysFromExchange(secretkey, fingerprint)
		if err != nil {
			return fmt.Errorf("loading public key for %s: %w", fingerprintText, err)
		}

		if len(keys) != 1 {
			return fmt.Errorf("key %s has %d public keys", fingerprintText, len(keys))
		}

		updated++

		err = SaveSecretExchange(secretkey, keys[0])
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
