package sdk

import (
	"github.com/julian7/redact/files"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
func RedactRepo(l *logrus.Logger) (*files.MasterKey, error) {
	masterkey, err := files.NewMasterKey(l)
	if err != nil {
		return nil, err
	}
	err = masterkey.Load()
	if err != nil {
		if _, err2 := masterkey.ExchangeDir(); err2 != nil {
			return nil, errors.New("repository is not using redact")
		}
		return nil, errors.Wrap(err, "repository is locked")
	}
	return masterkey, nil
}
