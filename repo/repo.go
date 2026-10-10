package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	commonDir              string
}

func (r *Repo) SetupRepo() error {
	repo, err := gitutil.DetectGitRepo()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	r.Workdir = NewOSFS(osfs.New(repo.Toplevel, osfs.WithBoundOS()))
	r.commonDir = repo.Common

	// common git dir can be outside of the work tree (linked worktrees,
	// --separate-git-dir), so it gets its own filesystem.
	commonfs := osfs.New(repo.Common, osfs.WithBoundOS())

	r.SecretKey, err = files.NewSecretKey(NewOSFS(commonfs))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) LoadSecretKey(ctx context.Context, _ *cli.Command) (context.Context, error) {
	if err := r.SetupRepo(); err != nil {
		return ctx, fmt.Errorf("detecting repo config: %w", err)
	}

	if err := r.Load(r.StrictPermissionChecks); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return ctx, fmt.Errorf("loading secret key: %w", err)
		}

		return ctx, fmt.Errorf("%w in %s", ErrRedactKeyNotFound, filepath.Join(r.commonDir, r.Keyfile()))
	}

	return ctx, nil
}
