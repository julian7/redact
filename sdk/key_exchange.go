package sdk

import (
	"io"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

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
