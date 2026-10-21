package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julian7/redact/ageutil"
)

func writeStrayKeyExchangeFile(t *testing.T) string {
	t.Helper()

	stray := filepath.Join(KeyExchangeDir, "notes"+ExtAgePublic)
	if err := os.WriteFile(stray, []byte("not a recipient name\n"), 0o600); err != nil {
		t.Fatalf("writing stray Key Exchange file: %v", err)
	}

	return stray
}

func TestStrayKeyExchangeFileIsReported(t *testing.T) {
	initRepo(t)

	_, recipient := newIdentity(t)
	grantCollection(t, testCollection, recipient)

	stray := writeStrayKeyExchangeFile(t)

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "list"}, nil)
	if code := ExitCode(err); code != Failure {
		t.Fatalf("list with a stray Key Exchange file: exit code %d, want %d (%s)", code, Failure, stderr)
	}

	if !strings.Contains(stdout, "age recipient: "+ageutil.RecipientID(recipient)) {
		t.Errorf("list did not print the registered recipient %s: %s", recipient, stdout)
	}

	if strings.Contains(stdout, "notes") {
		t.Errorf("list took the stray file for a recipient: %s", stdout)
	}

	// redact shows the failure to the user, so the file to look at has to be named
	// by it.
	if err == nil || !strings.Contains(err.Error(), stray) {
		t.Errorf("list did not name the stray file %s: %v", stray, err)
	}
}

func TestStrayKeyExchangeFileDoesNotBlockRevoke(t *testing.T) {
	root := initRepo(t)

	_, recipient := newIdentity(t)
	grantCollection(t, testCollection, recipient)

	writeStrayKeyExchangeFile(t)

	_, stderr, err := runExtension(t, []string{"redact-ext-age", "revoke", "recipient=" + recipient}, nil)
	if ExitCode(err) != Failure {
		t.Errorf("revoke with a stray Key Exchange file reported no failure: %v (%s)", err, stderr)
	}

	if !strings.Contains(stderr, "Removed 1 age recipient registration") {
		t.Errorf("revoke did not remove the recipient: %s", stderr)
	}

	secret := filepath.Join(root, KeyExchangeDir, ExchangeAgeSecretKeyFile(recipient))
	if _, err := os.Stat(secret); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the encrypted collection of the revoked recipient survives: %v", err)
	}
}
