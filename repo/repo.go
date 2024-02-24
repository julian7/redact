package repo

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/urfave/cli/v2"
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

	r.Workdir = NewOSFS(osfs.New(repo.Toplevel))

	r.SecretKey, err = files.NewSecretKey(chroot.New(r.Workdir, repo.Common))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) LoadSecretKey(_ *cli.Context) error {
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
