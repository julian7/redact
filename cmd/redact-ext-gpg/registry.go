package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
)

type recipient struct {
	key         *openpgp.Entity
	fingerprint []byte
	filename    string
	stub        string
}

func readRecipients(redactRepo *repo.Repo) ([]recipient, []error) {
	entries, err := redactRepo.Workdir.ReadDir(repo.DefaultKeyExchangeDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, []error{fmt.Errorf("reading Key Exchange directory: %w", err)}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	recipients := make([]recipient, 0, len(entries))

	var failures []error

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), repo.ExtKeyArmor) {
			continue
		}

		filename := filepath.Join(repo.DefaultKeyExchangeDir, entry.Name())

		file, err := redactRepo.Workdir.Open(filename)
		if err != nil {
			failures = append(failures, fmt.Errorf("opening public key %s: %w", filename, err))

			continue
		}

		keys, readErr := gpgutil.LoadPubKey(file, true)
		closeErr := file.Close()

		if readErr != nil {
			failures = append(failures, fmt.Errorf("loading public key %s: %w", filename, readErr))

			continue
		}

		if closeErr != nil {
			failures = append(failures, fmt.Errorf("closing public key %s: %w", filename, closeErr))

			continue
		}

		if len(keys) != 1 {
			failures = append(failures, fmt.Errorf("public key file %s contains %d keys", filename, len(keys)))

			continue
		}

		stub := strings.TrimSuffix(filename, repo.ExtKeyArmor)
		recipients = append(recipients, recipient{
			key:         keys[0],
			fingerprint: keys[0].PrimaryKey.Fingerprint,
			filename:    filename,
			stub:        stub,
		})
	}

	return recipients, failures
}

func (recipient recipient) fingerprintString() string {
	return hex.EncodeToString(recipient.fingerprint)
}

func (recipient recipient) matches(selector string) bool {
	selector = strings.ToLower(strings.TrimSpace(selector))
	selector = strings.TrimPrefix(selector, "0x")

	selector = strings.ReplaceAll(selector, " ", "")
	if selector == "" {
		return false
	}

	fingerprint := recipient.fingerprintString()
	if isHex(selector) && (strings.HasPrefix(fingerprint, selector) || strings.HasSuffix(fingerprint, selector)) {
		return true
	}

	for name := range recipient.key.Identities {
		if strings.Contains(strings.ToLower(name), selector) {
			return true
		}
	}

	return false
}

func isHex(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}

	return true
}

func matchingRecipients(recipients []recipient, selectors []string) []recipient {
	matched := make([]recipient, 0, len(recipients))
	for _, item := range recipients {
		for _, selector := range selectors {
			if item.matches(selector) {
				matched = append(matched, item)

				break
			}
		}
	}

	return matched
}

func readPublicKeys(filename string, armor bool) ([]*openpgp.Entity, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keys, err := gpgutil.LoadPubKey(file, armor)
	if err != nil {
		return nil, err
	}

	return keys, nil
}
