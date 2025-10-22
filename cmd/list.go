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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored passwords interactively",
	Long:  `Display an interactive list of all stored passwords with search functionality`,
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
			fmt.Println("No passwords stored yet. Use 'passvault add' to add one.")
			return
		}

		p := tea.NewProgram(initialListModel(entries, masterPassword))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type listModel struct {
	entries        []internal.PasswordEntry
	filteredItems  []internal.PasswordEntry
	cursor         int
	searchQuery    string
	masterPassword string
	viewMode       string
	selectedEntry  *internal.PasswordEntry
	decryptedPass  string
	err            error
}

func initialListModel(entries []internal.PasswordEntry, masterPassword string) listModel {
	return listModel{
		entries:        entries,
		filteredItems:  entries,
		cursor:         0,
		searchQuery:    "",
		masterPassword: masterPassword,
		viewMode:       "list",
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.viewMode {
		case "list":
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.filteredItems) > 0 {
					selected := m.filteredItems[m.cursor]
					decrypted, err := internal.DecryptPassword(selected.EncryptedPassword, m.masterPassword)
					if err != nil {
						m.err = err
					} else {
						m.selectedEntry = &selected
						m.decryptedPass = decrypted
						m.viewMode = "detail"
					}
				}
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterItems()
					if m.cursor >= len(m.filteredItems) {
						m.cursor = len(m.filteredItems) - 1
						if m.cursor < 0 {
							m.cursor = 0
						}
					}
				}
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.filterItems()
					m.cursor = 0
				}
			}
		case "detail":
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter", "backspace", "esc":
				m.viewMode = "list"
				m.selectedEntry = nil
				m.decryptedPass = ""
				m.err = nil
			}
		}
	}
	return m, nil
}

func (m *listModel) filterItems() {
	if m.searchQuery == "" {
		m.filteredItems = m.entries
		return
	}

	var filtered []internal.PasswordEntry
	query := strings.ToLower(m.searchQuery)
	for _, entry := range m.entries {
		if strings.Contains(strings.ToLower(entry.Service), query) ||
			strings.Contains(strings.ToLower(entry.Username), query) {
			filtered = append(filtered, entry)
		}
	}
	m.filteredItems = filtered
}

func (m listModel) View() string {
	if m.viewMode == "detail" {
		return m.renderDetail()
	}
	return m.renderList()
}

func (m listModel) renderList() string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var s strings.Builder

	if m.searchQuery != "" {
		s.WriteString(fmt.Sprintf("Search: %s\n\n", m.searchQuery))
	} else {
		s.WriteString(mutedStyle.Render("Type to search...") + "\n\n")
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	if len(m.filteredItems) == 0 {
		s.WriteString("No matching passwords found.\n")
	} else {
		for i, entry := range m.filteredItems {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("%s %s (%s)\n", cursor, entry.Service, entry.Username))
		}
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("↑/k up • ↓/j down • enter select • q quit"))

	return s.String()
}

func (m listModel) renderDetail() string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var s strings.Builder

	s.WriteString("Password Details\n\n")

	s.WriteString(fmt.Sprintf("Service: %s\n", m.selectedEntry.Service))
	s.WriteString(fmt.Sprintf("Username: %s\n", m.selectedEntry.Username))
	s.WriteString(fmt.Sprintf("Password: %s\n", m.decryptedPass))
	if m.selectedEntry.Notes != "" {
		s.WriteString(fmt.Sprintf("Notes: %s\n", m.selectedEntry.Notes))
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("Press enter to go back to list • q to quit"))

	return s.String()
}

func init() {
	rootCmd.AddCommand(listCmd)
}
