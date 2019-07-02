package main

import (
	"bytes"
	"io"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/log"
	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var accessGrantCmd = &cobra.Command{
	Use:   "grant [KEY...]",
	Short: "Grants access to collaborators with OpenPGP keys",
	RunE:  accessGrantDo,
}

func init() {
	flags := accessGrantCmd.Flags()
	flags.StringSliceP("openpgp", "p", nil, "import from OpenPGP file instead of gpg keyring")
	flags.StringSliceP("openpgp-armor", "a", nil, "import from OpenPGP ASCII Armored file instead of gpg keyring")
	accessCmd.AddCommand(accessGrantCmd)
	viper.BindPFlags(flags)
}

func accessGrantDo(cmd *cobra.Command, args []string) error {
	var keyEntries openpgp.EntityList

	pgpFiles := viper.GetStringSlice("openpgp")
	armorFiles := viper.GetStringSlice("opengpg-armor")

	if len(pgpFiles) > 0 {
		for _, pgpFile := range pgpFiles {
			entries, err := gpgutil.LoadPubKeyFromFile(pgpFile, false)
			if err != nil {
				log.Log().Warnf("loading public key: %v", err)
			}
			keyEntries = append(keyEntries, entries...)
		}
	}
	if len(armorFiles) > 0 {
		for _, pgpFile := range armorFiles {
			entries, err := gpgutil.LoadPubKeyFromFile(pgpFile, true)
			if err != nil {
				log.Log().Warnf("loading public key: %v", err)
			}
			keyEntries = append(keyEntries, entries...)
		}
	}

	if len(args) > 0 {
		out, err := gpgutil.ExportKey(args)
		if err != nil {
			cmdErrHandler(errors.Wrap(err, "exporting GPG key"))
			return nil
		}
		reader := bytes.NewReader(out)
		entries, err := gpgutil.LoadPubKey(reader, true)
		if err != nil {
			cmdErrHandler(errors.Wrap(err, "reading GPG key"))
		}
		keyEntries = append(keyEntries, entries...)
	}

	if len(keyEntries) <= 0 {
		return errors.New("nobody to grant access to")
	}

	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
	}

	saved := 0
	for _, key := range keyEntries {
		if err = saveKey(masterkey, key); err != nil {
			log.Log().Warnf("cannot save key: %v", err)
			continue
		}
		saved++
	}
	log.Log().Infof(
		"Added %d key%s. Don't forget to commit exchange files to the repository.",
		saved,
		map[bool]string{true: "", false: "s"}[saved == 1],
	)

	return nil
}

func saveKey(masterkey *files.MasterKey, key *openpgp.Entity) error {
	gpgutil.PrintKey(key)
	if err := saveMasterExchange(masterkey, key); err != nil {
		return err
	}
	if err := savePubkeyExchange(masterkey, key); err != nil {
		return err
	}
	return nil
}

func saveMasterExchange(masterkey *files.MasterKey, key *openpgp.Entity) error {
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

func savePubkeyExchange(masterkey *files.MasterKey, key *openpgp.Entity) error {
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
