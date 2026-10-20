package ext

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/julian7/redact/repo"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)

	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func privateEnvRepo(t *testing.T) *repo.Repo {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	runGit(t, root, "init", "-q")
	t.Chdir(root)

	r := &repo.Repo{}
	if err := r.SetupRepo(); err != nil {
		t.Fatal(err)
	}

	return r
}

func writePrivateEnv(t *testing.T, r *repo.Repo, contents string) string {
	t.Helper()

	dir := filepath.Join(r.CommonDir(), "redact")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(dir, "aws.env")
	if err := os.WriteFile(filename, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	return filename
}

func TestPrivateEnv(t *testing.T) {
	r := privateEnvRepo(t)
	t.Setenv("AWS_PROFILE", "parent")
	t.Setenv("REDACT_TEST_INHERITED", "inherited")
	t.Setenv("REDACT_TEST_EMPTY", "parent")
	writePrivateEnv(t, r, "# Private AWS settings\nexport AWS_PROFILE='private profile'\n"+
		"REDACT_TEST_QUOTED=\"value # with = signs\"\nREDACT_TEST_EMPTY=\n")

	for _, name := range []string{"aws", "azure"} {
		t.Run(name, func(t *testing.T) {
			extension := &Ext{name: name, repo: r}

			cmd, err := extension.cmd("list")
			if err != nil {
				t.Fatal(err)
			}

			env := map[string]string{}

			for _, entry := range cmd.Environ() {
				key, value, _ := strings.Cut(entry, "=")
				env[key] = value
			}

			wantProfile, wantEmpty := "parent", "parent"
			if name == "aws" {
				wantProfile, wantEmpty = "private profile", ""

				if env["REDACT_TEST_QUOTED"] != "value # with = signs" {
					t.Errorf("quoted value = %q", env["REDACT_TEST_QUOTED"])
				}
			}

			if env["AWS_PROFILE"] != wantProfile || env["REDACT_TEST_EMPTY"] != wantEmpty {
				t.Errorf("profile = %q, empty = %q", env["AWS_PROFILE"], env["REDACT_TEST_EMPTY"])
			}

			if env["REDACT_TEST_INHERITED"] != "inherited" {
				t.Error("inherited variable was lost")
			}
		})
	}

	if os.Getenv("AWS_PROFILE") != "parent" || os.Getenv("REDACT_TEST_EMPTY") != "parent" {
		t.Error("private configuration changed the parent environment")
	}
}

func TestPrivateEnvErrors(t *testing.T) {
	r := privateEnvRepo(t)
	extension := &Ext{name: "aws", repo: r}
	filename := writePrivateEnv(t, r, "SECRET='unterminated-private-value\n")

	_, err := extension.cmd("list")
	if !errors.Is(err, ErrInvalidEnvFile) {
		t.Fatalf("expected invalid dotenv error, got %v", err)
	}

	if strings.Contains(err.Error(), "unterminated-private-value") {
		t.Fatal("parse error exposed a private value")
	}

	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}

	// A directory is an unreadable configuration file, even when running as root.
	if err := os.Mkdir(filename, 0700); err != nil {
		t.Fatal(err)
	}

	if _, err := extension.cmd("list"); err == nil {
		t.Fatal("expected an error reading a directory")
	}

	extension.name = "../aws"
	if _, err := extension.cmd("list"); err == nil {
		t.Fatal("expected extension name traversal to be rejected")
	}
}

func TestPrivateEnvLinkedWorktree(t *testing.T) {
	r := privateEnvRepo(t)
	writePrivateEnv(t, r, "AWS_PROFILE=shared-private\n")
	runGit(t, r.Workdir.Root(), "-c", "user.name=test", "-c", "user.email=test@example.com",
		"commit", "-q", "--allow-empty", "-m", "init")
	worktree := filepath.Join(t.TempDir(), "worktree")
	runGit(t, r.Workdir.Root(), "worktree", "add", "-q", worktree)
	t.Chdir(worktree)

	linked := &repo.Repo{}
	if err := linked.SetupRepo(); err != nil {
		t.Fatal(err)
	}

	extension := &Ext{name: "aws", repo: linked}

	cmd, err := extension.cmd("get")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(cmd.Environ(), "AWS_PROFILE=shared-private") {
		t.Fatal("linked worktree did not load the common private environment")
	}
}

func TestPrivateEnvCommands(t *testing.T) {
	r := privateEnvRepo(t)
	writePrivateEnv(t, r, "AWS_PROFILE=private\n")
	output := filepath.Join(t.TempDir(), "output")
	t.Setenv("REDACT_TEST_OUTPUT", output)

	script := filepath.Join(t.TempDir(), "extension")

	contents := "#!/bin/sh\nprintf '%s:%s' \"$1\" \"$AWS_PROFILE\" > \"$REDACT_TEST_OUTPUT\"\n" +
		"case \"$1\" in\nget) printf key;;\nput) cat >> \"$REDACT_TEST_OUTPUT\";;\nesac\n"
	if err := os.WriteFile(script, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(script, 0700); err != nil {
		t.Fatal(err)
	}

	extension := &Ext{name: "aws", repo: r, Command: script}

	for _, command := range []string{"list", "get", "put"} {
		t.Run(command, func(t *testing.T) {
			switch command {
			case "list":
				if err := extension.List(); err != nil {
					t.Fatal(err)
				}
			case "get":
				key, err := extension.LoadKey()
				if err != nil || string(key) != "key" {
					t.Fatalf("LoadKey() = %q, %v", key, err)
				}
			case "put":
				if err := extension.SaveKey([]byte("key")); err != nil {
					t.Fatal(err)
				}
			}

			want := command + ":private"
			if command == "put" {
				want += "key"
			}

			got, err := os.ReadFile(output)
			if err != nil || string(got) != want {
				t.Fatalf("extension output = %q, %v; want %q", got, err, want)
			}
		})
	}
}
