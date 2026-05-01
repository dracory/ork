# Commands

Commands provide a fluent API for executing shell commands on remote servers. Unlike skills (which are idempotent and reusable), commands are for simple one-off shell command execution with optional configuration.

## Overview

Commands implement `CommandInterface` which extends `RunnableInterface`, allowing them to be executed on nodes, groups, and inventories using the standard `Run()` methods. Commands support:

- **Fluent API**: Chain methods for readable configuration
- **Required flag**: Control whether command failures should halt execution
- **Working directory**: Execute commands in specific directories
- **Privilege escalation**: Run commands as different users via sudo
- **Dry-run mode**: Preview commands without execution

## Basic Usage

### Creating and Running a Command

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Create a command
    command := ork.NewCommand().
        WithDescription("Check server uptime").
        WithCommand("uptime").
        WithRequired(true)

    // Run on a single node
    node := ork.NewNodeForHost("server.example.com")
    result := node.Run(command).FirstResult()

    if result.Error != nil {
        log.Printf("Command failed: %v\n", result.Error)
        return
    }

    log.Printf("Command succeeded: %s\n", result.Message)
    log.Printf("Output: %s\n", result.Details["output"])
}
```

### Running on Multiple Nodes

```go
// Create an inventory with multiple nodes
inventory := ork.NewInventory()
prodGroup := ork.NewGroup("production")

prodGroup.AddNode(ork.NewNodeForHost("server1.example.com"))
prodGroup.AddNode(ork.NewNodeForHost("server2.example.com"))
prodGroup.AddNode(ork.NewNodeForHost("server3.example.com"))

inventory.AddGroup(prodGroup)

// Create command to run on all nodes
command := ork.NewCommand().
    WithDescription("Restart application").
    WithCommand("pm2 restart app").
    WithRequired(true)

// Run on all nodes in inventory
results := inventory.Run(command)

// Check results
summary := results.Summary()
log.Printf("Total: %d, Changed: %d, Unchanged: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Unchanged, summary.Failed)

for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("Failed on %s: %v\n", host, result.Error)
    } else {
        log.Printf("Success on %s: %s\n", host, result.Message)
    }
}
```

## Command Configuration

### Required Flag

The `required` flag controls whether command failures should halt execution:

```go
// Command must succeed (default behavior for required)
command := ork.NewCommand().
    WithDescription("Critical operation").
    WithCommand("systemctl restart nginx").
    WithRequired(true)

result := node.Run(command).FirstResult()
if result.Error != nil {
    // Execution stops here - error is returned
    log.Fatal(result.Error)
}
```

```go
// Command can fail without stopping execution
command := ork.NewCommand().
    WithDescription("Non-critical operation").
    WithCommand("some-non-critical-command").
    WithRequired(false)

result := node.Run(command).FirstResult()
if result.Error != nil {
    // Error is logged but doesn't fail the operation
    log.Printf("Non-required command failed (continuing): %v\n", result.Error)
}
log.Printf("Command completed: %s\n", result.Message)
```

### Working Directory

Execute commands in a specific directory using `SetChdir` or `WithChdir`:

```go
// Run command in specific directory
command := ork.NewCommand().
    WithDescription("List files in web directory").
    WithCommand("ls -la").
    WithChdir("/var/www")

result := node.Run(command).FirstResult()

if result.Error != nil {
    log.Printf("Failed: %v\n", result.Error)
} else {
    log.Printf("Success: %s\n", result.Message)
    log.Printf("Output: %s\n", result.Details["output"])
}
```

When combined with privilege escalation, the order is: `cd <dir> && sudo -u <user> <command>`

### Privilege Escalation

Run commands as a different user using `WithBecomeUser`:

```go
// Run database command as postgres user
command := ork.NewCommand().
    WithDescription("Backup database as postgres user").
    WithCommand("pg_dump mydb").
    WithRequired(true).
    WithBecomeUser("postgres")

result := node.Run(command).FirstResult()

if result.Error != nil {
    log.Printf("Failed: %v\n", result.Error)
} else {
    log.Printf("Success: %s\n", result.Message)
}
```

```go
// Run web server command as www-data
command := ork.NewCommand().
    WithDescription("Restart nginx as www-data").
    WithCommand("systemctl restart nginx").
    WithRequired(true).
    WithBecomeUser("www-data")
```

### Dry-Run Mode

Preview commands without executing them:

```go
// Enable dry-run mode on the node
node := ork.NewNodeForHost("server.example.com")
node.SetDryRunMode(true)

// Create command
command := ork.NewCommand().
    WithDescription("Restart application").
    WithCommand("pm2 restart app").
    WithRequired(true)

// Run in dry-run mode - command is logged but not executed
result := node.Run(command).FirstResult()
log.Printf("Dry-run result: %s\n", result.Message)
// Output: "Would execute: pm2 restart app"
```

## Fluent API

Commands support a fluent API for method chaining. Use `With*` methods for consistent chaining:

```go
command := ork.NewCommand().
    WithID("check-uptime").
    WithDescription("Check server uptime").
    WithCommand("uptime").
    WithRequired(true).
    WithChdir("/home/user").
    WithBecomeUser("deploy").
    WithDryRun(false)
```

**Note**: Command-specific methods (`SetCommand`, `SetRequired`, `SetChdir`) return `CommandInterface`, while other methods return `RunnableInterface`. For consistent fluent chaining, use the `With*` variants.

## Arguments

Commands can use arguments like other runnables:

```go
node := ork.NewNodeForHost("server.example.com")
node.SetArg("app-name", "myapp")
node.SetArg("port", "3000")

command := ork.NewCommand().
    WithDescription("Restart application with args").
    WithCommand("pm2 restart ${app-name} --port ${port}").
    WithRequired(true)

result := node.Run(command).FirstResult()
```

Note: Command argument interpolation would need to be implemented in your command logic.

## API Reference

### CommandInterface

```go
type CommandInterface interface {
    types.RunnableInterface

    // Command-specific methods
    SetCommand(cmd string) CommandInterface
    SetRequired(required bool) CommandInterface
    WithCommand(cmd string) CommandInterface
    WithRequired(required bool) CommandInterface
    SetChdir(dir string) CommandInterface
    WithChdir(dir string) CommandInterface

    // Fluent chaining methods
    WithDescription(description string) CommandInterface
    WithID(id string) CommandInterface
    WithArg(key, value string) CommandInterface
    WithArgs(args map[string]string) CommandInterface
    WithNodeConfig(cfg types.NodeConfig) CommandInterface
    WithDryRun(dryRun bool) CommandInterface
    WithTimeout(timeout interface{}) CommandInterface
    WithBecomeUser(user string) CommandInterface
}
```

### Constructor

```go
func NewCommand() CommandInterface
```

Creates a new Command with default values:
- ID: "command"
- Description: "Execute shell command"
- Required: false

### Methods

**Command Configuration:**
- `SetCommand(cmd string) CommandInterface` - Sets the shell command to execute
- `SetRequired(required bool) CommandInterface` - Sets whether the command must succeed
- `SetChdir(dir string) CommandInterface` - Sets the working directory
- `WithCommand(cmd string) CommandInterface` - Fluent alternative to SetCommand
- `WithRequired(required bool) CommandInterface` - Fluent alternative to SetRequired
- `WithChdir(dir string) CommandInterface` - Fluent alternative to SetChdir

**Fluent Chaining:**
- `WithDescription(description string) CommandInterface` - Sets description
- `WithID(id string) CommandInterface` - Sets ID
- `WithArg(key, value string) CommandInterface` - Sets single argument
- `WithArgs(args map[string]string) CommandInterface` - Sets arguments map
- `WithNodeConfig(cfg types.NodeConfig) CommandInterface` - Sets node config
- `WithDryRun(dryRun bool) CommandInterface` - Sets dry-run mode
- `WithTimeout(timeout interface{}) CommandInterface` - Sets timeout
- `WithBecomeUser(user string) CommandInterface` - Sets become user

## Difference from Skills

| Feature | Commands | Skills |
|---------|----------|--------|
| **Purpose** | One-off shell commands | Reusable automation tasks |
| **Idempotency** | Not idempotent | Idempotent by design |
| **Check Method** | Always returns false | Implements actual checks |
| **Required Flag** | Supported | Not applicable |
| **Use Case** | Ad-hoc operations | Common patterns |

## Use Cases

**Use Commands for:**
- Ad-hoc shell commands
- One-off operations
- Quick checks and diagnostics
- Custom scripts that don't need to be reusable

**Use Skills for:**
- Common automation patterns
- Idempotent operations
- Reusable tasks across projects
- Operations that need check/run pattern

## Examples

### System Diagnostics

```go
// Check disk space
command := ork.NewCommand().
    WithDescription("Check disk space").
    WithCommand("df -h").
    WithRequired(true)

result := node.Run(command).FirstResult()
```

### Application Management

```go
// Restart application in specific directory
command := ork.NewCommand().
    WithDescription("Restart Node.js app").
    WithCommand("pm2 restart app").
    WithChdir("/var/www/myapp").
    WithRequired(true)

result := node.Run(command).FirstResult()
```

### Database Operations

```go
// Run database migration as postgres user
command := ork.NewCommand().
    WithDescription("Run database migration").
    WithCommand("psql -f /path/to/migration.sql").
    WithBecomeUser("postgres").
    WithRequired(true)

result := node.Run(command).FirstResult()
```

### Log Analysis

```go
// Check error logs (non-critical)
command := ork.NewCommand().
    WithDescription("Check for errors in logs").
    WithCommand("tail -100 /var/log/app.log | grep ERROR").
    WithRequired(false)

result := node.Run(command).FirstResult()
if result.Error != nil {
    log.Printf("Log check failed but continuing: %v\n", result.Error)
}
```

## See Also

- [Skills](skills.md) - Built-in automation tasks
- [Playbooks](playbooks.md) - Complex orchestration with full Go power
- [Privilege Escalation](privilege_escalation.md) - Running commands as different users
- [API Reference](livewiki/api_reference.md) - Complete API documentation
