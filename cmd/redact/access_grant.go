package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
)

func (rt *Runtime) accessGrantCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "grant [KEY...]",
		Args:    cobra.ArbitraryArgs,
		Short:   "Grants access to collaborators with OpenPGP keys",
		PreRunE: rt.RetrieveMasterKey,
		RunE:    rt.accessGrantDo,
	}

	flags := cmd.Flags()
	flags.StringSliceP("openpgp", "p", nil, "import from OpenPGP file instead of gpg keyring")
	flags.StringSliceP("openpgp-armor", "a", nil, "import from OpenPGP ASCII Armored file instead of gpg keyring")

	if err := rt.RegisterFlags("", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) accessGrantDo(cmd *cobra.Command, args []string) error { //nolint:funlen
	var keyEntries openpgp.EntityList

	rt.loadKeys(rt.Viper.GetStringSlice("openpgp"), false, &keyEntries)
	rt.loadKeys(rt.Viper.GetStringSlice("openpgp-armor"), true, &keyEntries)

	if len(args) > 0 {
		out, err := gpgutil.ExportKey(args)
		if err != nil {
			return fmt.Errorf("exporting GPG key: %w", err)
		}

		reader := bytes.NewReader(out)

		entries, err := gpgutil.LoadPubKey(reader, true)
		if err != nil {
			return fmt.Errorf("reading GPG key: %w", err)
		}

		keyEntries = append(keyEntries, entries...)
	}

	if len(keyEntries) == 0 {
		return errors.New("nobody to grant access to")
	}

	saved := 0

	for _, key := range keyEntries {
		if err := rt.saveKey(key); err != nil {
			rt.Logger.Warnf("cannot save key: %v", err)
			continue
		}
		saved++
	}

	rt.Logger.Infof(
		"Added %d key%s. Don't forget to commit exchange files to the repository.",
		saved,
		map[bool]string{true: "", false: "s"}[saved == 1],
	)

	return nil
}

func (rt *Runtime) loadKeys(pgpFiles []string, isArmor bool, keyEntries *openpgp.EntityList) {
	if len(pgpFiles) == 0 {
		return
	}

	for _, pgpFile := range pgpFiles {
		entries, err := gpgutil.LoadPubKeyFromFile(pgpFile, isArmor)
		if err != nil {
			rt.Logger.Warnf("loading public key: %v", err)
			continue
		}

		*keyEntries = append(*keyEntries, entries...)
	}
}

func (rt *Runtime) saveKey(key *openpgp.Entity) error {
	gpgutil.PrintKey(key)

	if err := sdk.SaveMasterExchange(rt.MasterKey, key); err != nil {
		return err
	}

	if err := sdk.SavePubkeyExchange(rt.MasterKey, key); err != nil {
		return err
	}

	return nil
}
