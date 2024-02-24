package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk"
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) accessGrantCmd() *cli.Command {
	return &cli.Command{
		Name:      "grant",
		Usage:     "Grants access to collaborators with OpenPGP keys",
		ArgsUsage: "[KEY...]",
		Before:    rt.LoadSecretKey,
		Action:    rt.accessGrantDo,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:      "openpgp",
				Aliases:   []string{"p"},
				Usage:     "import from OpenPGP file instead of gpg keyring",
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:    "openpgp-armor",
				Aliases: []string{"a"},
				Usage:   "import from OpenPGP ASCII Armored file instead of gpg keyring",
			},
		},
	}
}

func (rt *Runtime) accessGrantDo(ctx *cli.Context) error {
	var keyEntries openpgp.EntityList

	rt.loadKeys(ctx.StringSlice("openpgp"), false, &keyEntries)
	rt.loadKeys(ctx.StringSlice("openpgp-armor"), true, &keyEntries)

	args := ctx.Args()
	if args.Len() > 0 {
		out, err := gpgutil.ExportKey(args.Slice())
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

	err := sdk.SaveSecretExchange(rt.Repo, key, func(w io.Writer) {
		err := rt.SecretKey.SaveTo(w)
		if err != nil {
			rt.Warn(err)
		}
	})

	if err != nil {
		return err
	}

	if err := sdk.SavePubkeyExchange(rt.Repo, key); err != nil {
		return err
	}

	return nil
}
