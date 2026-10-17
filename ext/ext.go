package ext

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/julian7/redact/repo"
)

type Ext struct {
	name    string            `json:"-"`
	repo    *repo.Repo        `json:"-"`
	Command string            `json:"cmd,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
}

func (ext *Ext) cmd(operation string) (*exec.Cmd, error) {
	executable := ext.Command
	if executable == "" {
		executable = fmt.Sprintf("redact-ext-%s", ext.name)
	}

	keys := make([]string, 0, len(ext.Config))
	for key := range ext.Config {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	args := make([]string, 0, len(keys)+1)

	args = append(args, operation)
	for _, key := range keys {
		args = append(args, fmt.Sprintf("%s=%s", key, ext.Config[key]))
	}

	c := exec.Command(executable, args...)
	c.Dir = ext.repo.Workdir.Root()

	if err := ext.privateEnv(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (ext *Ext) result(operation string, err error) error {
	if err == nil {
		return nil
	}

	code := Failure

	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() >= NotApplicable && exit.ExitCode() <= Failure {
		code = exit.ExitCode()
	}

	return fmt.Errorf("extension %s %s: %w", ext.name, operation, &Error{Code: code, Err: err})
}

func (ext *Ext) Exec(operation string) error {
	c, err := ext.cmd(operation)
	if err != nil {
		return err
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return ext.result(operation, c.Run())
}

func (ext *Ext) List() error {
	return ext.Exec("list")
}

// Run supplies a Key Exchange collection in PEM format for put, grant, or update.
// Revoke receives no secret material. Diagnostics are inherited on stderr.
func (ext *Ext) Run(operation string, key []byte) error {
	c, err := ext.cmd(operation)
	if err != nil {
		return err
	}

	c.Stdin = bytes.NewReader(key)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return ext.result(operation, c.Run())
}

func (ext *Ext) SaveKey(key []byte) error {
	return ext.Run("put", key)
}

func (ext *Ext) LoadKey() ([]byte, error) {
	c, err := ext.cmd("get")
	if err != nil {
		return nil, err
	}

	c.Stderr = os.Stderr

	val, err := c.Output()
	if err != nil {
		return nil, ext.result("get", err)
	}

	return val, nil
}
