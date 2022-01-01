package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func (rt *Runtime) gitCleanCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "clean",
		Args:  cobra.NoArgs,
		Short: "Encoding file from STDIN, to STDOUT",
		Long: `This plumbing command allows fine-tuning encryption of individual files.
This command takes a cleartext file from standard input, and emits encoded
contents to standard out. You can set specific epoch (key number), encoding
type for this process, or you can also take an already existing, encrypted
file in the git repository, to be used as a template.

To enforce encoding type, set "git.clean.type" in the config file, or set
REDACT_GIT_CLEAN_TYPE environment variable with encryption name (case
insensitive).

Currently only AES256-GCM96 (default) and ChaCha20-Poly1305 encryptions are
supported. According to [Go's automatic cipher suite
ordering](https://go.dev/blog/tls-cipher-suites) blog post, the only two
viable encryptions are the aforementioned two. CPUs with AES-NI support
go just fine with AES256-GCM96, but when used in environments with no
hardware support, ChaCha20-Poly1305 is the better choice.
`,
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.gitCleanDo,
	}

	flags := cmd.Flags()
	flags.Uint32P("epoch", "e", 0, "Use specific key epoch (by default it uses the latest key)")
	flags.StringP("type", "t", "", "Use specific encoding type (aes256-gcm96 (default) or chacha20-poly1305)")
	flags.StringP("file", "f", "", "file path being filtered; --epoch and --type overwrites")

	if err := rt.RegisterFlags("git.clean", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) gitCleanDo(cmd *cobra.Command, args []string) error {
	var keyEpoch uint32

	var encType = encoder.TypeAES256GCM96

	if rt.Viper.IsSet("git.clean.epoch") {
		epoch, err := cast.ToUint32E(rt.Viper.Get("git.clean.epoch"))
		if err != nil {
			return err
		}

		if epoch > 0 {
			keyEpoch = epoch
		}
	} else {
		fname := rt.Viper.GetString("git.clean.file")
		hdr, err := rt.hdrByFilename(fname)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				rt.Warnf("unable to determine epoch from filename: %s", err.Error())
			}
		} else {
			keyEpoch = hdr.Epoch
			encType = hdr.Encoding
		}
	}

	if keyEpoch == 0 {
		keyEpoch = rt.SecretKey.LatestKey
	}

	encTypeName := rt.Viper.GetString("git.clean.type")
	if encTypeName != "" {
		var err error

		encType, err = encoder.FindEncoder(encTypeName)
		if err != nil {
			return fmt.Errorf("finding encoding %q: %w", encTypeName, err)
		}
	}

	if err := rt.SecretKey.Encode(encType, keyEpoch, os.Stdin, os.Stdout); err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) hdrByFilename(filename string) (*files.FileHeader, error) {
	if filename == "" {
		return nil, fs.ErrNotExist
	}

	files, err := gitutil.LsTree("HEAD", []string{filename})
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if diff := strings.Compare(f.Filename, filename); diff != 0 {
			continue
		}

		fReader, err := gitutil.Cat(f.ObjectID)
		if err != nil {
			return nil, err
		}

		hdr, err := rt.SecretKey.FileStatus(fReader)
		if err == nil {
			return hdr, nil
		}

		return nil, err
	}

	return nil, fs.ErrNotExist
}
