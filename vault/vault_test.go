package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVaultCreate(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create new vault
	v, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()

	// Verify file exists immediately
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	// Creating again should fail
	_, err = Create(vaultPath, password)
	if err == nil {
		t.Fatal("Expected error when creating vault that already exists")
	}
}

func TestVaultOpenNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "nonexistent.vault")
	password := "TestPass123"

	// Open missing vault should return ErrVaultNotFound
	_, err := Open(vaultPath, password)
	if err == nil {
		t.Fatal("Expected error when opening non-existent vault")
	}
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("Expected ErrVaultNotFound, got %v", err)
	}
}

func TestVaultOpenCreate(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create new vault
	vault, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer vault.Close()

	// Add some data
	if err := vault.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault.KeySet("KEY2", "value2"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Save
	if err := vault.Save(); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}
}

func TestVaultOpenLoad(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create and save vault
	vault1, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	if err := vault1.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault1.Save(); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}
	if err := vault1.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// Open existing vault
	vault2, err := Open(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}
	defer vault2.Close()

	// Verify data
	value, err := vault2.KeyGet("KEY1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
}

func TestVaultWrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	correctPassword := "CorrectPass123"
	wrongPassword := "WrongPass123"

	// Create and save vault
	vault1, err := Create(vaultPath, correctPassword)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault1.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault1.Save(); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}
	if err := vault1.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// Try to open with wrong password
	_, err = Open(vaultPath, wrongPassword)
	if err == nil {
		t.Fatal("Expected error when opening with wrong password")
	}
	if !errors.Is(err, ErrInvalidAuth) {
		t.Errorf("Expected ErrInvalidAuth, got %v", err)
	}
}

func TestVaultTryOpen(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	correctPassword := "CorrectPass123"
	wrongPassword := "WrongPass123"

	// Create and save vault
	vault1, err := Create(vaultPath, correctPassword)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault1.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault1.Save(); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}
	if err := vault1.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// TryOpen with correct password
	vault, ok, err := TryOpen(vaultPath, correctPassword)
	if err != nil {
		t.Fatalf("TryOpen with correct password failed: %v", err)
	}
	if !ok {
		t.Fatal("TryOpen with correct password returned ok=false")
	}
	if vault == nil {
		t.Fatal("TryOpen with correct password returned nil vault")
	}
	vault.Close()

	// TryOpen with wrong password
	vault, ok, err = TryOpen(vaultPath, wrongPassword)
	if err != nil {
		t.Fatalf("TryOpen with wrong password returned error: %v", err)
	}
	if ok {
		t.Fatal("TryOpen with wrong password returned ok=true")
	}
	if vault != nil {
		t.Fatal("TryOpen with wrong password returned non-nil vault")
	}
}

func TestVaultGetSetDelete(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	vault, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer vault.Close()

	// Set key
	if err := vault.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Get key
	value, err := vault.KeyGet("KEY1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	// Delete key
	if err := vault.KeyDelete("KEY1"); err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// Verify deleted
	_, err = vault.KeyGet("KEY1")
	if err == nil {
		t.Fatal("Expected error when getting deleted key")
	}
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestVaultExists(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	vault, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer vault.Close()

	// Key doesn't exist
	if vault.KeyExists("KEY1") {
		t.Fatal("Exists returned true for non-existent key")
	}

	// Set key
	if err := vault.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Key exists
	if !vault.KeyExists("KEY1") {
		t.Fatal("Exists returned false for existing key")
	}
}

func TestVaultList(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	vault, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer vault.Close()

	// Empty vault
	keys := vault.KeyList()
	if len(keys) != 0 {
		t.Errorf("Expected empty list, got %d keys", len(keys))
	}

	// Add keys
	if err := vault.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault.KeySet("KEY2", "value2"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault.KeySet("KEY3", "value3"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// List keys
	keys = vault.KeyList()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestVaultChangePassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	oldPassword := "OldPass123"
	newPassword := "NewPass123"

	// Create vault with old password
	vault1, err := Create(vaultPath, oldPassword)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault1.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault1.Save(); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}
	if err := vault1.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// Open with old password and change
	vault2, err := Open(vaultPath, oldPassword)
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}
	if err := vault2.ChangePassword(newPassword); err != nil {
		t.Fatalf("Failed to change password: %v", err)
	}
	if err := vault2.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// Old password should not work
	_, err = Open(vaultPath, oldPassword)
	if err == nil {
		t.Fatal("Expected error with old password after change")
	}

	// New password should work
	vault3, err := Open(vaultPath, newPassword)
	if err != nil {
		t.Fatalf("Failed to open vault with new password: %v", err)
	}
	defer vault3.Close()

	// Verify data
	value, err := vault3.KeyGet("KEY1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
}

func TestVaultClose(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create vault
	vault, err := Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	// Re-open existing vault, modify, and close
	vault, err = Open(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}
	if err := vault.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := vault.Close(); err != nil {
		t.Fatalf("Failed to close vault with modifications: %v", err)
	}

	// Verify data persisted
	vault, err = Open(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to open vault: %v", err)
	}
	defer vault.Close()

	value, err := vault.KeyGet("KEY1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"Strong1A", false},
		{"TestPass123", false},
		{"short1A", true}, // too short
		{"onlylowercase", true},
		{"ONLYUPPERCASE", true},
		{"12345678", true},
		{"NoDigitsHere", true},
		{"", true},
		{"123", true},
		{"a", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePassword(%q) expected error, got nil", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePassword(%q) unexpected error: %v", tt.password, err)
			}
		})
	}
}

func TestCreateWeakPassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")

	_, err := Create(vaultPath, "weak")
	if err == nil {
		t.Fatal("Expected error for weak password, got nil")
	}
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("Expected ErrWeakPassword, got %v", err)
	}
}

func TestChangePasswordWeak(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")

	v, err := Create(vaultPath, "Strong1A")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()

	err = v.ChangePassword("weak")
	if err == nil {
		t.Fatal("Expected error for weak password, got nil")
	}
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("Expected ErrWeakPassword, got %v", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	data := []byte(`{"KEY1":"value1","KEY2":"value2"}`)
	password := "TestPass123"

	// Encrypt
	encrypted, err := encrypt(data, password)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Decrypt
	decrypted, err := decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify data
	if decrypted["KEY1"] != "value1" {
		t.Errorf("Expected value1, got %s", decrypted["KEY1"])
	}
	if decrypted["KEY2"] != "value2" {
		t.Errorf("Expected value2, got %s", decrypted["KEY2"])
	}
}

func TestEncryptDecryptWrongPassword(t *testing.T) {
	data := []byte(`{"KEY1":"value1"}`)
	correctPassword := "CorrectPass123"
	wrongPassword := "WrongPass123"

	// Encrypt with correct password
	encrypted, err := encrypt(data, correctPassword)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Decrypt with wrong password
	_, err = decrypt(encrypted, wrongPassword)
	if err == nil {
		t.Fatal("Expected error when decrypting with wrong password")
	}
	if !errors.Is(err, ErrInvalidAuth) {
		t.Errorf("Expected ErrInvalidAuth, got %v", err)
	}
}

func TestInvalidFormat(t *testing.T) {
	// Invalid base64 string
	invalidData := "invalid-base64!!!"

	_, err := decrypt(invalidData, "password")
	if err == nil {
		t.Fatal("Expected error with invalid format")
	}
	// Should get an error (invalid format)
}

func TestCompression(t *testing.T) {
	// Create JSON data
	vaultData := make(map[string]string)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("KEY_%d", i)
		value := strings.Repeat("VALUE_", 50) // ~300 bytes per entry
		vaultData[key] = value
	}
	data, err := json.Marshal(vaultData)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	password := "TestPass123"

	// Encrypt (always compresses)
	encrypted, err := encrypt(data, password)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Decrypt
	decrypted, err := decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify data
	if len(decrypted) != len(vaultData) {
		t.Errorf("Expected %d keys, got %d", len(vaultData), len(decrypted))
	}
	// Verify one key
	if decrypted["KEY_0"] != vaultData["KEY_0"] {
		t.Errorf("Data mismatch")
	}
}

func TestCRLFHandling(t *testing.T) {
	data := []byte(`{"KEY1":"value1"}`)
	password := "TestPass123"

	// Encrypt
	encrypted, err := encrypt(data, password)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Simulate Windows CRLF editing by replacing LF with CRLF
	encryptedCRLF := strings.ReplaceAll(encrypted, "\n", "\r\n")

	// Decrypt should still work despite CRLF line endings
	decrypted, err := decrypt(encryptedCRLF, password)
	if err != nil {
		t.Fatalf("Failed to decrypt CRLF data: %v", err)
	}

	if decrypted["KEY1"] != "value1" {
		t.Errorf("Expected value1, got %s", decrypted["KEY1"])
	}
}
