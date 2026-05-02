---
path: cheatsheet.md
page-type: reference
summary: Quick reference for common Ork operations and patterns.
tags: [cheatsheet, quick-reference, cookbook]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# Ork Cheatsheet

Quick reference for common operations.

## Installation

```bash
go get github.com/dracory/ork
```

## Basic Operations

### Create Node

```go
// From hostname (defaults: port 22, user root, key id_rsa)
node := ork.NewNodeForHost("server.example.com")

// From config
cfg := config.NodeConfig{
    SSHHost: "server.example.com",
    SSHPort: "2222",
    RootUser: "deploy",
    SSHKey: "production.prv",
}
node := ork.NewNodeFromConfig(cfg)

// Empty node (configure manually - NOT for SSH)
node := ork.NewNode()
```

### Configure Node (Fluent API)

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv").
    SetArg("username", "alice")
```

### Run Commands

```go
// One-time connection
results := node.RunCommand("uptime")
result := results.Results["server.example.com"]

// Persistent connection
node.Connect()
defer node.Close()
results1 := node.RunCommand("uptime")
results2 := node.RunCommand("df -h")
```

### Run Playbooks

```go
// Direct instance (preferred)
results := node.RunPlaybook(playbooks.NewPing())
results := node.RunPlaybook(playbooks.NewAptUpdate())

// By ID
results := node.RunPlaybookByID(playbooks.IDPing)

// Check mode (dry-run for single playbook)
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
```

## Working with Groups

```go
// Create group
group := ork.NewGroup("webservers")

// Add nodes
group.AddNode(node1)
group.AddNode(node2)

// Set group arguments
group.WithArg("env", "production")

// Run on all nodes
results := group.RunPlaybook(playbooks.NewAptUpdate())
```

## Working with Inventory

```go
// Create inventory
inv := ork.NewInventory()

// Add groups
inv.AddGroup(webGroup)
inv.AddGroup(dbGroup)

// Or add individual nodes
inv.AddNode(node)

// Configure concurrency (default: 10)
inv.SetMaxConcurrency(20)

// Run on all nodes across all groups
results := inv.RunPlaybook(playbooks.NewPing())
```

## Result Handling

```go
results := inv.RunPlaybook(playbooks.NewAptUpdate())

// Get summary
summary := results.Summary()
fmt.Printf("Total: %d, Changed: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Failed)

// Iterate results
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    } else if result.Changed {
        log.Printf("%s changed: %s", host, result.Message)
    } else {
        log.Printf("%s unchanged", host)
    }
}
```

## Common Playbooks

### System

```go
playbooks.NewPing()        // Check connectivity
playbooks.NewAptUpdate()   // Update package lists
playbooks.NewAptUpgrade()  // Upgrade packages
playbooks.NewAptStatus()   // Check for updates
playbooks.NewReboot()      // Reboot server
```

### Users

```go
node.SetArg("username", "alice")
node.SetArg("shell", "/bin/bash")
results := node.RunPlaybook(playbooks.NewUserCreate())

node.SetArg("username", "bob")
results := node.RunPlaybook(playbooks.NewUserDelete())

results := node.RunPlaybook(playbooks.NewUserStatus())
```

### Swap

```go
node.SetArg("size", "2")
node.SetArg("swappiness", "10")
results := node.RunPlaybook(playbooks.NewSwapCreate())

results := node.RunPlaybook(playbooks.NewSwapDelete())
results := node.RunPlaybook(playbooks.NewSwapStatus())
```

### MariaDB

```go
results := node.RunPlaybook(playbooks.NewMariadbInstall())

node.SetArg("root-password", "secret")
results := node.RunPlaybook(playbooks.NewSecure())

node.SetArg("database", "myapp")
results := node.RunPlaybook(playbooks.NewCreateDB())

node.SetArg("username", "appuser")
node.SetArg("password", "apppass")
results := node.RunPlaybook(playbooks.NewCreateUser())
```

### Security

```go
results := node.RunPlaybook(playbooks.NewSshHarden())
results := node.RunPlaybook(playbooks.NewKernelHarden())
results := node.RunPlaybook(playbooks.NewAideInstall())

node.SetArg("port", "2222")
results := node.RunPlaybook(playbooks.NewSshChangePort())
```

### Firewall

```go
results := node.RunPlaybook(playbooks.NewUfwInstall())
results := node.RunPlaybook(playbooks.NewUfwStatus())
results := node.RunPlaybook(playbooks.NewAllowMariaDB())
```

### Fail2ban

```go
results := node.RunPlaybook(playbooks.NewFail2banInstall())
results := node.RunPlaybook(playbooks.NewFail2banStatus())
```

## Dry-Run Mode

```go
// Node level
node := ork.NewNodeForHost("server.example.com").
    SetDryRunMode(true)

// Group level
group.SetDryRunMode(true)

// Inventory level
inv.SetDryRunMode(true)

// Check if enabled
if node.GetDryRunMode() {
    log.Println("Running in dry-run mode")
}
```

## Logging

```go
import "log/slog"

// Create logger
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Set on any level
node.SetLogger(logger)
group.SetLogger(logger)
inv.SetLogger(logger)
```

## Configuration Inspection

```go
// Get values
host := node.GetHost()
port := node.GetPort()
user := node.GetUser()
key := node.GetKey()
arg := node.GetArg("username")
args := node.GetArgs()

// Get full config
cfg := node.GetNodeConfig()
addr := cfg.SSHAddr()  // "host:port"
```

## Playbook IDs

```go
// System
playbooks.IDPing
playbooks.IDAptUpdate
playbooks.IDAptUpgrade
playbooks.IDAptStatus
playbooks.IDReboot

// Users
playbooks.IDUserCreate
playbooks.IDUserDelete
playbooks.IDUserStatus

// Swap
playbooks.IDSwapCreate
playbooks.IDSwapDelete
playbooks.IDSwapStatus

// Security
playbooks.IDSshHarden
playbooks.IDKernelHarden
playbooks.IDAideInstall
playbooks.IDAuditdInstall

// MariaDB
playbooks.IDMariadbInstall
playbooks.IDMariadbSecure
playbooks.IDMariadbCreateDB
// ... etc
```

## Complete Example

```go
package main

import (
    "log"
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/playbooks"
)

func main() {
    // Create node
    node := ork.NewNodeForHost("server.example.com").
        SetPort("2222").
        SetUser("deploy")
    
    // Check connectivity
    results := node.RunPlaybook(playbooks.NewPing())
    if results.Results["server.example.com"].Error != nil {
        log.Fatal("Connection failed")
    }
    
    // Update packages
    results = node.RunPlaybook(playbooks.NewAptUpdate())
    if results.Results["server.example.com"].Error != nil {
        log.Printf("Update failed: %v", results.Results["server.example.com"].Error)
    }
    
    // Create user (dry-run first)
    node.SetDryRunMode(true)
    node.SetArg("username", "alice")
    results = node.RunPlaybook(playbooks.NewUserCreate())
    
    if results.Results["server.example.com"].Changed {
        log.Println("Would create user")
        
        // Actually create
        node.SetDryRunMode(false)
        results = node.RunPlaybook(playbooks.NewUserCreate())
        log.Println(results.Results["server.example.com"].Message)
    }
}
```

## See Also

- [Getting Started](getting_started.md) - Full tutorial
- [API Reference](api_reference.md) - Complete API docs
- [Playbooks](modules/playbooks.md) - All available playbooks
