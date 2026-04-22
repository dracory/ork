package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"unicode"
)

// ErrWeakPassword is returned when a password does not meet strength requirements.
var ErrWeakPassword = errors.New("password must be at least 8 characters and contain uppercase, lowercase, and digit")

// ValidatePassword checks that a password meets minimum strength requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: too short (%d characters, minimum 8)", ErrWeakPassword, len(password))
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: missing uppercase letter", ErrWeakPassword)
	}
	if !hasLower {
		return fmt.Errorf("%w: missing lowercase letter", ErrWeakPassword)
	}
	if !hasDigit {
		return fmt.Errorf("%w: missing digit", ErrWeakPassword)
	}

	return nil
}

var (
	ErrVaultNotFound = errors.New("vault file not found")
	ErrKeyNotFound   = errors.New("key not found in vault")
)

// Vault represents an encrypted vault
type Vault struct {
	path     string
	password string
	data     map[string]string
	modified bool
}

// Create creates a new vault file. Returns an error if the file already exists
// or the password does not meet strength requirements.
func Create(path string, password string) (*Vault, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("vault file already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check vault file: %w", err)
	}

	v := &Vault{
		path:     path,
		password: password,
		data:     make(map[string]string),
		modified: true,
	}

	if err := v.Save(); err != nil {
		return nil, fmt.Errorf("failed to save new vault: %w", err)
	}

	return v, nil
}

// Open opens an existing vault. Returns ErrVaultNotFound if the file does not exist.
func Open(path string, password string) (*Vault, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to check vault file: %w", err)
	}

	v := &Vault{
		path:     path,
		password: password,
		data:     make(map[string]string),
		modified: false,
	}

	if err := v.load(); err != nil {
		return nil, err
	}

	return v, nil
}

// TryOpen opens a vault, returning (vault, true, nil) on success or (nil, false, nil) on auth failure.
// Other errors (missing file, permission denied, etc.) are returned as (nil, false, err).
func TryOpen(path string, password string) (*Vault, bool, error) {
	vault, err := Open(path, password)
	if err != nil {
		if errors.Is(err, ErrInvalidAuth) || errors.Is(err, ErrInvalidFormat) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return vault, true, nil
}

// Close saves and closes the vault
func (v *Vault) Close() error {
	if v.modified {
		return v.Save()
	}
	return nil
}

// KeyGet retrieves a value from the vault
func (v *Vault) KeyGet(key string) (string, error) {
	if !v.KeyExists(key) {
		return "", ErrKeyNotFound
	}
	return v.data[key], nil
}

// KeySet stores a value in the vault
func (v *Vault) KeySet(key, value string) error {
	v.data[key] = value
	v.modified = true
	return nil
}

// KeyDelete removes a key from the vault
func (v *Vault) KeyDelete(key string) error {
	if !v.KeyExists(key) {
		return ErrKeyNotFound
	}
	delete(v.data, key)
	v.modified = true
	return nil
}

// KeyList returns all keys in the vault
func (v *Vault) KeyList() []string {
	keys := make([]string, 0, len(v.data))
	for key := range v.data {
		keys = append(keys, key)
	}
	return keys
}

// KeyExists checks if a key exists
func (v *Vault) KeyExists(key string) bool {
	_, exists := v.data[key]
	return exists
}

// Save writes the vault to disk
func (v *Vault) Save() error {
	// Serialize data to JSON
	data, err := json.Marshal(v.data)
	if err != nil {
		return fmt.Errorf("failed to serialize vault data: %w", err)
	}

	// Encrypt data
	encrypted, err := encrypt(data, v.password)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	// Write to file
	if err := os.WriteFile(v.path, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	v.modified = false
	return nil
}

// ChangePassword changes the vault password and re-encrypts with new password.
// Returns an error if the new password does not meet strength requirements.
func (v *Vault) ChangePassword(newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	v.password = newPassword
	v.modified = true
	return v.Save()
}

// load reads the vault from disk and decrypts it
func (v *Vault) load() error {
	// Read vault file
	data, err := os.ReadFile(v.path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrVaultNotFound
		}
		return fmt.Errorf("failed to read vault file: %w", err)
	}

	// Decrypt data
	vaultData, err := decrypt(string(data), v.password)
	if err != nil {
		return err
	}

	// Update in-memory data
	v.data = vaultData
	v.modified = false
	return nil
}
