package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) gitCleanCmd() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Encoding file from STDIN, to STDOUT",
		Description: `This plumbing command allows fine-tuning encryption of individual files.
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
		Before: rt.LoadSecretKey,
		Action: rt.gitCleanDo,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "epoch",
				Aliases: []string{"e"},
				Value:   0,
				Usage:   "Use specific key epoch (by default it uses the latest key)",
				Sources: cli.EnvVars("REDACT_GIT_CLEAN_EPOCH"),
			},
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "Use specific `encoding` type (aes256-gcm96 (default) or chacha20-poly1305)",
				Sources: cli.EnvVars("REDACT_GIT_CLEAN_TYPE"),
			},
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "file path being filtered; --epoch and --type overwrites",
			},
		},
	}
}

func (rt *Runtime) gitCleanDo(ctx context.Context, cmd *cli.Command) error {
	keyEpoch := uint32(0)
	encType := encoder.TypeAES256GCM96

	if cmd.IsSet("file") {
		fname := cmd.String("file")
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

	if cmd.IsSet("epoch") {
		keyEpoch = uint32(cmd.Uint("epoch"))
	}

	if keyEpoch == 0 {
		keyEpoch = rt.SecretKey.LatestKey
	}

	encTypeName := cmd.String("type")
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
