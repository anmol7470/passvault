package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit all passwords for security weaknesses",
	Long:  `Analyze all stored passwords and generate a security report showing password strength and weak passwords.`,
	Run: func(cmd *cobra.Command, args []string) {
		masterPassword, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		entries, err := internal.ListAllPasswords()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching passwords: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("No passwords to audit.")
			return
		}

		type auditResult struct {
			entry    internal.PasswordEntry
			password string
			strength internal.PasswordStrength
		}

		var results []auditResult
		var weakPasswords []auditResult
		var moderatePasswords []auditResult
		var strongPasswords []auditResult

		fmt.Printf("Auditing %d password(s)...\n\n", len(entries))

		for _, entry := range entries {
			decryptedPassword, err := internal.DecryptPassword(entry.EncryptedPassword, masterPassword)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not decrypt password for %s (%s): %v\n", entry.Service, entry.Username, err)
				continue
			}

			strength := internal.CheckPasswordStrength(decryptedPassword)

			result := auditResult{
				entry:    entry,
				password: decryptedPassword,
				strength: strength,
			}

			results = append(results, result)

			if strength.Score < 2 {
				weakPasswords = append(weakPasswords, result)
			} else if strength.Score < 3 {
				moderatePasswords = append(moderatePasswords, result)
			} else {
				strongPasswords = append(strongPasswords, result)
			}
		}

		sort.Slice(weakPasswords, func(i, j int) bool {
			return weakPasswords[i].strength.Score < weakPasswords[j].strength.Score
		})

		fmt.Println("═══════════════════════════════════════════════")
		fmt.Println("           PASSWORD SECURITY AUDIT")
		fmt.Println("═══════════════════════════════════════════════")
		fmt.Printf("\nTotal passwords: %d\n", len(results))
		fmt.Printf("Strong passwords (score 3-4): %d (%.1f%%)\n", len(strongPasswords), float64(len(strongPasswords))/float64(len(results))*100)
		fmt.Printf("Moderate passwords (score 2): %d (%.1f%%)\n", len(moderatePasswords), float64(len(moderatePasswords))/float64(len(results))*100)
		fmt.Printf("Weak passwords (score 0-1): %d (%.1f%%)\n", len(weakPasswords), float64(len(weakPasswords))/float64(len(results))*100)

		if len(weakPasswords) > 0 {
			fmt.Println("\n═══════════════════════════════════════════════")
			fmt.Println("⚠️ WEAK PASSWORDS - IMMEDIATE ACTION REQUIRED")
			fmt.Println("═══════════════════════════════════════════════")

			for i, result := range weakPasswords {
				fmt.Printf("\n%d. %s (%s)\n", i+1, result.entry.Service, result.entry.Username)
				fmt.Printf("   Score: %d/4\n", result.strength.Score)
				fmt.Printf("   Crack time: %s\n", result.strength.CrackTime)
				if result.strength.Feedback != "" {
					fmt.Printf("   Feedback: %s\n", result.strength.Feedback)
				}
			}
		}

		if len(moderatePasswords) > 0 {
			fmt.Println("\n═══════════════════════════════════════════════")
			fmt.Println("MODERATE PASSWORDS - CONSIDER STRENGTHENING")
			fmt.Println("═══════════════════════════════════════════════")

			for i, result := range moderatePasswords {
				fmt.Printf("\n%d. %s (%s)\n", i+1, result.entry.Service, result.entry.Username)
				fmt.Printf("   Score: %d/4\n", result.strength.Score)
				fmt.Printf("   Crack time: %s\n", result.strength.CrackTime)
				if result.strength.Feedback != "" {
					fmt.Printf("   Feedback: %s\n", result.strength.Feedback)
				}
			}
		}

		if len(strongPasswords) > 0 {
			fmt.Println("\n═══════════════════════════════════════════════")
			fmt.Println("STRONG PASSWORDS")
			fmt.Println("═══════════════════════════════════════════════")

			for i, result := range strongPasswords {
				fmt.Printf("\n%d. %s (%s)\n", i+1, result.entry.Service, result.entry.Username)
				fmt.Printf("   Score: %d/4\n", result.strength.Score)
				fmt.Printf("   Crack time: %s\n", result.strength.CrackTime)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
}
