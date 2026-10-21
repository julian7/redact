package kx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
)

var errExchangeWrite = errors.New("exchange write failed")

type failingFS struct {
	billy.Filesystem
	failure string
}

func TestExchangePaths(t *testing.T) {
	key, err := openpgp.NewEntity("recipient", "", "recipient@example.com", &packet.Config{RSABits: 1024})
	if err != nil {
		t.Fatal(err)
	}

	for _, memory := range []bool{false, true} {
		for _, nested := range []bool{false, true} {
			t.Run(fmt.Sprintf("memory=%t/nested=%t", memory, nested), func(t *testing.T) {
				root := t.TempDir()

				cwd := root
				if nested {
					cwd = filepath.Join(root, "sub")
					if err := os.Mkdir(cwd, 0755); err != nil {
						t.Fatal(err)
					}
				}

				t.Chdir(cwd)

				fs := osfs.New(root, osfs.WithBoundOS())
				if memory {
					fs = memfs.New()
				}

				checkExchangePaths(t, &repo.Repo{Workdir: fs}, key)
			})
		}
	}
}

func checkExchangePaths(t *testing.T, r *repo.Repo, key *openpgp.Entity) {
	t.Helper()

	if err := SaveGPGPubkeyToKX(r, key); err != nil {
		t.Fatal(err)
	}

	payload := []byte("updated redact keys")

	if err := SaveGPGKeyToKX(r, key, func(w io.Writer) error {
		_, err := w.Write(payload)

		return err
	}); err != nil {
		t.Fatalf("save exchange key: %v", err)
	}

	stub, err := r.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := util.ReadFile(r.Workdir, repo.ExchangeSecretKeyFile(stub))
	if err != nil {
		t.Fatal(err)
	}

	assertExchangePayload(t, key, false, ciphertext, payload)
}

func (fs *failingFS) TempFile(dir, prefix string) (billy.File, error) {
	file, err := fs.Filesystem.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}

	return &failingFile{File: file, failure: fs.failure}, nil
}

func (fs *failingFS) Rename(oldpath, newpath string) error {
	if fs.failure == "rename" {
		return errExchangeWrite
	}

	return fs.Filesystem.Rename(oldpath, newpath)
}

type failingFile struct {
	billy.File
	failure string
}

func (f *failingFile) Write(p []byte) (int, error) {
	if f.failure == "write" {
		return 0, errExchangeWrite
	}

	return f.File.Write(p)
}

func (f *failingFile) Close() error {
	err := f.File.Close()
	if f.failure == "close" {
		return errExchangeWrite
	}

	return err
}

func TestExchangeReplacement(t *testing.T) {
	key, err := openpgp.NewEntity("recipient", "", "recipient@example.com", &packet.Config{RSABits: 1024})
	if err != nil {
		t.Fatal(err)
	}

	for _, public := range []bool{false, true} {
		for _, failure := range []string{"", "write", "close", "rename", "serialize"} {
			if public && failure == "serialize" {
				continue
			}

			t.Run(fmt.Sprintf("public=%t/failure=%s", public, failure), func(t *testing.T) {
				runExchangeReplacement(t, key, public, failure)
			})
		}
	}
}

func runExchangeReplacement(t *testing.T, key *openpgp.Entity, public bool, failure string) {
	t.Helper()

	fs := &failingFS{Filesystem: memfs.New(), failure: failure}
	r := &repo.Repo{Workdir: fs}

	stub, err := r.GetExchangeFilenameStubFor(key.PrimaryKey.Fingerprint)
	if err != nil {
		t.Fatal(err)
	}

	name := repo.ExchangeSecretKeyFile(stub)
	if public {
		name = repo.ExchangePubKeyFile(stub)
	}

	previous := []byte("previous valid copy")
	if err := util.WriteFile(fs, name, previous, 0644); err != nil {
		t.Fatal(err)
	}

	payload := []byte("serialized redact key")

	if public {
		err = SaveGPGPubkeyToKX(r, key)
	} else {
		err = SaveGPGKeyToKX(r, key, func(w io.Writer) error {
			if failure == "serialize" {
				return errExchangeWrite
			}

			_, err := w.Write(payload)

			return err
		})
	}

	if failure != "" && !errors.Is(err, errExchangeWrite) {
		t.Fatalf("expected injected error, got %v", err)
	}

	if failure == "" && err != nil {
		t.Fatal(err)
	}

	data, err := util.ReadFile(fs, name)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := fs.ReadDir(repo.DefaultKeyExchangeDir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("expected destination and attributes only, got %v, %v", entries, err)
	}

	if failure != "" {
		if !bytes.Equal(data, previous) {
			t.Fatal("previous destination was changed")
		}

		return
	}

	assertExchangePayload(t, key, public, data, payload)
}

func assertExchangePayload(t *testing.T, key *openpgp.Entity, public bool, data, payload []byte) {
	t.Helper()

	if public {
		keys, err := gpgutil.LoadPubKey(bytes.NewReader(data), true)
		if err != nil || len(keys) != 1 {
			t.Fatalf("invalid public key: %v", err)
		}

		if !bytes.Equal(keys[0].PrimaryKey.Fingerprint, key.PrimaryKey.Fingerprint) {
			t.Fatal("unexpected public key")
		}
	} else {
		message, err := openpgp.ReadMessage(bytes.NewReader(data), openpgp.EntityList{key}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		plain, err := io.ReadAll(message.UnverifiedBody)
		if err != nil || !bytes.Equal(plain, payload) {
			t.Fatalf("invalid decrypted key: %q, %v", plain, err)
		}
	}
}
