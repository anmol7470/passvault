package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a password entry",
	Long:  `Search for a password entry and delete it after confirmation.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		query, _ := cmd.Flags().GetString("query")
		entry, err := internal.SearchAndSelectPassword(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("\n⚠️  WARNING: You are about to delete the following password:\n")
		fmt.Printf("Service: %s\n", entry.Service)
		fmt.Printf("Username: %s\n\n", entry.Username)
		fmt.Print("Are you sure you want to delete this password? (yes/no): ")

		confirmation, err := internal.PromptString("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
			os.Exit(1)
		}

		confirmation = strings.ToLower(strings.TrimSpace(confirmation))
		if confirmation != "yes" {
			fmt.Println("Deletion cancelled.")
			return
		}

		if err := internal.DeletePassword(entry.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting password: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nPassword for %s (%s) deleted successfully!\n", entry.Service, entry.Username)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("query", "q", "", "Search query for service or username")
}
