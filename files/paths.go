package files

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

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
	key := "kxdir_ensure"
	if _, ok := k.Cache[key]; ok {
		return nil
	}
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
	if err := k.ensureExchangeGitAttributes(kxdir); err != nil {
		return err
	}
	k.Cache[key] = kxdir
	return nil
}

// ExchangeDir returns key exchange directory if exists
func (k *MasterKey) ExchangeDir() (string, error) {
	key := "kxdir"
	if val, ok := k.Cache[key]; ok {
		return val, nil
	}
	kxdir := ExchangeDir(k.RepoInfo.Toplevel)
	st, err := k.Stat(kxdir)
	if err != nil {
		return "", errors.Wrap(err, "stat key exchange dir")
	}
	if !st.IsDir() {
		return "", errors.New("key exchange is not a directory")
	}
	k.Cache[key] = kxdir
	return kxdir, nil
}

func (k *MasterKey) ensureExchangeGitAttributes(kxdir string) error {
	key := "kxgitattrs"
	if _, ok := k.Cache[key]; ok {
		return nil
	}
	var data []byte
	gaFileName := filepath.Join(kxdir, gitAttributesFile)
	st, err := k.Stat(gaFileName)
	if err == nil {
		if st.IsDir() {
			return errors.Errorf("%s is not a normal file: %+v", gaFileName, st)
		}
		f, err := k.Open(gaFileName)
		if err != nil {
			return errors.Wrap(err, "opening .gitattributes file inside exchange dir")
		}
		defer f.Close()
		data, err = ioutil.ReadAll(f)
		if err != nil {
			return errors.Wrap(err, "reading .gitattributes file in key exchange dir")
		}
		if !bytes.Equal(data, []byte(kxGitAttributesContents)) {
			return nil
		}
		k.Logger.Warn("rewriting .gitattributes file in key exchange dir")
	}
	if err := ioutil.WriteFile(gaFileName, []byte(kxGitAttributesContents), 0644); err != nil {
		return errors.Wrap(err, "writing .gitattributes file in key exchange dir")
	}

	k.Cache[key] = kxdir
	return nil
}
