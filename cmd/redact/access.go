package main

import (
	"github.com/spf13/cobra"
)

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Key Exchange commands",
	Long: `Key Exchange commands

Key exchange allows contributors to have access to the master key inside
the git repo by storing them in OpenPGP-encrypted format for each individual.

With Key Exchange commands you can give or revoke access to the project
for contributors by their OpenPGP keys.`,
}

func init() {
	rootCmd.AddCommand(accessCmd)
}
