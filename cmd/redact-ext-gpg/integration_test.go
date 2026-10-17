package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gpgutil"
)

//nolint:gocognit,gocyclo // Keeps the isolated end-to-end collection lifecycle together.
func TestGPGCollectionLifecycle(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg is not installed")
	}

	root := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", output, err)
	}

	t.Chdir(root)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	home := t.TempDir()
	t.Setenv("GNUPGHOME", home)

	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("FF_GIT_EXT", "1")

	for _, name := range []string{"Alice One", "Alice Two"} {
		args := []string{
			"--batch", "--pinentry-mode", "loopback", "--passphrase", "",
			"--quick-generate-key", name + " <shared@example.test>", "rsa2048", "encr", "0",
		}
		gpgRun(t, args...)
	}

	fingerprints, warnings, err := gpgutil.GetSecretKeys("")
	if err != nil || len(fingerprints) != 2 {
		t.Fatalf("discover isolated GnuPG identities: fingerprints=%d warnings=%v err=%v", len(fingerprints), warnings, err)
	}

	key, err := files.NewSecretKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := key.Generate(); err != nil {
		t.Fatal(err)
	}

	if err := key.Generate(); err != nil {
		t.Fatal(err)
	}

	var collection bytes.Buffer
	if err := key.Export(&collection); err != nil {
		t.Fatal(err)
	}

	// An unset feature flag leaves the GPG extension skipped.
	os.Unsetenv("FF_GIT_EXT")

	_, _, err = runExtension(t, []string{"redact-ext-gpg", "list"}, nil)
	if ext.ExitCode(err) != ext.NotApplicable {
		t.Fatalf("unset transition flag exit = %v", err)
	}

	t.Setenv("FF_GIT_EXT", "1")

	dataDir := t.TempDir()
	binaryKey := gpgRunOutput(t, "--export", "--", fmt.Sprintf("%x", fingerprints[0]))

	armorKey, err := gpgutil.ExportKey([]string{fmt.Sprintf("%x", fingerprints[1])})
	if err != nil {
		t.Fatal(err)
	}

	binaryFile := filepath.Join(dataDir, "recipient.bin")
	armorFile := filepath.Join(dataDir, "recipient.asc")

	if err := os.WriteFile(binaryFile, binaryKey, 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(armorFile, armorKey, 0600); err != nil {
		t.Fatal(err)
	}

	_, grantStderr, err := runExtension(t, []string{
		"redact-ext-gpg", "grant",
		"recipient=" + fmt.Sprintf("%x", fingerprints[0]),
		"file=" + binaryFile,
		"armor=" + armorFile,
	}, collection.Bytes())
	if ext.ExitCode(err) != ext.Success || !strings.Contains(grantStderr, "Granted access to 2 recipients") {
		t.Fatalf("grant: stderr=%q err=%v", grantStderr, err)
	}

	if _, err := os.Stat(filepath.Join(root, ".git", "redact", "key")); !os.IsNotExist(err) {
		t.Fatalf("listing/grant required a local Redact key: %v", err)
	}

	listStdout, _, err := runExtension(t, []string{"redact-ext-gpg", "list"}, nil)
	if err != nil || strings.Count(listStdout, "fingerprint:") != 2 {
		t.Fatalf("list recipients: output=%q err=%v", listStdout, err)
	}

	selectionOut, selectionErr, err := runExtension(t, []string{"redact-ext-gpg", "get"}, nil)
	if ext.ExitCode(err) != ext.MissingData || selectionOut != "" ||
		!strings.Contains(selectionErr, "gpgkey=<identity>") || !strings.Contains(selectionErr, "Alice One") {
		t.Fatalf("multiple-key selection: stdout=%q stderr=%q err=%v", selectionOut, selectionErr, err)
	}

	selector := fmt.Sprintf("%x", fingerprints[0])

	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}

	t.Chdir(nested)

	got, _, err := runExtension(t, []string{"redact-ext-gpg", "get", "gpgkey=" + selector}, nil)

	equal := bytes.Equal([]byte(got), collection.Bytes())
	if err != nil || !equal {
		t.Fatalf("explicit get did not return only the exported collection: equal=%v err=%v", equal, err)
	}

	t.Chdir(root)

	secretFile := filepath.Join(root, ".redact", selector+".key")
	if err := os.WriteFile(secretFile, []byte("damaged ciphertext"), 0600); err != nil {
		t.Fatal(err)
	}

	badOut, _, err := runExtension(t, []string{"redact-ext-gpg", "get", "gpgkey=" + selector}, nil)
	if ext.ExitCode(err) != ext.Failure || badOut != "" {
		t.Fatalf("decryption failure: stdout=%q err=%v", badOut, err)
	}

	_, _, err = runExtension(t, []string{"redact-ext-gpg", "put"}, collection.Bytes())
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	_, _, err = runExtension(t, []string{"redact-ext-gpg", "update"}, collection.Bytes())
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	for _, fingerprint := range fingerprints {
		selector := fmt.Sprintf("%x", fingerprint)

		got, _, err := runExtension(t, []string{"redact-ext-gpg", "get", "gpgkey=" + selector}, nil)
		if err != nil || !bytes.Equal([]byte(got), collection.Bytes()) {
			t.Fatalf("updated recipient %s: equal=%v err=%v", selector, bytes.Equal([]byte(got), collection.Bytes()), err)
		}
	}

	_, revokeStderr, err := runExtension(t, []string{"redact-ext-gpg", "revoke", "recipient=shared@example.test"}, nil)
	if err != nil || !strings.Contains(revokeStderr, "Removed 2 GPG recipient registrations") {
		t.Fatalf("revoke matching shared identity: stderr=%q err=%v", revokeStderr, err)
	}

	_, _, err = runExtension(t, []string{"redact-ext-gpg", "update"}, collection.Bytes())
	if err != nil {
		t.Fatalf("update after revoke: %v", err)
	}

	for _, fingerprint := range fingerprints {
		stub := fmt.Sprintf("%x", fingerprint)
		for _, suffix := range []string{".asc", ".key"} {
			if _, err := os.Stat(filepath.Join(root, ".redact", stub+suffix)); !os.IsNotExist(err) {
				t.Fatalf("revoke/update recreated %s: %v", stub+suffix, err)
			}
		}
	}

	zeroOut, _, err := runExtension(t, []string{"redact-ext-gpg", "get"}, nil)
	if ext.ExitCode(err) != ext.NotApplicable || zeroOut != "" {
		t.Fatalf("zero registered local matches: stdout=%q err=%v", zeroOut, err)
	}
}

func runExtension(t *testing.T, args []string, input []byte) (string, string, error) {
	t.Helper()

	dir := t.TempDir()

	stdin, err := os.Create(filepath.Join(dir, "stdin"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := stdin.Write(input); err != nil {
		t.Fatal(err)
	}

	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	stdout, err := os.Create(filepath.Join(dir, "stdout"))
	if err != nil {
		t.Fatal(err)
	}

	stderr, err := os.Create(filepath.Join(dir, "stderr"))
	if err != nil {
		t.Fatal(err)
	}

	oldStdin, oldStdout, oldStderr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = stdin, stdout, stderr
	commandErr := app().Run(context.Background(), args)
	os.Stdin, os.Stdout, os.Stderr = oldStdin, oldStdout, oldStderr
	_ = stdin.Close()
	_ = stdout.Close()
	_ = stderr.Close()

	stdoutData, err := os.ReadFile(stdout.Name())
	if err != nil {
		t.Fatal(err)
	}

	stderrData, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}

	return string(stdoutData), string(stderrData), commandErr
}

func gpgRun(t *testing.T, args ...string) {
	t.Helper()

	if output, err := exec.Command("gpg", args...).CombinedOutput(); err != nil {
		t.Fatalf("gpg %v: %s: %v", args, output, err)
	}
}

func gpgRunOutput(t *testing.T, args ...string) []byte {
	t.Helper()

	output, err := exec.Command("gpg", args...).Output()
	if err != nil {
		t.Fatalf("gpg %v: %v", args, err)
	}

	return output
}
