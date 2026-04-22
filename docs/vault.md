# Vault (Secure Secrets Management)

Ork provides secure vault support for managing sensitive data like passwords and API keys. The vault uses AES-256-GCM encryption with Argon2id key derivation.

## Usage

### Programmatic API

```go
import "github.com/dracory/ork/vault"

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

## API Reference

### Vault Struct

```go
type Vault struct {
    // Internal fields
}
```

### Functions

#### Open
```go
func Open(path string, password string) (*Vault, error)
```
Opens an existing vault or creates a new one. Returns an error if the file cannot be read or if authentication fails.

#### TryOpen
```go
func TryOpen(path string, password string) (*Vault, bool, error)
```
Opens a vault, returning (vault, true, nil) on success or (nil, false, nil) on auth failure. Cleaner API for UI flows where you want to distinguish auth errors from other errors.

### Methods

#### Close
```go
func (v *Vault) Close() error
```
Saves and closes the vault. Only writes to disk if modifications were made.

#### Get
```go
func (v *Vault) Get(key string) (string, error)
```
Retrieves a value from the vault. Returns `ErrKeyNotFound` if the key doesn't exist.

#### Set
```go
func (v *Vault) Set(key, value string) error
```
Stores a value in the vault. Marks the vault as modified.

#### Delete
```go
func (v *Vault) Delete(key string) error
```
Removes a key from the vault. Returns `ErrKeyNotFound` if the key doesn't exist.

#### List
```go
func (v *Vault) List() []string
```
Returns all keys in the vault.

#### Exists
```go
func (v *Vault) Exists(key string) bool
```
Checks if a key exists in the vault.

#### Save
```go
func (v *Vault) Save() error
```
Writes the vault to disk. Called automatically by `Close()` if modifications were made.

#### ChangePassword
```go
func (v *Vault) ChangePassword(newPassword string) error
```
Changes the vault password and re-encrypts with the new password. Generates new salt and nonce.

## Security

- **Encryption**: AES-256-GCM with authenticated encryption
- **Key derivation**: Argon2id (time=3, memory=64MB, parallelism=4)
- **Random salts**: Unique salt per vault prevents rainbow table attacks
- **Random nonces**: Unique nonce per encryption prevents nonce reuse
- **Authentication**: GCM auth tag verifies data integrity
- **Compression**: Gzip compression always enabled for all vault data
- **CLI security**: Passwords never in command-line args (visible in process list, shell history)
- **CLI password input**: Read from stdin or file descriptor only

## File Format

The vault uses a single base64-encoded string wrapped to 80 characters (opaque format for security through obscurity):

```
aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890abcdef==
```

- **Concatenated data**: Binary magic number + salt + nonce + params + encrypted
- **Magic number**: 4-byte binary marker (embedded, not visible)
- **Salt**: 16-byte random salt
- **Nonce**: 12-byte random nonce
- **Argon2 params**: 12-byte Argon2 parameters (binary encoded)
- **Encrypted data**: GCM encrypted data (JSON)
- **Entire blob**: Base64 encoded to single string
- **Line wrapping**: Wrapped to 80 characters for git diff readability and copy-paste
- **No visible structure**: Attacker sees only wrapped base64 string

## Error Handling

The vault package defines these errors:

- `ErrVaultNotFound` - Vault file doesn't exist
- `ErrKeyNotFound` - Key doesn't exist in vault
- `ErrInvalidFormat` - Invalid vault file format
- `ErrInvalidAuth` - Authentication failed (wrong password or corrupted data)

## Best Practices

- Use strong passwords (at least 16 characters with mixed case, numbers, symbols)
- Call `defer vault.Close()` to ensure vault is saved
- Use `TryOpen()` for UI flows to distinguish auth failures from other errors
- Never hardcode passwords in source code
- Store vault files with restricted permissions (0600)
- Use `ChangePassword()` periodically for password rotation
- Never pass passwords via command-line args to the CLI
- Single base64 string format provides security through obscurity and git diff convenience

## Migration from envenc

This is a breaking change from the previous envenc-based vault. Old vault files are not compatible. To migrate:

1. Open old vault with envenc and export secrets
2. Create new vault with ork vault API
3. Import secrets into new vault
4. Delete old vault file

See the [Internal Vault Implementation proposal](../proposals/2026-04-22-internal-vault.md) for details.
