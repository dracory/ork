# Vault (Secure Secrets Management)

Ork provides secure vault support for managing sensitive data like passwords and API keys. The vault uses encrypted storage with two loading strategies.

## Load to Memory (Recommended for Security)

Secrets are loaded into a memory map, keeping them isolated from the rest of the process:

```go
// Simple one-liner with interactive password prompt
secrets, err := ork.VaultFileToKeysWithPrompt(".env.vault")
if err != nil {
    log.Fatal(err)
}
dbPassword := secrets["DATABASE_PASSWORD"]

// Or separate prompt and load for custom workflows
password, err := ork.PromptPassword("Vault password: ")
if err != nil {
    log.Fatal(err)
}
secrets, err := ork.VaultFileToKeys(".env.vault", password)
if err != nil {
    log.Fatal(err)
}
```

## Load to Environment (For .env Compatibility)

Secrets are loaded into environment variables for compatibility with tools that expect `.env` files:

```go
// Simple one-liner with interactive password prompt
if err := ork.VaultFileToEnvWithPrompt(".env.vault"); err != nil {
    log.Fatal(err)
}
// Secrets now available via os.Getenv()

// Or separate prompt and load
password, err := ork.PromptPassword("Vault password: ")
if err != nil {
    log.Fatal(err)
}
if err := ork.VaultFileToEnv(".env.vault", password); err != nil {
    log.Fatal(err)
}
```

## Available Functions

### Load to Memory

- `VaultFileToKeys(filePath, password)` - Load from file to map
- `VaultContentToKeys(content, password)` - Load from string to map
- `VaultFileToKeysWithPrompt(filePath)` - Load from file with interactive prompt
- `VaultContentToKeysWithPrompt(content)` - Load from string with interactive prompt

### Load to Environment

- `VaultFileToEnv(filePath, password)` - Load from file to environment
- `VaultContentToEnv(content, password)` - Load from string to environment
- `VaultFileToEnvWithPrompt(filePath)` - Load from file with interactive prompt
- `VaultContentToEnvWithPrompt(content)` - Load from string with interactive prompt

### Utilities

- `PromptPassword(prompt)` - Securely prompt for password from stdin (no echo)

## Creating a Vault File

Use the `envenc` CLI tool to create encrypted vault files:

```bash
# Initialize a new vault
envenc init .env.vault

# Set secrets
envenc set DATABASE_PASSWORD "my-secret-password"
envenc set API_KEY "my-api-key"

# List keys
envenc list .env.vault
```

## Design Trade-offs

### ToKeys functions

- Secrets stored in memory map
- No global exposure
- Recommended for security-sensitive applications
- Caller manages secret lifecycle

### ToEnv functions

- Secrets dumped to environment variables
- Global exposure via `os.Getenv`
- For compatibility with .env-based tooling
- Use when external tools expect environment variables
