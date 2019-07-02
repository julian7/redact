package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/log"
	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlocks repository",
	Run:   unlockDo,
}

func init() {
	flags := unlockCmd.Flags()
	flags.StringP("gpgkey", "k", "", "Use specific GPG key")
	rootCmd.AddCommand(unlockCmd)
	viper.BindPFlags(flags)
}

type keyFound struct {
	key  [20]byte
	stub string
}

func unlockDo(cmd *cobra.Command, args []string) {
	toplevel, err := gitutil.TopLevel()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	masterkey, err := files.NewMasterKey()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "building master key"))
		return
	}
	keys, err := gpgutil.GetSecretKeys(viper.GetString("gpgkey"))
	if err != nil {
		cmdErrHandler(err)
		return
	}
	availableKeys := []*keyFound{}

	for _, key := range keys {
		stub, err := masterkey.GetExchangeFilenameStubFor(toplevel, key)
		if err != nil {
			cmdErrHandler(err)
			return
		}
		pubkeyFilename := files.ExchangePubKeyFile(stub)
		st, err := masterkey.Stat(pubkeyFilename)
		if err != nil || st.IsDir() {
			continue
		}
		item := &keyFound{stub: stub}
		copy(item.key[:], key[:])

		availableKeys = append(availableKeys, item)
	}

	if len(availableKeys) > 2 {
		l := log.Log()
		fmt.Println("Multiple keys found. Please specify one:")
		for _, key := range availableKeys {
			pubKey, err := gpgutil.LoadPubKeyFromFile(files.ExchangePubKeyFile(key.stub), true)
			if err != nil {
				l.Warnf("error loading public key for %x: %v", key.key, err)
				continue
			}
			for _, entity := range pubKey {
				gpgutil.PrintKey(entity)
			}
		}
		return
	} else if len(availableKeys) < 1 {
		cmdErrHandler(errors.New("no appropriate key found for unlock"))
		return
	}
	reader, err := gpgutil.DecryptWithKey(files.ExchangeMasterKeyFile(availableKeys[0].stub), availableKeys[0].key)
	if err != nil {
		cmdErrHandler(err)
		return
	}
	defer reader.Close()
	if err := masterkey.Read(reader); err != nil {
		cmdErrHandler(err)
	}
	if err := masterkey.Save(); err != nil {
		cmdErrHandler(err)
	}
	if err := sdk.SaveGitSettings(); err != nil {
		cmdErrHandler(err)
	}
	if err := sdk.TouchUp(masterkey); err != nil {
		cmdErrHandler(err)
	}
	fmt.Println("Key is unlocked.")
}
