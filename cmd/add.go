package cmd

import (
	"fmt"
	"os"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new password entry",
	Long:  `Add a new password entry to the vault with service name, username, and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		masterPassword, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		service, _ := cmd.Flags().GetString("service")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if service == "" {
			service, err = internal.PromptString("Service: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading service: %v\n", err)
				os.Exit(1)
			}
		}

		if username == "" {
			username, err = internal.PromptString("Username: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
				os.Exit(1)
			}
		}

		if password == "" {
			password, err = internal.PromptString("Password: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
		}

		if service == "" || username == "" || password == "" {
			fmt.Fprintf(os.Stderr, "Error: service, username, and password are required\n")
			os.Exit(1)
		}

		encryptedPassword, err := internal.EncryptPassword(password, masterPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encrypting password: %v\n", err)
			os.Exit(1)
		}

		if err := internal.AddPassword(service, username, encryptedPassword, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving password: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Password for %s (%s) added successfully!\n", service, username)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringP("service", "s", "", "Service name")
	addCmd.Flags().StringP("username", "u", "", "Username")
	addCmd.Flags().StringP("password", "p", "", "Password")
}
