package ext

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		Input: []string{program, "source", "success", literal},
	}

	got, err := extension.LoadKey()
	if err != nil || string(got) != "private|"+literal+"|private|"+r.Workdir.Root() {
		t.Fatalf("pipeline output = %q, err=%v", got, err)
	}

	for _, code := range []string{"1", "2", "77"} {
		t.Run("input exit "+code, func(t *testing.T) {
			extension.Input = []string{program, "source", "fail", code}

			got, err := extension.LoadKey()
			if ExitCode(err) != Failure || len(got) != 0 {
				t.Fatalf("failed producer: output=%q code=%d", got, ExitCode(err))
			}
		})
	}

	for _, command := range [][]string{{""}, {filepath.Join(t.TempDir(), "missing")}} {
		extension.Input = command

		if got, err := extension.LoadKey(); err == nil || len(got) != 0 {
			t.Fatalf("invalid input command: output=%q err=%v", got, err)
		}
	}

	extension.Input = []string{program, "source", "block"}
	extension.Config = map[string]string{"mode": "fail"}

	if got, err := extension.LoadKey(); ExitCode(err) != MissingData || len(got) != 0 {
		t.Fatalf("early target failure: output=%q code=%d", got, ExitCode(err))
	}

	extension.Config = nil
	extension.Input = []string{filepath.Join(t.TempDir(), "must-not-run")}

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
	if err := conf.AddExt("gpg", Ext{Input: input, Config: map[string]string{"private-key": "stdin"}}); err != nil {
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
	if err != nil || strings.Join(resolved.Input, "\n") != strings.Join(input, "\n") {
		t.Fatalf("input was lost during configuration round trip: %v", err)
	}
}
