package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "passvault",
	Short: "A secure CLI-based password manager",
	Long:  "PassVault is a secure command-line password manager built with Go.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
