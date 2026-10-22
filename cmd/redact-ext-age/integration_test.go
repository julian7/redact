package main

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/julian7/redact/ageutil"
)

//nolint:gocognit,gocyclo // Keeps the isolated end-to-end collection lifecycle together.
func TestAgeKeyExchangeLifecycle(t *testing.T) {
	root := initRepo(t)

	identity1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating first identity: %v", err)
	}

	identity2, err := age.GenerateHybridIdentity()
	if err != nil {
		t.Fatalf("generating second identity: %v", err)
	}

	identityFile := filepath.Join(t.TempDir(), "identities")
	if err := os.WriteFile(identityFile, []byte(identity1.String()+"\n"+identity2.String()+"\n"), 0o600); err != nil {
		t.Fatalf("writing identities file: %v", err)
	}

	t.Setenv("AGE_IDENTITIES_FILE", identityFile)

	recipientFile := filepath.Join(t.TempDir(), "recipients")
	if err := os.WriteFile(recipientFile, []byte(identity2.Recipient().String()+"\n"), 0o600); err != nil {
		t.Fatalf("writing recipients file: %v", err)
	}

	// The payload redact hands over, shared with the passphrase tests.
	collection := testCollection

	// grant accepts one recipient= and one file= entry and encrypts the
	// collection for both.
	_, stderr, err := runExtension(t, []string{
		"redact-ext-age", "grant",
		"recipient=" + identity1.Recipient().String(),
		"file=" + recipientFile,
	}, collection)
	if err != nil {
		t.Fatalf("grant: %v (%s)", err, stderr)
	}

	if !strings.Contains(stderr, "Granted access to 2 recipient") {
		t.Errorf("grant did not report two recipients: %s", stderr)
	}

	checkForeignFiles(t, root)

	// The extension must never fall back to the in-tree Key Exchange key.
	if _, err := os.Stat(filepath.Join(root, ".git", "redact", "key")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("in-tree Key Exchange key was created: %v", err)
	}

	// redact knows nothing about the age files, so the extension maintains their
	// git attributes entry itself: an encrypted collection must not be converted.
	attributes, err := os.ReadFile(filepath.Join(root, KeyExchangeDir, GitAttributesFile))
	if err != nil {
		t.Fatalf("reading git attributes: %v", err)
	}

	if !strings.Contains(string(attributes), GitAttributesEntry) {
		t.Errorf("git attributes lack %q: %s", GitAttributesEntry, attributes)
	}

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "list"}, nil)
	if err != nil {
		t.Fatalf("list: %v (%s)", err, stderr)
	}

	if got := strings.Count(stdout, "age recipient:"); got != 2 {
		t.Errorf("list printed %d recipients, want 2: %s", got, stdout)
	}

	standard := identity1.Recipient().String()
	hybrid := identity2.Recipient().String()

	if !strings.Contains(stdout, "("+standard+")") {
		t.Errorf("list omitted the full standard recipient: %s", stdout)
	}

	if !strings.Contains(stdout, "("+hybrid[:len(standard)]+"...)") || strings.Contains(stdout, hybrid) {
		t.Errorf("list did not shorten the hybrid recipient to standard recipient length: %s", stdout)
	}

	// Two local identities have encrypted copies, so a bare get cannot pick
	// one and must ask for an explicit selection without leaking the key.
	stdout, stderr, err = runExtension(t, []string{"redact-ext-age", "get"}, nil)
	if code := ExitCode(err); code != MissingData {
		t.Fatalf("ambiguous get: exit code %d, want %d", code, MissingData)
	}

	if stdout != "" {
		t.Errorf("ambiguous get printed a collection: %s", stdout)
	}

	if !strings.Contains(stderr, "agekey=<recipient>") {
		t.Errorf("ambiguous get did not suggest agekey=: %s", stderr)
	}

	for _, recipient := range []string{identity1.Recipient().String(), identity2.Recipient().String()} {
		id := ageutil.RecipientID(recipient)
		if !strings.Contains(stderr, id) {
			t.Errorf("ambiguous get did not list candidate %s: %s", id, stderr)
		}

		publicPath := filepath.Join(root, KeyExchangeDir, id+ExtAgePublic)

		public, err := os.ReadFile(publicPath)
		if err != nil || strings.TrimSpace(string(public)) != recipient {
			t.Fatalf("public recipient was not retained: %v", err)
		}

		if len(id) != 40 {
			t.Fatalf("recipient ID has %d characters", len(id))
		}

		decrypted, diagnostics, err := runExtension(t,
			[]string{"redact-ext-age", "get", "agekey=" + id[:12]}, nil)
		if err != nil || decrypted != string(collection) {
			t.Fatalf("get by fingerprint: %v (%s)", err, diagnostics)
		}
	}

	// An explicit recipient unlocks the collection byte for byte.
	stdout, stderr, err = runExtension(t, []string{
		"redact-ext-age", "get", "agekey=" + ageutil.RecipientID(identity1.Recipient().String()),
	}, nil)
	if err != nil {
		t.Fatalf("get with agekey: %v (%s)", err, stderr)
	}

	if stdout != string(collection) {
		t.Errorf("get exported %q, want the original collection", stdout)
	}

	t.Run("stdin identities", func(t *testing.T) {
		checkStdinIdentity(t, identity1, collection)
	})

	t.Run("nested directory", func(t *testing.T) {
		nested := filepath.Join(root, "nested")
		if err := os.Mkdir(nested, 0700); err != nil {
			t.Fatal(err)
		}

		t.Chdir(nested)

		stdout, stderr, err := runExtension(t,
			[]string{"redact-ext-age", "get", "agekey=" + identity1.Recipient().String()}, nil)
		if err != nil || stdout != string(collection) {
			t.Fatalf("get from nested directory: %v (%s)", err, stderr)
		}
	})

	// A damaged encrypted copy fails without printing the collection.
	secretPath := filepath.Join(root, KeyExchangeDir, ExchangeAgeSecretKeyFile(identity1.Recipient().String()))
	if err := os.WriteFile(secretPath, []byte("not an age file\n"), 0o600); err != nil {
		t.Fatalf("damaging encrypted key: %v", err)
	}

	stdout, _, err = runExtension(t, []string{"redact-ext-age", "get", "agekey=" + identity1.Recipient().String()}, nil)
	if code := ExitCode(err); code != Failure {
		t.Errorf("get with damaged key: exit code %d, want %d", code, Failure)
	}

	if stdout != "" {
		t.Errorf("get with damaged key printed a collection: %s", stdout)
	}

	// put and update re-encrypt the supplied collection for every registered
	// recipient, repairing the damaged copy.
	if _, stderr, err := runExtension(t, []string{"redact-ext-age", "put"}, collection); err != nil {
		t.Fatalf("put: %v (%s)", err, stderr)
	}

	stdout, stderr, err = runExtension(t, []string{
		"redact-ext-age", "get", "agekey=" + identity1.Recipient().String(),
	}, nil)
	if err != nil {
		t.Fatalf("get after put: %v (%s)", err, stderr)
	}

	if stdout != string(collection) {
		t.Errorf("get after put exported %q, want the original collection", stdout)
	}

	if _, stderr, err := runExtension(t, []string{"redact-ext-age", "update"}, collection); err != nil {
		t.Fatalf("update: %v (%s)", err, stderr)
	}

	// revoke removes both registrations and their encrypted copies.
	_, stderr, err = runExtension(t, []string{
		"redact-ext-age", "revoke",
		"recipient=" + identity1.Recipient().String(),
		"recipient=" + ageutil.RecipientID(identity2.Recipient().String())[:12],
	}, nil)
	if err != nil {
		t.Fatalf("revoke: %v (%s)", err, stderr)
	}

	if !strings.Contains(stderr, "Removed 2 age recipient registration") {
		t.Errorf("revoke did not report two registrations: %s", stderr)
	}

	assertNoAgeFiles(t, filepath.Join(root, KeyExchangeDir))

	// With no registrations left, update succeeds while reporting nothing to
	// do, and a bare get is not applicable.
	if _, _, err := runExtension(t, []string{"redact-ext-age", "update"}, collection); err != nil {
		t.Fatalf("update after revoke: %v", err)
	}

	stdout, _, err = runExtension(t, []string{"redact-ext-age", "get"}, nil)
	if code := ExitCode(err); code != NotApplicable {
		t.Errorf("get after revoke: exit code %d, want %d", code, NotApplicable)
	}

	if stdout != "" {
		t.Errorf("get after revoke printed a collection: %s", stdout)
	}
}

func checkStdinIdentity(t *testing.T, identity *age.X25519Identity, collection []byte) {
	t.Helper()

	key := []byte(identity.String() + "\n")

	const passphrase = "age stdin test passphrase" //nolint:gosec // Test fixture.

	protected := passphraseContainer(t, string(key), passphrase)
	for _, scenario := range []struct {
		name       string
		input      []byte
		passphrase string
		wantCode   int
	}{
		{"plain", key, "", Success},
		{"encrypted", protected, passphrase, Success},
		{"wrong passphrase", protected, "incorrect secret", Failure},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Setenv("REDACT_AGE_PASSPHRASE", scenario.passphrase)

			stdout, stderr, err := runExtension(t,
				[]string{"redact-ext-age", "get", "private-key=stdin"}, scenario.input)
			if ExitCode(err) != scenario.wantCode {
				t.Fatalf("stdin get outcome: %v (%s)", err, stderr)
			}

			if err == nil && stdout != string(collection) {
				t.Fatal("stdin get changed the payload")
			}

			if err != nil && stdout != "" {
				t.Fatal("failed stdin get exposed a payload")
			}
		})
	}
}

func checkForeignFiles(t *testing.T, root string) {
	t.Helper()

	for _, name := range []string{"something.key", "something.asc", "something.other.pub"} {
		filename := filepath.Join(root, KeyExchangeDir, name)
		if err := os.WriteFile(filename, []byte("other extension data"), 0600); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			data, err := os.ReadFile(filename)
			if err != nil || string(data) != "other extension data" {
				t.Errorf("age modified another extension's file %s: %v", name, err)
			}
		})
	}
}

// testCollection is what redact hands over as the Key Exchange collection: an
// opaque payload this extension encrypts and decrypts without parsing. It has
// the shape of redact's armored collection, so a round trip stays byte-exact
// without depending on redact's API.
var testCollection = []byte(`-----BEGIN REDACT KEY EXCHANGE-----

dGVzdCBtYXRlcmlhbDogdGhlIGNvbGxlY3Rpb24gaXMgb3BhcXVlIHRvIHRoaXMgZXh0ZW5zaW9u

-----END REDACT KEY EXCHANGE-----
`)

func initRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	out, err := exec.Command("git", "init", "-q", root).CombinedOutput()
	if err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}

	t.Chdir(root)

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	return root
}

func assertNoAgeFiles(t *testing.T, dir string) {
	t.Helper()

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ExtAgeSecret) ||
			strings.HasSuffix(entry.Name(), ExtAgePublic)) {
			t.Errorf("revoked age file survives: %s", path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("scanning Key Exchange directory: %v", err)
	}
}

func runExtension(t *testing.T, args []string, input []byte) (string, string, error) {
	t.Helper()

	stdinFile, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatalf("creating stdin file: %v", err)
	}
	defer stdinFile.Close() // nolint:errcheck

	if _, err := stdinFile.Write(input); err != nil {
		t.Fatalf("writing stdin file: %v", err)
	}

	if _, err := stdinFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("rewinding stdin file: %v", err)
	}

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("creating stdout file: %v", err)
	}
	defer stdoutFile.Close() // nolint:errcheck

	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("creating stderr file: %v", err)
	}
	defer stderrFile.Close() // nolint:errcheck

	oldIn, oldOut, oldErr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = stdinFile, stdoutFile, stderrFile

	defer func() {
		os.Stdin, os.Stdout, os.Stderr = oldIn, oldOut, oldErr
	}()

	runErr := app().Run(context.Background(), args)

	return readCaptured(t, stdoutFile), readCaptured(t, stderrFile), runErr
}

func readCaptured(t *testing.T, file *os.File) string {
	t.Helper()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("rewinding captured output: %v", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("reading captured output: %v", err)
	}

	return string(data)
}
