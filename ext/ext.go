package ext

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"sort"

	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
)

const getOperation = "get"

type Operation struct {
	Config map[string]string `json:"config,omitempty"`
	Filter []string          `json:"filter,omitempty"`
}

type Ext struct {
	name    string            `json:"-"`
	repo    *repo.Repo        `json:"-"`
	options map[string]string `json:"-"`
	Command string            `json:"cmd,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	Get     *Operation        `json:"get,omitempty"`
	Put     *Operation        `json:"put,omitempty"`
	Update  *Operation        `json:"update,omitempty"`
	Grant   *Operation        `json:"grant,omitempty"`
	Revoke  *Operation        `json:"revoke,omitempty"`
	Listing *Operation        `json:"list,omitempty"`
}

func (ext *Ext) operation(name string) *Operation {
	switch name {
	case "get":
		return ext.Get
	case "put":
		return ext.Put
	case "update":
		return ext.Update
	case "grant":
		return ext.Grant
	case "revoke":
		return ext.Revoke
	case "list":
		return ext.Listing
	default:
		return nil
	}
}

// Name returns the configured extension name.
func (ext *Ext) Name() string { return ext.name }

func (ext *Ext) cmd(operation string) (*exec.Cmd, error) {
	executable := ext.Command
	if executable == "" {
		executable = fmt.Sprintf("redact-ext-%s", ext.name)
	}

	config := map[string]string{}
	maps.Copy(config, ext.Config)

	if settings := ext.operation(operation); settings != nil {
		maps.Copy(config, settings.Config)
	}

	maps.Copy(config, ext.options)

	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	args := make([]string, 0, len(keys)+1)

	args = append(args, operation)
	for _, key := range keys {
		args = append(args, fmt.Sprintf("%s=%s", key, config[key]))
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

func (ext *Ext) logStart(operation string) {
	logger.Debugf("Using extension %q for operation %q", ext.name, operation)
}

func (ext *Ext) logResult(operation string, err error) {
	logger.Debugf("Extension %q operation %q finished with outcome %d", ext.name, operation, ExitCode(err))
}

func (ext *Ext) Exec(operation string) error {
	if operation == getOperation && ext.Get != nil && len(ext.Get.Filter) > 0 {
		data, err := ext.LoadKey()
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(data)

		return err
	}

	c, err := ext.cmd(operation)
	if err != nil {
		return err
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	ext.logStart(operation)
	err = ext.run(operation, c)
	ext.logResult(operation, err)

	return err
}

// ExecAll dispatches a command in preference order using the extension outcome
// contract. Success and not-applicable advance; missing data and failure stop.
func ExecAll(items []Ext, operation string) error {
	var skipped []error

	successful := false

	for _, item := range items {
		err := item.Exec(operation)
		switch ExitCode(err) {
		case Success:
			if operation == getOperation {
				return nil
			}

			successful = true
		case NotApplicable:
			skipped = append(skipped, err)
		default:
			return err
		}
	}

	if !successful {
		return errors.Join(skipped...)
	}

	return nil
}

func (ext *Ext) List() error {
	return ext.Exec("list")
}

// Run supplies a Key Exchange collection in PEM format for put, grant, or update.
// Revoke receives no secret material. Diagnostics are inherited on stderr.
func (ext *Ext) Run(operation string, key []byte) error {
	if operation == getOperation && ext.Get != nil && len(ext.Get.Filter) > 0 {
		return ext.Exec(operation)
	}

	c, err := ext.cmd(operation)
	if err != nil {
		return err
	}

	c.Stdin = bytes.NewReader(key)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	ext.logStart(operation)
	err = ext.run(operation, c)
	ext.logResult(operation, err)

	return err
}

func (ext *Ext) SaveKey(key []byte) error {
	return ext.Run("put", key)
}

func (ext *Ext) LoadKey() ([]byte, error) {
	c, err := ext.cmd("get")
	if err != nil {
		return nil, err
	}

	c.Stdin = os.Stdin
	c.Stderr = os.Stderr

	ext.logStart("get")

	var output bytes.Buffer

	c.Stdout = &output

	err = ext.run(getOperation, c)
	if err != nil {
		ext.logResult("get", err)

		return nil, err
	}

	ext.logResult("get", nil)

	return output.Bytes(), nil
}
