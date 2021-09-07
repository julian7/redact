package sdk

import (
	"errors"
	"fmt"

	"github.com/julian7/redact/files"
)

const (
	// AttrName defines name used in .gitattribute file's attribute
	// like: `*.key filter=AttrName diff=AttrName`
	AttrName = "redact"
)

// RedactRepo returns the loaded master key if it is unlocked. Otherwise,
// it returns appropriate error:
//
// - not a git repository: if git repository cannot be detected
// - detecting top level directory: ...: if git rev-parse command returns with failure
// - repository is not using redact: when there're no key exchange dir in the repo
// - repository is locked: when there is an exchange dir
//
// Checks key file permissions optionally.
func RedactRepo(masterkey *files.MasterKey, strict bool) error {
	if err := masterkey.Load(strict); err != nil {
		if _, err2 := masterkey.ExchangeDir(); err2 != nil {
			return errors.New("repository is not using redact")
		}

		return fmt.Errorf("repository is locked: %w", err)
	}

	return nil
}
