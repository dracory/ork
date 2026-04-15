# Quick Start

The core concept is the **Node** - a representation of a remote server:

```go
package main

import (
    "log"
    "github.com/dracory/ork"
    "github.com/dracory/ork/types"
)

func main() {
    // Create a node (remote server) - multiple ways:
    
    // Option 1: From host (most common)
    node := ork.NewNodeForHost("server.example.com")
    
    // Option 2: From config (useful for complex setups)
    cfg := types.NodeConfig{
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

## Skills

Run pre-built automation tasks (skills) against a node. For a complete list of available skills, see [Skills Documentation](skills.md).

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

results := node.Run(skills.NewUserCreate())
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
results := inv.Run(skills.NewPing())

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
