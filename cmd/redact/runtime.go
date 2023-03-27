package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
)

type Runtime struct {
	*logger.Logger
	*repo.Repo
}

func (rt *Runtime) SaveGitSettings() error {
	argv0, err := fullPath()
	if err != nil {
		return err
	}

	if err := rt.Repo.SaveGitSettings(argv0, func(attr string) {
		rt.Logger.Debugf("Setting up filter/diff git config of %s to %s", attr, argv0)
	}); err != nil {
		return fmt.Errorf("setting git config: %w", err)
	}

	return nil
}

func fullPath() (string, error) {
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
