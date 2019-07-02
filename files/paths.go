package files

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
)

const (
	// DefaultKeyDir contains standard key directory name inside .git/ directory
	DefaultKeyDir = "redact"
	// DefaultKeyFile contains standard key file name inside key directory
	DefaultKeyFile = "key"
	// DefaultKeyExchangeDir is where key exchange files are stored
	DefaultKeyExchangeDir   = ".redact"
	gitAttributesFile       = ".gitattributes"
	kxGitAttributesContents = `# This file has been created by redact
# DO NOT EDIT!
* !filter !diff
*.gpg binary
`
)

func buildKeyDir(gitdir string) string {
	return filepath.Join(gitdir, DefaultKeyDir)
}

func buildKeyFileName(path string) string {
	return filepath.Join(path, DefaultKeyFile)
}

func checkFileMode(name string, fileinfo os.FileInfo, expected os.FileMode) error {
	mode := fileinfo.Mode().Perm()
	if mode&^expected != 0 {
		return errors.Errorf("excessive rights on %s", name)
	}
	if mode&expected != expected {
		return errors.Errorf("insufficient rights on %s", name)
	}
	return nil
}

// ExchangeDir returns Key Exchange dir inside the git repo
func ExchangeDir(toplevel string) string {
	return filepath.Join(toplevel, DefaultKeyExchangeDir)
}

func (k *MasterKey) ensureExchangeDir(kxdir string) error {
	st, err := k.Stat(kxdir)
	if err != nil {
		err = k.Mkdir(kxdir, 0755)
		if err != nil {
			return errors.Wrap(err, "creating key exchange dir")
		}
		st, err = k.Stat(kxdir)
	}
	if err != nil {
		return errors.Wrap(err, "stat key exchange dir")
	}
	if !st.IsDir() {
		return errors.New("key exchange is not a directory")
	}
	return k.ensureExchangeGitAttributes(kxdir)
}

// ExchangeDir returns key exchange directory if exists
func (k *MasterKey) ExchangeDir(toplevel string) (string, error) {
	kxdir := ExchangeDir(toplevel)
	st, err := k.Stat(kxdir)
	if err != nil {
		return "", errors.Wrap(err, "stat key exchange dir")
	}
	if !st.IsDir() {
		return "", errors.New("key exchange is not a directory")
	}
	return kxdir, nil
}

func (k *MasterKey) ensureExchangeGitAttributes(kxdir string) error {
	var data []byte
	gaFileName := filepath.Join(kxdir, gitAttributesFile)
	st, err := k.Stat(gaFileName)
	if err == nil {
		if st.IsDir() {
			return errors.Errorf("%s is not a normal file: %+v", gaFileName, st)
		}
		data, err = ioutil.ReadFile(gaFileName)
		if err != nil {
			return errors.Wrap(err, "reading .gitattributes file in key exchange dir")
		}
		if bytes.Compare(data, []byte(kxGitAttributesContents)) == 0 {
			return nil
		}
		log.Log().Warn("rewriting .gitattributes file in key exchange dir")
	}
	return errors.Wrap(
		ioutil.WriteFile(gaFileName, []byte(kxGitAttributesContents), 0644),
		"writing .gitattributes file in key exchange dir",
	)
}
