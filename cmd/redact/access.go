package main

import (
	"github.com/spf13/cobra"
)

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Key Exchange commands",
	Long: `Key Exchange allows contributors to have access to the master key by
storing them in OpenPGP-encrypted format for each individual.

Key Exchange commands allows you to add or remove contributors to/from the
project.`,
}

func init() {
	rootCmd.AddCommand(accessCmd)
}
