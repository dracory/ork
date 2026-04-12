# Ork

Ork is a Go package for SSH-based server automation playbooks. It provides common utilities for connecting to remote servers via SSH and running automation tasks.

## Installation

```bash
go get github.com/dracory/ork
```

## Simplified API (Recommended)

Ork provides a clean, intuitive top-level API that requires only a single import for common operations. This is the recommended way to use Ork for most use cases.

### Simple SSH Command Execution

Execute commands on remote servers with minimal code:

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Execute a command with default settings (port 22, user root, key id_rsa)
    output, err := ork.RunSSH("server.example.com", "uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
}
```

### Simple SSH with Custom Configuration

Use functional options for flexible configuration:

```go
output, err := ork.RunSSH("server.example.com", "uptime",
    ork.WithPort("2222"),
    ork.WithUser("deploy"),
    ork.WithKey("production.prv"),
)
```

### Simple Playbook Execution

Run pre-registered automation tasks by name:

```go
// Update package database
err := ork.RunPlaybook("apt-update", "server.example.com")
if err != nil {
    log.Fatal(err)
}

// Create a user with arguments
err = ork.RunPlaybook("user-create", "server.example.com",
    ork.WithArg("username", "alice"),
    ork.WithArg("shell", "/bin/bash"),
)
```

### Fluent Node API

Use the fluent builder pattern for readable, chainable configuration:

```go
// Create and configure a node
node := ork.NewNode("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")

// Execute commands
output, err := node.Run("uptime")
if err != nil {
    log.Fatal(err)
}
log.Println(output)

// Execute playbooks
err = node.SetArg("username", "alice").Playbook("user-create")
```

### Persistent Connections

Maintain persistent SSH connections for multiple operations:

```go
node := ork.NewNode("server.example.com").
    SetPort("2222").
    SetUser("deploy")

// Establish persistent connection
if err := node.Connect(); err != nil {
    log.Fatal(err)
}
defer node.Close()

// Multiple operations reuse the same connection (more efficient)
output1, _ := node.Run("uptime")
output2, _ := node.Run("df -h")
node.Playbook("apt-status")
```

### Inspecting Configuration

Use getter methods to inspect node configuration:

```go
node := ork.NewNode("server.example.com").
    SetPort("2222").
    SetUser("deploy")

fmt.Printf("Host: %s\n", node.GetHost())
fmt.Printf("Port: %s\n", node.GetPort())
fmt.Printf("User: %s\n", node.GetUser())

// Get full config for integration with internal packages
cfg := node.GetConfig()
```

### Code Comparison: Simplified API vs Internal Packages

**Before (Internal Packages):**
```go
import (
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/playbooks"
)

cfg := config.Config{
    SSHHost:  "server.example.com",
    SSHPort:  "2222",
    RootUser: "deploy",
    SSHKey:   "production.prv",
}
output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
```

**After (Simplified API):**
```go
import "github.com/dracory/ork"

output, err := ork.RunSSH("server.example.com", "uptime",
    ork.WithPort("2222"),
    ork.WithUser("deploy"),
    ork.WithKey("production.prv"),
)
```

### Custom Playbooks

Extend Ork with custom automation tasks:

```go
import (
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/config"
)

// Create a custom playbook
customPlaybook := playbook.NewSimplePlaybook(
    "custom-task",
    "Performs a custom automation task",
    func(cfg config.Config) error {
        // Your custom automation logic here
        return nil
    },
)

// Register it globally
ork.RegisterPlaybook(customPlaybook)

// Now use it like any built-in playbook
err := ork.RunPlaybook("custom-task", "server.example.com")
```

### Implementing Custom NodeInterface

For advanced use cases, implement the `NodeInterface` for custom behavior:

```go
type MyCustomNode struct {
    // Your custom fields
}

func (n *MyCustomNode) SetPort(port string) ork.NodeInterface {
    // Your custom implementation
    return n
}

// Implement all other NodeInterface methods...
```

### Discovering Available Playbooks

```go
// List all registered playbooks
names := ork.ListPlaybooks()
for _, name := range names {
    fmt.Println(name)
}

// Get a specific playbook
pb, ok := ork.GetPlaybook("apt-update")
if ok {
    fmt.Printf("%s: %s\n", pb.Name(), pb.Description())
}
```

### Backward Compatibility

The simplified API is fully backward compatible. All internal packages (`config`, `ssh`, `playbook`, `playbooks`) remain accessible and unchanged. You can use both APIs in the same codebase and migrate gradually.

## Internal Packages (Advanced)

For advanced use cases or when you need fine-grained control, you can use the internal packages directly:

- `ssh` - SSH connection utilities and command execution
- `config` - Configuration types for remote operations
- `playbook` - Base interfaces and registry for organizing playbooks
- `playbooks` - Reusable playbook implementations (ping, apt, reboot, swap, user)

## Quick Start (Internal Packages)

```go
package main

import (
    "log"
    
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
)

func main() {
    // Create config
    cfg := config.Config{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }
    
    // Run a command
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
}
```

## Using Reusable Playbooks

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

### Available Playbooks

| Playbook | Description | Args |
|----------|-------------|------|
| `ping` | Check SSH connectivity | - |
| `apt-update` | Refresh package database | - |
| `apt-upgrade` | Install available updates | - |
| `apt-status` | Show available updates | - |
| `reboot` | Reboot server | - |
| `swap-create` | Create swap file | `size` (GB, default 1) |
| `swap-delete` | Remove swap file | - |
| `swap-status` | Show swap status | - |
| `user-create` | Create user with sudo | `username` |
| `user-delete` | Delete user | `username` |
| `user-status` | Show user info | `username` (optional) |

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
