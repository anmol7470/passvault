package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anmol7470/passvault/internal"
	"github.com/atotto/clipboard"
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
		if query == "" && len(args) > 0 {
			query = args[0]
		}

		var entry *internal.PasswordEntry

		if query != "" {
			aliasEntry, err := internal.GetPasswordByAlias(query)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking alias: %v\n", err)
				os.Exit(1)
			}

			if aliasEntry != nil {
				entry = aliasEntry
				decryptedPassword, err := internal.DecryptPassword(entry.EncryptedPassword, masterPassword)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error decrypting password: %v\n", err)
					os.Exit(1)
				}

				err = clipboard.WriteAll(decryptedPassword)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
					os.Exit(1)
				}

				fmt.Printf("✓ Password for %s (%s) copied to clipboard!\n", entry.Service, entry.Username)
				return
			}
		}

		entry, err = internal.SearchAndSelectPassword(query)
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
		if entry.Alias != "" {
			fmt.Printf("Alias: %s\n", entry.Alias)
		}

		fmt.Print("\nCopy password to clipboard? (yes/no): ")
		copyChoice, err := internal.PromptString("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading choice: %v\n", err)
			os.Exit(1)
		}

		if strings.ToLower(copyChoice) == "yes" {
			err = clipboard.WriteAll(decryptedPassword)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Password copied to clipboard!")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("query", "q", "", "Search query for service or username")
}
