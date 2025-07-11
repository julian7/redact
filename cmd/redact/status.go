package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) statusCmd() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "Shows redact status",
		ArgsUsage: "[files...]",
		Description: `Show encryption status of repo files

This command lists all files in the repository (filtering is possible with
--encrypted, --repo, --unencrypted, and --quiet options) showing encrypted
and not encrypted files. It also detects possible problems with file
statuses, when a file was wrongly encrypted, or not encrypted even it should
have been.

It also shows if a file is encrypted with an older key. While re-encryption
as-is is possible with --rekey option, it's strongly recommended to replace
these secrets instead.`,
		Before: rt.LoadSecretKey,
		Action: rt.statusDo,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "Show repo status only",
			},
			&cli.BoolFlag{
				Name:    "encrypted",
				Aliases: []string{"e"},
				Value:   false,
				Usage:   "Show encrypted files only",
			},
			&cli.BoolFlag{
				Name:    "unencrypted",
				Aliases: []string{"u"},
				Value:   false,
				Usage:   "Show plaintext files only",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Value:   false,
				Usage:   "Quiet mode (report only issues)",
			},
			&cli.BoolFlag{
				Name:    "fix",
				Aliases: []string{"f"},
				Value:   false,
				Usage:   "Fix problems (doesn't affect files encrypted with older keys)",
			},
			&cli.BoolFlag{
				Name:    "rekey",
				Aliases: []string{"R"},
				Value:   false,
				Usage:   "Rekey files (NOT RECOMMENDED; update for latest encryption key)",
			},
		},
	}
}

type statusOptions struct {
	Logger     *logger.Logger
	repoOnly   bool
	encOnly    bool
	plainOnly  bool
	quiet      bool
	fixRepo    bool
	rekeyFiles bool
	key        *files.SecretKey
	args       []string
	toFix      []string
	toRekey    []string
}

func (rt *Runtime) statusDo(_ context.Context, cmd *cli.Command) error {
	opts := statusOptions{
		Logger:     rt.Logger,
		repoOnly:   cmd.Bool("repo"),
		encOnly:    cmd.Bool("encrypted"),
		plainOnly:  cmd.Bool("unencrypted"),
		quiet:      cmd.Bool("quiet"),
		fixRepo:    cmd.Bool("fix"),
		rekeyFiles: cmd.Bool("rekey"),
		args:       cmd.Args().Slice(),
	}
	if err := opts.validate(); err != nil {
		return err
	}

	opts.key = rt.SecretKey

	files, err := gitutil.LsFiles(opts.args)
	if err != nil {
		return err
	}

	if err := files.CheckAttrs(); err != nil {
		return err
	}

	for _, entry := range files.Errors {
		rt.Warn(entry.Error())
	}

	for _, entry := range files.Items {
		if entry.Filter == repo.AttrName && entry.Status != gitutil.StatusOther {
			if opts.encOnly || !opts.plainOnly {
				opts.handleFileEntry(entry, true)
			}
		} else {
			if !opts.encOnly {
				opts.handleFileEntry(entry, false)
			}
		}
	}

	if opts.fixRepo || opts.rekeyFiles {
		if err := rt.ForceReencrypt(opts.rekeyFiles, func(err error) {
			rt.Warn(err.Error())
		}); err != nil {
			return fmt.Errorf("fixing problems: %w", err)
		}
	}

	return nil
}

func (opts *statusOptions) handleFileEntry(entry *gitutil.FileEntry, shouldBeEncrypted bool) {
	var isEncrypted bool

	var encKeyVersion, encType uint32

	reader, err := gitutil.Cat(entry.SHA1[:])
	if err != nil {
		opts.Logger.Warnf("git cat-file %s: %w", entry.Name, err)

		return
	}

	defer reader.Close()

	hdr, err := opts.key.FileStatus(reader)
	if err == nil {
		encKeyVersion = hdr.Epoch
		encType = hdr.Encoding
		isEncrypted = true
	}

	msg := []string{}

	baseName := filepath.Base(entry.Name)
	if strings.HasPrefix(entry.Name, repo.DefaultKeyExchangeDir+"/") || baseName == repo.GitAttributesFile {
		if isEncrypted {
			msg = append(msg, "should NEVER be encrypted")
			opts.toFix = append(opts.toFix, entry.Name)
			shouldBeEncrypted = false
		}
	} else if isEncrypted != shouldBeEncrypted {
		if isEncrypted {
			msg = append(msg, "should NOT be encrypted")
		} else {
			msg = append(msg, "should be encrypted")
		}

		opts.toFix = append(opts.toFix, entry.Name)
	}

	if isEncrypted {
		msg = append(msg, fmt.Sprintf("encoded with %s", encoder.Name(encType)))
		if encKeyVersion != opts.key.LatestKey {
			msg = append(msg, fmt.Sprintf("encrypted with key epoch %d, update to %d", encKeyVersion, opts.key.LatestKey))
			opts.toRekey = append(opts.toRekey, entry.Name)
		}

		if _, err := opts.key.Key(encKeyVersion); err != nil {
			msg = append(msg, err.Error())
		}
	}

	if !opts.repoOnly && (!opts.quiet || len(msg) > 0) {
		printFileEntry(entry, isEncrypted, shouldBeEncrypted, strings.Join(msg, "; "))
	}
}

func printFileEntry(entry *gitutil.FileEntry, isEncrypted bool, shouldBeEncrypted bool, msg string) {
	encryptedString := map[bool]string{
		false: "   ",
		true:  "enc",
	}
	fixString := map[bool]string{
		false: "   ",
		true:  "fix",
	}
	data := fmt.Sprintf(
		"%s %s %s",
		encryptedString[isEncrypted],
		fixString[isEncrypted != shouldBeEncrypted],
		entry.Name,
	)

	if len(msg) > 0 {
		data = fmt.Sprintf("%s NOTE: %s", data, msg)
	}

	fmt.Println(data)
}

func (opts statusOptions) validate() error {
	if opts.repoOnly {
		if opts.encOnly || opts.plainOnly {
			return fmt.Errorf("%w: --encrypted and --unencrypted options cannot be use with --repo", ErrOptions)
		}

		if opts.fixRepo {
			return fmt.Errorf("%w: --fix option cannot be used with --repo", ErrOptions)
		}

		if len(opts.args) > 0 {
			return fmt.Errorf("%w: files cannot be specified when --repo is used", ErrOptions)
		}
	}

	if opts.encOnly && opts.plainOnly {
		return fmt.Errorf("%w: --encrypted and --unencrypted are mutually exclusive options", ErrOptions)
	}

	if opts.fixRepo && (opts.encOnly || opts.plainOnly) {
		return fmt.Errorf("%w: --encrypted and --unencrypted cannot be used with --fix", ErrOptions)
	}

	return nil
}
