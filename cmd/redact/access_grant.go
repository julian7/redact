package main

import (
	"bytes"
	"crypto"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

var accessGrantCmd = &cobra.Command{
	Use:   "grant [KEY...]",
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

	if len(args) > 0 {
		out, err := gpgutil.ExportKey(args)
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

	if len(keyEntries) <= 0 {
		return errors.New("nobody to grant access to")
	}

	masterkey, err := basicDo()
	if err != nil {
		cmdErrHandler(err)
	}
	toplevel, err := gitutil.TopLevel()
	if err != nil {
		cmdErrHandler(err)
	}

	saved := 0
	for _, key := range keyEntries {
		if err = saveKey(masterkey, key, toplevel); err != nil {
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

func saveKey(masterkey *files.MasterKey, key *openpgp.Entity, toplevel string) error {
	printKey(key)
	kxstub, err := masterkey.GetExchangeFilenameStubFor(toplevel, key.PrimaryKey.Fingerprint)
	if err != nil {
		return err
	}
	if err := saveMasterExchange(masterkey, kxstub, key); err != nil {
		return err
	}
	if err := savePubkeyExchange(masterkey, kxstub, key); err != nil {
		return err
	}
	return nil
}

func printKey(key *openpgp.Entity) {
	fmt.Printf(
		"KeyID: %s, fingerprint: %x\n",
		key.PrimaryKey.KeyIdShortString(),
		key.PrimaryKey.Fingerprint,
	)
	for _, id := range key.Identities {
		var expires string
		if id.SelfSignature.SigType != packet.SigTypePositiveCert {
			continue
		}
		sig := id.SelfSignature
		if sig.KeyLifetimeSecs == nil {
			expires = "no expiration"
		} else {
			expiry := sig.CreationTime.Add(time.Duration(*sig.KeyLifetimeSecs) * time.Second)
			if !expiry.After(time.Now()) {
				expires = fmt.Sprintf("%s (expired)", expiry)
			} else {
				expires = expiry.String()
			}
		}
		fmt.Printf(
			"  identity: %s, expires: %s\n",
			id.Name,
			expires,
		)
	}
}

func saveMasterExchange(masterkey *files.MasterKey, kxstub string, key *openpgp.Entity) error {
	masterName := files.ExchangeMasterKeyFile(kxstub)
	masterKeyReader, err := os.Open(masterkey.KeyFile())
	if err != nil {
		return errors.Wrap(err, "opening master key file")
	}
	defer masterKeyReader.Close()
	masterWriter, err := os.OpenFile(masterName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "opening exchange master key")
	}
	defer masterWriter.Close()
	hints := openpgp.FileHints{IsBinary: true}
	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: 9,
		},
		RSABits: 4096,
	}
	plain, err := openpgp.Encrypt(masterWriter, []*openpgp.Entity{key}, nil, &hints, &config)
	if err != nil {
		return errors.Wrap(err, "creating encryption stream")
	}
	defer plain.Close()
	_, err = io.Copy(plain, masterKeyReader)
	if err != nil {
		return errors.Wrap(err, "writing master key to encryption stream")
	}
	return nil
}

func savePubkeyExchange(masterkey *files.MasterKey, kxstub string, key *openpgp.Entity) error {
	pubkeyName := files.ExchangePubKeyFile(kxstub)
	pubkeyWriter, err := os.OpenFile(pubkeyName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "opening exchange pubkey file")
	}
	defer pubkeyWriter.Close()
	arm, err := armor.Encode(pubkeyWriter, openpgp.PublicKeyType, make(map[string]string))
	if err != nil {
		return errors.Wrap(err, "creating armor stream")
	}
	defer arm.Close()
	if err := key.Serialize(arm); err != nil {
		return errors.Wrap(err, "serializing public key to exchange store")
	}
	return nil
}
