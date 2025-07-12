package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5/util"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) accessListCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List collaborators to secrets in git repo",
		Action: rt.accessListDo,
	}
}

func (rt *Runtime) accessListDo(ctx context.Context, cmd *cli.Command) error {
	_, _ = rt.LoadSecretKey(ctx, cmd)
	kxdir := rt.ExchangeDir()

	extConfig, err := ext.Load(rt.Repo)
	if err != nil {
		return fmt.Errorf("loading extension config: %w", err)
	}

	if extConfig != nil {
		extConfig.List()
	}

	err = util.Walk(rt.Workdir, kxdir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil // nolint:nilerr
		}

		if !strings.HasSuffix(path, repo.ExtKeyArmor) {
			return nil
		}

		entities, err := gpgutil.LoadPubKeyFromFile(path, true)
		if err != nil {
			rt.Warnf("cannot load public key: %v", err)

			return nil
		}

		if len(entities) != 1 {
			rt.Warnf("multiple entities in key file %s", path)

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
