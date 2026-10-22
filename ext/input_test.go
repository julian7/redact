package ext

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// The test executable doubles as both pipeline stages on every supported OS.
func TestMain(m *testing.M) {
	if os.Getenv("REDACT_INPUT_HELPER") == "1" {
		os.Exit(inputHelper())
	}

	os.Exit(m.Run())
}

func inputHelper() int {
	if os.Args[1] == "source" {
		switch os.Args[2] {
		case "fail":
			fmt.Print("untrusted partial output")

			code, _ := strconv.Atoi(os.Args[3])

			return code
		case "pass":
			_, err := io.Copy(os.Stdout, os.Stdin)
			if err != nil {
				return Failure
			}

			return Success
		case "block":
			for {
				if _, err := os.Stdout.Write(make([]byte, 8192)); err != nil {
					return 1
				}
			}
		default:
			fmt.Printf("%s|%s", os.Getenv("AWS_PROFILE"), os.Args[3])

			return 0
		}
	}

	if len(os.Args) > 2 && os.Args[2] == "mode=fail" {
		return MissingData
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Failure
	}

	if os.Args[1] != "get" {
		fmt.Print(string(data))

		return 0
	}

	cwd, _ := os.Getwd()
	fmt.Printf("%s|%s|%s", data, os.Getenv("AWS_PROFILE"), cwd)

	return 0
}

func TestInputCommand(t *testing.T) {
	r := privateEnvRepo(t)
	writePrivateEnv(t, r, "AWS_PROFILE=private\n")
	t.Setenv("REDACT_INPUT_HELPER", "1")
	t.Setenv("AWS_PROFILE", "parent")

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	const literal = "$HOME ' ; | literal=value"

	extension := &Ext{
		name: "aws", repo: r, Command: program,
		Get: &Operation{Filter: []string{program, "source", "success", literal}},
	}

	got, err := extension.LoadKey()
	if err != nil || string(got) != "private|"+literal+"|private|"+r.Workdir.Root() {
		t.Fatalf("pipeline output = %q, err=%v", got, err)
	}

	for _, code := range []string{"1", "2", "77"} {
		t.Run("input exit "+code, func(t *testing.T) {
			extension.Get.Filter = []string{program, "source", "fail", code}

			got, err := extension.LoadKey()
			if ExitCode(err) != Failure || len(got) != 0 {
				t.Fatalf("failed producer: output=%q code=%d", got, ExitCode(err))
			}
		})
	}

	for _, command := range [][]string{{""}, {filepath.Join(t.TempDir(), "missing")}} {
		extension.Get.Filter = command

		if got, err := extension.LoadKey(); err == nil || len(got) != 0 {
			t.Fatalf("invalid input command: output=%q err=%v", got, err)
		}
	}

	extension.Get.Filter = []string{program, "source", "block"}
	extension.Get.Config = map[string]string{"mode": "fail"}

	if got, err := extension.LoadKey(); ExitCode(err) != MissingData || len(got) != 0 {
		t.Fatalf("early target failure: output=%q code=%d", got, ExitCode(err))
	}

	extension.Get.Filter = []string{filepath.Join(t.TempDir(), "must-not-run")}

	for _, operation := range []string{"list", "put", "update", "grant", "revoke"} {
		if err := extension.Run(operation, nil); err != nil {
			t.Fatalf("%s ran the input command: %v", operation, err)
		}
	}

	if os.Getenv("AWS_PROFILE") != "parent" {
		t.Fatal("pipeline changed the parent environment")
	}
}

func TestInputConfigurationRoundTrip(t *testing.T) {
	r := privateEnvRepo(t)

	conf, err := Load(r)
	if err != nil {
		t.Fatal(err)
	}

	input := []string{"provider", "get", "key=op://vault/item/field"}

	item := Ext{Get: &Operation{Filter: input, Config: map[string]string{"private-key": "stdin"}}}
	if err := conf.AddExt("gpg", item); err != nil {
		t.Fatal(err)
	}

	if err := conf.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(r)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := loaded.Resolve("gpg", "", nil)
	if err != nil || resolved.Get == nil || strings.Join(resolved.Get.Filter, "\n") != strings.Join(input, "\n") {
		t.Fatalf("input was lost during configuration round trip: %v", err)
	}

	if resolved.Get.Config["private-key"] != "stdin" || len(resolved.Config) != 0 {
		t.Fatalf("get settings escaped their operation: %+v", resolved)
	}
}

func TestOperationSettings(t *testing.T) {
	r := privateEnvRepo(t)

	conf, err := Load(r)
	if err != nil {
		t.Fatal(err)
	}

	settings := func(name string) *Operation {
		return &Operation{Config: map[string]string{"identity": name, "only-" + name: "yes"}}
	}

	original := Ext{
		Command: "fixture",
		Config:  map[string]string{"identity": "shared", "shared": "yes"},
		Get:     settings("get"),
		Put:     settings("put"),
		Update:  settings("update"),
		Grant:   settings("grant"),
		Revoke:  settings("revoke"),
		Listing: settings("list"),
	}
	if err := conf.AddExt("fixture", original); err != nil {
		t.Fatal(err)
	}

	if err := conf.Save(); err != nil {
		t.Fatal(err)
	}

	conf, err = Load(r)
	if err != nil {
		t.Fatal(err)
	}

	for _, override := range []bool{false, true} {
		options := map[string]string{}
		if override {
			options["identity"] = "invocation"
		}

		item, err := conf.Resolve("fixture", "", options)
		if err != nil {
			t.Fatal(err)
		}

		for _, name := range []string{"get", "put", "update", "grant", "revoke", "list"} {
			command, err := item.cmd(name)
			if err != nil {
				t.Fatal(err)
			}

			identity := name
			if override {
				identity = "invocation"
			}

			want := []string{"fixture", name, "identity=" + identity, "only-" + name + "=yes", "shared=yes"}
			if !reflect.DeepEqual(command.Args, want) {
				t.Fatalf("%s override=%t: args=%v, want %v", name, override, command.Args, want)
			}
		}
	}

	stored, _ := conf.Ext("fixture")
	if stored.Config["identity"] != "shared" || stored.Get.Config["identity"] != "get" {
		t.Fatal("invocation modified configured settings")
	}
}

func TestOperationFilters(t *testing.T) {
	r := privateEnvRepo(t)
	t.Setenv("REDACT_INPUT_HELPER", "1")

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	settings := &Operation{Filter: []string{program, "source", "pass"}}
	item := Ext{
		name: "fixture", repo: r, Command: program,
		Get: settings, Put: settings, Update: settings,
		Grant: settings, Revoke: settings, Listing: settings,
	}

	for _, name := range []string{"put", "update", "grant", "revoke", "list"} {
		command, err := item.cmd(name)
		if err != nil {
			t.Fatal(err)
		}

		var output bytes.Buffer

		command.Stdin = strings.NewReader("operation input")
		command.Stdout = &output

		if err := item.run(name, command); err != nil || output.String() != "operation input" {
			t.Fatalf("%s filter: output=%q err=%v", name, output.String(), err)
		}
	}
}
