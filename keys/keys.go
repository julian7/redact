package keys

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/julian7/redact/gitutil"
	keyV0 "github.com/julian7/redact/keys/key_v0"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// KeyMagic magic string the key file starts with
	KeyMagic = "\000REDACT\000"
	// KeyCurrentVersion current key file version
	KeyCurrentVersion = 0
)

// MasterKey contains master key in a git repository
type MasterKey struct {
	afero.Fs
	Key    KeyHandler
	KeyDir string
}

// NewMasterKey generates a new repo key in the OS' filesystem
func NewMasterKey() (*MasterKey, error) {
	gitdir, err := gitutil.GitDir()
	if err != nil {
		return nil, errors.Wrap(err, "cannot build repo key")
	}
	return &MasterKey{Fs: afero.NewOsFs(), KeyDir: buildKeyDir(gitdir)}, nil
}

// Load loads existing key
func (k *MasterKey) Load() error {
	err := k.checkKeyDir()
	if err != nil {
		return err
	}
	keyfile := buildKeyFileName(k.KeyDir)
	fs, err := k.Fs.Stat(keyfile)
	if err != nil {
		return err
	}
	err = checkFileMode("key file", fs, 0600)
	if err != nil {
		return err
	}
	f, err := k.Fs.OpenFile(keyfile, os.O_RDONLY, 0600)
	if err != nil {
		return errors.Wrap(err, "opening key file for reading")
	}
	defer f.Close()
	readbuf := make([]byte, len(KeyMagic))
	_, err = f.Read(readbuf)
	if err != nil {
		return errors.Wrap(err, "reading preamble from key file")
	}
	if bytes.Compare(readbuf, []byte(KeyMagic)) != 0 {
		return errors.New("invalid key file preamble")
	}
	var version uint32
	err = binary.Read(f, binary.BigEndian, &version)
	if err != nil {
		return errors.Wrap(err, "reading key version")
	}
	if version != KeyCurrentVersion {
		return errors.New("invalid key version")
	}
	var key keyV0.KeyV0
	err = binary.Read(f, binary.BigEndian, &key)
	if err != nil {
		return errors.Wrap(err, "reading key data")
	}
	k.Key = &key
	return nil
}

func (k *MasterKey) checkKeyDir() error {
	fs, err := k.Fs.Stat(k.KeyDir)
	if err != nil {
		return errors.Wrap(err, "keydir not available")
	}
	if !fs.IsDir() {
		return errors.New("keydir is not a directory")
	}
	err = checkFileMode("key dir", fs, 0700)
	if err != nil {
		return err
	}
	return nil
}

func (k *MasterKey) String() string {
	return fmt.Sprintf(
		"%s (%s)",
		k.KeyDir,
		k.Key,
	)
}
