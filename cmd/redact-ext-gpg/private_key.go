package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/repo"
)

// loadStdinExchangeKey uses supplied private material entirely in memory,
// independently of the local GnuPG keyring and agent.
func loadStdinExchangeKey(r *repo.Repo, input io.Reader, selector string) ([]byte, error) {
	identities, err := openpgp.ReadArmoredKeyRing(input)
	if err != nil {
		return nil, failed(fmt.Errorf("reading armored private key from stdin: %w", err))
	}

	var available openpgp.EntityList

	for _, identity := range identities {
		item := recipient{key: identity, fingerprint: identity.PrimaryKey.Fingerprint}
		if selector != "" && !item.matches(selector) {
			continue
		}

		if !hasPrivateKey(identity) {
			return nil, failed(fmt.Errorf("stdin identity contains no private key"))
		}

		filename := filepath.Join(repo.DefaultKeyExchangeDir, item.fingerprintString()+repo.ExtSecret)

		info, err := r.Workdir.Stat(filename)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, failed(fmt.Errorf("checking encrypted Key Exchange file: %w", err))
		}

		if info.Mode().IsRegular() {
			available = append(available, identity)
		}
	}

	if len(available) == 0 {
		return nil, notApplicable(fmt.Errorf("no supplied private identity has a matching Key Exchange copy"))
	}

	if len(available) > 1 {
		return nil, missing(fmt.Errorf("multiple supplied private identities match; retry with gpgkey=<identity>"))
	}

	return decryptSuppliedKey(r, available[0])
}

func hasPrivateKey(identity *openpgp.Entity) bool {
	if identity.PrivateKey != nil && !identity.PrivateKey.Dummy() {
		return true
	}

	for _, subkey := range identity.Subkeys {
		if subkey.PrivateKey != nil && !subkey.PrivateKey.Dummy() {
			return true
		}
	}

	return false
}

func decryptSuppliedKey(r *repo.Repo, identity *openpgp.Entity) ([]byte, error) {
	stub := fmt.Sprintf("%x%s", identity.PrimaryKey.Fingerprint, repo.ExtSecret)
	filename := filepath.Join(repo.DefaultKeyExchangeDir, stub)

	ciphertext, err := r.Workdir.Open(filename)
	if err != nil {
		return nil, failed(fmt.Errorf("opening encrypted Key Exchange file: %w", err))
	}
	defer ciphertext.Close()

	message, err := openpgp.ReadMessage(ciphertext, openpgp.EntityList{identity}, privateKeyPrompt(), nil)
	if err != nil {
		if ext.ExitCode(err) == ext.MissingData {
			return nil, err
		}

		return nil, failed(fmt.Errorf("decrypting Key Exchange collection: %w", err))
	}

	if !message.IsEncrypted {
		return nil, failed(fmt.Errorf("key exchange copy is not encrypted"))
	}

	// Read through EOF before exposing plaintext: OpenPGP integrity checks
	// happen at the end of the stream.
	plaintext, err := io.ReadAll(message.UnverifiedBody)
	if err != nil {
		return nil, failed(fmt.Errorf("validating decrypted Key Exchange collection: %w", err))
	}

	if message.SignatureError != nil {
		return nil, failed(fmt.Errorf("verifying Key Exchange signature: %w", message.SignatureError))
	}

	return plaintext, nil
}
