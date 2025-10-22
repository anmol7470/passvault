package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var changeMasterPasswordCmd = &cobra.Command{
	Use:   "change-master-password",
	Short: "Change the master password",
	Long:  `Change the master password. All stored passwords will be re-encrypted with the new master password.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Changing master password...")

		currentPassword, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		entries, err := internal.ListAllPasswords()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading passwords: %v\n", err)
			os.Exit(1)
		}

		type DecryptedEntry struct {
			ID       int
			Service  string
			Username string
			Password string
			Notes    string
		}

		var decryptedEntries []DecryptedEntry
		for _, entry := range entries {
			decrypted, err := internal.DecryptPassword(entry.EncryptedPassword, currentPassword)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decrypting password for %s: %v\n", entry.Service, err)
				os.Exit(1)
			}
			decryptedEntries = append(decryptedEntries, DecryptedEntry{
				ID:       entry.ID,
				Service:  entry.Service,
				Username: entry.Username,
				Password: decrypted,
				Notes:    entry.Notes,
			})
		}

		fmt.Print("Enter new master password: ")
		newPassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}

		if len(newPassword) < 8 {
			fmt.Fprintf(os.Stderr, "Error: master password must be at least 8 characters\n")
			os.Exit(1)
		}

		fmt.Print("Confirm new master password: ")
		confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
			os.Exit(1)
		}

		if string(newPassword) != string(confirmPassword) {
			fmt.Fprintf(os.Stderr, "Error: passwords do not match\n")
			os.Exit(1)
		}

		newPasswordStr := string(newPassword)

		hashedPassword, err := internal.HashMasterPassword(newPasswordStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error hashing new password: %v\n", err)
			os.Exit(1)
		}

		var reencryptedEntries []internal.PasswordEntry
		for _, entry := range decryptedEntries {
			encrypted, err := internal.EncryptPassword(entry.Password, newPasswordStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encrypting password for %s: %v\n", entry.Service, err)
				os.Exit(1)
			}
			reencryptedEntries = append(reencryptedEntries, internal.PasswordEntry{
				ID:                entry.ID,
				Service:           entry.Service,
				Username:          entry.Username,
				EncryptedPassword: encrypted,
				Notes:             entry.Notes,
			})
		}

		if err := internal.UpdateMasterPassword(hashedPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating master password: %v\n", err)
			os.Exit(1)
		}

		if err := internal.UpdateAllEncryptedPasswords(reencryptedEntries); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating encrypted passwords: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nMaster password changed successfully!\n")
		fmt.Printf("Re-encrypted %d passwords.\n", len(decryptedEntries))
	},
}

func init() {
	rootCmd.AddCommand(changeMasterPasswordCmd)
}
