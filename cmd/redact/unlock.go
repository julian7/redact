package main

import (
	"fmt"
	"io"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (rt *Runtime) unlockCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "unlock [KEYFILE]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Unlocks repository",
		Long: `Unlock repository

This command is able to unlock a repository, or to obtain a new version of
the master key.

By default, it detects your GnuPG keys by running gpg -K, and tries to match
them to the available encrypted keys in the key exchange directory. This
process won't make decisions for you, if you have multiple keys available. In
this case, you have to provide the appropriate key with the --gpgkey option.

Alternatively, an unlocked master key can be provided. This allows unlocking
the repository where GnuPG (or the private key) is not available.`,
		RunE: rt.unlockDo,
	}

	flags := cmd.Flags()
	flags.StringP("gpgkey", "k", "", "Use specific GPG key")

	if err := rt.RegisterFlags("", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) unlockDo(cmd *cobra.Command, args []string) error {
	masterkey, err := files.NewMasterKey(rt.Logger)
	if err != nil {
		return errors.Wrap(err, "building master key")
	}

	if len(args) == 1 {
		err = loadKeyFromFile(masterkey, args[0])
	} else {
		err = rt.loadKeyFromGPG(masterkey, rt.Viper.GetString("gpgkey"))
	}

	if err != nil {
		if err != io.EOF {
			return err
		}

		return nil
	}

	if err := sdk.SaveGitSettings(); err != nil {
		return err
	}

	if err := sdk.TouchUp(masterkey); err != nil {
		return err
	}

	fmt.Println("Key is unlocked.")

	return nil
}

func loadKeyFromFile(masterkey *files.MasterKey, keyfile string) error {
	f, err := masterkey.Fs.OpenFile(keyfile, os.O_RDONLY, 0600)
	if err != nil {
		return errors.Wrapf(err, "loading secret key from %s", keyfile)
	}
	defer f.Close()

	return sdk.LoadMasterKeyFromReader(masterkey, f)
}

func (rt *Runtime) loadKeyFromGPG(masterkey *files.MasterKey, keyname string) error {
	keys, err := gpgutil.GetSecretKeys(keyname)
	if err != nil {
		return err
	}

	availableKeys := make([]int, 0, len(keys))

	for idx, key := range keys {
		stub, err := masterkey.GetExchangeFilenameStubFor(key)
		if err != nil {
			rt.Logger.Warnf("cannot get exchange filename for %x: %v", key, err)
			continue
		}

		masterFilename := files.ExchangeMasterKeyFile(stub)

		st, err := masterkey.Stat(masterFilename)
		if err != nil || st.IsDir() {
			continue
		}

		availableKeys = append(availableKeys, idx)
	}

	if len(availableKeys) > 1 {
		fmt.Println("Multiple keys found. Please specify one:")

		for _, idx := range availableKeys {
			pubKey, err := sdk.LoadPubkeysFromExchange(masterkey, keys[idx])
			if err != nil {
				rt.Logger.Warnf("%v", err)
				continue
			}

			for _, entity := range pubKey {
				gpgutil.PrintKey(entity)
			}
		}

		return io.EOF
	}

	if len(availableKeys) < 1 {
		return errors.New("no appropriate key found for unlock")
	}

	return sdk.LoadMasterKeyFromExchange(masterkey, keys[availableKeys[0]])
}
