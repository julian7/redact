package main

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/gitutil"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func (rt *Runtime) gitCleanCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "clean",
		Args:    cobra.NoArgs,
		Short:   "Encoding file from STDIN, to STDOUT",
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.gitCleanDo,
	}

	flags := cmd.Flags()
	flags.Int32P("epoch", "e", -1, "Use specific key epoch (by default it uses the latest key)")
	flags.StringP("file", "f", "", "file path being filtered; ignored when --epoch is set")

	if err := rt.RegisterFlags("git.clean", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) gitCleanDo(cmd *cobra.Command, args []string) error {
	var keyEpoch uint32

	if rt.Viper.IsSet("git.clean.epoch") {
		epoch, err := cast.ToInt32E(rt.Viper.Get("git.clean.epoch"))
		if err != nil {
			return err
		}

		if epoch > 0 {
			keyEpoch = uint32(epoch)
		}
	} else {
		fname := rt.Viper.GetString("git.clean.file")
		epoch, err := rt.epochByFilename(fname)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				rt.Warnf("unable to determine epoch from filename: %s", err.Error())
			}
		} else {
			keyEpoch = epoch
		}
	}

	if keyEpoch == 0 {
		keyEpoch = rt.SecretKey.LatestKey
	}

	if err := rt.SecretKey.Encode(encoder.TypeAES256GCM96, keyEpoch, os.Stdin, os.Stdout); err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) epochByFilename(filename string) (uint32, error) {
	if filename == "" {
		return 0, fs.ErrNotExist
	}

	files, err := gitutil.LsTree("HEAD", []string{filename})
	if err != nil {
		return 0, err
	}

	for _, f := range files {
		if diff := strings.Compare(f.Filename, filename); diff != 0 {
			continue
		}

		fReader, err := gitutil.Cat(f.ObjectID)
		if err != nil {
			return 0, err
		}

		epoch, err := rt.SecretKey.FileStatus(fReader)
		if err == nil {
			return epoch, nil
		}

		return 0, err
	}

	return 0, fs.ErrNotExist
}
