package main

import (
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git commands",
	Long: `Git commands

Redact interacts with git using gitattributes(5), through filter and diff
settings. Unlocked repositories are also configured to run these redact
commands for data conversion.`,
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
