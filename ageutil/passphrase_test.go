package ageutil

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// testPassphrase is the passphrase of the identities files the tests create.
// It is also what a test checks an error message does not repeat.
const testPassphrase = "correct horse battery staple"

// passphraseRecorder answers a passphrase prompt and records what the user was
// asked, so that a test can tell how often a passphrase was asked for, and what
// the prompt said.
type passphraseRecorder struct {
	prompts []string
	answer  string
	err     error
}

func (recorder *passphraseRecorder) ask(prompt string) (string, error) {
	recorder.prompts = append(recorder.prompts, prompt)

	return recorder.answer, recorder.err
}

func writeIdentitiesFile(t *testing.T, contents []byte) string {
	t.Helper()

	filename := filepath.Join(t.TempDir(), "identities")
	if err := os.WriteFile(filename, contents, 0o600); err != nil {
		t.Fatalf("writing identities file: %v", err)
	}

	return filename
}

func plainIdentitiesText(t *testing.T) []byte {
	t.Helper()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating an age identity: %v", err)
	}

	return []byte("# a comment\n" + identity.String() + "\n")
}

// Use a low work factor to avoid the default one-second derivation cost in tests.
func scryptRecipient(t *testing.T, passphrase string) age.Recipient {
	t.Helper()

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		t.Fatalf("creating a passphrase recipient: %v", err)
	}

	recipient.SetWorkFactor(15)

	return recipient
}

func passphraseContainer(t *testing.T, plaintext []byte, passphrase string, armored bool) []byte {
	t.Helper()

	return encryptContainer(t, plaintext, scryptRecipient(t, passphrase), armored)
}

func publicKeyContainer(t *testing.T, plaintext []byte, armored bool) []byte {
	t.Helper()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating an age identity: %v", err)
	}

	return encryptContainer(t, plaintext, identity.Recipient(), armored)
}

func encryptContainer(t *testing.T, plaintext []byte, recipient age.Recipient, armored bool) []byte {
	t.Helper()

	container := bytes.Buffer{}
	if err := Encrypt(bytes.NewReader(plaintext), &container, recipient); err != nil {
		t.Fatalf("encrypting an age container: %v", err)
	}

	if !armored {
		return container.Bytes()
	}

	return armorBytes(t, container.Bytes())
}

func armorBytes(t *testing.T, ciphertext []byte) []byte {
	t.Helper()

	armored := bytes.Buffer{}

	writer := armor.NewWriter(&armored)
	if _, err := writer.Write(ciphertext); err != nil {
		t.Fatalf("writing an armored container: %v", err)
	}

	// Close is not deferred: it writes the armored trailer, and the buffer is
	// read right after it.
	if err := writer.Close(); err != nil {
		t.Fatalf("closing an armored container: %v", err)
	}

	return armored.Bytes()
}

// Decrypt a payload to prove the parsed identity retained its private key.
func assertUnlocks(t *testing.T, identity age.Identity, payload string) {
	t.Helper()

	text, err := RecipientString(identity)
	if err != nil {
		t.Fatalf("deriving the recipient of an identity: %v", err)
	}

	recipient, err := ParseRecipient(text)
	if err != nil {
		t.Fatalf("parsing recipient %s: %v", text, err)
	}

	ciphertext := bytes.Buffer{}
	if err := Encrypt(strings.NewReader(payload), &ciphertext, recipient); err != nil {
		t.Fatalf("encrypting to %s: %v", text, err)
	}

	plaintext := bytes.Buffer{}
	if err := Decrypt(bytes.NewReader(ciphertext.Bytes()), &plaintext, identity); err != nil {
		t.Fatalf("decrypting with the identity of %s: %v", text, err)
	}

	if plaintext.String() != payload {
		t.Errorf("decrypted %q, want %q", plaintext.String(), payload)
	}
}

func TestParseIdentitiesFilePlain(t *testing.T) {
	recorder := &passphraseRecorder{answer: testPassphrase}

	identities, err := ParseIdentitiesFile(writeIdentitiesFile(t, plainIdentitiesText(t)), recorder.ask)
	if err != nil {
		t.Fatalf("parsing a plain identities file: %v", err)
	}

	if len(recorder.prompts) != 0 {
		t.Errorf("a plain identities file asked %d passphrases, want none: %q",
			len(recorder.prompts), recorder.prompts)
	}

	if len(identities) != 1 {
		t.Fatalf("parsed %d identities, want 1", len(identities))
	}

	assertUnlocks(t, identities[0], "a private key of a plain identities file")
}

func TestParseIdentitiesFilePassphrase(t *testing.T) {
	for _, armored := range []bool{false, true} {
		kind := "binary"
		if armored {
			kind = "armored"
		}

		t.Run(kind, func(t *testing.T) {
			contents := passphraseContainer(t, plainIdentitiesText(t), testPassphrase, armored)
			filename := writeIdentitiesFile(t, contents)

			recorder := &passphraseRecorder{answer: testPassphrase}

			identities, err := ParseIdentitiesFile(filename, recorder.ask)
			if err != nil {
				t.Fatalf("parsing a passphrase protected identities file: %v", err)
			}

			if len(recorder.prompts) != 1 {
				t.Fatalf("an encrypted identities file asked %d passphrases, want 1: %q",
					len(recorder.prompts), recorder.prompts)
			}

			if !strings.Contains(recorder.prompts[0], filename) {
				t.Errorf("the prompt does not name %s: %s", filename, recorder.prompts[0])
			}

			if len(identities) != 1 {
				t.Fatalf("parsed %d identities, want 1", len(identities))
			}

			assertUnlocks(t, identities[0], "a private key of an encrypted identities file")
		})
	}
}

func TestParseIdentitiesFileRefusedPassphrase(t *testing.T) {
	refused := []struct {
		name        string
		answer      string
		promptErr   error
		noPrompt    bool
		want        error
		wantPrompts int
	}{
		{
			name:        "wrong passphrase",
			answer:      "not the passphrase of this file",
			want:        ErrIncorrectPassphrase,
			wantPrompts: 1,
		},
		{
			name:        "empty passphrase",
			answer:      "",
			want:        ErrNoPassphraseEntered,
			wantPrompts: 1,
		},
		{
			name:        "no terminal to ask on",
			promptErr:   ErrPassphraseUnavailable,
			want:        ErrPassphraseUnavailable,
			wantPrompts: 1,
		},
		{
			name:        "nobody to ask",
			noPrompt:    true,
			want:        ErrPassphraseUnavailable,
			wantPrompts: 0,
		},
	}

	for _, armored := range []bool{false, true} {
		kind := "binary"
		if armored {
			kind = "armored"
		}

		for _, one := range refused {
			t.Run(kind+"/"+one.name, func(t *testing.T) {
				contents := passphraseContainer(t, plainIdentitiesText(t), testPassphrase, armored)

				recorder := &passphraseRecorder{answer: one.answer, err: one.promptErr}

				var passphrase PassphraseFunc = recorder.ask
				if one.noPrompt {
					passphrase = nil
				}

				_, err := ParseIdentitiesFile(writeIdentitiesFile(t, contents), passphrase)
				if !errors.Is(err, one.want) {
					t.Fatalf("got %v, want %v", err, one.want)
				}

				if len(recorder.prompts) != one.wantPrompts {
					t.Errorf("asked %d passphrases, want %d: %q",
						len(recorder.prompts), one.wantPrompts, recorder.prompts)
				}

				// A passphrase is a secret: reporting that it was refused must not
				// repeat it. What the user typed is not the only leak to check, so
				// the whole message is compared against it.
				if one.answer != "" && strings.Contains(err.Error(), one.answer) {
					t.Errorf("the error repeats the passphrase %q: %v", one.answer, err)
				}
			})
		}
	}
}

func TestParseIdentitiesFileEncryptedToPublicKey(t *testing.T) {
	for _, armored := range []bool{false, true} {
		kind := "binary"
		if armored {
			kind = "armored"
		}

		t.Run(kind, func(t *testing.T) {
			contents := publicKeyContainer(t, plainIdentitiesText(t), armored)

			recorder := &passphraseRecorder{answer: testPassphrase}

			_, err := ParseIdentitiesFile(writeIdentitiesFile(t, contents), recorder.ask)
			if !errors.Is(err, errNotPassphraseEncrypted) {
				t.Fatalf("got %v, want %v", err, errNotPassphraseEncrypted)
			}

			if len(recorder.prompts) != 0 {
				t.Errorf("a public key container asked %d passphrases, want none: %q",
					len(recorder.prompts), recorder.prompts)
			}
		})
	}
}

func TestParseIdentitiesFileTooLarge(t *testing.T) {
	// The limit is applied before the container is opened, so a file that only
	// looks like a container of the wrong size is enough to test it.
	ciphertext := append([]byte("age-encryption"), bytes.Repeat([]byte("x"), maxIdentityFileSize)...)

	recorder := &passphraseRecorder{answer: testPassphrase}

	_, err := ParseIdentitiesFile(writeIdentitiesFile(t, ciphertext), recorder.ask)
	if err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("got %v, want an error about the size of the file", err)
	}

	if len(recorder.prompts) != 0 {
		t.Errorf("an oversized container asked %d passphrases, want none: %q",
			len(recorder.prompts), recorder.prompts)
	}
}

func TestParseIdentitiesFileWithoutKeys(t *testing.T) {
	for _, name := range []string{"comments only", "missing file"} {
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "identities")
			if name == "comments only" {
				filename = writeIdentitiesFile(t, []byte("# nothing to unlock here\n\n"))
			}

			recorder := &passphraseRecorder{answer: testPassphrase}

			identities, err := ParseIdentitiesFile(filename, recorder.ask)
			if err != nil {
				t.Fatalf("got %v, want no error", err)
			}

			if len(identities) != 0 {
				t.Errorf("parsed %d identities, want none", len(identities))
			}

			if len(recorder.prompts) != 0 {
				t.Errorf("asked %d passphrases, want none: %q", len(recorder.prompts), recorder.prompts)
			}
		})
	}
}

func TestParseIdentitiesFileUnreadable(t *testing.T) {
	recorder := &passphraseRecorder{answer: testPassphrase}

	_, err := ParseIdentitiesFile(writeIdentitiesFile(t, []byte("not an age private key\n")), recorder.ask)
	if err == nil {
		t.Fatal("a file of keys that are not keys parsed")
	}

	for _, sentinel := range []error{ErrIncorrectPassphrase, ErrNoPassphraseEntered, ErrPassphraseUnavailable} {
		if errors.Is(err, sentinel) {
			t.Errorf("an unreadable identities file reports %v, want an error to skip: %v", sentinel, err)
		}
	}

	if len(recorder.prompts) != 0 {
		t.Errorf("asked %d passphrases, want none: %q", len(recorder.prompts), recorder.prompts)
	}
}
