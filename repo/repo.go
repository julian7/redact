package repo

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/julian7/redact/files"
	"github.com/urfave/cli/v2"
)

type Repo struct {
	*files.SecretKey
	*git.Repository
	Workdir                billy.Filesystem
	StrictPermissionChecks bool
}

func (r *Repo) SetupRepo() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("setup repo: %w", err)
	}

	repo, err := git.PlainOpenWithOptions(
		pwd,
		&git.PlainOpenOptions{DetectDotGit: true},
	)
	if err != nil {
		return fmt.Errorf("setup repo: open git repo: %w", err)
	}

	r.Repository = repo

	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("setup repo: detect workdir: %w", err)
	}

	r.Workdir = NewOSFS(workTree.Filesystem)

	c, ok := repo.Storer.(*filesystem.Storage)
	if !ok {
		return errors.New("setup repo: storage is not a filesystem")
	}

	r.SecretKey, err = files.NewSecretKey(c.Filesystem())
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) LoadSecretKey(ctx *cli.Context) error {
	if err := r.SetupRepo(); err != nil {
		return fmt.Errorf("detecting repo config: %w", err)
	}

	if err := r.SecretKey.Load(r.StrictPermissionChecks); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("loading secret key: %w", err)
		}

		return errors.New("repository is not redacted")
	}

	return nil
}
