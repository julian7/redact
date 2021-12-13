package sdk

import (
	"fmt"

	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/sdk/git"
)

var configItems = map[string]string{
	"filter.%s.clean":  `"%s" git clean --file=%%f`,
	"filter.%s.smudge": `"%s" git smudge`,
	"diff.%s.textconv": `"%s" git diff`,
}

// SaveGitSettings sets filter / diff settings into git repository config
func SaveGitSettings(argv0 string, cb func(string)) error {
	for key, val := range configItems {
		attr := fmt.Sprintf(key, git.AttrName)
		val := fmt.Sprintf(val, argv0)

		if err := gitutil.GitConfig(attr, val); err != nil {
			return err
		}

		if cb != nil {
			cb(attr)
		}
	}

	return nil
}

// RemoveGitSettings removes filter / diff settings from git repository config
func RemoveGitSettings(cb func(string)) error {
	for key := range configItems {
		attr := fmt.Sprintf(key, git.AttrName)

		if err := gitutil.GitConfig("--unset", attr); err != nil {
			return err
		}

		if cb != nil {
			cb(attr)
		}
	}

	return nil
}
