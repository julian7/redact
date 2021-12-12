package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/logger"
	sdkfs "github.com/julian7/redact/sdk/fs"
	"github.com/spf13/afero"
)

const (
	// DefaultKeyDir contains standard key directory name inside .git/ directory
	DefaultKeyDir = "redact"
	// DefaultKeyFile contains standard key file name inside key directory
	DefaultKeyFile = "key"
	// DefaultKeyExchangeDir is where key exchange files are stored
	DefaultKeyExchangeDir = ".redact"
)

type Repo struct {
	afero.Fs
	*logger.Logger
	Common   string
	Toplevel string
}

func NewRepo(l *logger.Logger) (*Repo, error) {
	repo, err := gitutil.DetectGitRepo()
	if err != nil {
		return nil, errors.New("not a git repository")
	}

	return &Repo{
		Fs:       sdkfs.NewOsFs(),
		Logger:   l,
		Common:   repo.Common,
		Toplevel: repo.Toplevel,
	}, nil
}

func (r *Repo) Keydir() string {
	return filepath.Join(r.Common, DefaultKeyDir)
}

func (r *Repo) Keyfile() string {
	return filepath.Join(r.Common, DefaultKeyDir, DefaultKeyFile)
}

func (r *Repo) ExchangeDir() string {
	return filepath.Join(r.Toplevel, DefaultKeyExchangeDir)
}

func (r *Repo) TouchFile(filePath string) error {
	touchTime := time.Now()
	if err := r.Fs.Chtimes(filePath, touchTime, touchTime); err != nil {
		return fmt.Errorf("touch file: %w", err)
	}

	return nil
}
