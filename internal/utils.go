package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

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
