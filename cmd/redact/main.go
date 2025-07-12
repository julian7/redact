package main

import (
	"context"
	"os"

	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func main() {
	rt := &Runtime{
		Repo:   &repo.Repo{},
		Logger: logger.New(),
	}

	if err := rt.app().Run(context.Background(), os.Args); err != nil {
		rt.Fatal(err.Error())
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
