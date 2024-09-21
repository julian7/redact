package ext

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/julian7/redact/repo"
)

type Ext struct {
	name    string            `json:"-"`
	Command string            `json:"cmd,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	repo    *repo.Repo        `json:"-"`
}

func (ext *Ext) cmd(cmd string) *exec.Cmd {
	executable := ext.Command
	if len(executable) <= 0 {
		executable = fmt.Sprintf("redact-ext-%s", ext.name)
	}
	args := make([]string, 0, len(ext.Config)+1)
	args = append(args, cmd)
	for key, val := range ext.Config {
		args = append(args, fmt.Sprintf("%s=%s", key, val))
	}
	return exec.Command(executable, args...)
}

func (ext *Ext) Exec(cmd string) error {
	c := ext.cmd(cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}

	return nil
}

func (ext *Ext) List() error {
	return ext.Exec("list")
}

func (ext *Ext) SaveKey(key []byte) error {
	c := ext.cmd("put")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	writer, err := c.StdinPipe()
	if err != nil {
		return fmt.Errorf("savekey setting up extension stdin: %w", err)
	}
	if err = c.Start(); err != nil {
		return fmt.Errorf("savekey running extension: %w", err)
	}
	writer.Write(key)
	writer.Close()

	return c.Wait()
}

func (ext *Ext) LoadKey() ([]byte, error) {
	c := ext.cmd("get")
	c.Stderr = os.Stderr
	val, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("loadkey reading extension output: %w", err)
	}
	return val, nil
}
