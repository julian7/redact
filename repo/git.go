package repo

import (
	"fmt"
	"time"

	"github.com/julian7/redact/gitutil"

	"github.com/go-git/go-billy/v5"
)

type configItem struct {
	sect string
	key  string
	val  string
}

var configItems = []configItem{
	{"filter", "clean", "%q git clean --file=%%f"},
	{"filter", "smudge", "%q git smudge"},
	{"diff", "textconv", "%q git diff"},
}

const (
	// AttrName defines name used in .gitattribute file's attribute
	// like: `*.key filter=AttrName diff=AttrName`
	AttrName = "redact"
	// DefaultKeyExchangeDir is where key exchange files are stored
	DefaultKeyExchangeDir = ".redact"
)

func (r *Repo) ExchangeDir() string {
	return DefaultKeyExchangeDir
}

func (r *Repo) SaveGitSettings(argv0 string, cb func(string)) error {
	for _, opt := range configItems {
		attr := fmt.Sprintf("%s.%s.%s", opt.sect, AttrName, opt.key)
		val := fmt.Sprintf(opt.val, argv0)

		if err := gitutil.GitConfig(attr, val); err != nil {
			return fmt.Errorf("saving git settings: %w", err)
		}

		if cb != nil {
			cb(attr)
		}
	}

	return nil
}

// RemoveGitSettings removes filter / diff settings from git repository config
func (r *Repo) RemoveGitSettings(cb func(string)) error {
	for _, opt := range configItems {
		attr := fmt.Sprintf("%s.%s.%s", opt.sect, AttrName, opt.key)

		if err := gitutil.GitConfig("--unset", attr); err != nil {
			return fmt.Errorf("unsetting git settings: %w", err)
		}

		if cb != nil {
			cb(attr)
		}
	}

	return nil
}

func (r *Repo) TouchFile(filePath string) error {
	touchTime := time.Now()

	if chfs, ok := r.Workdir.(billy.Change); ok {
		if err := chfs.Chtimes(filePath, touchTime, touchTime); err != nil {
			return fmt.Errorf("touch file: %w", err)
		}
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

		touched = append(touched, entry)
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
