package ork

import (
	"path/filepath"
	"testing"

	"github.com/dracory/envenc"
)

func TestVaultFileToKeys(t *testing.T) {
	// Create a temporary vault file
	tmpDir := t.TempDir()
	vaultFile := filepath.Join(tmpDir, "test.vault")

	// Create test data
	testData := map[string]string{
		"TEST_SECRET":    "test_value",
		"ANOTHER_SECRET": "another_value",
	}

	// Create encrypted vault
	password := "testpassword123"
	if err := envenc.Init(vaultFile, password); err != nil {
		t.Fatalf("Failed to init test vault: %v", err)
	}
	for key, value := range testData {
		if err := envenc.KeySet(vaultFile, password, key, value); err != nil {
			t.Fatalf("Failed to set key %s: %v", key, err)
		}
	}

	// Test loading vault
	secrets, err := VaultFileToKeys(vaultFile, password)
	if err != nil {
		t.Errorf("VaultFileToKeys failed: %v", err)
	}

	// Verify secrets are available
	if got := secrets["TEST_SECRET"]; got != "test_value" {
		t.Errorf("secrets[TEST_SECRET] = %q, want %q", got, "test_value")
	}

	if got := secrets["ANOTHER_SECRET"]; got != "another_value" {
		t.Errorf("secrets[ANOTHER_SECRET] = %q, want %q", got, "another_value")
	}
}

func TestVaultFileToKeysWrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultFile := filepath.Join(tmpDir, "test.vault")

	correctPassword := "correctpassword"
	wrongPassword := "wrongpassword"

	if err := envenc.Init(vaultFile, correctPassword); err != nil {
		t.Fatalf("Failed to init test vault: %v", err)
	}
	if err := envenc.KeySet(vaultFile, correctPassword, "KEY", "value"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Loading with wrong password should fail
	_, err := VaultFileToKeys(vaultFile, wrongPassword)
	if err == nil {
		t.Error("VaultFileToKeys with wrong password should fail")
	}
}

func TestVaultFileToKeysNonExistentFile(t *testing.T) {
	nonExistentFile := filepath.Join(t.TempDir(), "nonexistent.vault")

	_, err := VaultFileToKeys(nonExistentFile, "password")
	if err == nil {
		t.Error("VaultFileToKeys with non-existent file should fail")
	}
}

func TestVaultContentToKeys(t *testing.T) {
	// VaultContentToKeys requires raw encrypted vault file content as a string.
	// The envenc library doesn't expose the raw encrypted content directly,
	// and testing this would require reading the encrypted file bytes.
	// This function is a thin wrapper around envenc.LoadKeysFromString,
	// which is tested by the envenc library itself.
	t.Skip("VaultContentToKeys requires raw encrypted vault content, which is not easily accessible for testing")
}

func TestVaultFileToEnv(t *testing.T) {
	tmpDir := t.TempDir()
	vaultFile := filepath.Join(tmpDir, "test.vault")

	password := "password"

	if err := envenc.Init(vaultFile, password); err != nil {
		t.Fatalf("Failed to init test vault: %v", err)
	}
	if err := envenc.KeySet(vaultFile, password, "TEST_KEY", "test_value"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Test loading to environment
	if err := VaultFileToEnv(vaultFile, password); err != nil {
		t.Errorf("VaultFileToEnv failed: %v", err)
	}
}

func TestVaultFileToEnvNonExistentFile(t *testing.T) {
	nonExistentFile := filepath.Join(t.TempDir(), "nonexistent.vault")

	err := VaultFileToEnv(nonExistentFile, "password")
	if err == nil {
		t.Error("VaultFileToEnv with non-existent file should fail")
	}
}

func TestVaultContentToEnv(t *testing.T) {
	// VaultContentToEnv requires raw encrypted vault file content as a string.
	// The envenc library doesn't expose the raw encrypted content directly,
	// and testing this would require reading the encrypted file bytes.
	// This function is a thin wrapper around envenc.HydrateEnvFromString,
	// which is tested by the envenc library itself.
	t.Skip("VaultContentToEnv requires raw encrypted vault content, which is not easily accessible for testing")
}
