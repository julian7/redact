package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) gitMergeCmd() *cli.Command {
	return &cli.Command{
		Name:      "merge",
		Usage:     "Three-way merge of encrypted files",
		ArgsUsage: "BASE OURS THEIRS [MARKER-SIZE [PATHNAME]]",
		Description: `This plumbing command is a git merge driver (see gitattributes(5)).
It decrypts base, ours, and theirs versions of a file, merges them with
"git merge-file", then encrypts the result back into OURS. It exits with
non-zero status if the merge resulted in conflicts.

Configured as:

	merge.redact.driver = redact git merge %O %A %B %L %P`,
		Before: rt.LoadSecretKey,
		Action: rt.gitMergeDo,
	}
}

func (rt *Runtime) gitMergeDo(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() < 3 || args.Len() > 5 {
		return fmt.Errorf("%w: redact git merge requires BASE, OURS, and THEIRS arguments", ErrOptions)
	}

	base, ours, theirs := args.Get(0), args.Get(1), args.Get(2)

	markerSize := 0

	if args.Len() > 3 {
		var err error

		markerSize, err = strconv.Atoi(args.Get(3))
		if err != nil {
			return fmt.Errorf("%w: invalid marker size %q", ErrOptions, args.Get(3))
		}
	}

	pathname := args.Get(4)

	tmpdir, err := os.MkdirTemp(rt.CommonDir(), "redact-merge-")
	if err != nil {
		return fmt.Errorf("creating temporary directory: %w", err)
	}

	defer os.RemoveAll(tmpdir)

	plainBase := filepath.Join(tmpdir, "base")
	plainOurs := filepath.Join(tmpdir, "ours")
	plainTheirs := filepath.Join(tmpdir, "theirs")

	for _, item := range []struct{ src, dst string }{
		{base, plainBase},
		{ours, plainOurs},
		{theirs, plainTheirs},
	} {
		if err := rt.decodeToFile(item.src, item.dst); err != nil {
			return err
		}
	}

	conflicts, err := gitutil.MergeFile(
		plainOurs,
		plainBase,
		plainTheirs,
		[3]string{mergeLabel(pathname, "ours"), mergeLabel(pathname, "base"), mergeLabel(pathname, "theirs")},
		markerSize,
	)
	if err != nil {
		return err
	}

	if err := rt.encodeMergeResult(plainOurs, ours); err != nil {
		return err
	}

	if conflicts > 0 {
		rt.Warnf("%d conflict(s) in %s", conflicts, pathname)

		return cli.Exit("", 1)
	}

	return nil
}

func mergeLabel(pathname, side string) string {
	if pathname == "" {
		return side
	}

	return pathname + " (" + side + ")"
}

// decodeToFile decrypts src into dst. Unencrypted input is copied verbatim.
func (rt *Runtime) decodeToFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading merge input: %w", err)
	}

	var out bytes.Buffer

	if err := rt.Decode(bytes.NewReader(data), &out); err != nil {
		if !isUnencrypted(err) {
			return fmt.Errorf("decrypting merge input: %w", err)
		}

		out.Reset()
		out.Write(data)
	}

	if err := os.WriteFile(dst, out.Bytes(), 0o600); err != nil {
		return fmt.Errorf("writing decrypted merge input: %w", err)
	}

	return nil
}

// encodeMergeResult encrypts src into dst, keeping dst's key epoch and
// encoding type, if dst was encrypted.
func (rt *Runtime) encodeMergeResult(src, dst string) error {
	keyEpoch := rt.LatestKey
	encType := encoder.TypeAES256GCM96

	if hdr := rt.fileHeader(dst); hdr != nil {
		keyEpoch = hdr.Epoch
		encType = hdr.Encoding
	}

	reader, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("reading merge result: %w", err)
	}

	defer reader.Close()

	var out bytes.Buffer

	if err := rt.Encode(encType, keyEpoch, reader, &out); err != nil {
		return fmt.Errorf("encrypting merge result: %w", err)
	}

	if err := os.WriteFile(dst, out.Bytes(), 0o600); err != nil {
		return fmt.Errorf("writing merge result: %w", err)
	}

	return nil
}

func isUnencrypted(err error) bool {
	return errors.Is(err, files.ErrInvalidPreamble) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}

func (rt *Runtime) fileHeader(filename string) *files.FileHeader {
	reader, err := os.Open(filename)
	if err != nil {
		return nil
	}

	defer reader.Close()

	hdr, err := rt.FileStatus(reader)
	if err != nil {
		return nil
	}

	return hdr
}
