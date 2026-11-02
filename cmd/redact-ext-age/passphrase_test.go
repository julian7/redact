package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/julian7/redact/ageutil"
)

func TestGetWithPassphraseProtectedIdentitiesFile(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	// Swapped before the extension runs: the real prompt reads the controlling
	// terminal of whoever runs the test. The count below covers the grant too,
	// which must not need the passphrase of an identities file at all.
	prompts := answerPassphrase(t, passphrase, nil)
	filename := passphraseRepo(t, passphrase)

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "get"}, nil)
	if err != nil {
		t.Fatalf("get with a passphrase protected identities file: %v (%s)", err, stderr)
	}

	if stdout != string(testCollection) {
		t.Errorf("get exported %q, want the original collection", stdout)
	}

	if len(*prompts) != 1 {
		t.Fatalf("the extension asked %d passphrases, want 1: %q", len(*prompts), *prompts)
	}

	if !strings.Contains((*prompts)[0], filename) {
		t.Errorf("the prompt does not name %s: %s", filename, (*prompts)[0])
	}
}

func TestGetRefusesWrongPassphrase(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	typed := "not the passphrase of this file"

	prompts := answerPassphrase(t, typed, nil)
	passphraseRepo(t, passphrase)

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "get"}, nil)
	if code := ExitCode(err); code != Failure {
		t.Errorf("get with a wrong passphrase: exit code %d, want %d", code, Failure)
	}

	if stdout != "" {
		t.Errorf("get with a wrong passphrase printed a collection: %s", stdout)
	}

	if len(*prompts) != 1 {
		t.Errorf("the extension asked %d passphrases, want 1: %q", len(*prompts), *prompts)
	}

	// redact reads the collection on stdout, and a passphrase is a secret: what
	// the user typed belongs in neither stream, nor in a message they may have
	// to paste somewhere.
	if strings.Contains(stdout+stderr, typed) {
		t.Errorf("the extension repeated the passphrase the user typed: %s", stderr)
	}
}

func TestGetWithoutTerminalForPassphrase(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	prompts := answerPassphrase(t, "", ageutil.ErrPassphraseUnavailable)
	filename := passphraseRepo(t, passphrase)

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "get"}, nil)
	if code := ExitCode(err); code != NotApplicable {
		t.Errorf("get without a terminal: exit code %d, want %d", code, NotApplicable)
	}

	if stdout != "" {
		t.Errorf("get without a terminal printed a collection: %s", stdout)
	}

	if !strings.Contains(stderr, "no terminal to type its passphrase") {
		t.Errorf("get without a terminal did not say how to get one: %s", stderr)
	}

	if !strings.Contains(stderr, filename) {
		t.Errorf("get without a terminal did not name %s: %s", filename, stderr)
	}

	if len(*prompts) != 1 {
		t.Errorf("the extension asked %d passphrases, want 1: %q", len(*prompts), *prompts)
	}
}

func TestGetIgnoresUnusableIdentitiesFile(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	unusable := []struct {
		name        string
		contents    func(t *testing.T) []byte
		promptErr   error
		wantPrompts int
		wantStderr  string
	}{
		{
			name:       "not an age identities file",
			contents:   func(*testing.T) []byte { return []byte("not an age private key\n") },
			wantStderr: "Ignoring age identities file",
		},
		{
			name: "a passphrase nobody is there to type",
			contents: func(t *testing.T) []byte {
				keyText, _ := newIdentity(t)

				return passphraseContainer(t, keyText, passphrase)
			},
			promptErr:   ageutil.ErrPassphraseUnavailable,
			wantPrompts: 1,
			wantStderr:  "no terminal to type its passphrase",
		},
	}

	for _, one := range unusable {
		t.Run(one.name, func(t *testing.T) {
			initRepo(t)

			usableKey, usableRecipient := newIdentity(t)
			usableFile := identitiesFile(t, []byte(usableKey))
			unusableFile := identitiesFile(t, one.contents(t))

			prompts := answerPassphrase(t, "", one.promptErr)

			grantCollection(t, testCollection, usableRecipient)

			stdout, stderr, err := runExtension(t, []string{
				"redact-ext-age", "get",
				"keyfile=" + unusableFile, "keyfile=" + usableFile,
			}, nil)
			if err != nil {
				t.Fatalf("get with %s among the identities files: %v (%s)", one.name, err, stderr)
			}

			if stdout != string(testCollection) {
				t.Errorf("get exported %q, want the original collection", stdout)
			}

			if !strings.Contains(stderr, one.wantStderr) {
				t.Errorf("get did not report the unusable identities file (%s): %s", one.wantStderr, stderr)
			}

			if !strings.Contains(stderr, unusableFile) {
				t.Errorf("get did not name %s: %s", unusableFile, stderr)
			}

			if len(*prompts) != one.wantPrompts {
				t.Errorf("the extension asked %d passphrases, want %d: %q",
					len(*prompts), one.wantPrompts, *prompts)
			}
		})
	}
}

func TestGetPromptNamesTheSourceOfTheKey(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	initRepo(t)

	keyText, recipient := newIdentity(t)
	grantCollection(t, testCollection, recipient)

	prompts := answerPassphrase(t, passphrase, nil)

	stdout, stderr, err := runExtension(t,
		[]string{"redact-ext-age", "get", "private-key=stdin", "prompt=thekey.sh"},
		passphraseContainer(t, keyText, passphrase))
	if err != nil {
		t.Fatalf("get with a labelled stdin key: %v (%s)", err, stderr)
	}

	if stdout != string(testCollection) {
		t.Errorf("get exported %q, want the original collection", stdout)
	}

	if len(*prompts) != 1 {
		t.Fatalf("the extension asked %d passphrases, want 1: %q", len(*prompts), *prompts)
	}

	// A key that arrives through a pipe has no name of its own, so without the
	// label the request cannot say which key it is about.
	if !strings.Contains((*prompts)[0], "thekey.sh") {
		t.Errorf("the prompt does not name the source of the key: %s", (*prompts)[0])
	}

	if strings.Contains((*prompts)[0], "stdin") {
		t.Errorf("the prompt names stdin although the config labelled the key: %s", (*prompts)[0])
	}
}

func TestGetPromptWithoutLabelNamesStdin(t *testing.T) {
	const passphrase = "the passphrase of this identities file"

	initRepo(t)

	keyText, recipient := newIdentity(t)
	grantCollection(t, testCollection, recipient)

	prompts := answerPassphrase(t, passphrase, nil)

	_, stderr, err := runExtension(t, []string{"redact-ext-age", "get", "private-key=stdin"},
		passphraseContainer(t, keyText, passphrase))
	if err != nil {
		t.Fatalf("get with an unlabelled stdin key: %v (%s)", err, stderr)
	}

	if len(*prompts) != 1 {
		t.Fatalf("the extension asked %d passphrases, want 1: %q", len(*prompts), *prompts)
	}

	if !strings.Contains((*prompts)[0], defaultPrompt) {
		t.Errorf("the prompt does not name %s: %s", defaultPrompt, (*prompts)[0])
	}
}

// A label belongs to the source it names: without stdin key input the identities
// file already has a name, and silently ignoring the setting would let a
// misspelled one look like it worked.
func TestGetRejectsPromptWithoutStdinKey(t *testing.T) {
	prompts := answerPassphrase(t, "", nil)

	stdout, stderr, err := runExtension(t, []string{"redact-ext-age", "get", "prompt=thekey.sh"}, nil)
	if code := ExitCode(err); code != Failure {
		t.Errorf("get with a label and no stdin key: exit code %d, want %d", code, Failure)
	}

	if stdout != "" {
		t.Errorf("get with a label and no stdin key printed a collection: %s", stdout)
	}

	// The command reports the failure itself; the harness keeps it as the error.
	if err == nil || !strings.Contains(err.Error(), "use prompt with private-key=stdin") {
		t.Errorf("get did not say how to use the label: %v (%s)", err, stderr)
	}

	if len(*prompts) != 0 {
		t.Errorf("the extension asked %d passphrases, want none: %q", len(*prompts), *prompts)
	}
}

// The label comes from a repository's committed extension config and is written to
// the terminal as-is, so it must not carry what a terminal would act on.
func TestGetRejectsUnusablePromptLabel(t *testing.T) {
	unusable := []struct {
		name    string
		label   string
		wantErr string
	}{
		{name: "a newline that could forge another prompt", label: "key\nEnter admin password:", wantErr: "printable"},
		{name: "a tab", label: "key\tsource", wantErr: "printable"},
		{name: "an escape sequence", label: "key\x1b[2J", wantErr: "printable"},
		{
			name:    "a label over the length limit",
			label:   strings.Repeat("k", maxPromptLength+1),
			wantErr: "at most 80 characters",
		},
	}

	for _, one := range unusable {
		t.Run(one.name, func(t *testing.T) {
			prompts := answerPassphrase(t, "", nil)

			stdout, stderr, err := runExtension(t,
				[]string{"redact-ext-age", "get", "private-key=stdin", "prompt=" + one.label}, nil)
			if code := ExitCode(err); code != Failure {
				t.Fatalf("get with an unusable label: exit code %d, want %d", code, Failure)
			}

			// The command reports the rejection itself; the harness keeps it as the
			// error the process prints on exit.
			if err == nil || !strings.Contains(err.Error(), one.wantErr) {
				t.Errorf("get did not say why the label is unusable: %v (%s)", err, stderr)
			}

			if stdout != "" {
				t.Errorf("get with an unusable label printed a collection: %s", stdout)
			}

			if len(*prompts) != 0 {
				t.Errorf("the extension asked %d passphrases, want none: %q", len(*prompts), *prompts)
			}
		})
	}
}

// Replace the prompt so tests never block on the controlling terminal.
func answerPassphrase(t *testing.T, answer string, promptErr error) *[]string {
	t.Helper()

	prompts := &[]string{}
	previous := promptPassphrase

	promptPassphrase = func(prompt string) (string, error) {
		*prompts = append(*prompts, prompt)

		return answer, promptErr
	}

	t.Cleanup(func() { promptPassphrase = previous })

	return prompts
}

func passphraseRepo(t *testing.T, passphrase string) string {
	t.Helper()

	initRepo(t)

	keyText, recipient := newIdentity(t)
	filename := identitiesFile(t, passphraseContainer(t, keyText, passphrase))

	t.Setenv("AGE_IDENTITIES_FILE", filename)

	grantCollection(t, testCollection, recipient)

	return filename
}

func grantCollection(t *testing.T, collection []byte, recipients ...string) {
	t.Helper()

	args := make([]string, 0, len(recipients)+2)
	args = append(args, "redact-ext-age", "grant")

	for _, recipient := range recipients {
		args = append(args, "recipient="+recipient)
	}

	if _, stderr, err := runExtension(t, args, collection); err != nil {
		t.Fatalf("grant: %v (%s)", err, stderr)
	}
}

func newIdentity(t *testing.T) (keyText, recipient string) {
	t.Helper()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating an age identity: %v", err)
	}

	text, err := ageutil.RecipientString(identity)
	if err != nil {
		t.Fatalf("deriving the recipient of an age identity: %v", err)
	}

	return identity.String() + "\n", text
}

func identitiesFile(t *testing.T, contents []byte) string {
	t.Helper()

	filename := filepath.Join(t.TempDir(), "identities")
	if err := os.WriteFile(filename, contents, 0o600); err != nil {
		t.Fatalf("writing identities file: %v", err)
	}

	return filename
}

func passphraseContainer(t *testing.T, keyText, passphrase string) []byte {
	t.Helper()

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		t.Fatalf("creating a passphrase recipient: %v", err)
	}

	// age derives the key of a container for about a second by default, which a
	// test does not need.
	recipient.SetWorkFactor(15)

	container := bytes.Buffer{}
	if err := ageutil.Encrypt(strings.NewReader(keyText), &container, recipient); err != nil {
		t.Fatalf("encrypting identities to a passphrase: %v", err)
	}

	return container.Bytes()
}
