//+build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type buildTarget struct {
	os   string
	arch string
	ext  string
	out  string
}

var (
	hasUPX        = false
	packageName   = "redact"
	targetDirName = "target"
	buildParam    = "./cmd/redact"
	ldflags       = `-s -w -X main.version=$VERSION`
	targets       = map[string]*buildTarget{
		"macos": &buildTarget{
			os:   "darwin",
			arch: "amd64",
		},
		"linux": &buildTarget{
			os:   "linux",
			arch: "amd64",
		},
		"windows": &buildTarget{
			os:   "windows",
			arch: "amd64",
			ext:  ".exe",
		},
	}
	versionTag string
)

func withEnv(osType, arch string) map[string]string {
	return map[string]string{
		"PACKAGENAME": packageName,
		"GOOS":        osType,
		"GOARCH":      arch,
		"VERSION":     versionTag,
	}
}

func (tar *buildTarget) SetOutput() {
	outFile := fmt.Sprintf("$PACKAGENAME-%s-%s-$VERSION%s", tar.os, tar.arch, tar.ext)
	tar.out = filepath.Join(targetDirName, outFile)
}

func build(target string) error {
	tar, ok := targets[target]
	if !ok {
		tar = nil
		for _, target := range targets {
			if target.os == runtime.GOOS {
				tar = target
				break
			}
		}
	}
	if tar == nil {
		return errors.New("Could not find appropriate build target for your OS.")
	}
	if len(tar.out) == 0 {
		tar.out = fmt.Sprintf("%s%s", packageName, tar.ext)
	}
	steps := []func(*buildTarget) error{(*buildTarget).compile, (*buildTarget).upx}
	for _, step := range steps {
		if err := step(tar); err != nil {
			return err
		}
	}
	env := withEnv(tar.os, tar.arch)
	out := os.Expand(tar.out, func(s string) string {
		s2, ok := env[s]
		if ok {
			return s2
		}
		return os.Getenv(s)
	})
	fmt.Printf("Built %s for %s %s\n", out, tar.os, tar.arch)
	return nil
}

func (tar *buildTarget) compile() error {
	if len(tar.out) == 0 {
		tar.out = packageName
	}
	return sh.RunWith(
		withEnv(tar.os, tar.arch),
		mg.GoCmd(),
		"build",
		"-o",
		tar.out,
		"-ldflags",
		ldflags,
		buildParam,
	)
}

func (tar *buildTarget) upx() error {
	if hasUPX {
		return sh.RunWith(
			withEnv(tar.os, tar.arch),
			"upx",
			tar.out,
		)
	}
	return nil
}

// Target creates target subdirectory
func Target() error {
	st, err := os.Stat(targetDirName)
	if err == nil && st.IsDir() {
		return nil
	}
	return os.Mkdir(targetDirName, 0755)
}

// All builds for all possible targets, put into target subfolder
func All() {
	step("all")
	mg.Deps(Target, Environment)
	for name, target := range targets {
		step(name)
		target.SetOutput()
		build(name)
	}
}

func Build() error {
	step("build")
	mg.Deps(Environment)
	return build("")
}

// Darwin builds for MacOS
func Darwin() error {
	step("darwin")
	mg.Deps(Environment)
	return build("macos")
}

// Linux builds for linux
func Linux() error {
	step("linux")
	mg.Deps(Environment)
	return build("linux")
}

// Windows builds for windows
func Windows() error {
	step("windows")
	mg.Deps(Environment)
	return build("windows")
}

// Environment sets up environment (like calling version and checkUPX)
func Environment() {
	mg.Deps(CheckUPX, Version)
}

// CheckUPX checks whether upx is available
func CheckUPX() {
	_, err := exec.LookPath("upx")
	if err == nil {
		hasUPX = true
	}
}

// Version calculates app version by git describe output
func Version() error {
	var err error
	versionTag, err = sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return err
	}
	return nil
}

// AllTests runs go fmt, go vet, test, and cover
func AllTests() {
	mg.SerialDeps(Lint, Test, Cover)
}

// Test runs go test
func Test() error {
	step("test")
	return sh.RunV(mg.GoCmd(), "test", "./...")
}

// Lint runs lint tool
func Lint() error {
	step("lint")
	if _, err := exec.LookPath("golangci-lint"); err == nil {
		return sh.RunV("golangci-lint", "run", "-v", "./...")
	}
	if _, err := exec.LookPath("golint"); err == nil {
		return sh.RunV("golint", "./...")
	}
	if err := sh.RunV(mg.GoCmd(), "fmt", "./..."); err != nil {
		return err
	}
	if err := sh.RunV(mg.GoCmd(), "vet", "./..."); err != nil {
		return err
	}
	return nil
}

// Cover runs coverity profile
func Cover() error {
	step("cover")
	err := sh.RunV(mg.GoCmd(), "test", "-coverprofile", "sum.cov", "./...")
	if err != nil {
		return err
	}
	return sh.RunV(mg.GoCmd(), "tool", "cover", "-func", "sum.cov")
}

func step(str string) {
	fmt.Printf("----> %s\n", str)
}
