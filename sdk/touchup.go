package sdk

import (
	"path/filepath"
	"time"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// TouchUp touches and checks out encrypted files, generally fixing them.
func TouchUp(masterkey *files.MasterKey) error {
	files, err := gitutil.LsFiles(nil)
	if err != nil {
		return errors.Wrap(err, "list git files")
	}
	err = files.CheckAttrs()
	if err != nil {
		return errors.Wrap(err, "check git files' attributes")
	}
	affectedFiles := make([]string, 0, len(files))
	for _, entry := range files {
		if entry.Filter == AttrName && entry.Status != gitutil.StatusOther {
			affectedFiles = append(affectedFiles, entry.Name)
		}
	}
	return TouchUpFiles(masterkey, affectedFiles)
}

// TouchUpFiles force-checkouts specific files in a repo
func TouchUpFiles(masterkey *files.MasterKey, files []string) error {
	if len(files) < 1 {
		return nil
	}
	l := log.Log()
	touched := make([]string, 0, len(files))
	for _, entry := range files {
		fullpath := filepath.Join(masterkey.RepoInfo.Toplevel, entry)
		if err := TouchFile(masterkey.Fs, fullpath); err != nil {
			l.Warnf("%s: %v", entry, err)
			continue
		}
		touched = append(touched, fullpath)
	}
	if len(touched) > 0 {
		return gitutil.Checkout(touched)
	}
	return nil
}

// TouchFile touches a single file
func TouchFile(filesystem afero.Fs, fullpath string) error {
	touchTime := time.Now()
	if err := filesystem.Chtimes(fullpath, touchTime, touchTime); err != nil {
		return errors.Wrap(err, "touch file")
	}
	return nil
}
