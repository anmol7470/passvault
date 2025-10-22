package cmd

import (
	"fmt"
	"os"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Search and retrieve a specific password",
	Long:  `Search for passwords by service name or username.`,
	Run: func(cmd *cobra.Command, args []string) {
		masterPassword, err := internal.PromptMasterPassword()
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

		decryptedPassword, err := internal.DecryptPassword(entry.EncryptedPassword, masterPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decrypting password: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nPassword Retrieved")
		fmt.Printf("Service: %s\n", entry.Service)
		fmt.Printf("Username: %s\n", entry.Username)
		fmt.Printf("Password: %s\n", decryptedPassword)
		if entry.Notes != "" {
			fmt.Printf("Notes: %s\n", entry.Notes)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("query", "q", "", "Search query for service or username")
}
