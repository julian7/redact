package main

import (
	"context"
	"errors"
	"os"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func main() {
	rt := &Runtime{
		Repo: &repo.Repo{},
	}

	if err := rt.app().Run(context.Background(), os.Args); err != nil {
		var outcome *ext.Error
		if errors.As(err, &outcome) {
			logger.Error(err.Error())
			os.Exit(ext.ExitCode(err))
		}

		logger.Fatal(err.Error())
	}
}

func openFileToRead(filename string) (*os.File, error) {
	if filename == "" || filename == "-" {
		return os.Stdin, nil
	}

	return os.Open(filename)
}

func commands(cmds ...*cli.Command) []*cli.Command {
	result := make([]*cli.Command, 0, len(cmds))

	for _, item := range cmds {
		if item != nil {
			result = append(result, item)
		}
	}

	return result
}
