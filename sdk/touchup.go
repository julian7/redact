package sdk

import (
	"path/filepath"
	"time"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
)

// TouchUp touches and checks out encrypted files, generally fixing them.
func TouchUp(masterkey *files.MasterKey) error {
	toplevel, err := gitutil.TopLevel()
	if err != nil {
		return errors.Wrap(err, "top level dir")
	}
	files, err := gitutil.LsFiles(nil)
	if err != nil {
		return errors.Wrap(err, "list git files")
	}
	err = files.CheckAttrs()
	if err != nil {
		return errors.Wrap(err, "check git files' attributes")
	}
	touchTime := time.Now()
	l := log.Log()
	affectedFiles := make([]string, 0, len(files))
	for _, entry := range files {
		if entry.Filter == AttrName && entry.Status != gitutil.StatusOther {
			fullpath := filepath.Join(toplevel, entry.Name)
			err = masterkey.Chtimes(fullpath, touchTime, touchTime)
			if err != nil {
				l.Warnf("cannot touch %s: %v", entry.Name, err)
				continue
			}
			affectedFiles = append(affectedFiles, fullpath)
		}
	}
	if len(affectedFiles) > 0 {
		return gitutil.Checkout(affectedFiles)
	}
	return nil
}
