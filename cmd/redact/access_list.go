package main

import (
	"os"
	"strings"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk/git"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) accessListCmd() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List collaborators to secrets in git repo",
		ArgsUsage: " ",
		Action:    rt.accessListDo,
	}
}

func (rt *Runtime) accessListDo(ctx *cli.Context) error {
	_ = rt.LoadSecretKey(ctx)

	kxdir := rt.Repo.ExchangeDir()
	err := afero.Walk(rt.Repo.Fs, kxdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}
		if !strings.HasSuffix(path, git.ExtKeyArmor) {
			return nil
		}
		entities, err := gpgutil.LoadPubKeyFromFile(path, true)
		if err != nil {
			rt.Logger.Warnf("cannot load public key: %v", err)

			return nil
		}
		if len(entities) != 1 {
			rt.Logger.Warnf("multiple entities in key file %s", path)

			return nil
		}
		gpgutil.PrintKey(entities[0])

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
