package main

import (
	"io"
	"os"

	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var gitDiffCmd = &cobra.Command{
	Use:   "diff FILENAME",
	Short: "Decoding file from arg, to STDOUT",
	Run:   gitDiffDo,
}

func init() {
	gitCmd.AddCommand(gitDiffCmd)
}

func gitDiffDo(cmd *cobra.Command, args []string) {
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	if len(args) != 1 {
		cmdErrHandler(errors.New("redact git diff requires a single argument"))
		return
	}
	reader, err := os.Open(args[0])
	if err != nil {
		cmdErrHandler(err)
		return
	}
	defer reader.Close()
	err = masterkey.Decode(reader, os.Stdout)
	if err == nil {
		return
	}
	n, err := reader.Seek(0, os.SEEK_SET)
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "re-reading file from beginning"))
	}
	if n != 0 {
		cmdErrHandler(errors.Errorf("cannot return to beginning of file: returned to position %d instead", n))
		return
	}
	readbuf := make([]byte, 1024)
	for {
		n, err := reader.Read(readbuf)
		if err != nil {
			if err == io.EOF {
				return
			}
			cmdErrHandler(errors.Wrap(err, "reading file"))
			return
		}
		os.Stdout.Write(readbuf[:n])
	}

}
