# Internal Vault Implementation

**Status:** Implemented
**Created:** 2026-04-22
**Author:** System

**Note:** Implementation complete. Single base64 string format with 80-char wrapping, AES-256-GCM + Argon2id encryption, gzip compression always enabled, and interactive web UI implemented.

## Problem Statement

Ork currently depends on the external `envenc` library (github.com/dracory/envenc v1.4.1) for vault functionality. This creates several issues:

- **External dependency** - Adds another dependency to the project
- **Tight coupling** - `vault.go` is a thin wrapper around envenc's API
- **Poor API design** - Functional API with multiple variants (ToKeys, ToEnv, WithPrompt) is confusing
- **Limited functionality** - Basic key-value storage only, no advanced features
- **Operational overhead** - Requires separate CLI tool to create vault files
- **Testing limitations** - Difficult to test content-based functions due to envenc not exposing raw encrypted content
- **No team features** - No access control, audit trails, or secret sharing
- Single password model - One password for entire vault, no per-secret encryption
- Weak crypto - envenc uses simple password fortification instead of Argon2id, no authentication tag
- No web UI - envenc has no built-in web interface for interactive vault management

## Proposed Solution

Build an internal vault system within Ork with a clean, idiomatic Go API designed from scratch. This is a breaking change - no backward compatibility with envenc vault files or API. Focus on best practices: simple API, secure defaults, and extensibility.

## Web UI

Interactive web-based UI for vault management:

```bash
ork vault ui <vault-file> [address]
```

**Features:**
- Login with vault password
- List all keys with truncated values
- Add new keys via modal dialog
- Edit existing keys via modal dialog
- Delete keys with confirmation dialog
- Success/failure notifications via Notiflix
- Vue.js 3 reactive frontend
- RESTful JSON API backend

**Technology Stack:**
- Vue 3 (CDN) for reactive UI
- Notiflix for toast notifications
- Pure CSS for styling (no external CSS framework)
- Go HTTP handlers for API endpoints

## Core Design

### Encryption Strategy

Use standard Go crypto libraries (golang.org/x/crypto) for encryption:

- **Algorithm**: AES-256-GCM for authenticated encryption
- **Key derivation**: Argon2id for password-based key derivation (memory-hard, resistant to GPU/ASIC attacks)
- **File format**: Single base64-encoded string (opaque, no visible structure)
- **Compression**: Gzip compression always enabled for all vault data

### API Design

Clean, idiomatic Go API designed from scratch without envenc compatibility constraints:

```go
// Vault represents an encrypted vault
type Vault struct {
    path     string
    password string
    data     map[string]string
}

// Open opens an existing vault or creates a new one
func Open(path string, password string) (*Vault, error)

// TryOpen opens a vault, returning (vault, true, nil) on success or (nil, false, nil) on auth failure
// Cleaner API for UI flows where you want to distinguish auth errors from other errors
func TryOpen(path string, password string) (*Vault, bool, error)

// Close saves and closes the vault
func (v *Vault) Close() error

// Get retrieves a value from the vault
func (v *Vault) Get(key string) (string, error)

// Set stores a value in the vault
func (v *Vault) Set(key, value string) error

// Delete removes a key from the vault
func (v *Vault) Delete(key string) error

// List returns all keys in the vault
func (v *Vault) List() []string

// Exists checks if a key exists
func (v *Vault) Exists(key string) bool

// Save writes the vault to disk
func (v *Vault) Save() error

// ChangePassword changes the vault password and re-encrypts with new password
func (v *Vault) ChangePassword(newPassword string) error
```

### File Format

Single base64-encoded string wrapped to 80 characters (opaque format for security through obscurity):

```
aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890abcdef==
```

- **Concatenated data**: header + salt + nonce + params + encrypted
- **Header**: 4-byte magic number (binary, not base64)
- **Salt**: 16-byte random salt
- **Nonce**: 12-byte random nonce
- **Params**: 12-byte Argon2 parameters (binary encoded)
- **Encrypted**: GCM encrypted data (gzip compressed JSON)
- **Entire blob**: Base64 encoded to single string
- **Line wrapping**: Wrapped to 80 characters for git diff readability and copy-paste
- **No visible structure**: Attacker sees only wrapped base64 string

### Internal Structure

```go
package vault

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "golang.org/x/crypto/argon2"
)

const (
    magicNumber    = [4]byte{'O', 'R', 'K', 'V'}
    lineWrapWidth = 80
    saltSize       = 16
    nonceSize      = 12
    paramsSize     = 12
    keySize        = 32 // AES-256
)

type VaultFile struct {
    Data string // Single base64-encoded string
}

// VaultParams represents Argon2 parameters
type VaultParams struct {
    TimeCost    uint32
    MemoryCost  uint32
    Parallelism uint32
}

func deriveKey(password string, salt []byte) []byte {
    return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, keySize)
}

func encrypt(data []byte, password string) (string, error) {
    // Generate salt and nonce
    // Derive key using Argon2id
    // Encrypt using AES-256-GCM
    // Concatenate: magic + salt + nonce + params + encrypted
    // Base64 encode entire blob
    // Return single string
}

func decrypt(data string, password string) (map[string]string, error) {
    // Base64 decode entire blob
    // Parse: magic + salt + nonce + params + encrypted
    // Derive key using Argon2id
    // Decrypt using AES-256-GCM
    // Verify auth tag
    // Return key-value map
}
```

## Usage

### Programmatic API

```go
// Open or create vault
vault, err := vault.Open("secrets.vault", "my-password")
if err != nil {
    log.Fatal(err)
}
defer vault.Close()

// Set secrets
vault.Set("DB_HOST", "db.example.com")
vault.Set("DB_PASSWORD", "secret123")
vault.Set("API_KEY", "abc123xyz")

// Get secrets
dbHost, err := vault.Get("DB_HOST")
dbPassword, err := vault.Get("DB_PASSWORD")

// Check if key exists
if vault.Exists("API_KEY") {
    apiKey, _ := vault.Get("API_KEY")
}

// List all keys
keys := vault.List()

// Delete a key
vault.Delete("OLD_SECRET")

// Change password (re-encrypts vault)
if err := vault.ChangePassword("new-secure-password"); err != nil {
    log.Fatal(err)
}

// Save to disk
if err := vault.Save(); err != nil {
    log.Fatal(err)
}

// TryOpen for UI flows (distinguish auth failure from other errors)
vault, ok, err := vault.TryOpen("secrets.vault", "password")
if err != nil {
    // File error, permission error, etc.
    log.Fatal(err)
}
if !ok {
    // Wrong password
    fmt.Println("Invalid password")
    return
}
// Vault opened successfully
```

### CLI (ork vault subcommand)

```bash
# Create new vault (password read from stdin, never command-line args)
ork vault init secrets.vault
# Prompts: Enter password: (no echo)
# Prompts: Confirm password:

# Set secrets
ork vault set secrets.vault DB_HOST db.example.com
# Prompts: Enter vault password:

ork vault set secrets.vault DB_PASSWORD secret123
# Prompts: Enter vault password:

# Get secrets
ork vault get secrets.vault DB_HOST
# Prompts: Enter vault password:

# List all keys
ork vault list secrets.vault
# Prompts: Enter vault password:

# Delete a key
ork vault delete secrets.vault OLD_SECRET
# Prompts: Enter vault password:

# Change password
ork vault changepassword secrets.vault
# Prompts: Enter current password:
# Prompts: Enter new password:
# Prompts: Confirm new password:

# For automation, password can be read from file descriptor
ork vault get secrets.vault DB_HOST < password.txt
# OR
echo "password" | ork vault get secrets.vault DB_HOST
```

## Implementation Plan

### Phase 1: Core Encryption
1. Implement `deriveKey()` using Argon2id
2. Implement `encrypt()` and `decrypt()` using AES-256-GCM
3. Define text-based file format with base64 encoding
4. Add unit tests for encryption/decryption

### Phase 2: Vault API
1. Implement `Vault` struct with `Open()`, `TryOpen()`, `Close()`, `Save()`
2. Implement `Get()`, `Set()`, `Delete()`, `List()`, `Exists()`, `ChangePassword()` methods
3. Add file read/write operations
4. Add unit tests for Vault operations

### Phase 3: Integration
1. Delete existing `vault.go` (envenc wrapper)
2. Delete existing `vault_test.go`
3. Create new `vault/vault.go` with internal implementation
4. Create new `vault/vault_test.go`
5. Update documentation to reflect new API
6. Remove envenc dependency from go.mod

### Phase 4: CLI Tool (Optional)
1. Create `ork vault` subcommand for vault management
2. Commands: init, set, get, delete, list, changepassword
3. Interactive password prompts (read from stdin/file descriptor, never command-line args)
4. Integrated into main ork binary (no separate distribution)

## Benefits

- **No external dependency** - Reduces dependency tree
- **Full control** - Can customize encryption parameters and features
- **Better testing** - Full control over internal implementation
- **Simpler deployment** - No need to install envenc CLI
- **Extensible** - Easy to add features like per-secret encryption, versioning, etc.
- **Standard crypto** - Uses well-vetted Go crypto libraries
- **Clean design** - Breaking change allows optimal design without legacy constraints
- **Password rotation** - ChangePassword() method for secure password updates without re-creating vault
- Cleaner UI flows - TryOpen() distinguishes auth failures from other errors
- Web UI included - Built-in interactive vault management interface (no external tool needed)
- **Secure CLI** - Passwords never in command-line args, only stdin/file descriptor
- **Security through obscurity** - Single base64 string reveals no structure or metadata
- **Git-diff friendly** - 80-char line wrapping enables git diff and copy-paste convenience

## Trade-offs

- **Maintenance burden** - Need to maintain crypto code
- **Security audit** - Need to ensure implementation is secure
- **Breaking change** - Users must recreate vault files with new format
- **API change** - Existing code using envenc wrapper functions must be updated


## Security Considerations

- Use Argon2id with appropriate parameters (time=3, memory=64MB, parallelism=4)
- Always use random salts and nonces
- Verify GCM authentication tag before decryption
- Zero out sensitive data from memory when done
- Never log passwords or secret values
- Use constant-time comparison for password verification (if adding password check)
- ChangePassword should generate new salt and nonce, re-encrypt all data
- CLI never reads passwords from command-line args (visible in process list, shell history)
- CLI reads passwords from stdin or file descriptor only

## Future Enhancements

- [x] Web UI for interactive vault management
- [ ] Per-secret encryption with different keys
- [ ] Secret versioning and history
- [ ] Binary secret support (certificates, keys)
- [ ] Vault file integrity verification (HMAC)
- [x] Compression for large vaults (gzip always enabled)
- [ ] Remote vault support (HTTP, S3)

## Open Questions

1. What should the default Argon2 parameters be?
2. Should compression be enabled by default?

## Success Metrics

- [x] New vault tests pass with comprehensive coverage
- [x] No security vulnerabilities in crypto implementation
- [x] Performance comparable to or better than envenc
- [x] API is clean, idiomatic Go with clear semantics
- [x] Web UI renders correctly with Vue.js
- [x] UI tests pass for HTML/CSS/JS generation
- [x] Notifications work via Notiflix
