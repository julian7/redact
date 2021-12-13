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
	// AttrName defines name used in .gitattribute file's attribute
	// like: `*.key filter=AttrName diff=AttrName`
	AttrName = "redact"
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
	filePath = filepath.Join(r.Toplevel, filePath)
	touchTime := time.Now()

	if err := r.Fs.Chtimes(filePath, touchTime, touchTime); err != nil {
		return fmt.Errorf("touch file: %w", err)
	}

	return nil
}

func (r *Repo) TouchUp(files []string, rekey bool, softErrHandler func(error)) error {
	if len(files) < 1 {
		return nil
	}

	touched := make([]string, 0, len(files))

	for _, entry := range files {
		if err := r.TouchFile(entry); err != nil && softErrHandler != nil {
			softErrHandler(err)

			continue
		}

		touched = append(touched, filepath.Join(r.Toplevel, entry))
	}

	if len(touched) > 0 {
		issues, err := gitutil.Checkout(touched, rekey)
		if len(issues) > 0 && softErrHandler != nil {
			for _, issue := range issues {
				softErrHandler(issue)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) ForceReencrypt(rekey bool, softErrHandler func(error)) error {
	files, err := gitutil.LsFiles(nil)
	if err != nil {
		return fmt.Errorf("list git files: %w", err)
	}

	err = files.CheckAttrs()
	if err != nil {
		return fmt.Errorf("check git files' attributes: %w", err)
	}

	affectedFiles := make([]string, 0, len(files.Items))

	for _, entry := range files.Items {
		if entry.Filter == AttrName && entry.Status != gitutil.StatusOther {
			affectedFiles = append(affectedFiles, entry.Name)
		}
	}

	if softErrHandler != nil {
		for _, err := range files.Errors {
			softErrHandler(err)
		}
	}

	return r.TouchUp(affectedFiles, rekey, softErrHandler)
}
