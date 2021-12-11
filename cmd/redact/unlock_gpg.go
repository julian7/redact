package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
)

func (rt *Runtime) unlockGpgCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "gpg",
		Args:  cobra.NoArgs,
		Short: "Unlocks repository",
		Long: `Unlock repository

This command unlocks the repository using a GPG key.

By default, it detects your GnuPG keys by running gpg -K, and tries to match
them to the available encrypted keys in the key exchange directory. This
process won't make decisions for you, if you have multiple keys available. In
this case, you have to provide the appropriate key with the --gpgkey option.`,
		RunE: rt.unlockGpgDo,
	}

	flags := cmd.Flags()
	flags.StringP("gpgkey", "k", "", "Use specific GPG key")

	if err := rt.RegisterFlags("", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) unlockGpgDo(cmd *cobra.Command, args []string) error {
	var err error

	rt.MasterKey, err = files.NewMasterKey(rt.Logger)
	if err != nil {
		return fmt.Errorf("building master key: %w", err)
	}

	var key string

	key, err = rt.loadKeyFromGPG(rt.Viper.GetString("gpgkey"))
	if err != nil {
		if err == io.EOF {
			return nil
		}

		if key != "" {
			fmt.Printf("Hint: try to unlock by hand:\n\n")
			fmt.Printf(
				"  gpg -o key -u %s -d .redact/%s.key\n",
				key,
				key,
			)
			fmt.Println("  redact unlock key")
			fmt.Println("  rm key")
		}

		return err
	}

	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	err = sdk.TouchUp(rt.MasterKey, func(err error) {
		rt.Logger.Warn(err.Error())
	})
	if err != nil {
		return err
	}

	fmt.Println("Key is unlocked.")

	return nil
}

func (rt *Runtime) loadKeyFromFile(keyfile string) error {
	f, err := rt.MasterKey.Fs.OpenFile(keyfile, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("loading secret key from %s: %w", keyfile, err)
	}
	defer f.Close()

	return sdk.LoadMasterKeyFromReader(rt.MasterKey, f)
}

func (rt *Runtime) selectKey(keyname string) (*[]byte, error) {
	keys, warns, err := gpgutil.GetSecretKeys(keyname)
	if err != nil {
		return nil, err
	}

	if len(warns) > 0 {
		for _, item := range warns {
			rt.Logger.Warn(item)
		}
	}

	availableKeys := make([]int, 0, len(keys))

	for idx, key := range keys {
		stub, err := rt.MasterKey.GetExchangeFilenameStubFor(key)
		if err != nil {
			rt.Logger.Warnf("cannot get exchange filename for %x: %v", key, err)
			continue
		}

		masterFilename := files.ExchangeMasterKeyFile(stub)

		st, err := rt.MasterKey.Stat(masterFilename)
		if err != nil || st.IsDir() {
			continue
		}

		availableKeys = append(availableKeys, idx)
	}

	if len(availableKeys) > 1 {
		fmt.Println("Multiple keys found. Please specify one:")

		for _, idx := range availableKeys {
			pubKey, err := sdk.LoadPubkeysFromExchange(rt.MasterKey, keys[idx])
			if err != nil {
				rt.Logger.Warnf("%v", err)
				continue
			}

			for _, entity := range pubKey {
				gpgutil.PrintKey(entity)
			}
		}

		return nil, io.EOF
	}

	if len(availableKeys) < 1 {
		return nil, errors.New("no appropriate key found for unlock")
	}

	return &keys[availableKeys[0]], nil
}

func (rt *Runtime) loadKeyFromGPG(gpgkey string) (string, error) {
	key, err := rt.selectKey(gpgkey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", *key), sdk.LoadMasterKeyFromExchange(rt.MasterKey, *key)
}
