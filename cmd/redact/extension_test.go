package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
)

func extensionRepo(t *testing.T) (string, string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("test extension uses a POSIX shell")
	}

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("FF_GIT_EXT", "1")
	t.Setenv("REDACT_UNLOCK_KEY", "")
	t.Setenv("REDACT_UNLOCK_EXPORTED_KEY", "")
	t.Setenv("REDACT_UNLOCK_GPG_KEY", "")

	root := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", output, err)
	}

	t.Chdir(root)
	t.Setenv("TEST_EXTENSION_CALLS", filepath.Join(root, "calls"))
	t.Setenv("TEST_EXTENSION_KEY", filepath.Join(root, "key.pem"))
	program := filepath.Join(t.TempDir(), "redact-ext-fixture")
	script := `#!/bin/sh
operation=$1
shift
code=0
name=fixture
for setting in "$@"; do
 case "$setting" in
  code=*) code=${setting#code=} ;;
  name=*) name=${setting#name=} ;;
 esac
done
printf '%s:%s %s\n' "$name" "$operation" "$*" >> "$TEST_EXTENSION_CALLS"
if [ "$code" != 0 ]; then
 echo 'choose a key with --option identity=NAME and retry' >&2
 exit "$code"
fi
case "$operation" in
 get) cat "$TEST_EXTENSION_KEY" ;;
 put|update|grant) cat > "$TEST_EXTENSION_KEY" ;;
 revoke) cat > revoke.stdin ;;
esac
`
	// Executable fixture contains no secret material.
	if err := os.WriteFile(program, []byte(script), 0700); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	return root, program
}

func extensionRuntime() *Runtime {
	logger.SetOutput(io.Discard)

	return &Runtime{Repo: &repo.Repo{}}
}

func runRedact(args ...string) error {
	app := extensionRuntime().app()
	app.Writer, app.ErrWriter = io.Discard, io.Discard

	return app.Run(context.Background(), append([]string{"redact"}, args...))
}

func setExtensionConfig(t *testing.T, program string, codes map[string]int, order []string) *ext.Config {
	t.Helper()

	rt := extensionRuntime()
	if err := rt.SetupRepo(); err != nil {
		t.Fatal(err)
	}

	conf, err := ext.Load(rt.Repo)
	if err != nil {
		t.Fatal(err)
	}

	conf.Exts = map[string]ext.Ext{}
	conf.Order = order

	for name, code := range codes {
		item := ext.Ext{Command: program, Config: map[string]string{"code": strconv.Itoa(code), "name": name}}
		if err := conf.AddExt(name, item); err != nil {
			t.Fatal(err)
		}
	}

	if err := conf.Save(); err != nil {
		t.Fatal(err)
	}

	return conf
}

func fixtureCalls(t *testing.T, root string) []string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil {
		t.Fatal(err)
	}

	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestExtensionScenarios(t *testing.T) {
	root, program := extensionRepo(t)
	setExtensionConfig(t, program, map[string]int{"aws": 0, "azure": 0}, []string{"azure", "aws"})

	if err := runRedact("init"); err != nil {
		t.Fatal(err)
	}

	keyfile := filepath.Join(root, ".git", "redact", "key")

	key, err := os.ReadFile(keyfile)
	if err != nil {
		t.Fatal(err)
	}

	exported, err := os.ReadFile(filepath.Join(root, "key.pem"))
	if err != nil {
		t.Fatal(err)
	}

	initCalls, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil {
		t.Fatal(err)
	}

	for _, scenario := range []struct {
		name  string
		run   func(*testing.T, string, string)
		calls []byte
	}{
		{"CollectionDispatchAndConfiguration", checkCollectionDispatchAndConfiguration, initCalls},
		{"UnlockExtensionPolicy", checkUnlockExtensionPolicy, nil},
		{"ForcedUnconfiguredUnlockAndStderr", checkForcedUnconfiguredUnlockAndStderr, nil},
		{"ExplicitUnlockNeverFallsBack", checkExplicitUnlockNeverFallsBack, nil},
		{"InvalidGetDataStopsUnlock", checkInvalidGetDataStopsUnlock, nil},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			// Restore an unlocked repository without repeating Git or redact init.
			t.Chdir(root)

			for filename, data := range map[string][]byte{
				keyfile:                        key,
				filepath.Join(root, "key.pem"): exported,
				filepath.Join(root, "calls"):   scenario.calls,
			} {
				if err := os.WriteFile(filename, data, 0600); err != nil {
					t.Fatal(err)
				}
			}

			if err := os.Remove(filepath.Join(root, ".redact", "config.json")); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}

			scenario.run(t, root, program)
		})
	}
}

func checkUnlockExtensionPolicy(t *testing.T, root, program string) {
	conf := setExtensionConfig(t, program, map[string]int{"first": 0, "second": 0}, []string{"first", "second"})

	for _, code := range []int{0, 1, 2, 3, 77} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			conf.Exts["first"].Config["code"] = strconv.Itoa(code)
			if err := conf.Save(); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(root, "calls"), nil, 0600); err != nil {
				t.Fatal(err)
			}

			err := runRedact("unlock")

			want := ext.Success
			if code >= 2 {
				want = min(code, ext.Failure)
			}

			if ext.ExitCode(err) != want {
				t.Fatalf("exit %d: err=%v", code, err)
			}

			calls := fixtureCalls(t, root)

			count := 1
			if code == 1 {
				count = 2
			}

			if len(calls) != count || !strings.HasPrefix(calls[0], "first:get") {
				t.Fatalf("unlock calls = %v", calls)
			}
		})
	}
}

func checkForcedUnconfiguredUnlockAndStderr(t *testing.T, root, program string) {
	if err := runRedact("key", "save", "--ext", "fixture", "--cmd", program); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(root, ".git", "redact", "key")); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", filepath.Dir(program)+string(os.PathListSeparator)+os.Getenv("PATH"))

	subdir := filepath.Join(root, "nested")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Chdir(subdir)

	if err := runRedact("unlock", "--extension", "fixture", "--option", "identity=chosen"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, ".redact", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("forced invocation created config: %v", err)
	}

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()

	previous := os.Stderr
	os.Stderr = stderr

	t.Cleanup(func() { os.Stderr = previous })

	err = runRedact("unlock", "--ext", "fixture", "--option", "code=2")
	if ext.ExitCode(err) != ext.MissingData {
		t.Fatalf("missing data lost: %v", err)
	}

	data, readErr := os.ReadFile(stderr.Name())
	if readErr != nil || !strings.Contains(string(data), "--option identity=NAME") {
		t.Fatalf("actionable stderr lost: %q, %v", data, readErr)
	}
}

func checkCollectionDispatchAndConfiguration(t *testing.T, root, program string) {
	setExtensionConfig(t, program, map[string]int{"aws": 0, "azure": 0}, []string{"azure", "aws"})

	for _, args := range [][]string{
		{"key", "save"},
		{"key", "generate"},
		{"grant", "--option", "recipient=alice"},
		{"revoke", "--ext", "aws", "--option", "recipient=alice"},
		{"ext", "update", "--name", "aws", "--option", "default=kept"},
	} {
		if err := runRedact(args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}

	calls := fixtureCalls(t, root)

	operations := make([]string, 0, len(calls))
	for _, call := range calls {
		operations = append(operations, strings.Fields(call)[0])
	}

	want := []string{
		"azure:put", "aws:put", "azure:put", "aws:put", "azure:update", "aws:update",
		"azure:grant", "aws:grant", "aws:revoke", "aws:list",
	}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("dispatch = %v, want %v", operations, want)
	}

	rt := extensionRuntime()
	if _, err := rt.LoadSecretKey(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if len(rt.Keys) != 2 {
		t.Fatalf("rotation lost retained epochs: %d", len(rt.Keys))
	}

	data, err := os.ReadFile(filepath.Join(root, "revoke.stdin"))
	if err != nil || len(data) != 0 {
		t.Fatalf("revoke received secret data: %q, %v", data, err)
	}

	data, err = os.ReadFile(filepath.Join(root, ".redact", "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if string(raw["_order_"]) != `["azure","aws"]` || strings.Contains(string(data), "recipient") {
		t.Fatalf("config metadata/defaults were changed by invocation: %s", data)
	}
	// Cloud writes and exchange changes are left for inspection; no Git commit.
	output, err := exec.Command("git", "rev-list", "--all", "--count").CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "0" {
		t.Fatalf("unexpected Git publication: %q, %v", output, err)
	}
}

func checkExplicitUnlockNeverFallsBack(t *testing.T, root, program string) {
	setExtensionConfig(t, program, map[string]int{"first": 1, "second": 0}, []string{"first", "second"})

	if err := runRedact("unlock", "--ext", "first"); ext.ExitCode(err) != ext.NotApplicable {
		t.Fatalf("explicit outcome = %v", err)
	}

	if got := fixtureCalls(t, root); len(got) != 1 || !strings.HasPrefix(got[0], "first:get") {
		t.Fatalf("explicit unlock fell back: %v", got)
	}

	setExtensionConfig(t, program, map[string]int{"first": 1, "second": 1}, []string{"first", "second"})

	if err := runRedact("unlock"); !errors.Is(err, ErrNoSuitableKey) {
		t.Fatalf("all not applicable = %v", err)
	}
}

func checkInvalidGetDataStopsUnlock(t *testing.T, root, program string) {
	setExtensionConfig(t, program, map[string]int{"first": 0, "second": 0}, []string{"first", "second"})

	if err := os.WriteFile(filepath.Join(root, "key.pem"), []byte("invalid Key Exchange data"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := runRedact("unlock"); ext.ExitCode(err) != ext.Failure {
		t.Fatalf("invalid successful output was accepted: %v", err)
	}

	if got := fixtureCalls(t, root); len(got) != 1 {
		t.Fatalf("invalid successful output caused fallback: %v", got)
	}
}
