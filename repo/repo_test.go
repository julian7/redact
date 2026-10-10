package repo_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/julian7/redact/repo"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestSetupRepoGitContext(t *testing.T) { //nolint:funlen
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	main := filepath.Join(tmp, "main")
	wt := filepath.Join(tmp, "wt")
	sub := filepath.Join(main, "sub")

	git(t, tmp, "init", "-q", main)
	git(t, main, "-c", "user.name=test", "-c", "user.email=test@example.com",
		"commit", "-q", "--allow-empty", "-m", "init")
	git(t, main, "worktree", "add", "-q", wt)

	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}

	t.Chdir(main)

	r := &repo.Repo{}
	if err := r.SetupRepo(); err != nil {
		t.Fatal(err)
	}

	if err := r.Generate(); err != nil {
		t.Fatal(err)
	}

	if err := r.Save(); err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name string
		dir  string
		env  map[string]string
	}{
		{name: "toplevel", dir: main},
		{name: "subdirectory", dir: sub},
		{
			name: "absolute GIT_DIR and GIT_WORK_TREE",
			dir:  main,
			env: map[string]string{
				"GIT_DIR":       filepath.Join(main, ".git"),
				"GIT_WORK_TREE": main,
			},
		},
		{
			name: "relative GIT_DIR and GIT_WORK_TREE",
			dir:  main,
			env:  map[string]string{"GIT_DIR": ".git", "GIT_WORK_TREE": "."},
		},
		{name: "linked worktree", dir: wt},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(tc.dir)

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			r := &repo.Repo{}
			if _, err := r.LoadSecretKey(context.Background(), nil); err != nil {
				t.Fatalf("loading secret key: %v", err)
			}
		})
	}
}

func TestRemoveGitSettingsMissingKeys(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	tmp := t.TempDir()
	git(t, tmp, "init", "-q")
	// emulate a repo unlocked by an earlier version, without merge driver
	git(t, tmp, "config", "filter.redact.clean", "redact git clean")
	t.Chdir(tmp)

	r := &repo.Repo{}

	missing, err := r.MissingGitSettings()
	if err != nil {
		t.Fatalf("checking git settings: %v", err)
	}

	if len(missing) != 3 || missing[len(missing)-1] != "merge.redact.driver" {
		t.Errorf("unexpected missing settings: %v", missing)
	}

	if err := r.RemoveGitSettings(nil); err != nil {
		t.Fatalf("removing git settings: %v", err)
	}

	cmd := exec.Command("git", "config", "filter.redact.clean")
	if err := cmd.Run(); err == nil {
		t.Error("filter.redact.clean is still set")
	}
}
