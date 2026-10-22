package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"filippo.io/age"

	"github.com/julian7/redact/ageutil"
)

// recipient is one age registration found in the Key Exchange directory: the
// parsed recipient, its canonical bech32 text, and the file it came from.
type recipient struct {
	key      age.Recipient
	text     string
	filename string
	stub     string
}

func readRecipients() ([]recipient, []error) {
	directory, err := exchangeDirectory()
	if err != nil {
		return nil, []error{err}
	}

	entries, err := os.ReadDir(directory)
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ExtAgePublic) {
			continue
		}

		filename := filepath.Join(directory, entry.Name())

		keys, err := ageutil.ParseRecipientsFile(filename)
		if err != nil {
			failures = append(failures, fmt.Errorf("reading recipient %s: %w", filename, err))

			continue
		}

		if len(keys) != 1 {
			failures = append(failures, fmt.Errorf("recipient file %s must contain one public key", filename))

			continue
		}

		key := keys[0]

		text, err := ageutil.RecipientText(key)
		if err != nil {
			failures = append(failures, fmt.Errorf("identifying recipient of %s: %w", filename, err))

			continue
		}

		if strings.TrimSuffix(entry.Name(), ExtAgePublic) != ageutil.RecipientID(text) {
			failures = append(failures, fmt.Errorf("recipient fingerprint does not match %s", filename))

			continue
		}

		recipients = append(recipients, recipient{
			key:      key,
			text:     text,
			filename: filename,
			stub:     strings.TrimSuffix(filename, ExtAgePublic),
		})
	}

	return recipients, failures
}

func (recipient recipient) matches(selector string) bool {
	return matchesRecipient(recipient.text, selector)
}

// age recipients have no user IDs; support short recipient fragments for selection.
func matchesRecipient(text, selector string) bool {
	selector = strings.ToLower(strings.TrimSpace(selector))
	if selector == "" {
		return false
	}

	if strings.HasPrefix(ageutil.RecipientID(text), selector) {
		return true
	}

	if isBech32Fragment(selector) && (strings.HasPrefix(text, selector) || strings.HasSuffix(text, selector)) {
		return true
	}

	return strings.Contains(text, selector)
}

func isBech32Fragment(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'z') {
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
