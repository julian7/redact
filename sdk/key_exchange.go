package sdk

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"golang.org/x/crypto/openpgp"
)

// LoadMasterKeyFromExchange loads data from an encrypted master key into provided
// object, and then saves it.
func LoadMasterKeyFromExchange(masterkey *files.MasterKey, fingerprint [20]byte) error {
	stub, err := masterkey.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return errors.Wrap(err, "finding key in exchange dir")
	}
	reader, err := gpgutil.DecryptWithKey(
		files.ExchangeMasterKeyFile(stub),
		fingerprint,
	)
	defer reader.Close()
	if err := masterkey.Read(reader); err != nil {
		return errors.Wrap(err, "reading unencrypted master key")
	}
	if err := masterkey.Save(); err != nil {
		return errors.Wrap(err, "saving master key")
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
		return errors.Wrap(err, "opening exchange master key")
	}
	defer masterWriter.Close()
	r, w := io.Pipe()
	go func(writer io.WriteCloser) {
		masterkey.SaveTo(writer)
		writer.Close()
	}(w)
	return gpgutil.Encrypt(r, masterWriter, key)
}

// LoadPubkeysFromExchange loads a public key from key exchange
func LoadPubkeysFromExchange(masterkey *files.MasterKey, fingerprint [20]byte) (openpgp.EntityList, error) {
	stub, err := masterkey.GetExchangeFilenameStubFor(fingerprint)
	if err != nil {
		return nil, errors.Wrap(err, "finding exchange public key")
	}
	pubkey, err := gpgutil.LoadPubKeyFromFile(files.ExchangePubKeyFile(stub), true)
	if err != nil {
		return nil, errors.Wrapf(err, "loading public key for %x", fingerprint)
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
		return errors.Wrap(err, "opening exchange pubkey file")
	}
	defer pubkeyWriter.Close()
	if err := gpgutil.SavePubKey(pubkeyWriter, key, true); err != nil {
		return errors.Wrap(err, "serializing public key to exchange store")
	}
	return nil
}

// UpdateMasterExchangeKeys updates all key exchange master keys with new data
func UpdateMasterExchangeKeys(masterkey *files.MasterKey) (int, error) {
	kxdir, err := masterkey.ExchangeDir()
	if err != nil {
		return 0, errors.Wrap(err, "fetching key exchange dir")
	}
	updated := 0
	afero.Walk(masterkey.Fs, kxdir, func(path string, info os.FileInfo, err error) error {
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
			return errors.Wrapf(err, "loading public key for %s", fingerprintText)
		}
		if len(keys) != 1 {
			return errors.Errorf("key %s has %d public keys", fingerprintText, len(keys))
		}
		updated++
		return errors.Wrapf(SaveMasterExchange(masterkey, keys[0]), "saving master key encrypted with key %s", fingerprintText)
	})
	return updated, nil
}
