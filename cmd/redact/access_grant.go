package main

import (
	"bytes"

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
	Args:  cobra.ArbitraryArgs,
	Short: "Grants access to collaborators with OpenPGP keys",
	RunE:  accessGrantDo,
}

func init() {
	flags := accessGrantCmd.Flags()
	flags.StringSliceP("openpgp", "p", nil, "import from OpenPGP file instead of gpg keyring")
	flags.StringSliceP("openpgp-armor", "a", nil, "import from OpenPGP ASCII Armored file instead of gpg keyring")
	accessCmd.AddCommand(accessGrantCmd)
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
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
	if err := sdk.SaveMasterExchange(masterkey, key); err != nil {
		return err
	}
	if err := sdk.SavePubkeyExchange(masterkey, key); err != nil {
		return err
	}
	return nil
}
