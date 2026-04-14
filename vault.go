package ork

import (
	"fmt"
	"os"

	"github.com/dracory/envenc"
)

// VaultContentToKeys loads keys from a vault content string.
func VaultContentToKeys(vaultContent, vaultPassword string) (keys map[string]string, err error) {
	return envenc.LoadKeysFromString(vaultContent, vaultPassword)
}

// VaultFileToKeys loads keys from a vault file.
func VaultFileToKeys(vaultFilePath string, vaultPassword string) (keys map[string]string, err error) {
	if _, err = os.Stat(vaultFilePath); err != nil {
		return nil, fmt.Errorf("vault file not found: %w", err)
	}
	return envenc.LoadKeysFromFile(vaultFilePath, vaultPassword)
}

// VaultContentToEnv hydrates environment variables from a vault content string.
func VaultContentToEnv(vaultContent, vaultPassword string) (err error) {
	return envenc.HydrateEnvFromString(vaultContent, vaultPassword)
}

// VaultFileToEnv decrypts vault and hydrates environment variables.
func VaultFileToEnv(vaultFilePath, vaultPassword string) (err error) {
	if _, err = os.Stat(vaultFilePath); err != nil {
		return fmt.Errorf("vault file not found: %w", err)
	}

	if err = envenc.HydrateEnvFromFile(vaultFilePath, vaultPassword); err != nil {
		return fmt.Errorf("failed to load vault: %w", err)
	}

	return nil
}

// VaultFileToKeysWithPrompt prompts for password and loads keys from a vault file.
func VaultFileToKeysWithPrompt(vaultFilePath string) (keys map[string]string, err error) {
	password, err := PromptPassword("Vault password: ")
	if err != nil {
		return nil, err
	}
	return VaultFileToKeys(vaultFilePath, password)
}

// VaultFileToEnvWithPrompt prompts for password and hydrates environment variables from a vault file.
func VaultFileToEnvWithPrompt(vaultFilePath string) (err error) {
	password, err := PromptPassword("Vault password: ")
	if err != nil {
		return err
	}
	return VaultFileToEnv(vaultFilePath, password)
}

// VaultContentToKeysWithPrompt prompts for password and loads keys from a vault content string.
func VaultContentToKeysWithPrompt(vaultContent string) (keys map[string]string, err error) {
	password, err := PromptPassword("Vault password: ")
	if err != nil {
		return nil, err
	}
	return VaultContentToKeys(vaultContent, password)
}

// VaultContentToEnvWithPrompt prompts for password and hydrates environment variables from a vault content string.
func VaultContentToEnvWithPrompt(vaultContent string) (err error) {
	password, err := PromptPassword("Vault password: ")
	if err != nil {
		return err
	}
	return VaultContentToEnv(vaultContent, password)
}
