package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing password entry",
	Long:  `Search for a password entry and update its service, username, password, or notes.`,
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

		fmt.Printf("\nUpdating password for %s (%s)\n", entry.Service, entry.Username)
		fmt.Println("Press Enter to keep current value")
		fmt.Println()

		newService, err := promptWithDefault("Service", entry.Service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading service: %v\n", err)
			os.Exit(1)
		}

		newUsername, err := promptWithDefault("Username", entry.Username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
			os.Exit(1)
		}

		newPassword, err := promptWithDefault("Password", decryptedPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}

		newNotes, err := promptWithDefault("Notes", entry.Notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading notes: %v\n", err)
			os.Exit(1)
		}

		encryptedPassword, err := internal.EncryptPassword(newPassword, masterPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encrypting password: %v\n", err)
			os.Exit(1)
		}

		if err := internal.UpdatePassword(entry.ID, newService, newUsername, encryptedPassword, newNotes); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating password: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nPassword for %s (%s) updated successfully!\n", newService, newUsername)
	},
}

func promptWithDefault(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("query", "q", "", "Search query for service or username")
}
