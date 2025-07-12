package repo

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-git/go-billy/v5/util"

	"github.com/julian7/redact/logger"
)

const (
	// ExtKeyArmor is public key ASCII armor file extension in Key Exchange folder
	ExtKeyArmor = ".asc"
	// ExtSecret is encrypted secret key file extension in Key Exchange folder
	ExtSecret = ".key"
	// DefaultKeyExchangeDir is where key exchange files are stored
	GitAttributesFile       = ".gitattributes"
	kxGitAttributesContents = `# This file has been created by redact
# DO NOT EDIT!
* !filter !diff
*.gpg binary
`
)

// GetExchangeFilename returns full path to file name in Key Exchange
func (r *Repo) GetExchangeFilename(filename string, log *logger.Logger) (string, error) {
	if err := r.ensureExchangeDir(log); err != nil {
		return "", err
	}

	return filepath.Join(r.ExchangeDir(), filename), nil
}

// GetExchangeFilenameStubFor returns file name stub of the Key Exchange for an
// OpenPGP key identified by its full public key ID.
//
// Add extensions for files:
//
// - .asc: Public key ASCII armor file
// - .key: Secret key encryped with public key
func (r *Repo) GetExchangeFilenameStubFor(fingerprint []byte, log *logger.Logger) (string, error) {
	return r.GetExchangeFilename(fmt.Sprintf("%x", fingerprint), log)
}

func (r *Repo) CheckExchangeDir() error {
	kxdir := r.ExchangeDir()

	st, err := r.Workdir.Stat(kxdir)

	if err != nil {
		return fmt.Errorf("checking exchange dir: %w", err)
	}

	if !st.IsDir() {
		return ErrExchangeIsNotDir
	}

	return nil
}

func (r *Repo) ensureExchangeDir(log *logger.Logger) error {
	kxdir := r.ExchangeDir()

	st, err := r.Workdir.Stat(kxdir)
	if err != nil {
		err = r.Workdir.MkdirAll(kxdir, 0755)
		if err != nil {
			return fmt.Errorf("creating key exchange dir: %w", err)
		}

		st, err = r.Workdir.Stat(kxdir)
	}

	if err != nil {
		return fmt.Errorf("stat key exchange dir: %w", err)
	}

	if !st.IsDir() {
		return fmt.Errorf("%q: %w", kxdir, ErrExchangeIsNotDir)
	}

	if err := r.ensureExchangeGitAttributes(log); err != nil {
		return err
	}

	return nil
}

func (r *Repo) ensureExchangeGitAttributes(log *logger.Logger) error {
	kxdir := r.ExchangeDir()

	var data []byte

	gaFileName := filepath.Join(kxdir, GitAttributesFile)

	st, err := r.Workdir.Stat(gaFileName)
	if err == nil {
		if st.IsDir() {
			return fmt.Errorf("%s is not a normal file: %+v", gaFileName, st)
		}

		f, err := r.Workdir.Open(gaFileName)
		if err != nil {
			return fmt.Errorf("opening .gitattributes file inside exchange dir: %w", err)
		}

		defer f.Close()

		data, err = io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("reading .gitattributes file in key exchange dir: %w", err)
		}

		if bytes.Equal(data, []byte(kxGitAttributesContents)) {
			return nil
		}

		if log != nil {
			log.Warn("rewriting .gitattributes file in key exchange dir")
		}
	}

	if err := util.WriteFile(r.Workdir, gaFileName, []byte(kxGitAttributesContents), 0644); err != nil {
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
