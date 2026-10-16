package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/julian7/redact/kx"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
)

func TestGPGListPaths(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	root := t.TempDir()

	cmd := exec.Command("git", "init", "-q", root)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", output, err)
	}

	t.Chdir(root)

	r := &repo.Repo{}
	if err := r.SetupRepo(); err != nil {
		t.Fatal(err)
	}

	key, err := openpgp.NewEntity("recipient", "", "recipient@example.com", &packet.Config{RSABits: 1024})
	if err != nil {
		t.Fatal(err)
	}

	if err := kx.SaveGPGPubkeyToKX(r, key); err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}

	for _, cwd := range []string{root, sub} {
		t.Run(filepath.Base(cwd), func(t *testing.T) {
			t.Chdir(cwd)
			checkGPGList(t, fmt.Sprintf("%x", key.PrimaryKey.Fingerprint))
		})
	}
}

func checkGPGList(t *testing.T, fingerprint string) {
	t.Helper()

	output, err := os.CreateTemp(t.TempDir(), "list-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	previous := os.Stdout
	os.Stdout = output

	t.Cleanup(func() { os.Stdout = previous })

	rt := &Runtime{Repo: &repo.Repo{}, Logger: logger.With(nil)}
	if err := rt.accessListDo(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "fingerprint: "+fingerprint) {
		t.Fatalf("recipient missing from listing: %q", data)
	}
}
