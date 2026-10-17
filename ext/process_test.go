package ext_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/repo"
)

func processConfig(t *testing.T) (*ext.Config, string, string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("test extension uses a POSIX shell")
	}

	root := t.TempDir()
	program := filepath.Join(t.TempDir(), "redact-ext-test")
	script := `#!/bin/sh
operation=$1
shift
code=0
store=key.pem
for setting in "$@"; do
 case "$setting" in
  code=*) code=${setting#code=} ;;
  store=*) store=${setting#store=} ;;
 esac
done
printf '%s\n' "$operation $*" >> calls
pwd -P > cwd
if [ "$code" != 0 ]; then
 echo 'select an identity with --option identity=NAME and retry' >&2
 exit "$code"
fi
case "$operation" in
 get) cat "$store" ;;
 put|grant|update) cat > "$store" ;;
 revoke) cat > revoke.stdin ;;
esac
`
	// Executable fixture contains no secret material.
	if err := os.WriteFile(program, []byte(script), 0700); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	conf, err := ext.Load(&repo.Repo{Workdir: osfs.New(root)})
	if err != nil {
		t.Fatal(err)
	}

	return conf, root, program
}

func TestProcessOutcomes(t *testing.T) {
	conf, _, program := processConfig(t)
	for _, code := range []int{0, 1, 2, 3, 44} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			item, err := conf.Resolve("test", program, map[string]string{"code": strconv.Itoa(code)})
			if err != nil {
				t.Fatal(err)
			}

			want := min(code, ext.Failure)
			if got := ext.ExitCode(item.Run("update", []byte("collection"))); got != want {
				t.Fatalf("exit %d: got %d, want %d", code, got, want)
			}
		})
	}

	item, err := conf.Resolve("absent", filepath.Join(t.TempDir(), "missing"), nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := ext.ExitCode(item.SaveKey(nil)); got != ext.Failure {
		t.Fatalf("launch failure code = %d", got)
	}
}

func TestPEMRoundTripAndInvocationSettings(t *testing.T) {
	conf, root, program := processConfig(t)

	defaultCommand := filepath.Join(t.TempDir(), "default-command")

	itemConfig := ext.Ext{
		Command: defaultCommand,
		Config:  map[string]string{"store": "default.pem", "identity": "default"},
	}
	if err := conf.AddExt("test", itemConfig); err != nil {
		t.Fatal(err)
	}

	item, err := conf.Resolve("test", program, map[string]string{"store": "invocation.pem", "identity": "chosen=value"})
	if err != nil {
		t.Fatal(err)
	}

	key, err := files.NewSecretKey(osfs.New(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}

	if err := key.Generate(); err != nil {
		t.Fatal(err)
	}

	var exported bytes.Buffer
	if err := key.Export(&exported); err != nil {
		t.Fatal(err)
	}

	if err := item.SaveKey(exported.Bytes()); err != nil {
		t.Fatal(err)
	}

	got, err := item.LoadKey()
	if err != nil || !bytes.Equal(got, exported.Bytes()) {
		t.Fatalf("PEM round trip: equal=%v, err=%v", bytes.Equal(got, exported.Bytes()), err)
	}

	if err := conf.Save(); err != nil {
		t.Fatal(err)
	}

	saved := readExtsFile(t, &repo.Repo{Workdir: osfs.New(root)}, filepath.Join(".redact", "config.json"))
	if saved["test"].Command != defaultCommand ||
		saved["test"].Config["store"] != "default.pem" || saved["test"].Config["identity"] != "default" {
		t.Fatalf("invocation changed defaults: %+v", saved)
	}

	cwd, err := os.ReadFile(filepath.Join(root, "cwd"))
	if err != nil {
		t.Fatal(err)
	}

	// Compare physical paths because macOS temporary directories may use symlinks.
	wantDir, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(cwd)) != wantDir {
		t.Fatalf("extension cwd = %q, want directory %q", cwd, wantDir)
	}

	calls, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil || !strings.Contains(string(calls), "identity=chosen=value") {
		t.Fatalf("invocation settings missing: %q, %v", calls, err)
	}
}

func TestOrderedConfigurationRoundTrip(t *testing.T) {
	conf, root, program := processConfig(t)

	conf.Order = []string{"gpg", "aws", "gpg", "absent"}
	for _, name := range []string{"azure", "gpg", "aws"} {
		if err := conf.AddExt(name, ext.Ext{Command: program}); err != nil {
			t.Fatal(err)
		}
	}

	if err := conf.Save(); err != nil {
		t.Fatal(err)
	}
	// Exercise the actual JSON parser with a new Config instance.
	loaded, err := ext.Load(&repo.Repo{Workdir: osfs.New(root)})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(loaded.Order, conf.Order) || len(loaded.Exts) != 3 {
		t.Fatalf("round trip lost metadata: %+v", loaded)
	}

	conf = loaded
	if !reflect.DeepEqual(conf.Names(), []string{"gpg", "aws", "azure"}) {
		t.Fatalf("order = %v", conf.Names())
	}

	if err := conf.AddExt("_order_", ext.Ext{}); !errors.Is(err, ext.ErrReservedName) {
		t.Fatalf("reserved name accepted: %v", err)
	}
}

func TestCollectionBatchContinuesAfterFailures(t *testing.T) {
	conf, root, program := processConfig(t)

	conf.Order = []string{"failed", "skipped", "saved"}
	for name, code := range map[string]string{"failed": "3", "skipped": "1", "saved": "0"} {
		if err := conf.AddExt(name, ext.Ext{Command: program, Config: map[string]string{"code": code}}); err != nil {
			t.Fatal(err)
		}
	}

	if got := ext.ExitCode(conf.SaveKey([]byte("retained collection"))); got != ext.Failure {
		t.Fatalf("batch code = %d", got)
	}

	calls, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil || string(calls) != "put code=3\nput code=1\nput code=0\n" {
		t.Fatalf("destinations = %q, err=%v", calls, err)
	}

	data, err := os.ReadFile(filepath.Join(root, "key.pem"))
	if err != nil || string(data) != "retained collection" {
		t.Fatalf("successful result lost: %q, %v", data, err)
	}

	conf.DelExt("failed")

	if err := conf.SaveKey(data); err != nil {
		t.Fatalf("unsupported destination failed mixed batch: %v", err)
	}

	conf.DelExt("saved")

	if got := ext.ExitCode(conf.SaveKey(data)); got != ext.NotApplicable {
		t.Fatalf("all unsupported code = %d", got)
	}
}
