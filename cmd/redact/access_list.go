package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5/util"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) accessListCmd() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List collaborators to secrets in git repo",
		ArgsUsage: " ",
		Action:    rt.accessListDo,
	}
}

func (rt *Runtime) accessListDo(ctx context.Context, cmd *cli.Command) error {
	_ = rt.LoadSecretKey(ctx, cmd)

	kxdir := rt.Repo.ExchangeDir()
	err := util.Walk(rt.Repo.Workdir, kxdir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}

		if !strings.HasSuffix(path, repo.ExtKeyArmor) {
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
