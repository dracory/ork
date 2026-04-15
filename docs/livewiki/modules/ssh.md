---
path: modules/ssh.md
page-type: module
summary: SSH client utilities and connection management for remote server operations.
tags: [module, ssh, connection, remote]
created: 2025-04-14
updated: 2026-04-15
version: 2.1.0
---

## Changelog
- **v2.1.0** (2026-04-15): Updated config package references to types package, updated terminology from playbooks to skills
- **v1.0.0** (2025-04-14): Initial creation

# ssh Package

SSH connectivity utilities for remote server automation.

## Purpose

The `ssh` package wraps `github.com/sfreiberg/simplessh` with a simplified API for skill-style operations. It provides connection management, command execution, and dry-run safety.

## Key Files

| File | Purpose |
|------|---------|
| `ssh.go` | `Client` struct and methods |
| `functions.go` | Utility functions (`Run`, `PrivateKeyPath`) |
| `ssh_test.go` | SSH tests |

## Client

SSH connection wrapper.

```go
type Client struct {
    host    string    // Hostname or IP
    port    string    // SSH port
    user    string    // Username
    keyPath string    // Full path to private key
    client  *simplessh.Client
}
```

### Constructor

```go
func NewClient(host, port, user, key string) *Client
```

Parameters:
- `host`: Hostname or IP (e.g., "server.example.com")
- `port`: SSH port (e.g., "22" or "2222"). Empty defaults to "22"
- `user`: Username (e.g., "root", "deploy")
- `key`: Key filename (e.g., "id_rsa", "production.prv"). Resolved to `~/.ssh/<key>`

```go
client := ssh.NewClient("server.example.com", "2222", "deploy", "production.prv")
```

### Connect

Establishes the SSH connection.

```go
func (c *Client) Connect() error
```

Must be called before `Run()` or `Close()`. Returns error if host is empty.

```go
client := ssh.NewClient("server.example.com", "22", "root", "id_rsa")
if err := client.Connect(); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Run

Executes a command on the remote server.

```go
func (c *Client) Run(cmd string) (string, error)
```

Returns combined stdout/stderr output. Returns error if not connected.

```go
output, err := client.Run("uptime")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output)
```

### Close

Closes the SSH connection.

```go
func (c *Client) Close() error
```

Safe to call multiple times. Returns nil if already closed.

## Utility Functions

### PrivateKeyPath

Constructs the absolute path to an SSH private key file.

```go
func PrivateKeyPath(sshKey string) string
```

Returns empty string if user lookup fails.

```go
path := ssh.PrivateKeyPath("id_rsa")
// Returns: "/home/username/.ssh/id_rsa"

path2 := ssh.PrivateKeyPath("production.prv")
// Returns: "/home/username/.ssh/production.prv"
```

### Run (with config)

Connects using `NodeConfig` and executes a command. **Includes dry-run safety check**.

```go
func Run(cfg types.NodeConfig, cmd string) (string, error)
```

This is the recommended function for skills because it respects `cfg.IsDryRunMode`.

**SAFETY**: When `cfg.IsDryRunMode` is true, this function will:
1. Log the command via `cfg.Logger`
2. Return `"[dry-run]"` as output
3. NOT execute any commands on the server

```go
cfg := types.NodeConfig{
    SSHHost:      "server.example.com",
    SSHPort:      "22",
    RootUser:     "root",
    SSHKey:       "id_rsa",
    IsDryRunMode: true,  // Safety enabled
}

output, err := ssh.Run(cfg, "apt-get upgrade -y")
// output = "[dry-run]"
// err = nil
// Nothing executed on server!
```

## Usage Patterns

### Pattern 1: Using Run() with Config

For skill development, use Run() with NodeConfig:

```go
cfg := types.NodeConfig{
    SSHHost:  "server.example.com",
    SSHPort:  "22",
    SSHLogin: "root",
    SSHKey:   "id_rsa",
}
output, err := ssh.Run(cfg, "uptime")
```

### Pattern 2: Persistent Connection

For multiple operations on the same server:

```go
client := ssh.NewClient("server.example.com", "22", "root", "id_rsa")

if err := client.Connect(); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Reuse connection
output1, _ := client.Run("uptime")
output2, _ := client.Run("df -h")
output3, _ := client.Run("free -m")
```

### Pattern 3: Config-Based with Dry-Run Safety (Run)

Used by skills for automatic dry-run support:

```go
cfg := types.NodeConfig{
    SSHHost:  "server.example.com",
    SSHPort:  "22",
    SSHLogin: "root",
    SSHKey:   "id_rsa",
    // IsDryRunMode set by parent
}

output, err := ssh.Run(cfg, "command")
// Automatically respects cfg.IsDryRunMode
```

## Integration with Node

The `ork` package uses `ssh` internally:

```go
// node_implementation.go

// One-time connection via Run
output, err := ssh.Run(n.cfg, types.Command{Command: cmd})

// Or persistent connection via stored client
output, err := n.sshClient.Run(cmd)
```

Testing uses `SetRunFunc` for mocking:

```go
// In tests
ssh.SetRunFunc(func(cfg types.NodeConfig, cmd string) (string, error) {
    return "mocked output", nil
})
defer ssh.SetRunFunc(nil)
```

## Error Handling

Common errors:

| Error | Cause | Solution |
|-------|-------|----------|
| "host cannot be empty" | Empty host passed to Connect | Check host parameter |
| "not connected, call Connect() first" | Run() called before Connect() | Call Connect() first |
| "failed to connect to ..." | SSH connection failed | Check host, port, credentials |
| "failed to execute ..." | Command execution failed | Check command syntax, permissions |

## Security Considerations

1. **Key-based authentication only**: No password support
2. **Private key permissions**: Ensure `~/.ssh/` and keys have proper permissions (600)
3. **Dry-run safety**: `Run()` checks `IsDryRunMode` before executing
4. **No password storage**: Passwords not stored in config

## Examples

### Basic Command Execution

```go
package main

import (
    "fmt"
    "log"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/types"
)

func main() {
    cfg := types.NodeConfig{
        SSHHost:  "server.example.com",
        SSHPort:  "22",
        SSHLogin: "root",
        SSHKey:   "id_rsa",
    }
    output, err := ssh.Run(cfg, "uptime")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(output)
}
```

### Multiple Commands

```go
client := ssh.NewClient("server.example.com", "22", "root", "id_rsa")
if err := client.Connect(); err != nil {
    log.Fatal(err)
}
defer client.Close()

commands := []string{"uptime", "df -h", "free -m"}
for _, cmd := range commands {
    output, err := client.Run(cmd)
    if err != nil {
        log.Printf("Command '%s' failed: %v", cmd, err)
        continue
    }
    fmt.Printf("$ %s\n%s\n", cmd, output)
}
```

### In a Skill

```go
func (s *MySkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    
    // Check dry-run - ssh.Run handles this automatically
    output, err := ssh.Run(cfg, "my-command")
    
    if err != nil {
        return types.Result{
            Changed: false,
            Error:   err,
        }
    }
    
    // Check for dry-run marker
    if output == "[dry-run]" {
        return types.Result{
            Changed: true,
            Message: "Would execute: my-command",
        }
    }
    
    return types.Result{
        Changed: true,
        Message: "Executed successfully",
        Details: map[string]string{
            "output": output,
        },
    }
}
```

## See Also

- [ork](ork.md) - Uses ssh package internally
- [skills](skills.md) - Skills use ssh.Run()
- [types](types.md) - NodeConfig used by ssh.Run()
- [Troubleshooting](../troubleshooting.md) - SSH connection issues
