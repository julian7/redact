package keys

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/julian7/redact/gitutil"
	keyV0 "github.com/julian7/redact/keys/key_v0"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// KeyMagic magic string the key file starts with
	KeyMagic = "\000REDACT\000"
	// KeyCurrentType current key file version
	KeyCurrentType = 0
)

// MasterKey contains master key in a git repository
type MasterKey struct {
	afero.Fs
	KeyDir    string
	Keys      map[uint32]KeyHandler
	LatestKey uint32
}

// NewMasterKey generates a new repo key in the OS' filesystem
func NewMasterKey() (*MasterKey, error) {
	gitdir, err := gitutil.GitDir()
	if err != nil {
		return nil, errors.Wrap(err, "cannot build repo key")
	}
	return &MasterKey{
		Fs:     afero.NewOsFs(),
		KeyDir: buildKeyDir(gitdir),
	}, nil
}

// Generate generates a new master key
func (k *MasterKey) Generate() error {
	epoch := k.LatestKey
	k.ensureKeys()
	epoch++
	k.Keys[epoch] = keyV0.NewKey(epoch)
	k.LatestKey = epoch
	return k.Keys[epoch].Generate()
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
	var keyType uint32
	err = binary.Read(f, binary.BigEndian, &keyType)
	if err != nil {
		return errors.Wrap(err, "reading key type")
	}
	if keyType != KeyCurrentType {
		return errors.New("invalid key type")
	}
	epoch := uint32(0)
	k.ensureKeys()
	for {
		key := new(keyV0.KeyV0)
		err = binary.Read(f, binary.BigEndian, key)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.Wrap(err, "reading key data")
		}
		if epoch >= key.Version() {
			return errors.Errorf(
				"invalid epoch in keys: %d (previous: %d)",
				key.Version(),
				epoch,
			)
		}
		epoch = key.Version()
		k.Keys[epoch] = key
		k.LatestKey = epoch
	}
}

// Save saves key
func (k *MasterKey) Save() error {
	err := k.getOrCreateKeyDir()
	if err != nil {
		return err
	}
	keyfile := buildKeyFileName(k.KeyDir)
	_, err = k.Fs.Stat(keyfile)
	if err == nil {
		return errors.Errorf("key file (%s) already exists", keyfile)
	}
	f, err := k.Fs.OpenFile(keyfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "saving key file")
	}
	defer f.Close()
	_, err = f.WriteString(KeyMagic)
	if err != nil {
		return errors.Wrap(err, "writing key preamble")
	}
	typeData := make([]byte, 4)
	binary.BigEndian.PutUint32(typeData, KeyCurrentType)
	_, err = f.Write(typeData)
	if err != nil {
		return errors.Wrap(err, "writing key type header")
	}
	for _, key := range k.Keys {
		err = binary.Write(f, binary.BigEndian, key)
		if err != nil {
			return errors.Wrap(err, "writing key contents")
		}
	}
	return nil
}

// Key returns the a key handler with a certain epoch. If epoch is 0,
// it returns the latest key.
func (k *MasterKey) Key(epoch uint32) (KeyHandler, error) {
	if k.Keys == nil || k.LatestKey == 0 {
		return nil, errors.New("no keys loaded")
	}
	if epoch == 0 {
		epoch = k.LatestKey
	}
	key, ok := k.Keys[epoch]
	if !ok {
		return nil, errors.Errorf("key version %d not found", epoch)
	}
	return key, nil
}

func (k *MasterKey) ensureKeys() {
	if k.Keys == nil {
		k.Keys = make(map[uint32]KeyHandler)
	}
}

func (k *MasterKey) getOrCreateKeyDir() error {
	fs, err := k.Fs.Stat(k.KeyDir)
	if err != nil {
		k.Fs.Mkdir(k.KeyDir, 0700)
		fs, err = k.Fs.Stat(k.KeyDir)
	}
	if err != nil {
		return errors.Wrap(err, "keydir not available")
	}
	if !fs.IsDir() {
		return errors.New("keydir is not a directory")
	}
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
	var keymsg string
	key, err := k.Key(0)
	if err != nil {
		keymsg = err.Error()
	} else {
		keymsg = key.String()
	}
	return fmt.Sprintf(
		"%s (%s)",
		k.KeyDir,
		keymsg,
	)
}
