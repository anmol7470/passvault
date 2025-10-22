package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	passvaultDir := filepath.Join(homeDir, ".passvault")
	if err := os.MkdirAll(passvaultDir, 0700); err != nil {
		return fmt.Errorf("failed to create .passvault directory: %w", err)
	}

	dbPath := filepath.Join(passvaultDir, "passvault.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	masterPasswordTable := `
	CREATE TABLE IF NOT EXISTS master_password (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	passwordsTable := `
	CREATE TABLE IF NOT EXISTS passwords (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT NOT NULL,
		username TEXT NOT NULL,
		encrypted_password TEXT NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(service, username)
	);
	`

	if _, err := DB.Exec(masterPasswordTable); err != nil {
		return fmt.Errorf("failed to create master_password table: %w", err)
	}

	if _, err := DB.Exec(passwordsTable); err != nil {
		return fmt.Errorf("failed to create passwords table: %w", err)
	}

	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func IsMasterPasswordSet() (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM master_password").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check master password: %w", err)
	}
	return count > 0, nil
}

func SetMasterPassword(hashedPassword string) error {
	isSet, err := IsMasterPasswordSet()
	if err != nil {
		return err
	}

	if isSet {
		return fmt.Errorf("master password is already set")
	}

	_, err = DB.Exec("INSERT INTO master_password (id, password_hash) VALUES (1, ?)", hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to set master password: %w", err)
	}

	return nil
}

func GetMasterPasswordHash() (string, error) {
	var hash string
	err := DB.QueryRow("SELECT password_hash FROM master_password WHERE id = 1").Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("master password not set")
		}
		return "", fmt.Errorf("failed to get master password hash: %w", err)
	}
	return hash, nil
}

func UpdateMasterPassword(hashedPassword string) error {
	result, err := DB.Exec("UPDATE master_password SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1", hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update master password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("master password not found")
	}

	return nil
}

func AddPassword(service, username, encryptedPassword, notes string) error {
	_, err := DB.Exec(
		"INSERT INTO passwords (service, username, encrypted_password, notes) VALUES (?, ?, ?, ?)",
		service, username, encryptedPassword, notes,
	)
	if err != nil {
		return fmt.Errorf("failed to add password: %w", err)
	}
	return nil
}

type PasswordEntry struct {
	ID                int
	Service           string
	Username          string
	EncryptedPassword string
	Notes             string
	CreatedAt         string
	UpdatedAt         string
}

func ListAllPasswords() ([]PasswordEntry, error) {
	rows, err := DB.Query("SELECT id, service, username, encrypted_password, notes, created_at, updated_at FROM passwords ORDER BY service, username")
	if err != nil {
		return nil, fmt.Errorf("failed to query passwords: %w", err)
	}
	defer rows.Close()

	var entries []PasswordEntry
	for rows.Next() {
		var entry PasswordEntry
		if err := rows.Scan(&entry.ID, &entry.Service, &entry.Username, &entry.EncryptedPassword, &entry.Notes, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan password entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passwords: %w", err)
	}

	return entries, nil
}

func SearchPasswords(query string) ([]PasswordEntry, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"
	rows, err := DB.Query(
		"SELECT id, service, username, encrypted_password, notes, created_at, updated_at FROM passwords WHERE LOWER(service) LIKE ? OR LOWER(username) LIKE ? OR LOWER(notes) LIKE ? ORDER BY service, username",
		searchPattern, searchPattern, searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search passwords: %w", err)
	}
	defer rows.Close()

	var entries []PasswordEntry
	for rows.Next() {
		var entry PasswordEntry
		if err := rows.Scan(&entry.ID, &entry.Service, &entry.Username, &entry.EncryptedPassword, &entry.Notes, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan password entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passwords: %w", err)
	}

	return entries, nil
}

func UpdatePassword(id int, service, username, encryptedPassword, notes string) error {
	result, err := DB.Exec(
		"UPDATE passwords SET service = ?, username = ?, encrypted_password = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		service, username, encryptedPassword, notes, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password entry not found")
	}

	return nil
}

func DeletePassword(id int) error {
	result, err := DB.Exec("DELETE FROM passwords WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password entry not found")
	}

	return nil
}

func ResetDatabase() error {
	if _, err := DB.Exec("DELETE FROM passwords"); err != nil {
		return fmt.Errorf("failed to delete passwords: %w", err)
	}

	if _, err := DB.Exec("DELETE FROM master_password"); err != nil {
		return fmt.Errorf("failed to delete master password: %w", err)
	}

	return nil
}

func UpdateAllEncryptedPasswords(entries []PasswordEntry) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, entry := range entries {
		_, err := tx.Exec(
			"UPDATE passwords SET encrypted_password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			entry.EncryptedPassword, entry.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update password for %s: %w", entry.Service, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
