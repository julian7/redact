//go:build darwin

package ageutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultIdentityFileUnderHome(t *testing.T) {
	home := t.TempDir()

	// Neither the identities file named by the environment nor one named by
	// XDG_CONFIG_HOME is there, so the per-user default below the home decides.
	t.Setenv("AGE_IDENTITIES_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)

	got := DefaultIdentityFile()

	want := filepath.Join(home, ".config", "age", "identities")
	if got != want {
		t.Errorf("DefaultIdentityFile() = %q, want %q", got, want)
	}

	if strings.Contains(got, "Library") {
		t.Errorf("DefaultIdentityFile() looks below the macOS Application Support folders: %q", got)
	}

	// XDG_CONFIG_HOME overrides the default here as it does on the unixes, where
	// os.UserConfigDir reads that variable itself.
	xdg := filepath.Join(home, "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)

	want = filepath.Join(xdg, "age", "identities")
	if got := DefaultIdentityFile(); got != want {
		t.Errorf("DefaultIdentityFile() with XDG_CONFIG_HOME = %q, want %q", got, want)
	}
}
