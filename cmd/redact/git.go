package main

import (
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git commands",
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
