# Ork

Ork is a Go package for SSH-based server automation. Think of it like Ansible, but in Go - you define **Nodes** (remote servers), organize them into **Groups**, and run commands or playbooks against them individually or at scale via **Inventory**.

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

## Playbooks

Run pre-built automation tasks (playbooks) against a node:

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

results := node.RunPlaybook(playbooks.NewUserCreate())
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatalf("Playbook failed: %v", result.Error)
}
if result.Changed {
    log.Printf("User created: %s", result.Message)
} else {
    log.Println("User already exists - no changes made")
}
```

### Available Playbooks

| `ork` Package | `playbook` Package | String | Args | Description |
|---------------|-------------------|--------|------|-------------|
| `PlaybookPing` | `IDPing` | `ping` | - | Check SSH connectivity |
| `PlaybookAptUpdate` | `IDAptUpdate` | `apt-update` | - | Refresh package database |
| `PlaybookAptUpgrade` | `IDAptUpgrade` | `apt-upgrade` | - | Install available updates |
| `PlaybookAptStatus` | `IDAptStatus` | `apt-status` | - | Show available updates |
| `PlaybookReboot` | `IDReboot` | `reboot` | - | Reboot server |
| `PlaybookSwapCreate` | `IDSwapCreate` | `swap-create` | `size` (GB) | Create swap file |
| `PlaybookSwapDelete` | `IDSwapDelete` | `swap-delete` | - | Remove swap file |
| `PlaybookSwapStatus` | `IDSwapStatus` | `swap-status` | - | Show swap status |
| `PlaybookUserCreate` | `IDUserCreate` | `user-create` | `username` | Create user with sudo |
| `PlaybookUserDelete` | `IDUserDelete` | `user-delete` | `username` | Delete user |
| `PlaybookUserStatus` | `IDUserStatus` | `user-status` | `username` (opt) | Show user info |
| `PlaybookFail2banInstall` | `IDFail2banInstall` | `fail2ban-install` | - | Install and configure fail2ban |
| `PlaybookFail2banStatus` | `IDFail2banStatus` | `fail2ban-status` | - | Show fail2ban status |
| `PlaybookUfwInstall` | `IDUfwInstall` | `ufw-install` | - | Install UFW firewall |
| `PlaybookUfwStatus` | `IDUfwStatus` | `ufw-status` | - | Show UFW status |
| `PlaybookUfwAllowMariadb` | `IDUfwAllowMariadb` | `ufw-allow-mariadb` | - | Allow MariaDB through UFW |
| `PlaybookMariadbInstall` | `IDMariadbInstall` | `mariadb-install` | - | Install MariaDB server |
| `PlaybookMariadbStatus` | `IDMariadbStatus` | `mariadb-status` | - | Show MariaDB status |
| `PlaybookMariadbSecure` | `IDMariadbSecure` | `mariadb-secure` | - | Secure MariaDB installation |
| `PlaybookMariadbBackup` | `IDMariadbBackup` | `mariadb-backup` | `database` (opt) | Backup database |
| `PlaybookMariadbBackupEncrypt` | `IDMariadbBackupEncrypt` | `mariadb-backup-encrypt` | - | Encrypted backup |
| `PlaybookMariadbChangePort` | `IDMariadbChangePort` | `mariadb-change-port` | `port` | Change MariaDB port |
| `PlaybookMariadbCreateDB` | `IDMariadbCreateDB` | `mariadb-create-db` | `database` | Create database |
| `PlaybookMariadbCreateUser` | `IDMariadbCreateUser` | `mariadb-create-user` | `username`, `password` | Create DB user |
| `PlaybookMariadbEnableEncryption` | `IDMariadbEnableEncryption` | `mariadb-enable-encryption` | - | Enable encryption at rest |
| `PlaybookMariadbEnableSSL` | `IDMariadbEnableSSL` | `mariadb-enable-ssl` | - | Enable SSL connections |
| `PlaybookMariadbListDBs` | `IDMariadbListDBs` | `mariadb-list-dbs` | - | List databases |
| `PlaybookMariadbListUsers` | `IDMariadbListUsers` | `mariadb-list-users` | - | List DB users |
| `PlaybookMariadbSecurityAudit` | `IDMariadbSecurityAudit` | `mariadb-security-audit` | - | Run security audit |
| `PlaybookSecurityAideInstall` | `IDSecurityAideInstall` | `security-aide-install` | - | Install AIDE IDS |
| `PlaybookSecurityAuditdInstall` | `IDSecurityAuditdInstall` | `security-auditd-install` | - | Install audit daemon |
| `PlaybookSecurityKernelHarden` | `IDSecurityKernelHarden` | `security-kernel-harden` | - | Apply kernel hardening |
| `PlaybookSecuritySSHChangePort` | `IDSecuritySSHChangePort` | `security-ssh-change-port` | `port` | Change SSH port |
| `PlaybookSecuritySSHHarden` | `IDSecuritySSHHarden` | `security-ssh-harden` | - | Harden SSH config |

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

// Run playbook on entire inventory
results := inv.RunPlaybook(playbooks.NewPing())

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

All playbooks support idempotent execution. Use `CheckPlaybook()` to preview changes:

```go
// Check if changes would be made (dry-run)
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would upgrade packages: %s", result.Message)
    // Now actually run it
    results = node.RunPlaybook(playbooks.NewAptUpgrade())
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

### Direct Playbook Access (Advanced)

For programmatic playbook handling, use the `playbook` package directly:

```go
import (
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/playbooks"
)

// Execute directly with config
aptUpgrade := playbooks.NewAptUpgrade()
aptUpgrade.SetConfig(cfg)
result := aptUpgrade.Run()

// Or check before running via CheckPlaybook
results := node.CheckPlaybook(playbooks.NewSwapCreate())
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
results := node.RunPlaybook(playbooks.NewAptUpgrade())
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

### Detecting Dry-Run in Playbooks

```go
func (p *MyPlaybook) Run() playbook.Result {
    output, _ := ssh.Run(p.cfg, "apt-get upgrade -y")

    if output == "[dry-run]" {
        return playbook.Result{
            Changed: true,
            Message: "Would run: apt-get upgrade -y",
        }
    }
    // Normal execution handling...
}
```

**Note:** Even if a playbook doesn't check for the `[dry-run]` marker, **safety is guaranteed** - no commands execute on the server when dry-run mode is enabled.

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

### Custom Playbooks

Extend Ork with custom automation tasks by implementing the `Playbook` interface:

#### Registering Custom Playbooks

Register your custom playbooks to use them via `node.RunPlaybookByID("custom-id")`:

```go
import (
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbook"
)

// Create a custom playbook
customPb := playbook.NewBasePlaybook()
customPb.SetID("install-docker")
customPb.SetDescription("Install Docker on the server")

// Register it globally
registry, err := ork.GetGlobalPlaybookRegistry()
if err != nil {
    log.Fatalf("Failed to get registry: %v", err)
}
if err := registry.PlaybookRegister(customPb); err != nil {
    log.Fatalf("Failed to register playbook: %v", err)
}

// Now use it like any built-in playbook
node := ork.NewNodeForHost("server.example.com")
result := node.RunPlaybookByID("install-docker")
```

### Custom Playbooks with Full Idempotency

For full idempotency support, implement all methods:

```go
type MyCustomPlaybook struct{}

func (p *MyCustomPlaybook) GetID() string { return "my-task" }
func (p *MyCustomPlaybook) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (p *MyCustomPlaybook) Check(cfg config.NodeConfig) (bool, error) {
    // Check if already configured
    // Playbooks implement Check for use with CheckPlaybook()
    output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "cat /etc/my-config")
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (p *MyCustomPlaybook) Run(cfg config.NodeConfig) playbook.Result {
    needsChange, _ := p.Check(cfg)
    if !needsChange {
        return playbook.Result{
            Changed: false,
            Message: "Already configured",
        }
    }

    // Apply changes...
    _, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "setup-command")
    if err != nil {
        return playbook.Result{Changed: false, Error: err}
    }

    return playbook.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}
```

## Internal Packages

For advanced use cases or when you need fine-grained control, you can use the internal packages directly:

- `ssh` - SSH connection utilities and command execution
- `config` - Configuration types for remote operations
- `playbook` - Base interfaces and registry for organizing playbooks
- `playbooks` - Reusable playbook implementations (ping, apt, reboot, swap, user)

For advanced use cases, use the internal packages directly:

```go
package main

import (
    "log"

    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbooks"
)

func main() {
    cfg := config.NodeConfig{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := playbooks.NewPing()
    ping.SetConfig(cfg)
    result := ping.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Update packages
    aptUpdate := playbooks.NewAptUpdate()
    aptUpdate.SetConfig(cfg)
    result = aptUpdate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := playbooks.NewSwapCreate()
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
- `playbook` - Playbook interface and registry
- `playbooks` - Built-in playbook implementations

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
