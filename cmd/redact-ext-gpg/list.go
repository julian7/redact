package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/julian7/redact/gpgutil"
	"github.com/urfave/cli/v3"
)

const (
	ExtKeyArmor = ".asc"
)

func cmdList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "Shows AWS param store configuration",
		ArgsUsage: "keyid=kms_key_id_or_alias param=/path/to/param",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err := fs.WalkDir(os.DirFS("."), ".redact", func(path string, _ fs.DirEntry, err error) error {
				if err != nil {
					return nil // nolint:nilerr
				}

				if !strings.HasSuffix(path, ExtKeyArmor) {
					return nil
				}

				entities, err := gpgutil.LoadPubKeyFromFile(path, true)
				if err != nil {
					return fmt.Errorf("loading public key: %w", err)
				}

				if len(entities) != 1 {
					return fmt.Errorf("loading public key from file %s: %w", path, errors.New("multiple entities in key file"))
				}

				gpgutil.PrintKey(entities[0])

				return nil
			})

			if err != nil {
				return err
			}
			return nil
		},
	}
}
