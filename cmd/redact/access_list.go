package main

import (
	"os"
	"strings"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/sdk/git"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func (rt *Runtime) accessListCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List collaborators to secrets in git repo",
		RunE:  rt.accessListDo,
	}

	return cmd, nil
}

func (rt *Runtime) accessListDo(cmd *cobra.Command, args []string) error {
	_ = rt.LoadSecretKey(cmd, args)

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
