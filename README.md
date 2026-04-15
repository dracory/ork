# Ork

Ork is a Go package for SSH-based server automation. Think of it like Ansible, but in Go - you define **Nodes** (remote servers), organize them into **Groups**, and run commands or skills against them individually or at scale via **Inventory**.

## Installation

```bash
go get github.com/dracory/ork
```

## Quick Start

The core concept is the **Node** - a representation of a remote server:

```go
package main

import (
    "log"
    "github.com/dracory/ork"
    "github.com/dracory/ork/config"
)

func main() {
    // Create a node (remote server) - multiple ways:
    
    // Option 1: From host (most common)
    node := ork.NewNodeForHost("server.example.com")
    
    // Option 2: From config (useful for complex setups)
    cfg := config.NodeConfig{
        SSHHost: "server.example.com",
        SSHPort: "22",
    }
    node := ork.NewNodeFromConfig(cfg)
    
    // Run a command
    results := node.RunCommand("uptime")
    result := results.Results["server.example.com"]
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    log.Println(result.Message)
}
```

## Vault (Secure Secrets Management)

Ork provides secure vault support for managing sensitive data like passwords and API keys. The vault uses encrypted storage with two loading strategies:

### Load to Memory (Recommended for Security)

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

### Load to Environment (For .env Compatibility)

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

### Available Functions

**Load to Memory:**
- `VaultFileToKeys(filePath, password)` - Load from file to map
- `VaultContentToKeys(content, password)` - Load from string to map
- `VaultFileToKeysWithPrompt(filePath)` - Load from file with interactive prompt
- `VaultContentToKeysWithPrompt(content)` - Load from string with interactive prompt

**Load to Environment:**
- `VaultFileToEnv(filePath, password)` - Load from file to environment
- `VaultContentToEnv(content, password)` - Load from string to environment
- `VaultFileToEnvWithPrompt(filePath)` - Load from file with interactive prompt
- `VaultContentToEnvWithPrompt(content)` - Load from string with interactive prompt

**Utilities:**
- `PromptPassword(prompt)` - Securely prompt for password from stdin (no echo)

### Creating a Vault File

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

### Design Trade-offs

**ToKeys functions:**
- Secrets stored in memory map
- No global exposure
- Recommended for security-sensitive applications
- Caller manages secret lifecycle

**ToEnv functions:**
- Secrets dumped to environment variables
- Global exposure via `os.Getenv`
- For compatibility with .env-based tooling
- Use when external tools expect environment variables

## Configuration

Use the fluent API to configure connection settings:

```go
// From host
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")

results := node.RunCommand("uptime")
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatal(result.Error)
}
output := result.Message
```

## Persistent Connections

For multiple operations, establish a persistent connection:

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy")

if err := node.Connect(); err != nil {
    log.Fatal(err)
}
defer node.Close()

// These commands reuse the same SSH connection
results1 := node.RunCommand("uptime")
results2 := node.RunCommand("df -h")
output1 := results1.Results["server.example.com"].Message
output2 := results2.Results["server.example.com"].Message
```

## Skills

Run pre-built automation tasks (skills) against a node:

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

results := node.RunSkill(skills.NewUserCreate())
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatalf("Skill failed: %v", result.Error)
}
if result.Changed {
    log.Printf("User created: %s", result.Message)
} else {
    log.Println("User already exists - no changes made")
}
```

### Available Skills

| `ork` Package | `skill` Package | String | Args | Description |
|---------------|----------------|--------|------|-------------|
| `SkillPing` | `IDPing` | `ping` | - | Check SSH connectivity |
| `SkillAptUpdate` | `IDAptUpdate` | `apt-update` | - | Refresh package database |
| `SkillAptUpgrade` | `IDAptUpgrade` | `apt-upgrade` | - | Install available updates |
| `SkillAptStatus` | `IDAptStatus` | `apt-status` | - | Show available updates |
| `SkillReboot` | `IDReboot` | `reboot` | - | Reboot server |
| `SkillSwapCreate` | `IDSwapCreate` | `swap-create` | `size` (GB) | Create swap file |
| `SkillSwapDelete` | `IDSwapDelete` | `swap-delete` | - | Remove swap file |
| `SkillSwapStatus` | `IDSwapStatus` | `swap-status` | - | Show swap status |
| `SkillUserCreate` | `IDUserCreate` | `user-create` | `username` | Create user with sudo |
| `SkillUserDelete` | `IDUserDelete` | `user-delete` | `username` | Delete user |
| `SkillUserStatus` | `IDUserStatus` | `user-status` | `username` (opt) | Show user info |
| `SkillFail2banInstall` | `IDFail2banInstall` | `fail2ban-install` | - | Install and configure fail2ban |
| `SkillFail2banStatus` | `IDFail2banStatus` | `fail2ban-status` | - | Show fail2ban status |
| `SkillUfwInstall` | `IDUfwInstall` | `ufw-install` | - | Install UFW firewall |
| `SkillUfwStatus` | `IDUfwStatus` | `ufw-status` | - | Show UFW status |
| `SkillUfwAllowMariadb` | `IDUfwAllowMariadb` | `ufw-allow-mariadb` | - | Allow MariaDB through UFW |
| `SkillMariadbInstall` | `IDMariadbInstall` | `mariadb-install` | - | Install MariaDB server |
| `SkillMariadbStatus` | `IDMariadbStatus` | `mariadb-status` | - | Show MariaDB status |
| `SkillMariadbSecure` | `IDMariadbSecure` | `mariadb-secure` | - | Secure MariaDB installation |
| `SkillMariadbBackup` | `IDMariadbBackup` | `mariadb-backup` | `database` (opt) | Backup database |
| `SkillMariadbBackupEncrypt` | `IDMariadbBackupEncrypt` | `mariadb-backup-encrypt` | - | Encrypted backup |
| `SkillMariadbChangePort` | `IDMariadbChangePort` | `mariadb-change-port` | `port` | Change MariaDB port |
| `SkillMariadbCreateDB` | `IDMariadbCreateDB` | `mariadb-create-db` | `database` | Create database |
| `SkillMariadbCreateUser` | `IDMariadbCreateUser` | `mariadb-create-user` | `username`, `password` | Create DB user |
| `SkillMariadbEnableEncryption` | `IDMariadbEnableEncryption` | `mariadb-enable-encryption` | - | Enable encryption at rest |
| `SkillMariadbEnableSSL` | `IDMariadbEnableSSL` | `mariadb-enable-ssl` | - | Enable SSL connections |
| `SkillMariadbListDBs` | `IDMariadbListDBs` | `mariadb-list-dbs` | - | List databases |
| `SkillMariadbListUsers` | `IDMariadbListUsers` | `mariadb-list-users` | - | List DB users |
| `SkillMariadbSecurityAudit` | `IDMariadbSecurityAudit` | `mariadb-security-audit` | - | Run security audit |
| `SkillSecurityAideInstall` | `IDSecurityAideInstall` | `security-aide-install` | - | Install AIDE IDS |
| `SkillSecurityAuditdInstall` | `IDSecurityAuditdInstall` | `security-auditd-install` | - | Install audit daemon |
| `SkillSecurityKernelHarden` | `IDSecurityKernelHarden` | `security-kernel-harden` | - | Apply kernel hardening |
| `SkillSecuritySSHChangePort` | `IDSecuritySSHChangePort` | `security-ssh-change-port` | `port` | Change SSH port |
| `SkillSecuritySSHHarden` | `IDSecuritySSHHarden` | `security-ssh-harden` | - | Harden SSH config |

## Inventory (Multi-Node Operations)

Manage multiple servers with the same API:

```go
// Create inventory
inv := ork.NewInventory()

// Add nodes to groups
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))
webGroup.SetArg("env", "production")
inv.AddGroup(webGroup)

// Run skill on entire inventory
results := inv.RunSkill(skills.NewPing())

// Check summary
summary := results.Summary()
fmt.Printf("Total: %d, Changed: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Failed)

// Check individual results
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    }
}
```

## Idempotency

All skills support idempotent execution. Use `CheckSkill()` to preview changes:

```go
// Check if changes would be made (dry-run)
results := node.CheckSkill(skills.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would upgrade packages: %s", result.Message)
    // Now actually run it
    results = node.RunSkill(skills.NewAptUpgrade())
}
```

### Result Structure

Results are returned as `types.Results` with per-node access:

```go
type Results struct {
    Results map[string]Result  // Key is node hostname
}

func (r Results) Summary() Summary

type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}

type Summary struct {
    Total     int
    Changed   int
    Unchanged int
    Failed    int
}
```

### Direct Skill Access (Advanced)

For programmatic skill handling, use the `skill` package directly:

```go
import (
    "github.com/dracory/ork/skill"
    "github.com/dracory/ork/skills"
)

// Execute directly with config
aptUpgrade := skills.NewAptUpgrade()
aptUpgrade.SetConfig(cfg)
result := aptUpgrade.Run()

// Or check before running via CheckSkill
results := node.CheckSkill(skills.NewSwapCreate())
result := results.Results["server.example.com"]
if !result.Changed {
    log.Println("Swap already exists, skipping...")
    return
}
```

## Dry-Run Mode

Preview what changes would be made without actually executing commands on the server. Safety is enforced at the SSH execution layer - **no commands execute on the server in dry-run mode**.

### Enable Dry-Run

```go
// At node level
node := ork.NewNodeForHost("server.example.com").
    SetDryRunMode(true)
results := node.RunSkill(skills.NewAptUpgrade())
// Commands are logged but not executed

// At group level
webGroup := ork.NewGroup("webservers")
webGroup.SetDryRunMode(true)
webGroup.AddNode(node1)
webGroup.AddNode(node2)
// All nodes inherit dry-run mode

// At inventory level
inv := ork.NewInventory()
inv.SetDryRunMode(true)
inv.AddGroup(webGroup)
results := inv.RunCommand("uptime")
// All groups and nodes inherit dry-run mode
```

### How It Works

1. **Safety at execution layer**: `ssh.Run()` checks `cfg.IsDryRunMode` and returns `"[dry-run]"` without executing commands
2. **Automatic propagation**: Dry-run mode propagates from Inventory → Groups → Nodes at execution time
3. **Thread-safe**: Uses mutex protection for concurrent access to dry-run state

### Detecting Dry-Run in Skills

```go
func (s *MySkill) Run() skill.Result {
    output, _ := ssh.Run(s.cfg, "apt-get upgrade -y")

    if output == "[dry-run]" {
        return skill.Result{
            Changed: true,
            Message: "Would run: apt-get upgrade -y",
        }
    }
    // Normal execution handling...
}
```

**Note:** Even if a skill doesn't check for the `[dry-run]` marker, **safety is guaranteed** - no commands execute on the server when dry-run mode is enabled.

## Advanced Usage

### Inspecting Configuration

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy")

fmt.Printf("Host: %s\n", node.GetHost())
fmt.Printf("Port: %s\n", node.GetPort())
fmt.Printf("User: %s\n", node.GetUser())

// Get full config for integration with internal packages
cfg := node.GetNodeConfig()
```

### Custom Skills

Extend Ork with custom automation tasks by implementing the `Skill` interface:

#### Registering Custom Skills

Register your custom skills to use them via `node.RunSkillByID("custom-id")`:

```go
import (
    "github.com/dracory/ork"
    "github.com/dracory/ork/skill"
)

// Create a custom skill
customSkill := skill.NewBaseSkill()
customSkill.SetID("install-docker")
customSkill.SetDescription("Install Docker on the server")

// Register it globally
registry, err := ork.GetGlobalSkillRegistry()
if err != nil {
    log.Fatalf("Failed to get registry: %v", err)
}
if err := registry.SkillRegister(customSkill); err != nil {
    log.Fatalf("Failed to register skill: %v", err)
}

// Now use it like any built-in skill
node := ork.NewNodeForHost("server.example.com")
result := node.RunSkillByID("install-docker")
```

### Custom Skills with Full Idempotency

For full idempotency support, implement all methods:

```go
type MyCustomSkill struct{}

func (s *MyCustomSkill) GetID() string { return "my-task" }
func (s *MyCustomSkill) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (s *MyCustomSkill) Check(cfg config.NodeConfig) (bool, error) {
    // Check if already configured
    // Skills implement Check for use with CheckSkill()
    output, _ := ssh.Run(cfg, types.Command{Command: "cat /etc/my-config"})
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (s *MyCustomSkill) Run(cfg config.NodeConfig) skill.Result {
    needsChange, _ := s.Check(cfg)
    if !needsChange {
        return skill.Result{
            Changed: false,
            Message: "Already configured",
        }
    }

    // Apply changes...
    _, err := ssh.Run(cfg, types.Command{Command: "setup-command"})
    if err != nil {
        return skill.Result{Changed: false, Error: err}
    }

    return skill.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}
```

## Internal Packages

For advanced use cases or when you need fine-grained control, you can use the internal packages directly:

- `ssh` - SSH connection utilities and command execution
- `config` - Configuration types for remote operations
- `skill` - Base interfaces and registry for organizing skills
- `skills` - Reusable skill implementations (ping, apt, reboot, swap, user)

For advanced use cases, use the internal packages directly:

```go
package main

import (
    "log"

    "github.com/dracory/ork/config"
    "github.com/dracory/ork/skills"
)

func main() {
    cfg := config.NodeConfig{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := skills.NewPing()
    ping.SetConfig(cfg)
    result := ping.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Update packages
    aptUpdate := skills.NewAptUpdate()
    aptUpdate.SetConfig(cfg)
    result = aptUpdate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := skills.NewSwapCreate()
    swapCreate.SetConfig(cfg)
    result = swapCreate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }
}
```

### Package Overview

- `ork` - Main API: `NodeInterface`, `InventoryInterface`, `GroupInterface`, `RunnableInterface`
- `types` - Shared types: `Result`, `Results`, `Summary`
- `runnable` - `RunnableInterface` for Node, Group, and Inventory
- `config` - Configuration types
- `ssh` - SSH client with connection management
- `skill` - Skill interface and registry
- `skills` - Built-in skill implementations

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
