package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/ProtonMail/go-crypto/openpgp/s2k"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/repo"
)

func armoredIdentity(t *testing.T, identity *openpgp.Entity, private bool) []byte {
	t.Helper()

	var output bytes.Buffer

	kind := openpgp.PublicKeyType
	if private {
		kind = openpgp.PrivateKeyType
	}

	writer, err := armor.Encode(&output, kind, nil)
	if err != nil {
		t.Fatal(err)
	}

	if private {
		err = identity.SerializePrivateWithoutSigning(writer, nil)
	} else {
		err = identity.Serialize(writer)
	}

	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	return output.Bytes()
}

func TestPrivateKeyInput(t *testing.T) {
	t.Setenv(passphraseEnv, "unused passphrase")

	originalPrompt := promptPassphrase
	promptPassphrase = func(string) ([]byte, error) {
		t.Fatal("unencrypted key prompted for a passphrase")

		return nil, nil
	}

	t.Cleanup(func() { promptPassphrase = originalPrompt })

	// All keys, encrypted copies, and recovered collections stay in memory.
	r := &repo.Repo{Workdir: memfs.New()}

	identity, err := openpgp.NewEntity("recipient", "", "recipient@example.test",
		&packet.Config{Algorithm: packet.PubKeyAlgoEdDSA})
	if err != nil {
		t.Fatal(err)
	}

	private := armoredIdentity(t, identity, true)

	plaintext := []byte("opaque payload without a Redact preamble or PEM wrapper\x00\xff")

	var ciphertext bytes.Buffer

	writer, err := openpgp.Encrypt(&ciphertext, openpgp.EntityList{identity}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := writer.Write(plaintext); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	filename := fmt.Sprintf(".redact/%x.key", identity.PrimaryKey.Fingerprint)

	if err := r.Workdir.MkdirAll(".redact", 0700); err != nil {
		t.Fatal(err)
	}

	if err := util.WriteFile(r.Workdir, filename, ciphertext.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	for _, selector := range []string{"", fmt.Sprintf("%x", identity.PrimaryKey.Fingerprint), "recipient@example.test"} {
		got, err := loadStdinExchangeKey(r, bytes.NewReader(private), selector)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, plaintext) {
			t.Fatal("decryption changed the opaque payload")
		}
	}

	for _, input := range [][]byte{nil, []byte("not a private key"), armoredIdentity(t, identity, false)} {
		if _, err := loadStdinExchangeKey(r, bytes.NewReader(input), ""); ext.ExitCode(err) != ext.Failure {
			t.Fatalf("invalid private key outcome: %v", err)
		}
	}

	_, err = loadStdinExchangeKey(r, bytes.NewReader(private), "other@example.test")
	if ext.ExitCode(err) != ext.NotApplicable {
		t.Fatalf("mismatched identity outcome: %v", err)
	}

	const passphrase = "private key test secret"

	config := &packet.Config{S2KConfig: &s2k.Config{S2KMode: s2k.IteratedSaltedS2K, S2KCount: 65536}}
	if err := identity.EncryptPrivateKeys([]byte(passphrase), config); err != nil {
		t.Fatal(err)
	}

	protected := armoredIdentity(t, identity, true)
	checkEncryptedPrivateKey(t, r, protected, plaintext, passphrase)

	// Corrupt the final integrity bytes, after the entire collection payload.
	corrupt := bytes.Clone(ciphertext.Bytes())
	corrupt[len(corrupt)-1] ^= 1

	if err := util.WriteFile(r.Workdir, filename, corrupt, 0600); err != nil {
		t.Fatal(err)
	}

	if got, err := loadStdinExchangeKey(r, bytes.NewReader(private), ""); err == nil || got != nil {
		t.Fatal("corrupt ciphertext exposed a collection")
	}
}

func checkEncryptedPrivateKey(t *testing.T, r *repo.Repo, protected, plaintext []byte, passphrase string) {
	t.Helper()

	for _, scenario := range []struct {
		name       string
		env        string
		unset      bool
		answer     string
		promptErr  error
		wantCode   int
		wantPrompt int
	}{
		{name: "environment", env: passphrase},
		{name: "wrong environment", env: "wrong secret", wantCode: ext.Failure},
		{name: "empty environment", wantCode: ext.Failure},
		{name: "terminal", unset: true, answer: passphrase, wantPrompt: 1},
		{name: "wrong terminal answer", unset: true, answer: "wrong secret", wantCode: ext.Failure, wantPrompt: 1},
		{name: "no terminal", unset: true, promptErr: ErrPassphraseUnavailable, wantCode: ext.MissingData, wantPrompt: 1},
		{name: "terminal read failure", unset: true, promptErr: io.EOF, wantCode: ext.Failure, wantPrompt: 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Setenv(passphraseEnv, scenario.env)

			if scenario.unset {
				if err := os.Unsetenv(passphraseEnv); err != nil {
					t.Fatal(err)
				}
			}

			prompts := 0
			previous := promptPassphrase
			promptPassphrase = func(string) ([]byte, error) {
				prompts++

				return []byte(scenario.answer), scenario.promptErr
			}

			t.Cleanup(func() { promptPassphrase = previous })

			got, err := loadStdinExchangeKey(r, bytes.NewReader(protected), "")
			if ext.ExitCode(err) != scenario.wantCode || prompts != scenario.wantPrompt {
				t.Fatalf("outcome=%d prompts=%d, want outcome=%d prompts=%d",
					ext.ExitCode(err), prompts, scenario.wantCode, scenario.wantPrompt)
			}

			if err == nil && !bytes.Equal(got, plaintext) {
				t.Fatal("encrypted key did not recover the original payload")
			}

			if err != nil && len(got) != 0 {
				t.Fatal("failed decryption exposed plaintext")
			}

			if err != nil && (strings.Contains(err.Error(), passphrase) || strings.Contains(err.Error(), "wrong secret")) {
				t.Fatal("error exposed the supplied passphrase")
			}
		})
	}
}
