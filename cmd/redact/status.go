package main

import (
	"errors"
	"fmt"

	"github.com/julian7/redact/gitutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status [files...]",
	Short: "Shows redact status",
	Run:   statusDo,
}

func init() {
	flags := statusCmd.Flags()
	flags.BoolP("repo", "r", false, "Show repo status only")
	flags.BoolP("encrypted", "e", false, "Show encrypted files only")
	flags.BoolP("unencrypted", "u", false, "Show plaintext files only")
	flags.BoolP("fix", "f", false, "Fix problems")
	viper.BindPFlags(flags)
	rootCmd.AddCommand(statusCmd)
}

type statusOptions struct {
	repoOnly  bool
	encOnly   bool
	plainOnly bool
	fixRepo   bool
	args      []string
}

func statusDo(cmd *cobra.Command, args []string) {
	opts := statusOptions{
		repoOnly:  viper.GetBool("repo"),
		encOnly:   viper.GetBool("encrypted"),
		plainOnly: viper.GetBool("unencrypted"),
		fixRepo:   viper.GetBool("fix"),
		args:      args,
	}
	if err := opts.validate(); err != nil {
		cmdErrHandler(err)
		return
	}
	files, err := gitutil.LsFiles(opts.args)
	if err != nil {
		cmdErrHandler(err)
		return
	}
	err = files.CheckAttrs()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	for _, entry := range files {
		if entry.Filter == AttrName {
			if opts.encOnly || !opts.plainOnly {
				printFileEntry(entry, true)
			}
		} else {
			if !opts.encOnly {
				printFileEntry(entry, false)
			}
		}
	}
}

func printFileEntry(entry *gitutil.FileEntry, shouldBeEncrypted bool) {
	encryptedString := map[bool]string{
		false: "plaintext",
		true:  "encrypted",
	}
	fmt.Printf("%s: %s\n", encryptedString[shouldBeEncrypted], entry.Name)
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
		cmdErrHandler(errors.New("--encrypted and --unencrypted are mutually exclusive options"))
	}
	if opts.fixRepo && (opts.encOnly || opts.plainOnly) {
		cmdErrHandler(errors.New("--encrypted and --unencrypted cannot be used with --fix"))
	}
	return nil
}
