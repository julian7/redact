package sdk

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/spf13/afero"
	"golang.org/x/crypto/openpgp"
)

// LoadMasterKeyFromExchange loads data from an encrypted master key into provided
// object, and then saves it.
func LoadMasterKeyFromExchange(masterkey *files.MasterKey, fingerprint [20]byte) error {
	stub, err := masterkey.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return fmt.Errorf("finding key in exchange dir: %w", err)
	}

	reader, err := gpgutil.DecryptWithKey(
		files.ExchangeMasterKeyFile(stub),
		fingerprint,
	)
	if err != nil {
		return fmt.Errorf("decrypt master key from exchange dir: %w", err)
	}

	defer reader.Close()

	return LoadMasterKeyFromReader(masterkey, reader)
}

// LoadMasterKeyFromReader reads master key from an io.Reader, and saves it into its
// place.
func LoadMasterKeyFromReader(masterkey *files.MasterKey, reader io.Reader) error {
	if err := masterkey.Read(reader); err != nil {
		return fmt.Errorf("reading unencrypted master key: %w", err)
	}

	if err := masterkey.Save(); err != nil {
		return fmt.Errorf("saving master key: %w", err)
	}

	return nil
}

// SaveMasterExchange saves master key into key exchange, encrypted with OpenPGP key
func SaveMasterExchange(masterkey *files.MasterKey, key *openpgp.Entity) error {
	kxstub, err := masterkey.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}

	masterName := files.ExchangeMasterKeyFile(kxstub)

	masterWriter, err := os.OpenFile(masterName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exchange master key: %w", err)
	}
	defer masterWriter.Close()

	r, w := io.Pipe()

	go func(writer io.WriteCloser) {
		_ = masterkey.SaveTo(writer)
		writer.Close()
	}(w)

	return gpgutil.Encrypt(r, masterWriter, key)
}

// LoadPubkeysFromExchange loads a public key from key exchange
func LoadPubkeysFromExchange(masterkey *files.MasterKey, fingerprint [20]byte) (openpgp.EntityList, error) {
	stub, err := masterkey.GetExchangeFilenameStubFor(fingerprint)
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
func SavePubkeyExchange(masterkey *files.MasterKey, key *openpgp.Entity) error {
	kxstub, err := masterkey.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
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

// UpdateMasterExchangeKeys updates all key exchange master keys with new data
func UpdateMasterExchangeKeys(masterkey *files.MasterKey) (int, error) {
	kxdir, err := masterkey.ExchangeDir()
	if err != nil {
		return 0, fmt.Errorf("fetching key exchange dir: %w", err)
	}

	updated := 0

	err = afero.Walk(masterkey.Fs, kxdir, func(path string, info os.FileInfo, err error) error {
		var fingerprint [20]byte
		if err != nil {
			return nil
		}

		if !strings.HasSuffix(path, files.ExtKeyArmor) {
			return nil
		}

		fingerprintText := strings.TrimRight(filepath.Base(path), files.ExtKeyArmor)

		fingerprintData, err := hex.DecodeString(fingerprintText)
		if err != nil {
			return nil
		}

		copy(fingerprint[:], fingerprintData)

		keys, err := LoadPubkeysFromExchange(masterkey, fingerprint)
		if err != nil {
			return fmt.Errorf("loading public key for %s: %w", fingerprintText, err)
		}

		if len(keys) != 1 {
			return fmt.Errorf("key %s has %d public keys", fingerprintText, len(keys))
		}

		updated++

		err = SaveMasterExchange(masterkey, keys[0])
		if err != nil {
			return fmt.Errorf(
				"saving master key encrypted with key %s: %w",
				fingerprintText,
				err,
			)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("updating master key in exchange dir: %w", err)
	}

	return updated, nil
}
