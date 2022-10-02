package files

import (
	"bytes"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	keyV0 "github.com/julian7/redact/files/key_v0"
	"github.com/julian7/redact/sdk/git"
	"github.com/spf13/afero"
)

const (
	// KeyMagic magic string the key file starts with
	KeyMagic = "\000REDACT\000"
	// KeyCurrentType current key file version
	KeyCurrentType = 0
	PEMType        = "REDACT SECRET KEY"
)

// SecretKey contains secret key in a git repository
type SecretKey struct {
	*git.Repo
	Keys      map[uint32]KeyHandler
	LatestKey uint32
	Cache     map[string]string
}

// NewSecretKey generates a new repo key in the OS' filesystem
func NewSecretKey(repo *git.Repo) (*SecretKey, error) {
	return &SecretKey{
		Repo:  repo,
		Cache: make(map[string]string),
	}, nil
}

// Generate generates a new secret key
func (k *SecretKey) Generate() error {
	epoch := k.LatestKey
	k.ensureKeys()
	epoch++
	k.Keys[epoch] = keyV0.NewKey(epoch)
	k.LatestKey = epoch

	return k.Keys[epoch].Generate()
}

// Read loads key from reader stream
func (k *SecretKey) Read(f io.Reader) error {
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

// Load loads existing key. Optionally it enforces strict file permissions.
func (k *SecretKey) Load(strict bool) error {
	err := k.checkKeyDir(strict)
	if err != nil {
		return err
	}

	keyfile := k.Repo.Keyfile()

	err = checkFileMode(k.Repo.Fs, "key file", keyfile, 0600, strict)
	if err != nil {
		return err
	}

	f, err := k.Repo.OpenFile(keyfile, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening key file for reading: %w", err)
	}

	defer f.Close()

	return k.Read(f)
}

func (k *SecretKey) Import(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("importing: %w", err)
	}

	blk, _ := pem.Decode(data)
	if blk == nil {
		return errors.New("no export found")
	}

	if blk.Type != PEMType {
		return errors.New("PEM type mismatch")
	}

	return k.Read(bytes.NewReader(blk.Bytes))
}

func (k *SecretKey) Export(writer io.Writer) error {
	blk := &pem.Block{
		Type: PEMType,
	}

	buf := bytes.Buffer{}
	if err := k.SaveTo(&buf); err != nil {
		return fmt.Errorf("export: %w", err)
	}

	blk.Bytes = buf.Bytes()

	if err := pem.Encode(writer, blk); err != nil {
		return fmt.Errorf("encoding: %w", err)
	}

	return nil
}

// SaveTo saves secret key into IO stream
func (k *SecretKey) SaveTo(writer io.Writer) error {
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
func (k *SecretKey) Save() error {
	err := k.getOrCreateKeyDir()
	if err != nil {
		return err
	}

	f, err := afero.TempFile(k.Repo.Fs, k.Repo.Keydir(), "temp")
	if err != nil {
		return fmt.Errorf("saving key file: %w", err)
	}

	if err = k.SaveTo(f); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("closing temp file for key: %w", err)
	}

	keyFileName := k.Repo.Keyfile()
	if err = k.Repo.Rename(f.Name(), keyFileName); err != nil {
		return fmt.Errorf("placing secret key: %w", err)
	}

	if err := k.Repo.Fs.Chmod(keyFileName, 0600); err != nil {
		return fmt.Errorf("setting permissions on secret key: %w", err)
	}

	return nil
}

// Key returns the a key handler with a certain epoch. If epoch is 0,
// it returns the latest key.
func (k *SecretKey) Key(epoch uint32) (KeyHandler, error) {
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

func (k *SecretKey) ensureKeys() {
	if k.Keys == nil {
		k.Keys = make(map[uint32]KeyHandler)
	}
}

func (k *SecretKey) getOrCreateKeyDir() error {
	fs, err := k.Repo.Stat(k.Repo.Keydir())
	if err != nil {
		if err := k.Repo.Mkdir(k.Repo.Keydir(), 0700); err != nil {
			return fmt.Errorf("creating keydir: %w", err)
		}

		fs, err = k.Repo.Stat(k.Repo.Keydir())
	}

	if err != nil {
		return fmt.Errorf("keydir not available: %w", err)
	}

	if !fs.IsDir() {
		return errors.New("keydir is not a directory")
	}

	return nil
}

func (k *SecretKey) checkKeyDir(strict bool) error {
	fs, err := k.Repo.Stat(k.Repo.Keydir())
	if err != nil {
		return fmt.Errorf("keydir not available: %w", err)
	}

	if !fs.IsDir() {
		return errors.New("keydir is not a directory")
	}

	err = checkFileMode(k.Repo.Fs, "key dir", k.Repo.Keydir(), 0700, strict)
	if err != nil {
		return err
	}

	return nil
}

func (k *SecretKey) String() string {
	var keymsg string

	key, err := k.Key(0)
	if err != nil {
		keymsg = err.Error()
	} else {
		keymsg = key.String()
	}

	return fmt.Sprintf(
		"%s (%s)",
		k.Repo.Keydir(),
		keymsg,
	)
}
