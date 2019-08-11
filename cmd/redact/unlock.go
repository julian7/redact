package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/log"
	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Args:  cobra.NoArgs,
	Short: "Unlocks repository",
	Long: `Unlock repository

This command is able to unlock a repository, or to obtain a new version of
the master key.

By default, it detects your GnuPG keys by running gpg -K, and tries to match
them to the available encrypted keys in the key exchange directory. This
process won't make decisions for you, if you have multiple keys available. In
this case, you have to provide the appropriate key with the --gpgkey option.`,
	Run: unlockDo,
}

func init() {
	flags := unlockCmd.Flags()
	flags.StringP("gpgkey", "k", "", "Use specific GPG key")
	rootCmd.AddCommand(unlockCmd)
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
}

func unlockDo(cmd *cobra.Command, args []string) {
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
	availableKeys := make([]int, 0, len(keys))

	l := log.Log()

	for idx, key := range keys {
		stub, err := masterkey.GetExchangeFilenameStubFor(key)
		if err != nil {
			l.Warnf("cannot get exchange filename for %x: %v", key, err)
			continue
		}
		masterFilename := files.ExchangeMasterKeyFile(stub)
		st, err := masterkey.Stat(masterFilename)
		if err != nil || st.IsDir() {
			continue
		}

		availableKeys = append(availableKeys, idx)
	}

	if len(availableKeys) > 2 {
		fmt.Println("Multiple keys found. Please specify one:")
		for _, idx := range availableKeys {
			pubKey, err := sdk.LoadPubkeysFromExchange(masterkey, keys[idx])
			if err != nil {
				l.Warnf("%v", err)
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
	if err := sdk.LoadMasterKeyFromExchange(masterkey, keys[availableKeys[0]]); err != nil {
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
