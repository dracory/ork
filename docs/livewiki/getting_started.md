---
path: getting_started.md
page-type: tutorial
summary: Step-by-step guide to installing Ork and running your first automation tasks.
tags: [tutorial, getting-started, installation, quickstart, privilege-escalation]
created: 2025-04-14
updated: 2026-04-15
version: 1.1.0
---

# Getting Started with Ork

This guide will walk you through installing Ork, configuring your first node, and running your first playbook.

## Prerequisites

Before you begin, ensure you have:

- **Go 1.25+** installed on your machine
- **SSH key pair** for authentication to remote servers
- **Root access** to the target servers (or a user with sudo privileges)

## Installation

Install Ork using `go get`:

```bash
go get github.com/dracory/ork
```

Or add it to your `go.mod`:

```go
require github.com/dracory/ork v1.0.0
```

## Your First Node

A `Node` represents a single remote server. Here's how to create and use one:

### Basic Node Creation

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Create a node with default settings (port 22, user root, key id_rsa)
    node := ork.NewNodeForHost("server.example.com")
    
    // Run a simple command
    results := node.RunCommand("uptime")
    result := results.Results["server.example.com"]
    
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    
    log.Println(result.Message)
}
```

### Custom Configuration

Configure your node using the fluent API:

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")
```

## Your First Playbook

Playbooks are reusable automation tasks. Ork includes many built-in playbooks.

### Ping (Connectivity Check)

```go
package main

import (
    "log"
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbooks"
)

func main() {
    node := ork.NewNodeForHost("server.example.com")
    
    // Check SSH connectivity
    results := node.RunPlaybook(playbooks.NewPing())
    result := results.Results["server.example.com"]
    
    if result.Error != nil {
        log.Fatalf("Connection failed: %v", result.Error)
    }
    
    log.Println(result.Message)
}
```

### Update Packages

```go
// Update the package database
results := node.RunPlaybook(playbooks.NewAptUpdate())
if result.Error != nil {
    log.Fatal(result.Error)
}

// Upgrade installed packages
results = node.RunPlaybook(playbooks.NewAptUpgrade())
```

### Create a User

```go
// Set arguments for the playbook
node.WithArg("username", "alice").
    WithArg("shell", "/bin/bash")

// Run the user creation playbook
results := node.RunPlaybook(playbooks.NewUserCreate())
result := results.Results["server.example.com"]

if result.Error != nil {
    log.Fatalf("User creation failed: %v", result.Error)
}

if result.Changed {
    log.Printf("User created: %s", result.Message)
} else {
    log.Println("User already exists - no changes made")
}
```

## Working with Groups

Groups allow you to manage multiple servers together:

```go
// Create a group
webGroup := ork.NewGroup("webservers").
    WithArg("env", "production")

// Add nodes to the group
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

// Run playbook on all nodes in the group
results := webGroup.RunPlaybook(playbooks.NewAptUpdate())

// Check results for all nodes
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    } else {
        log.Printf("%s: %s (changed: %v)", host, result.Message, result.Changed)
    }
}
```

## Working with Inventory

For large-scale operations, use an Inventory:

```go
// Create inventory
inv := ork.NewInventory()

// Create and add groups
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

// Add another group
dbGroup := ork.NewGroup("databases")
dbGroup.AddNode(ork.NewNodeForHost("db1.example.com"))

// Add groups to inventory
inv.AddGroup(webGroup)
inv.AddGroup(dbGroup)

// Run playbook across all nodes
results := inv.RunPlaybook(playbooks.NewPing())

// Get summary
summary := results.Summary()
log.Printf("Total: %d, Changed: %d, Failed: %d", 
    summary.Total, summary.Changed, summary.Failed)
```

## Dry-Run Mode

Always test your operations before running them:

```go
// Enable dry-run mode
node := ork.NewNodeForHost("server.example.com").
    SetDryRunMode(true)

// This will log what would happen without making changes
results := node.RunPlaybook(playbooks.NewAptUpgrade())
```

Dry-run mode works at all levels:
- **Node level**: `node.SetDryRunMode(true)`
- **Group level**: `group.SetDryRunMode(true)`
- **Inventory level**: `inv.SetDryRunMode(true)`

## Privilege Escalation (Become)

Ork supports running commands as a different user using `sudo`. This is useful when you need to perform operations that require elevated privileges.

### Basic Usage

```go
// Connect as a non-root user, but run commands as root
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecomeUser("root")

// This will run: sudo -u root apt-get update
results := node.RunCommand("apt-get update")
```

### Running as Specific Users

```go
// Run database commands as the postgres user
node.SetBecomeUser("postgres")
results := node.RunCommand("psql -c 'SELECT version()'")

// Run web server commands as www-data
node.SetBecomeUser("www-data")
results := node.RunCommand("systemctl restart nginx")
```

### Precedence Rules

The become user setting follows this precedence (highest to lowest):
1. **Skill/Playbook level**: Skill-specific setting
2. **Node level**: Node-specific setting
3. **Group level**: Group setting (propagated to all nodes)
4. **Inventory level**: Inventory setting (propagated to all groups and nodes)

```go
// Set at inventory level - applies to all
inv := ork.NewInventory()
inv.SetBecomeUser("root")

// Override at group level
dbGroup := ork.NewGroup("databases")
dbGroup.SetBecomeUser("postgres")  // Uses postgres, not root

// Override at node level
node := ork.NewNodeForHost("special.example.com")
node.SetBecomeUser("admin")  // Uses admin, not postgres or root
```

### With Skills

```go
// Set become user on the skill itself
skill := skills.NewAptUpdate()
skill.SetBecomeUser("root")

// The skill will run as root regardless of node/group/inventory settings
results := node.Run(skill)
```

### Security Considerations

- The connecting user must have `sudo` privileges to the target user
- No password support is currently implemented - ensure passwordless sudo is configured
- Use the principle of least privilege - only escalate when necessary

For detailed documentation, see [Privilege Escalation](../../privilege_escalation.md).

## Persistent Connections

For multiple operations, use a persistent connection:

```go
node := ork.NewNodeForHost("server.example.com")

// Establish connection
if err := node.Connect(); err != nil {
    log.Fatal(err)
}
defer node.Close()

// These commands reuse the same SSH connection
results1 := node.RunCommand("uptime")
results2 := node.RunCommand("df -h")
results3 := node.RunCommand("free -m")
```

## Checking Before Running

Use `CheckPlaybook` to preview if changes would be made:

```go
// Check if changes would be made
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Println("Would upgrade packages")
    // Now actually run it
    results = node.RunPlaybook(playbooks.NewAptUpgrade())
} else {
    log.Println("No upgrades needed")
}
```

## Next Steps

- Learn about [all available playbooks](modules/playbooks.md)
- Understand the [architecture](architecture.md)
- Explore the [API reference](api_reference.md)
- Read about [dry-run mode and idempotency](configuration.md)

## See Also

- [Overview](overview.md) - High-level introduction to Ork
- [Architecture](architecture.md) - System architecture and design patterns
- [API Reference](api_reference.md) - Complete API documentation
- [Modules](modules/ork.md) - Detailed module documentation
