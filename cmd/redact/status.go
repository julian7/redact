package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/julian7/redact/files"
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
	key       *files.MasterKey
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
	masterKey, err := basicDo()
	if err != nil {
		cmdErrHandler(err)
	}
	opts.key = masterKey
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
				opts.handleFileEntry(entry, true)
			}
		} else {
			if !opts.encOnly {
				opts.handleFileEntry(entry, false)
			}
		}
	}
}

func (opts statusOptions) handleFileEntry(entry *gitutil.FileEntry, shouldBeEncrypted bool) {
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
			shouldBeEncrypted = false
		}
	} else if isEncrypted != shouldBeEncrypted {
		if isEncrypted {
			msg = append(msg, "should NOT be encrypted")
		} else {
			msg = append(msg, "should be encrypted")
		}
	}
	if isEncrypted {
		if encKeyVersion != opts.key.LatestKey {
			msg = append(msg, fmt.Sprintf("encrypted with key epoch %d, update to %d", encKeyVersion, opts.key.LatestKey))
		}
		_, err := opts.key.Key(encKeyVersion)
		if err != nil {
			msg = append(msg, err.Error())
		}
	}
	printFileEntry(entry, shouldBeEncrypted, strings.Join(msg, "; "))
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
		cmdErrHandler(errors.New("--encrypted and --unencrypted are mutually exclusive options"))
	}
	if opts.fixRepo && (opts.encOnly || opts.plainOnly) {
		cmdErrHandler(errors.New("--encrypted and --unencrypted cannot be used with --fix"))
	}
	return nil
}
