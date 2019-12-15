package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
)

func (rt *Runtime) statusCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "status [files...]",
		Args:  cobra.ArbitraryArgs,
		Short: "Shows redact status",
		Long: `Show encryption status of repo files

This command lists all files in the repository (filtering is possible with
--encrypted, --repo, --unencrypted, and --quiet options) showing encrypted
and not encrypted files. It also detects possible problems with file
statuses, when a file was wrongly encrypted, or not encrypted even it should
have been.

It also shows if a file is encrypted with an older key. While re-encryption
as-is is possible with --rekey option, it's strongly recommended to replace
these secrets instead.`,
		PreRunE: rt.RetrieveMasterKey,
		RunE:    rt.statusDo,
	}

	flags := cmd.Flags()
	flags.BoolP("repo", "r", false, "Show repo status only")
	flags.BoolP("encrypted", "e", false, "Show encrypted files only")
	flags.BoolP("unencrypted", "u", false, "Show plaintext files only")
	flags.BoolP("quiet", "q", false, "Quiet mode (report only issues)")
	flags.BoolP("fix", "f", false, "Fix problems (doesn't affect files encrypted with older keys)")
	flags.BoolP("rekey", "R", false, "Rekey files (NOT RECOMMENDED; update for latest encryption key)")

	if err := rt.RegisterFlags("status", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

type statusOptions struct {
	repoOnly   bool
	encOnly    bool
	plainOnly  bool
	quiet      bool
	fixRepo    bool
	rekeyFiles bool
	key        *files.MasterKey
	args       []string
	toFix      []string
	toRekey    []string
}

func (rt *Runtime) statusDo(cmd *cobra.Command, args []string) error { //nolint:funlen
	opts := statusOptions{
		repoOnly:   rt.Viper.GetBool("status.repo"),
		encOnly:    rt.Viper.GetBool("status.encrypted"),
		plainOnly:  rt.Viper.GetBool("status.unencrypted"),
		quiet:      rt.Viper.GetBool("status.quiet"),
		fixRepo:    rt.Viper.GetBool("status.fix"),
		rekeyFiles: rt.Viper.GetBool("status.rekey"),
		args:       args,
	}
	if err := opts.validate(); err != nil {
		return err
	}

	opts.key = rt.MasterKey

	files, err := gitutil.LsFiles(opts.args)
	if err != nil {
		return err
	}

	if err := files.CheckAttrs(rt.MasterKey.Logger); err != nil {
		return err
	}

	for _, entry := range files {
		if entry.Filter == sdk.AttrName && entry.Status != gitutil.StatusOther {
			if opts.encOnly || !opts.plainOnly {
				opts.handleFileEntry(entry, true)
			}
		} else {
			if !opts.encOnly {
				opts.handleFileEntry(entry, false)
			}
		}
	}

	var msgFix string

	if opts.fixRepo {
		err := sdk.TouchUpFiles(rt.MasterKey, opts.toFix, func(entry string, err error) {
			rt.Logger.Warnf("%s: %v", entry, err)
		})

		if err != nil {
			return err
		}

		msgFix = "Fixed %d file%s.\n"
	} else {
		msgFix = "There are %d file%s to be fixed.\n"
	}

	l := len(opts.toFix)
	fmt.Printf(msgFix, l, map[bool]string{false: "s", true: ""}[l == 1])

	var msgRekey string

	if opts.rekeyFiles {
		err := sdk.TouchUpFiles(rt.MasterKey, opts.toRekey, func(entry string, err error) {
			rt.Logger.Warnf("%s: %v", entry, err)
		})
		if err != nil {
			return err
		}

		msgRekey = "Re-encrypted %d file%s.\n"
	} else {
		msgRekey = "There are %d file%s to be re-encrypted.\n"
	}

	l = len(opts.toRekey)
	fmt.Printf(msgRekey, l, map[bool]string{false: "s", true: ""}[l == 1])

	return nil
}

func (opts *statusOptions) handleFileEntry(entry *gitutil.FileEntry, shouldBeEncrypted bool) {
	var isEncrypted bool

	var encKeyVersion uint32

	reader, err := gitutil.Cat(entry.SHA1[:])
	if err == nil {
		isEncrypted, encKeyVersion = opts.key.FileStatus(reader)
		defer reader.Close()
	}

	msg := []string{}

	if strings.HasPrefix(entry.Name, files.DefaultKeyExchangeDir+"/") {
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
		if encKeyVersion != opts.key.LatestKey {
			msg = append(msg, fmt.Sprintf("encrypted with key epoch %d, update to %d", encKeyVersion, opts.key.LatestKey))
			opts.toRekey = append(opts.toRekey, entry.Name)
		}

		if _, err := opts.key.Key(encKeyVersion); err != nil {
			msg = append(msg, err.Error())
		}
	}

	if !opts.repoOnly && (!opts.quiet || len(msg) > 0) {
		printFileEntry(entry, shouldBeEncrypted, strings.Join(msg, "; "))
	}
}

func printFileEntry(entry *gitutil.FileEntry, shouldBeEncrypted bool, msg string) {
	encryptedString := map[bool]string{
		false: "           ",
		true:  "encrypted: ",
	}
	data := fmt.Sprintf("%s %s", encryptedString[shouldBeEncrypted], entry.Name)

	if len(msg) > 0 {
		data = fmt.Sprintf("%s WARNING: %s", data, msg)
	}

	fmt.Println(data)
}

func (opts statusOptions) validate() error {
	if opts.repoOnly {
		if opts.encOnly || opts.plainOnly {
			return errors.New("--encrypted and --unencrypted options cannot be use with --repo")
		}

		if opts.fixRepo {
			return errors.New("--fix option cannot be used with --repo")
		}

		if len(opts.args) > 0 {
			return errors.New("files cannot be specified when --repo is used")
		}
	}

	if opts.encOnly && opts.plainOnly {
		return errors.New("--encrypted and --unencrypted are mutually exclusive options")
	}

	if opts.fixRepo && (opts.encOnly || opts.plainOnly) {
		return errors.New("--encrypted and --unencrypted cannot be used with --fix")
	}

	return nil
}
