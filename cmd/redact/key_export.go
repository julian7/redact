package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (rt *Runtime) keyExportCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "export",
		Args:  cobra.NoArgs,
		Short: "exports key in PEM format",
		Long: `Exports secret key

This command exports a secret key in PEM format, allowing to store them in
markup files like YAML or JSON, or being provided as textual context.

The exported key can be provided as a parameter to the unlock command.`,
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.exportDo,
	}

	flags := cmd.Flags()
	flags.StringP("outfile", "f", "-", "Output filename. Empty string or '-' means standard output.")

	if err := rt.RegisterFlags("export", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) exportDo(cmd *cobra.Command, args []string) error {
	var err error

	var writer *os.File

	outFile := rt.Viper.GetString("export.outfile")

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
