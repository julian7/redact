package repo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/urfave/cli/v3"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
)

type Repo struct {
	*files.SecretKey
	Workdir                billy.Filesystem
	StrictPermissionChecks bool
}

func (r *Repo) SetupRepo() error {
	repo, err := gitutil.DetectGitRepo()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	fs := osfs.New(repo.Toplevel, osfs.WithBoundOS())
	r.Workdir = NewOSFS(fs)

	commonfs, err := fs.Chroot(repo.Common)
	if err != nil {
		return err
	}

	r.SecretKey, err = files.NewSecretKey(NewOSFS(commonfs))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) LoadSecretKey(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if err := r.SetupRepo(); err != nil {
		return ctx, fmt.Errorf("detecting repo config: %w", err)
	}

	if err := r.SecretKey.Load(r.StrictPermissionChecks); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return ctx, fmt.Errorf("loading secret key: %w", err)
		}

		return ctx, errors.New("repository is not redacted")
	}

	return ctx, nil
}
