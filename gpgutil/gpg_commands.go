package gpgutil

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// ExportKey exports GPG key in ASCII armor
func ExportKey(keyIDs []string) ([]byte, error) {
	args := make([]string, 0, 3+len(keyIDs))
	args = append(args,
		"--armor",
		"--export",
		"--",
	)
	args = append(args, keyIDs...)

	return exec.Command("gpg", args...).Output()
}

// GetSecretKeys retrieves secret keys from GnuPG
func GetSecretKeys(filter string) ([][]byte, []string, error) {
	args := []string{
		"--batch",
		"--with-colons",
		"--list-secret-keys",
		"--fingerprint",
	}
	if len(filter) > 0 {
		args = append(args, "--", filter)
	}

	out, err := exec.Command("gpg", args...).Output()
	if err != nil {
		return nil, nil, fmt.Errorf("fetching gpg secret keys: %w", err)
	}

	keys, warnings := parseSecretKeyListing(out)

	return keys, warnings, nil
}

func parseSecretKeyListing(out []byte) ([][]byte, []string) {
	var warnings []string

	var keys [][]byte

	primary := false

	for _, line := range strings.Split(string(out), "\n") {
		items := strings.Split(line, ":")
		switch items[0] {
		case "sec":
			primary = true
		case "ssb", "sub", "pub":
			primary = false
		case "fpr":
			// Exchange files are named for the primary key, even when an
			// encryption subkey supplies the private material for decryption.
			if !primary {
				continue
			}

			if len(items) < 10 {
				warnings = append(warnings, fmt.Sprintf("invalid private key entry: %q", line))

				continue
			}

			fingerprint, err := hex.DecodeString(items[9])
			if err != nil || len(fingerprint) == 0 {
				warnings = append(warnings, fmt.Sprintf("invalid fingerprint: %s", items[9]))

				continue
			}

			keys = append(keys, fingerprint)
		}
	}

	return keys, warnings
}

// DecryptWithKey decrypts ciphertext from a stream using the provided fingerprint.
func DecryptWithKey(ciphertext io.Reader, fingerprint []byte) (io.ReadCloser, error) {
	var stdout, stderr bytes.Buffer

	args := []string{
		"--quiet",
		"--pinentry-mode",
		"loopback",
		"-u",
		fmt.Sprintf("%x", fingerprint),
		"--decrypt",
	}
	cmd := exec.Command("gpg", args...)
	cmd.Stdin = ciphertext
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	messages := []string{}

	err := cmd.Run()
	bufreader := bufio.NewReader(&stderr)

	for {
		line, _, err := bufreader.ReadLine()
		if err != nil {
			break
		}

		messages = append(messages, string(line))
	}

	if err != nil {
		if len(messages) > 0 {
			return nil, fmt.Errorf("%s", strings.Join(messages, ""))
		}

		return nil, fmt.Errorf("%w: (no error message)", err)
	}

	return io.NopCloser(bufio.NewReader(&stdout)), nil
}
