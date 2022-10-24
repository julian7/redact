package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func (rt *Runtime) keyExportCmd() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "exports key in PEM format",
		ArgsUsage: " ",
		Description: `Exports secret key

This command exports a secret key in PEM format, allowing to store them in
markup files like YAML or JSON, or being provided as textual context.

The exported key can be provided as a parameter to the unlock command.`,
		Before: rt.LoadSecretKey,
		Action: rt.exportDo,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "outfile",
				Aliases: []string{"f"},
				Value:   "-",
				Usage:   "Output `FILENAME`. Empty string or '-' means standard output.",
			},
		},
	}
}

func (rt *Runtime) exportDo(ctx *cli.Context) error {
	var err error

	var writer *os.File

	outFile := ctx.Path("outfile")

	if outFile == "" || outFile == "-" {
		writer = os.Stdout
	} else {
		writer, err = os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("export: %w", err)
		}
	}

	if err := rt.SecretKey.Export(writer); err != nil {
		return fmt.Errorf("export: %w", err)
	}

	return nil
}
