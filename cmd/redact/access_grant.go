package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/julian7/redact/gpgutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

var accessGrantCmd = &cobra.Command{
	Use:   "grant [KEY]",
	Short: "Grant OpenPGP access",
	RunE:  accessGrantDo,
}

func init() {
	flags := accessCmd.Flags()
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
			out, err := os.Open(pgpFile)
			if err != nil {
				cmdErrHandler(errors.Wrap(err, "opening pgp file"))
				return nil
			}
			entries, err := openpgp.ReadKeyRing(out)
			if err != nil {
				cmdErrHandler(errors.Wrapf(err, "read keyring of pgp file %s", pgpFile))
			}
			keyEntries = append(keyEntries, entries...)
		}
	}
	if len(armorFiles) > 0 {
		for _, pgpFile := range armorFiles {
			out, err := os.Open(pgpFile)
			if err != nil {
				cmdErrHandler(errors.Wrap(err, "opening pgp asc file"))
				return nil
			}
			entries, err := openpgp.ReadArmoredKeyRing(out)
			if err != nil {
				cmdErrHandler(errors.Wrapf(err, "read keyring of pgp asc file %s", pgpFile))
			}
			keyEntries = append(keyEntries, entries...)
		}
	}
	if len(keyEntries) <= 0 {
		return errors.New("nobody to grant access to")
	}

	if len(args) > 0 {
		out, err := gpgutil.ExportKey(args[0])
		if err != nil {
			cmdErrHandler(errors.Wrap(err, "exporting GPG key"))
			return nil
		}
		reader := bytes.NewReader(out)
		entries, err := openpgp.ReadArmoredKeyRing(reader)
		if err != nil {
			cmdErrHandler(errors.Wrap(err, "reading GPG key"))
		}
		keyEntries = append(keyEntries, entries...)
	}
	for _, key := range keyEntries {
		fmt.Printf(
			"KeyID: %s, fingerprint: %x\n",
			key.PrimaryKey.KeyIdShortString(),
			key.PrimaryKey.Fingerprint,
		)
		for _, id := range key.Identities {
			if id.SelfSignature.SigType != packet.SigTypePositiveCert {
				continue
			}
			sig := id.SelfSignature
			expiry := sig.CreationTime.Add(time.Duration(*sig.KeyLifetimeSecs) * time.Second)
			fmt.Printf(
				"  identity: %s, expires: %s\n",
				id.Name,
				expiry,
			)
		}
	}
	return nil
}
