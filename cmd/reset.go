package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the entire database",
	Long:  `Delete all passwords and master password from the database. This action cannot be undone.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("\n⚠️  WARNING: This will delete ALL passwords and reset the master password.\n")
		fmt.Print("This action CANNOT be undone!\n\n")
		fmt.Print("Type 'DELETE' to confirm: ")

		confirmation, err := internal.PromptString("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
			os.Exit(1)
		}

		if strings.TrimSpace(confirmation) != "DELETE" {
			fmt.Println("Reset cancelled.")
			return
		}

		if err := internal.ResetDatabase(); err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nDatabase reset successfully!")
		fmt.Println("All passwords and master password have been deleted.")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
