package sdk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
)

var configItems = map[string]string{
	"filter.%s.clean":  `"%s" git clean`,
	"filter.%s.smudge": `"%s" git smudge`,
	"diff.%s.textconv": `"%s" git diff`,
}

// SaveGitSettings sets filter / diff settings into git repository config
func SaveGitSettings() error {
	argv0 := os.Args[0]
	if argv0[0] == '.' {
		var err error
		argv0, err = filepath.Abs(argv0)

		if err != nil {
			return errors.Wrap(err, "get absolute path of argv0")
		}
	}

	for key, val := range configItems {
		if err := gitutil.GitConfig(
			fmt.Sprintf(key, AttrName),
			fmt.Sprintf(val, argv0),
		); err != nil {
			return err
		}

		log.Log().Debugf("Setting up filter/diff git config of %s to %s", AttrName, argv0)
	}

	return nil
}

// RemoveGitSettings removes filter / diff settings from git repository config
func RemoveGitSettings() error {
	for key := range configItems {
		if err := gitutil.GitConfig(
			"--unset",
			fmt.Sprintf(key, AttrName),
		); err != nil {
			return err
		}

		log.Log().Debugf("Removing filter/diff git config of %s", AttrName)
	}

	return nil
}
