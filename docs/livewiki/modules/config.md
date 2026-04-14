---
path: modules/config.md
page-type: module
summary: Configuration types for SSH-based automation, including NodeConfig with connection settings.
tags: [module, config, configuration]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# config Package

Provides configuration types for SSH-based remote operations.

## Purpose

The `config` package defines the `NodeConfig` struct, which holds all configuration variables for remote server operations. It serves as the central configuration structure passed between packages.

## Key Files

| File | Purpose |
|------|---------|
| `node_config.go` | `NodeConfig` struct definition and methods |

## NodeConfig

Central configuration structure for remote operations.

```go
type NodeConfig struct {
    // SSH connection settings
    SSHHost  string            // Hostname or IP address
    SSHPort  string            // SSH port (default: "22")
    SSHLogin string            // SSH login user
    SSHKey   string            // Private key filename (resolved to ~/.ssh/)
    
    // User settings
    RootUser    string         // Root/admin user
    NonRootUser string         // Non-root user
    
    // Database settings
    DBPort         string      // Database port
    DBRootPassword string      // Database root password
    
    // Extra arguments for playbooks
    Args map[string]string
    
    // Logger for structured logging
    Logger *slog.Logger
    
    // Dry-run mode flag
    IsDryRunMode bool
}
```

## Methods

### SSHAddr

Returns the full SSH address as `host:port`.

```go
func (c NodeConfig) SSHAddr() string
```

Port defaults to "22" if not set.

```go
cfg := config.NodeConfig{
    SSHHost: "server.example.com",
    SSHPort: "2222",
}
addr := cfg.SSHAddr()  // "server.example.com:2222"

// Empty port defaults to "22"
cfg2 := config.NodeConfig{
    SSHHost: "server.example.com",
}
addr2 := cfg2.SSHAddr()  // "server.example.com:22"
```

### GetArg

Retrieves an argument from the `Args` map.

```go
func (c NodeConfig) GetArg(key string) string
```

Returns empty string if not found.

```go
cfg := config.NodeConfig{
    Args: map[string]string{
        "username": "alice",
    },
}

username := cfg.GetArg("username")  // "alice"
missing := cfg.GetArg("unknown")    // ""
```

### GetArgOr

Retrieves an argument with a default value fallback.

```go
func (c NodeConfig) GetArgOr(key, defaultValue string) string
```

```go
cfg := config.NodeConfig{
    Args: map[string]string{
        "size": "2",
    },
}

size := cfg.GetArgOr("size", "1")          // "2" (from args)
swappiness := cfg.GetArgOr("swappiness", "10")  // "10" (default)
```

### GetLoggerOrDefault

Returns the configured logger or `slog.Default()` if nil.

```go
func (c NodeConfig) GetLoggerOrDefault() *slog.Logger
```

```go
cfg := config.NodeConfig{}
logger := cfg.GetLoggerOrDefault()  // slog.Default()

// With custom logger
cfg2 := config.NodeConfig{
    Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
}
logger2 := cfg2.GetLoggerOrDefault()  // custom logger
```

## Usage

### Creating Config

```go
// Direct creation
cfg := config.NodeConfig{
    SSHHost:  "server.example.com",
    SSHPort:  "2222",
    RootUser: "deploy",
    SSHKey:   "production.prv",
    Args: map[string]string{
        "username": "alice",
    },
}
```

### From Node

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222")

cfg := node.GetNodeConfig()
addr := cfg.SSHAddr()
```

### With Playbook

```go
pb := playbooks.NewUserCreate()
pb.SetConfig(cfg)

// Access config within playbook
cfg := pb.GetConfig()
host := cfg.SSHHost
```

## Field Details

### SSHHost

The hostname or IP address of the remote server. This is the primary identifier for the node.

### SSHPort

The SSH port. While stored as a string for convenience, it should represent a valid port number.

- Default in constructors: "22"
- Common alternatives: "2222", "40022"

### SSHLogin

The SSH login user. Usually set from `RootUser` during connection establishment.

### SSHKey

The SSH private key filename (not full path). The SSH package resolves this to `~/.ssh/<filename>`.

- Default: "id_rsa"
- Must be a private key (not .pub)

### RootUser

The root or administrative user for privileged operations.

- Default: "root"
- Often set to "deploy", "admin", or similar

### NonRootUser

A non-privileged user for operations that don't require root access.

### DBPort

Database server port (e.g., "3306" for MariaDB/MySQL).

### DBRootPassword

The database root password, typically used by MariaDB playbooks.

### Args

Key-value map for playbook-specific arguments.

```go
cfg.Args = map[string]string{
    "username": "alice",
    "shell":    "/bin/bash",
    "size":     "2",
}
```

### Logger

Structured logger for output. Uses `slog.Logger` from the standard library.

### IsDryRunMode

When true, operations should not modify the server. The `ssh.Run()` function checks this flag and returns `"[dry-run]"` without executing commands.

## Thread Safety

`NodeConfig` is a value type (struct). It is safely copied between packages. However, the `Args` map and `Logger` pointer are shared across copies.

## Examples

### Complete Configuration

```go
cfg := config.NodeConfig{
    SSHHost:  "db.production.example.com",
    SSHPort:  "40022",
    RootUser: "admin",
    SSHKey:   "production_deploy.prv",
    Args: map[string]string{
        "username":    "appuser",
        "shell":       "/bin/bash",
        "ssh-key":     "ssh-rsa AAAAB3...",
        "database":    "myapp",
        "root-password": "secret123",
    },
    Logger:       slog.Default(),
    IsDryRunMode: false,
}

// Create node from config
node := ork.NewNodeFromConfig(cfg)
```

### Dynamic Configuration

```go
// Base config
baseConfig := config.NodeConfig{
    SSHPort:  "2222",
    RootUser: "deploy",
    SSHKey:   "deploy_key",
}

// Per-node customization
servers := []string{"web1", "web2", "web3"}
for _, server := range servers {
    cfg := baseConfig
    cfg.SSHHost = server + ".example.com"
    cfg.Args = map[string]string{
        "env": "production",
    }
    
    node := ork.NewNodeFromConfig(cfg)
    // ...
}
```

### Helper Usage

```go
cfg := node.GetNodeConfig()

// Build SSH address
addr := cfg.SSHAddr()
// Use for logging or other purposes
log.Printf("Connecting to %s", addr)

// Get arguments with defaults
username := cfg.GetArgOr("username", "defaultuser")
size := cfg.GetArgOr("size", "1")

// Safe logging
logger := cfg.GetLoggerOrDefault()
logger.Info("executing playbook", "host", cfg.SSHHost)
```

## See Also

- [ork](ork.md) - Uses NodeConfig for node configuration
- [playbook](playbook.md) - Playbooks receive NodeConfig
- [ssh](ssh.md) - Uses NodeConfig for connections
- [Configuration](../configuration.md) - Configuration guide
