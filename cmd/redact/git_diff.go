package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func (rt *Runtime) gitDiffCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "diff FILENAME",
		Args:    cobra.ExactArgs(1),
		Short:   "Decoding file from arg, to STDOUT",
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.gitDiffDo,
	}

	return cmd, nil
}

func (rt *Runtime) gitDiffDo(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("redact git diff requires a single argument")
	}

	reader, err := os.Open(args[0])
	if err != nil {
		return err
	}

	defer reader.Close()

	err = rt.SecretKey.Decode(reader, os.Stdout)
	if err == nil {
		return nil
	}

	n, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("re-reading file from beginning: %w", err)
	}

	if n != 0 {
		return fmt.Errorf("cannot return to beginning of file: returned to position %d instead", n)
	}

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	return nil
}
