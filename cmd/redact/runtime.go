package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/sdk"
	"github.com/julian7/redact/sdk/git"
	"github.com/urfave/cli/v2"
)

type Runtime struct {
	*logger.Logger
	*files.SecretKey
	*git.Repo
	Config                 string
	StrictPermissionChecks bool
}

func (rt *Runtime) FullPath() (string, error) {
	argv0 := os.Args[0]
	if argv0[0] == '.' {
		var err error
		argv0, err = filepath.Abs(argv0)

		if err != nil {
			return "", fmt.Errorf("get absolute path of argv0: %w", err)
		}
	}

	return argv0, nil
}

func (rt *Runtime) SetupRepo() error {
	var err error

	rt.Repo, err = git.NewRepo(rt.Logger)
	if err != nil {
		return err
	}

	rt.SecretKey, err = files.NewSecretKey(rt.Repo)
	if err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) LoadSecretKey(ctx *cli.Context) error {
	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("detecting repo config: %w", err)
	}

	if err := rt.SecretKey.Load(rt.StrictPermissionChecks); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("loading secret key: %w", err)
		}

		if err := rt.Repo.CheckExchangeDir(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("locating exchange dir: %w", err)
		}

		return errors.New("repository is not redacted")
	}

	return nil
}

func (rt *Runtime) SaveGitSettings() error {
	argv0, err := rt.FullPath()
	if err != nil {
		return err
	}

	err = sdk.SaveGitSettings(argv0, func(attr string) {
		rt.Logger.Debugf("Setting up filter/diff git config of %s to %s", attr, argv0)
	})
	if err != nil {
		return fmt.Errorf("setting git config: %w", err)
	}

	return nil
}
