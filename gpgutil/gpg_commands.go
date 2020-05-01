package gpgutil

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// ExportKey exports GPG key in ASCII armor
func ExportKey(keyIDs []string) ([]byte, error) {
	args := []string{
		"--armor",
		"--export",
		"--",
	}
	args = append(args, keyIDs...)

	return exec.Command("gpg", args...).Output()
}

// GetSecretKeys retrieves secret keys from GnuPG
func GetSecretKeys(filter string) ([][20]byte, []string, error) {
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

	buf := bytes.NewBuffer(out)
	warnings := []string{}

	var keys [][20]byte

	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, fmt.Errorf("reading gpg secret key listing output: %w", err)
		}

		if strings.HasPrefix(line, "fpr:") {
			items := strings.Split(line, ":")
			if len(items) < 9 {
				warnings = append(warnings, fmt.Sprintf("invalid private key entry: %q", line))
			} else {
				fingerprint, err := hex.DecodeString(items[9])
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("invalid fingerprint: %s", items[9]))
				}
				var key [20]byte
				copy(key[:], fingerprint)

				keys = append(keys, key)
			}
		}
	}

	return keys, warnings, nil
}

// DecryptWithKey decrypts a ciphertext, and stores into target path, using
// the provided fingerprint.
func DecryptWithKey(ciphertext string, fingerprint [20]byte) (io.ReadCloser, error) {
	var stdout, stderr bytes.Buffer

	args := []string{
		"--quiet",
		"--pinentry-mode",
		"loopback",
		"-u",
		fmt.Sprintf("%x", fingerprint),
		"--decrypt",
		ciphertext,
	}
	cmd := exec.Command("gpg", args...)
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

	return ioutil.NopCloser(bufio.NewReader(&stdout)), nil
}
