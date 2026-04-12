# Ork

Ork is a Go package for SSH-based server automation. Think of it like Ansible, but in Go - you define **Nodes** (remote servers) and run commands or playbooks against them.

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
)

func main() {
    // Create a node (remote server) - multiple ways:
    
    // Option 1: From host (most common)
    node := ork.NewNodeForHost("server.example.com")
    
    // Option 2: Empty node, configure later
    node := ork.NewNode().SetHost("server.example.com")
    
    // Option 3: From existing config
    node := ork.NewNodeFromConfig(cfg)
    
    // Run a command
    output, err := node.RunCommand("uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
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

output, err := node.RunCommand("uptime")
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
output1, _ := node.RunCommand("uptime")
output2, _ := node.RunCommand("df -h")
```

## Playbooks

Run pre-built automation tasks (playbooks) against a node:

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

result := node.RunPlaybook(ork.PlaybookUserCreate)
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
| `PlaybookPing` | `NamePing` | `ping` | - | Check SSH connectivity |
| `PlaybookAptUpdate` | `NameAptUpdate` | `apt-update` | - | Refresh package database |
| `PlaybookAptUpgrade` | `NameAptUpgrade` | `apt-upgrade` | - | Install available updates |
| `PlaybookAptStatus` | `NameAptStatus` | `apt-status` | - | Show available updates |
| `PlaybookReboot` | `NameReboot` | `reboot` | - | Reboot server |
| `PlaybookSwapCreate` | `NameSwapCreate` | `swap-create` | `size` (GB) | Create swap file |
| `PlaybookSwapDelete` | `NameSwapDelete` | `swap-delete` | - | Remove swap file |
| `PlaybookSwapStatus` | `NameSwapStatus` | `swap-status` | - | Show swap status |
| `PlaybookUserCreate` | `NameUserCreate` | `user-create` | `username` | Create user with sudo |
| `PlaybookUserDelete` | `NameUserDelete` | `user-delete` | `username` | Delete user |
| `PlaybookUserStatus` | `NameUserStatus` | `user-status` | `username` (opt) | Show user info |

## Idempotency

All playbooks now support idempotent execution. Use `RunPlaybook()` to see whether any changes were actually made:

```go
// RunPlaybook returns detailed result information
result := node.RunPlaybook(ork.PlaybookAptUpgrade)
if result.Error != nil {
    log.Fatal(result.Error)
}

if result.Changed {
    log.Printf("Changes made: %s", result.Message)
} else {
    log.Println("No changes needed - system already in desired state")
}
```

### Result Structure

```go
type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}
```

### Direct Playbook Access (Advanced)

For programmatic playbook handling, use the `playbook` package directly:

```go
import (
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/playbooks"
)

// Execute with helper function
result := playbook.Execute(playbooks.NewAptUpgrade(), cfg)

// Or check before running
pb := playbooks.NewSwapCreate()
needsChange, _ := pb.Check(cfg)
if !needsChange {
    log.Println("Swap already exists, skipping...")
    return
}
```

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
cfg := node.GetConfig()
```

### Custom Playbooks

Extend Ork with custom automation tasks by implementing the `Playbook` interface:

#### Custom Playbooks with Full Idempotency

For full idempotency support, implement all methods:

```go
type MyCustomPlaybook struct{}

func (p *MyCustomPlaybook) Name() string { return "my-task" }
func (p *MyCustomPlaybook) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (p *MyCustomPlaybook) Check(cfg config.Config) (bool, error) {
    // Check if already configured
    output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "cat /etc/my-config")
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (p *MyCustomPlaybook) Run(cfg config.Config) playbook.Result {
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
    cfg := config.Config{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := playbooks.NewPing()
    if err := ping.Run(cfg); err != nil {
        log.Fatal(err)
    }

    // Update packages
    aptUpdate := playbooks.NewAptUpdate()
    if err := aptUpdate.Run(cfg); err != nil {
        log.Fatal(err)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := playbooks.NewSwapCreate()
    if err := swapCreate.Run(cfg); err != nil {
        log.Fatal(err)
    }
}
```

### Package Overview

- `ork` - Main API: `NodeInterface`, `NewNode()`, `NewNodeForHost()`, `NewNodeFromConfig()`
- `config` - Configuration types
- `ssh` - SSH client with connection management
- `playbook` - Playbook interface and registry
- `playbooks` - Built-in playbook implementations

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
