package ext_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/repo"
)

func dotconfPath(r *repo.Repo) string {
	return r.Workdir.Join(ext.DotconfDir, ext.DotconfFile)
}

func redactPath(r *repo.Repo) string {
	return r.Workdir.Join(repo.DefaultKeyExchangeDir, ext.ConfigFilename)
}

func newTestRepo() *repo.Repo {
	return &repo.Repo{Workdir: memfs.New()}
}

func setupConfigFiles(t *testing.T, r *repo.Repo, writeDotconf, writeRedact bool) {
	t.Helper()

	if writeDotconf {
		writeExtsFile(t, r, dotconfPath(r), map[string]ext.Ext{
			"dotconfig": {Command: "dotconfig-cmd"},
		})
	}

	if writeRedact {
		writeExtsFile(t, r, redactPath(r), map[string]ext.Ext{
			"redact": {Command: "redact-cmd"},
		})
	}
}

func writeExtsFile(t *testing.T, r *repo.Repo, path string, exts map[string]ext.Ext) {
	t.Helper()

	dir := r.Workdir.Join(path, "..")
	if err := r.Workdir.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}

	fd, err := r.Workdir.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}

	if err := json.NewEncoder(fd).Encode(exts); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}

	_ = fd.Close()
}

func readExtsFile(t *testing.T, r *repo.Repo, path string) map[string]ext.Ext {
	t.Helper()

	fd, err := r.Workdir.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer fd.Close()

	exts := map[string]ext.Ext{}
	if err := json.NewDecoder(fd).Decode(&exts); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}

	return exts
}

func fileExists(t *testing.T, r *repo.Repo, path string) bool {
	t.Helper()

	_, err := r.Workdir.Stat(path)

	return err == nil
}

func assertExtSaved(t *testing.T, r *repo.Repo, path string) {
	t.Helper()

	saved := readExtsFile(t, r, path)

	if _, ok := saved["new"]; !ok {
		t.Errorf("expected new extension to be saved to %s", path)
	}
}

func assertNotCreated(t *testing.T, r *repo.Repo, pathFn func(r *repo.Repo) string) {
	t.Helper()

	if pathFn == nil {
		return
	}

	path := pathFn(r)

	if fileExists(t, r, path) {
		t.Errorf("did not expect %s to be created", path)
	}
}

func assertDotconfUntouched(t *testing.T, r *repo.Repo) {
	t.Helper()

	dotconf := readExtsFile(t, r, dotconfPath(r))

	if _, ok := dotconf["new"]; ok {
		t.Errorf("%s should not have been modified when %s exists", dotconfPath(r), redactPath(r))
	}

	if _, ok := dotconf["dotconfig"]; !ok {
		t.Errorf("expected %s content to remain untouched", dotconfPath(r))
	}
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name         string
		writeDotconf bool
		writeRedact  bool
		wantPresent  []string
		wantAbsent   []string
	}{
		{
			name:         "both exist: .redact/config.json takes precedence",
			writeDotconf: true,
			writeRedact:  true,
			wantPresent:  []string{"redact"},
			wantAbsent:   []string{"dotconfig"},
		},
		{
			name:         "only .config/redact.json exists",
			writeDotconf: true,
			writeRedact:  false,
			wantPresent:  []string{"dotconfig"},
		},
		{
			name:         "only .redact/config.json exists",
			writeDotconf: false,
			writeRedact:  true,
			wantPresent:  []string{"redact"},
		},
		{
			name:         "neither exists",
			writeDotconf: false,
			writeRedact:  false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRepo()
			setupConfigFiles(t, r, tc.writeDotconf, tc.writeRedact)

			conf, err := ext.Load(r)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			for _, name := range tc.wantPresent {
				if _, ok := conf.Ext(name); !ok {
					t.Errorf("expected extension %q to be loaded, but it was not", name)
				}
			}

			for _, name := range tc.wantAbsent {
				if _, ok := conf.Ext(name); ok {
					t.Errorf("expected extension %q not to be loaded, but it was", name)
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	tt := []struct {
		name             string
		writeDotconf     bool
		writeRedact      bool
		wantSavedAt      func(r *repo.Repo) string
		wantNotCreatedAt func(r *repo.Repo) string
	}{
		{
			name:             "neither exists: creates .redact/config.json",
			writeDotconf:     false,
			writeRedact:      false,
			wantSavedAt:      redactPath,
			wantNotCreatedAt: dotconfPath,
		},
		{
			name:             "only .config/redact.json exists: saved there",
			writeDotconf:     true,
			writeRedact:      false,
			wantSavedAt:      dotconfPath,
			wantNotCreatedAt: redactPath,
		},
		{
			name:             "only .redact/config.json exists: saved there",
			writeDotconf:     false,
			writeRedact:      true,
			wantSavedAt:      redactPath,
			wantNotCreatedAt: dotconfPath,
		},
		{
			name:             "both exist: .redact/config.json takes precedence, .config/redact.json left untouched",
			writeDotconf:     true,
			writeRedact:      true,
			wantSavedAt:      redactPath,
			wantNotCreatedAt: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRepo()
			setupConfigFiles(t, r, tc.writeDotconf, tc.writeRedact)

			conf, err := ext.Load(r)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if err := conf.AddExt("new", ext.Ext{Command: "new-cmd"}); err != nil {
				t.Fatalf("AddExt() error = %v", err)
			}

			if err := conf.Save(); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			assertExtSaved(t, r, tc.wantSavedAt(r))
			assertNotCreated(t, r, tc.wantNotCreatedAt)

			if tc.writeDotconf && tc.writeRedact {
				assertDotconfUntouched(t, r)
			}
		})
	}
}
