package git

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	// ExtKeyArmor is public key ASCII armor file extension in Key Exchange folder
	ExtKeyArmor = ".asc"
	// ExtSecret is encrypted secret key file extension in Key Exchange folder
	ExtSecret = ".key"
	// DefaultKeyExchangeDir is where key exchange files are stored
	gitAttributesFile       = ".gitattributes"
	kxGitAttributesContents = `# This file has been created by redact
# DO NOT EDIT!
* !filter !diff
*.gpg binary
`
)

// GetExchangeFilenameStubFor returns file name stub of the Key Exchange for an
// OpenPGP key identified by its full public key ID.
//
// Add extensions for files:
//
// - .asc: Public key ASCII armor file
// - .key: Secret key encryped with public key
func (r *Repo) GetExchangeFilenameStubFor(fingerprint []byte) (string, error) {
	if err := r.ensureExchangeDir(); err != nil {
		return "", err
	}

	return filepath.Join(r.ExchangeDir(), fmt.Sprintf("%x", fingerprint)), nil
}

func (r *Repo) CheckExchangeDir() error {
	kxdir := r.ExchangeDir()

	st, err := r.Stat(kxdir)

	if err != nil {
		return fmt.Errorf("checking exchange dir: %w", err)
	}

	if !st.IsDir() {
		return errors.New("exchange dir is not a directory")
	}

	return nil
}

func (r *Repo) ensureExchangeDir() error {
	kxdir := r.ExchangeDir()

	st, err := r.Stat(kxdir)
	if err != nil {
		err = r.Mkdir(kxdir, 0755)
		if err != nil {
			return fmt.Errorf("creating key exchange dir: %w", err)
		}

		st, err = r.Stat(kxdir)
	}

	if err != nil {
		return fmt.Errorf("stat key exchange dir: %w", err)
	}

	if !st.IsDir() {
		return errors.New("key exchange is not a directory")
	}

	if err := r.ensureExchangeGitAttributes(); err != nil {
		return err
	}

	return nil
}

func (r *Repo) ensureExchangeGitAttributes() error {
	kxdir := r.ExchangeDir()

	var data []byte

	gaFileName := filepath.Join(kxdir, gitAttributesFile)

	st, err := r.Stat(gaFileName)
	if err == nil {
		if st.IsDir() {
			return fmt.Errorf("%s is not a normal file: %+v", gaFileName, st)
		}

		f, err := r.Open(gaFileName)
		if err != nil {
			return fmt.Errorf("opening .gitattributes file inside exchange dir: %w", err)
		}

		defer f.Close()

		data, err = ioutil.ReadAll(f)
		if err != nil {
			return fmt.Errorf("reading .gitattributes file in key exchange dir: %w", err)
		}

		if bytes.Equal(data, []byte(kxGitAttributesContents)) {
			return nil
		}

		r.Logger.Warn("rewriting .gitattributes file in key exchange dir")
	}

	if err := afero.WriteFile(r.Fs, gaFileName, []byte(kxGitAttributesContents), 0644); err != nil {
		return fmt.Errorf("writing .gitattributes file in key exchange dir: %w", err)
	}

	return nil
}

// ExchangePubKeyFile returns full filename for Public key ASCII armor
func ExchangePubKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtKeyArmor)
}

// ExchangeSecretKeyFile returns full filename for Secret key exchange
func ExchangeSecretKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtSecret)
}
