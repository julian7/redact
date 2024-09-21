package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) gitCmd() *cli.Command {
	return &cli.Command{
		Name:  "git",
		Usage: "Git commands",
		Description: `Git commands

Redact interacts with git using gitattributes(5), through filter and diff
settings. Unlocked repositories are also configured to run these redact
commands for data conversion.`,
		Commands: []*cli.Command{
			rt.gitCleanCmd(),
			rt.gitDiffCmd(),
			rt.gitSmudgeCmd(),
		},
	}
}
