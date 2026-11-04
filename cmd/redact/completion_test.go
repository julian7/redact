package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julian7/redact/repo"
)

func TestXDGDir(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		home    string
		envVar  string
		sub     []string
		want    string
		wantErr string
	}{
		{
			name:   "the environment variable wins",
			env:    map[string]string{"XDG_DATA_HOME": "/xdg/data"},
			home:   "/home/user",
			envVar: "XDG_DATA_HOME",
			sub:    []string{".local", "share"},
			want:   "/xdg/data",
		},
		{
			name:   "the home directory holds the default",
			home:   "/home/user",
			envVar: "XDG_DATA_HOME",
			sub:    []string{".local", "share"},
			want:   "/home/user/.local/share",
		},
		{
			name:    "a relative path is rejected",
			env:     map[string]string{"XDG_DATA_HOME": "not/absolute"},
			home:    "/home/user",
			envVar:  "XDG_DATA_HOME",
			wantErr: `XDG_DATA_HOME is not an absolute path: "not/absolute"`,
		},
		{
			name:    "the home directory is needed for the default",
			envVar:  "XDG_CONFIG_HOME",
			wantErr: "neither XDG_CONFIG_HOME nor a home directory is set",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := func(name string) string { return test.env[name] }

			got, err := xdgDir(env, test.home, test.envVar, test.sub...)
			switch {
			case test.wantErr != "":
				if err == nil {
					t.Fatalf("xdgDir() error is nil, want %q", test.wantErr)
				}

				if err.Error() != test.wantErr {
					t.Errorf("xdgDir() error = %q, want %q", err, test.wantErr)
				}
			case err != nil:
				t.Fatalf("xdgDir() = %v", err)
			case got != test.want:
				t.Errorf("xdgDir() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCompletionDir(t *testing.T) {
	xdg := t.TempDir()
	dataHome := filepath.Join(xdg, "share")
	configHome := filepath.Join(xdg, "config")

	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	tests := map[string]string{
		"bash": filepath.Join(dataHome, "bash-completion", "completions"),
		"zsh":  filepath.Join(dataHome, "zsh", "site-functions"),
		"fish": filepath.Join(configHome, "fish", "completions"),
	}

	for shell, want := range tests {
		t.Run(shell, func(t *testing.T) {
			dir, err := completionTargets[shell].dir()
			if err != nil {
				t.Fatalf("%s completion dir: %v", shell, err)
			}

			if dir != want {
				t.Errorf("%s completion dir = %q, want %q", shell, dir, want)
			}
		})
	}
}

func TestCompletionDirDefaultsToHome(t *testing.T) {
	home := t.TempDir()

	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	tests := map[string]string{
		"bash": filepath.Join(home, ".local", "share", "bash-completion", "completions"),
		"zsh":  filepath.Join(home, ".local", "share", "zsh", "site-functions"),
		"fish": filepath.Join(home, ".config", "fish", "completions"),
	}

	for shell, want := range tests {
		t.Run(shell, func(t *testing.T) {
			dir, err := completionTargets[shell].dir()
			if err != nil {
				t.Fatalf("%s completion dir: %v", shell, err)
			}

			if dir != want {
				t.Errorf("%s completion dir = %q, want %q", shell, dir, want)
			}
		})
	}
}

func TestCompletionOutput(t *testing.T) {
	xdg := t.TempDir()

	t.Setenv("XDG_DATA_HOME", filepath.Join(xdg, "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(xdg, "config"))

	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			out := &bytes.Buffer{}
			runCompletion(t, out, "completion", shell)

			if out.Len() == 0 {
				t.Error("no completion script on the output")
			}

			entries, err := os.ReadDir(xdg)
			if err != nil {
				t.Fatalf("reading completion directories: %v", err)
			}

			if len(entries) != 0 {
				t.Errorf("completion was installed without --install: %v", entries)
			}
		})
	}
}

func TestCompletionInstall(t *testing.T) {
	xdg := t.TempDir()
	dataHome := filepath.Join(xdg, "share")
	configHome := filepath.Join(xdg, "config")

	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	tests := []struct {
		shell     string
		file      string
		signature string
	}{
		{shell: "bash", file: "redact", signature: "__redact_bash_autocomplete"},
		{shell: "zsh", file: "_redact", signature: "#compdef redact"},
		{shell: "fish", file: "redact.fish", signature: "__redact_perform_completion"},
	}

	for _, test := range tests {
		t.Run(test.shell, func(t *testing.T) {
			dir, err := completionTargets[test.shell].dir()
			if err != nil {
				t.Fatalf("%s completion dir: %v", test.shell, err)
			}

			out := &bytes.Buffer{}
			runCompletion(t, out, "completion", test.shell, "--install")

			script := readCompletion(t, filepath.Join(dir, test.file))
			if !strings.Contains(script, test.signature) {
				t.Errorf("installed %s completion does not contain %q:\n%s", test.shell, test.signature, script)
			}

			report := out.String()
			if !strings.Contains(report, "Installed "+test.shell+" completion in "+filepath.Join(dir, test.file)) {
				t.Errorf("installation of %s completion is not reported:\n%s", test.shell, report)
			}

			if strings.Contains(report, test.signature) {
				t.Errorf("the completion script is printed on the output as well:\n%s", report)
			}
		})
	}
}

func TestCompletionInstallHint(t *testing.T) {
	xdg := t.TempDir()

	t.Setenv("XDG_DATA_HOME", filepath.Join(xdg, "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(xdg, "config"))

	out := &bytes.Buffer{}
	runCompletion(t, out, "completion", "zsh", "--install")

	report := out.String()
	if !strings.Contains(report, "fpath=(") {
		t.Errorf("zsh installation doesn't tell about $fpath:\n%s", report)
	}

	out.Reset()

	runCompletion(t, out, "completion", "fish", "--install")

	if strings.Contains(out.String(), "fpath") {
		t.Errorf("fish installation mentions zsh configuration:\n%s", out)
	}
}

// runCompletion runs the application with the given arguments, capturing its
// output.
func runCompletion(t *testing.T, out io.Writer, args ...string) {
	t.Helper()

	rt := &Runtime{Repo: &repo.Repo{}}

	app := rt.app()
	app.Writer = out

	if err := app.Run(context.Background(), append([]string{"redact"}, args...)); err != nil {
		t.Fatalf("redact %v: %v", args, err)
	}
}

func readCompletion(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("reading installed completion: %v", err)
	}

	return string(content)
}
