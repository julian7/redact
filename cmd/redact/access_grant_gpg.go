package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/kx"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) accessGrantGPGCmd() *cli.Command {
	return &cli.Command{
		Name:      "gpg",
		Usage:     "Grants access to collaborators with OpenPGP keys",
		ArgsUsage: "[KEY...]",
		Before:    rt.LoadSecretKey,
		Action:    rt.accessGrantGPGDo,
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

func (rt *Runtime) accessGrantGPGDo(_ context.Context, cmd *cli.Command) error {
	var keyEntries openpgp.EntityList

	rt.loadKeys(cmd.StringSlice("openpgp"), false, &keyEntries)
	rt.loadKeys(cmd.StringSlice("openpgp-armor"), true, &keyEntries)

	args := cmd.Args()
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
		return ErrGPGKeyNotFound
	}

	saved := 0

	for _, key := range keyEntries {
		if err := rt.saveGPGKey(key); err != nil {
			rt.Warnf("cannot save key: %v", err)

			continue
		}

		saved++
	}

	rt.Infof(
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
			rt.Warnf("loading public key: %v", err)

			continue
		}

		*keyEntries = append(*keyEntries, entries...)
	}
}

func (rt *Runtime) saveGPGKey(key *openpgp.Entity) error {
	gpgutil.PrintKey(key)

	err := kx.SaveGPGKeyToKX(rt.Repo, key, func(w io.Writer) {
		err := rt.SaveTo(w)
		if err != nil {
			rt.Warn(err)
		}
	})

	if err != nil {
		return err
	}

	if err := kx.SaveGPGPubkeyToKX(rt.Repo, key); err != nil {
		return err
	}

	return nil
}
