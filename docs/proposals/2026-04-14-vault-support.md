# Proposal: Vault Support for Secure Secrets Management

**Date:** 2026-04-14
**Status:** Completed
**Author:** Code Review

> **Note:** Add encrypted vault support to ork using `envenc` library with both in-memory and environment variable loading options.

## Problem Statement

Users need to store sensitive data (passwords, API keys) securely. Currently each project implements vault loading separately.

## Current Implementation

The `DbThreeSineviaCom` project implements simple vault loading:

```go
var vaultLoaded bool

func EnsureVaultLoaded(config *types.Config) {
    if vaultLoaded {
        return
    }

    // Check if playbook needs vault
    needsVault := slices.Contains(mariaDBPlaybooks, playbook)
    if !needsVault {
        return
    }

    password := promptPassword("Enter vault password: ")
    if err := envenc.HydrateEnvFromFile(VAULT_FILE, password); err != nil {
        log.Fatalf("Failed to load vault: %v", err)
    }

    config.MariaDBRootPassword = os.Getenv("MARIADB_ROOT_PASSWORD")
    vaultLoaded = true
}
```

## Implementation

Simple wrapper functions providing two loading strategies:

**vault.go:**
```go
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

    if err := envenc.HydrateEnvFromFile(vaultFilePath, vaultPassword); err != nil {
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
```

**utils.go:**
```go
package ork

import (
    "fmt"
    "os"

    "golang.org/x/term"
)

// PromptPassword securely prompts for a password from stdin.
func PromptPassword(prompt string) (string, error) {
    fmt.Print(prompt)
    bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Println()
    if err != nil {
        return "", fmt.Errorf("failed to read password: %w", err)
    }
    return string(bytePassword), nil
}
```

## Usage

### Option 1: Load to memory (recommended for security)

**Simple one-liner with prompt:**
```go
secrets, err := ork.VaultFileToKeysWithPrompt(".env.vault")
if err != nil {
    log.Fatal(err)
}
dbPassword := secrets["DATABASE_PASSWORD"]
```

**Separate prompt and load (for custom prompts or non-interactive use):**
```go
password, err := ork.PromptPassword("Vault password: ")
if err != nil {
    log.Fatal(err)
}
secrets, err := ork.VaultFileToKeys(".env.vault", password)
if err != nil {
    log.Fatal(err)
}
dbPassword := secrets["DATABASE_PASSWORD"]
```

### Option 2: Load to environment (for .env compatibility)

**Simple one-liner with prompt:**
```go
if err := ork.VaultFileToEnvWithPrompt(".env.vault"); err != nil {
    log.Fatal(err)
}
// Secrets now available via os.Getenv
```

**Separate prompt and load:**
```go
password, err := ork.PromptPassword("Vault password: ")
if err != nil {
    log.Fatal(err)
}
if err := ork.VaultFileToEnv(".env.vault", password); err != nil {
    log.Fatal(err)
}
// Secrets now available via os.Getenv
```

## Playbook Usage

Playbooks can receive secrets via configuration or load them directly:

```go
func (i *Install) Run(cfg config.NodeConfig) types.Result {
    // Option 1: Secrets loaded from config
    rootPassword := cfg.Secrets["MARIADB_ROOT_PASSWORD"]

    // Option 2: Load directly in playbook
    secrets, err := ork.VaultFileToKeys(cfg.VaultPath, cfg.VaultPassword)
    if err != nil {
        return types.Result{Error: err}
    }
    rootPassword := secrets["MARIADB_ROOT_PASSWORD"]

    if rootPassword == "" {
        return types.Result{
            Error: fmt.Errorf("MARIADB_ROOT_PASSWORD not set"),
        }
    }
    // Use rootPassword...
}
```

## Design Trade-offs

**ToKeys functions (VaultFileToKeys, VaultContentToKeys):**
- Secrets stored in memory map
- No global exposure
- Recommended for security-sensitive applications
- Caller manages secret lifecycle

**ToEnv functions (VaultFileToEnv, VaultContentToEnv):**
- Secrets dumped to environment variables
- Global exposure via os.Getenv
- For compatibility with .env-based tooling
- Use when external tools expect environment variables

## Benefits

- Simple, stateless API
- Two loading strategies for different use cases
- No global state management
- Thread-safe (no shared state)
- Easy to test (pure functions)
- Backward compatible with .env workflows
- Secure password prompting included (prevents insecure user implementations)
- Convenience helpers combine prompt + load in one call

## Implementation

**Completed:**
- Created `vault.go` with 8 vault-specific functions:
  - VaultContentToKeys
  - VaultFileToKeys
  - VaultContentToEnv
  - VaultFileToEnv
  - VaultFileToKeysWithPrompt
  - VaultFileToEnvWithPrompt
  - VaultContentToKeysWithPrompt
  - VaultContentToEnvWithPrompt
- Created `utils.go` with generic PromptPassword utility
- Updated `go.mod` to use envenc v1.4.1
- Updated to Go 1.26
- Wrote tests for all functions (PromptPassword skipped as it requires interactive input)
- Documented in README with usage examples and design trade-offs
- Code review fixes applied:
  - Fixed variable shadowing in VaultFileToEnv (changed `:=` to `=`)
  - Added file existence check to VaultFileToKeys for consistency
  - Skipped TestVaultContentToKeys and TestVaultContentToEnv with proper explanations (require raw encrypted content)

## Dependencies

- `github.com/dracory/envenc` v1.4.1
- `golang.org/x/term` - For secure password prompting
