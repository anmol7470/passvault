package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func PromptMasterPassword() (string, error) {
	isSet, err := IsMasterPasswordSet()
	if err != nil {
		return "", err
	}

	if !isSet {
		return setupMasterPassword()
	}

	return verifyMasterPassword()
}

func setupMasterPassword() (string, error) {
	fmt.Println("No master password set. Let's create one.")
	fmt.Print("Enter master password: ")
	password, err := readPassword()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) < 8 {
		return "", fmt.Errorf("master password must be at least 8 characters")
	}

	fmt.Print("Confirm master password: ")
	confirm, err := readPassword()
	if err != nil {
		return "", fmt.Errorf("failed to read confirmation: %w", err)
	}

	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	hashedPassword, err := HashMasterPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	if err := SetMasterPassword(hashedPassword); err != nil {
		return "", fmt.Errorf("failed to save master password: %w", err)
	}

	fmt.Println("Master password set successfully!")
	return password, nil
}

func verifyMasterPassword() (string, error) {
	fmt.Print("Enter master password: ")
	password, err := readPassword()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	storedHash, err := GetMasterPasswordHash()
	if err != nil {
		return "", err
	}

	if err := VerifyMasterPassword(password, storedHash); err != nil {
		return "", fmt.Errorf("incorrect master password")
	}

	return password, nil
}

func readPassword() (string, error) {
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(password)), nil
}

func PromptString(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func SearchAndSelectPassword(query string) (*PasswordEntry, error) {
	if query == "" {
		var err error
		query, err = PromptString("Search query: ")
		if err != nil {
			return nil, fmt.Errorf("error reading query: %w", err)
		}
	}

	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	entries, err := SearchPasswords(query)
	if err != nil {
		return nil, fmt.Errorf("error searching passwords: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no passwords found matching '%s'", query)
	}

	if len(entries) == 1 {
		return &entries[0], nil
	}

	p := tea.NewProgram(initialSearchModel(entries, query))
	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running selection UI: %w", err)
	}

	finalModel := m.(searchModel)
	if finalModel.selectedEntry == nil {
		return nil, fmt.Errorf("no entry selected")
	}

	return finalModel.selectedEntry, nil
}

type searchModel struct {
	entries       []PasswordEntry
	cursor        int
	query         string
	selectedEntry *PasswordEntry
}

func initialSearchModel(entries []PasswordEntry, query string) searchModel {
	return searchModel{
		entries: entries,
		cursor:  0,
		query:   query,
	}
}

func (m searchModel) Init() tea.Cmd {
	return nil
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
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
				m.selectedEntry = &m.entries[m.cursor]
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m searchModel) View() string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var s strings.Builder

	s.WriteString("Search Results\n\n")
	s.WriteString(fmt.Sprintf("Query: %s\n", m.query))
	s.WriteString(fmt.Sprintf("Found %d match(es)\n\n", len(m.entries)))

	queryLower := strings.ToLower(m.query)
	for i, entry := range m.entries {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s (%s)", cursor, entry.Service, entry.Username))

		if entry.Notes != "" && strings.Contains(strings.ToLower(entry.Notes), queryLower) {
			s.WriteString(fmt.Sprintf(" - %s", entry.Notes))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("↑/k up • ↓/j down • enter select • q quit"))

	return s.String()
}
