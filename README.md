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

err := node.RunPlaybook(ork.PlaybookUserCreate)
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

Extend Ork with custom automation tasks:

```go
import (
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/config"
)

// Create a custom playbook
customPlaybook := playbook.NewSimplePlaybook(
    "custom-task",
    "Performs a custom automation task",
    func(cfg config.Config) error {
        // Your custom logic here
        return nil
    },
)

// Register it in the playbook registry
// (access via the playbook package)
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
