//go:build !debug

package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) extRunCmd() *cli.Command {
	return nil
}
