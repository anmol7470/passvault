package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/anmol7470/passvault/internal"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all passwords to JSON or CSV format",
	Long:  `Export all stored passwords in either JSON or CSV format.`,
	Run: func(cmd *cobra.Command, args []string) {
		masterPassword, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		entries, err := internal.ListAllPasswords()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading passwords: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("No passwords to export.")
			return
		}

		useJSON, _ := cmd.Flags().GetBool("json")
		useCSV, _ := cmd.Flags().GetBool("csv")

		var format string
		if useJSON {
			format = "json"
		} else if useCSV {
			format = "csv"
		} else {
			fmt.Print("Export format (json/csv): ")
			format, err = internal.PromptString("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading format: %v\n", err)
				os.Exit(1)
			}
			format = strings.ToLower(strings.TrimSpace(format))
		}

		type ExportEntry struct {
			Service  string `json:"service"`
			Username string `json:"username"`
			Password string `json:"password"`
			Notes    string `json:"notes,omitempty"`
		}

		var exportEntries []ExportEntry
		for _, entry := range entries {
			decrypted, err := internal.DecryptPassword(entry.EncryptedPassword, masterPassword)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decrypting password for %s: %v\n", entry.Service, err)
				continue
			}
			exportEntries = append(exportEntries, ExportEntry{
				Service:  entry.Service,
				Username: entry.Username,
				Password: decrypted,
				Notes:    entry.Notes,
			})
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		timestamp := time.Now()
		var filename string
		var fullPath string

		switch format {
		case "json":
			filename = fmt.Sprintf("passvault_export_%s.json", timestamp)
			fullPath = filepath.Join(cwd, filename)

			data, err := json.MarshalIndent(exportEntries, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
				os.Exit(1)
			}

			if err := os.WriteFile(fullPath, data, 0600); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}

			absPath, _ := filepath.Abs(fullPath)
			fmt.Printf("Exported %d passwords to: %s\n", len(exportEntries), absPath)

		case "csv":
			filename = fmt.Sprintf("passvault_export_%s.csv", timestamp)
			fullPath = filepath.Join(cwd, filename)

			file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			if err := writer.Write([]string{"Service", "Username", "Password", "Notes"}); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
				os.Exit(1)
			}

			for _, entry := range exportEntries {
				if err := writer.Write([]string{entry.Service, entry.Username, entry.Password, entry.Notes}); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing CSV row: %v\n", err)
					os.Exit(1)
				}
			}

			absPath, _ := filepath.Abs(fullPath)
			fmt.Printf("Exported %d passwords to: %s\n", len(exportEntries), absPath)

		default:
			fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Use 'json' or 'csv'\n", format)
			os.Exit(1)
		}

		fmt.Print("\nOpen the file? (yes/no): ")
		openFile, err := internal.PromptString("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		openFile = strings.ToLower(strings.TrimSpace(openFile))
		if openFile == "yes" {
			if err := openFileInDefaultApp(fullPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			}
		}
	},
}

func openFileInDefaultApp(filepath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filepath)
	case "linux":
		cmd = exec.Command("xdg-open", filepath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filepath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().Bool("json", false, "Export in JSON format")
	exportCmd.Flags().Bool("csv", false, "Export in CSV format")
}
