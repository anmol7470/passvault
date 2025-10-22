package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters for master password hashing
	hashTime    = 1
	hashMemory  = 64 * 1024 // 64 MB
	hashThreads = 4
	hashKeyLen  = 32

	// Argon2 parameters for encryption key derivation
	encTime    = 3
	encMemory  = 64 * 1024 // 64 MB
	encThreads = 4
	encKeyLen  = 32

	saltLen = 16
)

func HashMasterPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate a random salt
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey([]byte(password), salt, hashTime, hashMemory, hashThreads, hashKeyLen)

	// Combine salt + hash and encode as base64
	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

func VerifyMasterPassword(password, encodedHash string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if encodedHash == "" {
		return errors.New("hash cannot be empty")
	}

	// Decode the base64 hash
	combined, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	if len(combined) != saltLen+hashKeyLen {
		return errors.New("invalid hash format")
	}

	// Extract salt and hash
	salt := combined[:saltLen]
	storedHash := combined[saltLen:]

	// Hash the provided password with the same salt
	computedHash := argon2.IDKey([]byte(password), salt, hashTime, hashMemory, hashThreads, hashKeyLen)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(storedHash, computedHash) != 1 {
		return errors.New("invalid password")
	}

	return nil
}

// EncryptPassword encrypts a password using AES-256-GCM with a key derived from the master password
func EncryptPassword(password, masterPassword string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	if masterPassword == "" {
		return "", errors.New("master password cannot be empty")
	}

	// Generate a random salt for key derivation
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from master password using Argon2id
	key := argon2.IDKey([]byte(masterPassword), salt, encTime, encMemory, encThreads, encKeyLen)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the password
	ciphertext := gcm.Seal(nil, nonce, []byte(password), nil)

	// Combine salt + nonce + ciphertext and encode as base64
	combined := append(salt, nonce...)
	combined = append(combined, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptPassword decrypts a password using AES-256-GCM with a key derived from the master password
func DecryptPassword(encryptedPassword, masterPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", errors.New("encrypted password cannot be empty")
	}
	if masterPassword == "" {
		return "", errors.New("master password cannot be empty")
	}

	// Decode the base64 encrypted password
	combined, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted password: %w", err)
	}

	// Minimum length check: salt (16) + nonce (12) + at least some ciphertext + tag (16)
	if len(combined) < saltLen+12+16 {
		return "", errors.New("invalid encrypted password format")
	}

	// Extract salt
	salt := combined[:saltLen]
	remaining := combined[saltLen:]

	// Derive encryption key from master password using the same salt
	key := argon2.IDKey([]byte(masterPassword), salt, encTime, encMemory, encThreads, encKeyLen)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(remaining) < nonceSize {
		return "", errors.New("invalid encrypted password format")
	}

	// Extract nonce and ciphertext
	nonce := remaining[:nonceSize]
	ciphertext := remaining[nonceSize:]

	// Decrypt the password
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
