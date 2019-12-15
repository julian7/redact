package main

import (
	"github.com/spf13/cobra"
)

func (rt *Runtime) gitCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Git commands",
		Long: `Git commands

Redact interacts with git using gitattributes(5), through filter and diff
settings. Unlocked repositories are also configured to run these redact
commands for data conversion.`,
	}

	subcommands := []cmdFactory{
		rt.gitCleanCmd,
		rt.gitDiffCmd,
		rt.gitSmudgeCmd,
	}

	if err := rt.AddCmdTo(cmd, subcommands); err != nil {
		return nil, err
	}

	return cmd, nil
}
