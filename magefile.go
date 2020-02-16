//+build mage

package main

import (
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type buildTarget struct {
	os   string
	arch string
	ext  string
	out  string
}

func Build() error {
	return sh.RunV(mg.GoCmd(), "build", "./cmd/redact")
}

func All() error {
	return sh.RunV("goreleaser", "release", "--snapshot", "--rm-dist", "--skip-publish")
}

func Release() error {
	return sh.RunV("goreleaser", "release")
}

// AllTests runs go fmt, go vet, test, and cover
func AllTests() {
	mg.SerialDeps(Lint, Test, Cover)
}

// Test runs go test
func Test() error {
	return sh.RunV(mg.GoCmd(), "test", "./...")
}

// Lint runs lint tool
func Lint() error {
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
	err := sh.RunV(mg.GoCmd(), "test", "-coverprofile", "sum.cov", "./...")
	if err != nil {
		return err
	}
	return sh.RunV(mg.GoCmd(), "tool", "cover", "-func", "sum.cov")
}
