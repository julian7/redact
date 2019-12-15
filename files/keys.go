package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	keyV0 "github.com/julian7/redact/files/key_v0"
	"github.com/julian7/redact/gitutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// KeyMagic magic string the key file starts with
	KeyMagic = "\000REDACT\000"
	// KeyCurrentType current key file version
	KeyCurrentType = 0
)

var GitDirFunc = gitutil.GitDir

// MasterKey contains master key in a git repository
type MasterKey struct {
	afero.Fs
	*logrus.Logger
	RepoInfo  gitutil.GitRepoInfo
	KeyDir    string
	Keys      map[uint32]KeyHandler
	LatestKey uint32
	Cache     map[string]string
}

// NewMasterKey generates a new repo key in the OS' filesystem
func NewMasterKey(l *logrus.Logger) (*MasterKey, error) {
	var masterkey MasterKey

	err := GitDirFunc(&masterkey.RepoInfo)
	if err != nil {
		return nil, errors.New("not a git repository")
	}

	return &MasterKey{
		Fs:     afero.NewOsFs(),
		Logger: l,
		KeyDir: buildKeyDir(masterkey.RepoInfo.Common),
		Cache:  make(map[string]string),
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

// KeyFile returns master key's file name
func (k *MasterKey) KeyFile() string {
	return buildKeyFileName(k.KeyDir)
}

// Read loads key from reader stream
func (k *MasterKey) Read(f io.Reader) error {
	readbuf := make([]byte, len(KeyMagic))

	_, err := f.Read(readbuf)
	if err != nil {
		return fmt.Errorf("reading preamble from key file: %w", err)
	}

	if !bytes.Equal(readbuf, []byte(KeyMagic)) {
		return errors.New("invalid key file preamble")
	}

	var keyType uint32

	err = binary.Read(f, binary.BigEndian, &keyType)
	if err != nil {
		return fmt.Errorf("reading key type: %w", err)
	}

	if keyType != KeyCurrentType {
		return errors.New("invalid key type")
	}

	k.ensureKeys()

	for {
		key := new(keyV0.KeyV0)

		err = binary.Read(f, binary.BigEndian, key)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return fmt.Errorf("reading key data: %w", err)
		}

		epoch := key.Version()
		if _, ok := k.Keys[epoch]; ok {
			return fmt.Errorf(
				"invalid key: duplicate epoch number (%d)",
				epoch,
			)
		}

		k.Keys[epoch] = key
		if epoch > k.LatestKey {
			k.LatestKey = epoch
		}
	}
}

// Load loads existing key
func (k *MasterKey) Load() error {
	err := k.checkKeyDir()
	if err != nil {
		return err
	}

	keyfile := buildKeyFileName(k.KeyDir)

	fs, err := k.Stat(keyfile)
	if err != nil {
		return err
	}

	if err = checkFileMode("key file", fs, 0600); err != nil {
		return err
	}

	f, err := k.OpenFile(keyfile, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening key file for reading: %w", err)
	}

	defer f.Close()

	return k.Read(f)
}

// SaveTo saves master key into IO stream
func (k *MasterKey) SaveTo(writer io.Writer) error {
	if _, err := writer.Write([]byte(KeyMagic)); err != nil {
		return fmt.Errorf("writing key preamble: %w", err)
	}

	typeData := make([]byte, 4)
	binary.BigEndian.PutUint32(typeData, KeyCurrentType)

	if _, err := writer.Write(typeData); err != nil {
		return fmt.Errorf("writing key type header: %w", err)
	}

	err := EachKey(k.Keys, func(idx uint32, key KeyHandler) error {
		if err := binary.Write(writer, binary.BigEndian, key); err != nil {
			return fmt.Errorf("key #%d: %w", idx, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("writing key contents: %w", err)
	}

	return nil
}

// Save saves key
func (k *MasterKey) Save() error {
	err := k.getOrCreateKeyDir()
	if err != nil {
		return err
	}

	f, err := afero.TempFile(k.Fs, k.KeyDir, "temp")
	if err != nil {
		return fmt.Errorf("saving key file: %w", err)
	}

	if err = k.SaveTo(f); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("closing temp file for key: %w", err)
	}

	if err = k.Rename(f.Name(), buildKeyFileName(k.KeyDir)); err != nil {
		return fmt.Errorf("placing key file: %w", err)
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
		return nil, fmt.Errorf("key version %d not found", epoch)
	}

	return key, nil
}

func (k *MasterKey) ensureKeys() {
	if k.Keys == nil {
		k.Keys = make(map[uint32]KeyHandler)
	}
}

func (k *MasterKey) getOrCreateKeyDir() error {
	fs, err := k.Stat(k.KeyDir)
	if err != nil {
		k.Mkdir(k.KeyDir, 0700) // nolint:errcheck
		fs, err = k.Stat(k.KeyDir)
	}

	if err != nil {
		return fmt.Errorf("keydir not available: %w", err)
	}

	if !fs.IsDir() {
		return errors.New("keydir is not a directory")
	}

	return nil
}

func (k *MasterKey) checkKeyDir() error {
	fs, err := k.Stat(k.KeyDir)
	if err != nil {
		return fmt.Errorf("keydir not available: %w", err)
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
