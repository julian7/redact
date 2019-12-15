package main

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (rt *Runtime) gitDiffCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "diff FILENAME",
		Args:    cobra.ExactArgs(1),
		Short:   "Decoding file from arg, to STDOUT",
		PreRunE: rt.RetrieveMasterKey,
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
	err = rt.MasterKey.Decode(reader, os.Stdout)
	if err == nil {
		return nil
	}
	n, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "re-reading file from beginning")
	}
	if n != 0 {
		return errors.Errorf("cannot return to beginning of file: returned to position %d instead", n)
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return errors.Wrap(err, "reading file")
	}

	return nil
}
