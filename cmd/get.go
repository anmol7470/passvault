package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anmol7470/passvault/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Search and retrieve a specific password",
	Long:  `Search for passwords by service name or username`,
	Run: func(cmd *cobra.Command, args []string) {
		masterPassword, err := internal.PromptMasterPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			query, err = internal.PromptString("Search query: ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading query: %v\n", err)
				os.Exit(1)
			}
		}

		if query == "" {
			fmt.Fprintf(os.Stderr, "Error: search query cannot be empty\n")
			os.Exit(1)
		}

		entries, err := internal.SearchPasswords(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching passwords: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Printf("No passwords found matching '%s'\n", query)
			return
		}

		p := tea.NewProgram(initialGetModel(entries, masterPassword, query))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type getModel struct {
	entries        []internal.PasswordEntry
	cursor         int
	masterPassword string
	query          string
	selectedEntry  *internal.PasswordEntry
	decryptedPass  string
	err            error
}

func initialGetModel(entries []internal.PasswordEntry, masterPassword, query string) getModel {
	return getModel{
		entries:        entries,
		cursor:         0,
		masterPassword: masterPassword,
		query:          query,
	}
}

func (m getModel) Init() tea.Cmd {
	return nil
}

func (m getModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.selectedEntry != nil {
			switch msg.String() {
			case "ctrl+c", "q", "enter", "esc":
				return m, tea.Quit
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.entries) > 0 {
					selected := m.entries[m.cursor]
					decrypted, err := internal.DecryptPassword(selected.EncryptedPassword, m.masterPassword)
					if err != nil {
						m.err = err
					} else {
						m.selectedEntry = &selected
						m.decryptedPass = decrypted
					}
				}
			}
		}
	}
	return m, nil
}

func (m getModel) View() string {
	if m.selectedEntry != nil {
		return m.renderDetail()
	}
	return m.renderList()
}

func (m getModel) renderList() string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var s strings.Builder

	s.WriteString("Search Results\n\n")
	s.WriteString(fmt.Sprintf("Query: %s\n", m.query))
	s.WriteString(fmt.Sprintf("Found %d match(es)\n\n", len(m.entries)))

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	for i, entry := range m.entries {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s (%s)\n", cursor, entry.Service, entry.Username))
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("↑/k up • ↓/j down • enter select • q quit"))

	return s.String()
}

func (m getModel) renderDetail() string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var s strings.Builder

	s.WriteString("Password Retrieved\n\n")

	s.WriteString(fmt.Sprintf("Service: %s\n", m.selectedEntry.Service))
	s.WriteString(fmt.Sprintf("Username: %s\n", m.selectedEntry.Username))
	s.WriteString(fmt.Sprintf("Password: %s\n", m.decryptedPass))
	if m.selectedEntry.Notes != "" {
		s.WriteString(fmt.Sprintf("Notes: %s\n", m.selectedEntry.Notes))
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("Press any key to exit"))

	return s.String()
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("query", "q", "", "Search query for service or username")
}
