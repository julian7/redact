package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) gitDiffCmd() *cli.Command {
	return &cli.Command{
		Name:      "diff",
		Usage:     "Decoding file from FILENAME to standard out",
		ArgsUsage: "FILENAME",
		Before:    rt.LoadSecretKey,
		Action:    rt.gitDiffDo,
	}
}

func (rt *Runtime) gitDiffDo(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() != 1 {
		return fmt.Errorf("%w: redact git diff requires a single argument", ErrOptions)
	}

	reader, err := os.Open(args.First())
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
		return fmt.Errorf("%w: returned to position %d instead", ErrSeek, n)
	}

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	return nil
}
