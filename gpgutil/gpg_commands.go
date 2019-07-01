package gpgutil

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
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
func GetSecretKeys(filter string) ([][20]byte, error) {
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
	l := log.Log()
	if err != nil {
		return nil, errors.Wrap(err, "fetching gpg secret keys")
	}
	buf := bytes.NewBuffer(out)
	var keys [][20]byte
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "reading gpg secret key listing output")
		}
		if strings.HasPrefix(line, "fpr:") {
			items := strings.Split(line, ":")
			if len(items) < 9 {
				l.Warnf("invalid private key entry: %q", line)
			} else {
				fingerprint, err := hex.DecodeString(items[9])
				if err != nil {
					l.Warnf("invalid fingerprint: %s", items[9])
				}
				var key [20]byte
				copy(key[:], fingerprint)

				keys = append(keys, key)
			}
		}
	}
	return keys, nil
}

// DecryptWithKey decrypts a ciphertext, and stores into target path, using
// the provided fingerprint.
func DecryptWithKey(ciphertext string, fingerprint [20]byte) (io.ReadCloser, error) {
	args := []string{
		"--quiet",
		"-u",
		fmt.Sprintf("%x", fingerprint),
		"--decrypt",
		ciphertext,
	}
	cmd := exec.Command("gpg", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errStream, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func(reader io.ReadCloser) {
		bufreader := bufio.NewReader(reader)
		l := log.Log()
		for {
			line, _, err := bufreader.ReadLine()
			if err != nil {
				return
			}
			l.Warnf("decryption: %s", line)
		}
	}(errStream)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return out, nil
}