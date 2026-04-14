---
path: api_reference.md
page-type: reference
summary: Complete API reference for all public interfaces, functions, and types.
tags: [reference, api, interfaces]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# API Reference

Complete reference for all public APIs in Ork.

## Package Overview

| Package | Path | Purpose |
|---------|------|---------|
| `ork` | `github.com/dracory/ork` | Main API with Node, Group, Inventory |
| `config` | `github.com/dracory/ork/config` | Configuration types |
| `ssh` | `github.com/dracory/ork/ssh` | SSH client utilities |
| `playbook` | `github.com/dracory/ork/playbook` | Playbook interface and registry |
| `playbooks` | `github.com/dracory/ork/playbooks` | Built-in playbook implementations |
| `types` | `github.com/dracory/ork/types` | Shared result types |

## ork Package

### NodeInterface

The primary interface for managing a single remote server.

```go
type NodeInterface interface {
    RunnableInterface
    
    // Configuration getters
    GetArg(key string) string
    GetArgs() map[string]string
    GetNodeConfig() config.NodeConfig
    GetHost() string
    GetUser() string
    GetKey() string
    GetPort() string
    
    // Configuration setters (fluent)
    SetPort(port string) NodeInterface
    SetUser(user string) NodeInterface
    SetKey(key string) NodeInterface
    SetArg(key, value string) NodeInterface
    SetArgs(args map[string]string) NodeInterface
    
    // Connection management
    Connect() error
    Close() error
    IsConnected() bool
    
    // Playbook by ID (deprecated, use RunPlaybook)
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
}
```

#### Constructor Functions

```go
// Create node from hostname with defaults
func NewNodeForHost(host string) NodeInterface

// Create empty node for manual configuration
func NewNode() NodeInterface

// Create node from existing config
func NewNodeFromConfig(cfg config.NodeConfig) NodeInterface
```

### GroupInterface

Manages a collection of nodes.

```go
type GroupInterface interface {
    RunnableInterface
    
    GetName() string
    AddNode(node NodeInterface) GroupInterface
    GetNodes() []NodeInterface
    SetArg(key, value string) GroupInterface
    GetArg(key string) string
    GetArgs() map[string]string
}
```

#### Constructor

```go
func NewGroup(name string) GroupInterface
```

### InventoryInterface

Manages multiple groups for large-scale operations.

```go
type InventoryInterface interface {
    RunnableInterface
    
    AddGroup(group GroupInterface) InventoryInterface
    GetGroupByName(name string) GroupInterface
    AddNode(node NodeInterface) InventoryInterface
    GetNodes() []NodeInterface
    SetMaxConcurrency(max int) InventoryInterface
}
```

#### Constructor

```go
func NewInventory() InventoryInterface
```

### RunnableInterface

Base interface for all executable entities.

```go
type RunnableInterface interface {
    RunCommand(cmd string) types.Results
    RunPlaybook(pb playbook.PlaybookInterface) types.Results
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
    CheckPlaybook(pb playbook.PlaybookInterface) types.Results
    GetLogger() *slog.Logger
    SetLogger(logger *slog.Logger) RunnableInterface
    SetDryRunMode(dryRun bool) RunnableInterface
    GetDryRunMode() bool
}
```

### Registry Functions

```go
// Get the global playbook registry
func GetDefaultRegistry() *playbook.Registry
```

### Playbook Constants

```go
const (
    PlaybookPing              = "ping"
    PlaybookAptUpdate         = "apt-update"
    PlaybookAptUpgrade        = "apt-upgrade"
    PlaybookAptStatus         = "apt-status"
    PlaybookReboot            = "reboot"
    PlaybookSwapCreate        = "swap-create"
    PlaybookSwapDelete        = "swap-delete"
    PlaybookSwapStatus        = "swap-status"
    PlaybookUserCreate        = "user-create"
    PlaybookUserDelete        = "user-delete"
    PlaybookUserStatus        = "user-status"
    // ... see playbook/constants.go for full list
)
```

## config Package

### NodeConfig

Central configuration structure for all remote operations.

```go
type NodeConfig struct {
    // SSH connection settings
    SSHHost  string
    SSHPort  string
    SSHLogin string
    SSHKey   string
    
    // User settings
    RootUser    string
    NonRootUser string
    
    // Database settings
    DBPort         string
    DBRootPassword string
    
    // Extra arguments
    Args map[string]string
    
    // Logger for structured logging
    Logger *slog.Logger
    
    // Dry-run mode flag
    IsDryRunMode bool
}
```

#### Methods

```go
// SSHAddr returns host:port
func (c NodeConfig) SSHAddr() string

// GetArg retrieves from Args map
func (c NodeConfig) GetArg(key string) string

// GetArgOr retrieves with default value
func (c NodeConfig) GetArgOr(key, defaultValue string) string

// GetLoggerOrDefault returns logger or slog.Default()
func (c NodeConfig) GetLoggerOrDefault() *slog.Logger
```

## types Package

### Result

Individual operation result.

```go
type Result struct {
    Changed bool
    Message string
    Details map[string]string
    Error   error
}
```

### Results

Collection of results across multiple nodes.

```go
type Results struct {
    Results map[string]Result
}

// Methods
func (r Results) Summary() Summary
```

### Summary

Aggregated statistics.

```go
type Summary struct {
    Total     int
    Changed   int
    Unchanged int
    Failed    int
}
```

## playbook Package

### PlaybookInterface

All playbooks must implement this interface.

```go
type PlaybookInterface interface {
    // Identification
    GetID() string
    SetID(id string) PlaybookInterface
    GetDescription() string
    SetDescription(description string) PlaybookInterface
    
    // Configuration
    GetConfig() config.NodeConfig
    SetConfig(cfg config.NodeConfig) PlaybookInterface
    
    // Arguments
    GetArg(key string) string
    SetArg(key, value string) PlaybookInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) PlaybookInterface
    
    // Execution options
    IsDryRun() bool
    SetDryRun(dryRun bool) PlaybookInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) PlaybookInterface
    
    // Core operations
    Check() (bool, error)
    Run() Result
}
```

### Result

```go
type Result struct {
    Changed bool
    Message string
    Details map[string]string
    Error   error
}
```

### PlaybookOptions

```go
type PlaybookOptions struct {
    Args    map[string]string
    DryRun  bool
    Timeout time.Duration
}
```

### Registry

```go
type Registry struct {
    playbooks map[string]PlaybookInterface
    mu        sync.RWMutex
}

// Methods
func NewRegistry() *Registry
func (r *Registry) PlaybookRegister(pb PlaybookInterface) error
func (r *Registry) PlaybookFindByID(id string) (PlaybookInterface, bool)
func (r *Registry) PlaybookList() []PlaybookInterface
```

### BasePlaybook

Provides default implementation of PlaybookInterface.

```go
type BasePlaybook struct {
    // ... internal fields
}

// Constructor
func NewBasePlaybook() *BasePlaybook
```

## ssh Package

### Client

SSH connection wrapper.

```go
type Client struct {
    host    string
    port    string
    user    string
    keyPath string
    client  *simplessh.Client
}
```

#### Methods

```go
func NewClient(host, port, user, key string) *Client
func (c *Client) Connect() error
func (c *Client) Run(cmd string) (string, error)
func (c *Client) Close() error
```

### Utility Functions

```go
// RunOnce connects, runs command, and closes
func RunOnce(host, port, user, key, cmd string) (string, error)

// PrivateKeyPath returns full path to SSH key
func PrivateKeyPath(sshKey string) string

// Run with dry-run safety
func Run(cfg config.NodeConfig, cmd string) (string, error)
```

## playbooks Package

### System Playbooks

```go
// Ping
func ping.NewPing() playbook.PlaybookInterface

// Apt
func apt.NewAptUpdate() playbook.PlaybookInterface
func apt.NewAptUpgrade() playbook.PlaybookInterface
func apt.NewAptStatus() playbook.PlaybookInterface

// Reboot
func reboot.NewReboot() playbook.PlaybookInterface
```

### User Management

```go
func user.NewUserCreate() playbook.PlaybookInterface
func user.NewUserDelete() playbook.PlaybookInterface
func user.NewUserStatus() playbook.PlaybookInterface
```

### Swap Management

```go
func swap.NewSwapCreate() playbook.PlaybookInterface
func swap.NewSwapDelete() playbook.PlaybookInterface
func swap.NewSwapStatus() playbook.PlaybookInterface
```

### Security

```go
func security.NewSshHarden() playbook.PlaybookInterface
func security.NewKernelHarden() playbook.PlaybookInterface
func security.NewAideInstall() playbook.PlaybookInterface
func security.NewAuditdInstall() playbook.PlaybookInterface
func security.NewSshChangePort() playbook.PlaybookInterface
```

### Firewall

```go
func ufw.NewUfwInstall() playbook.PlaybookInterface
func ufw.NewUfwStatus() playbook.PlaybookInterface
func ufw.NewAllowMariaDB() playbook.PlaybookInterface
```

### Fail2ban

```go
func fail2ban.NewFail2banInstall() playbook.PlaybookInterface
func fail2ban.NewFail2banStatus() playbook.PlaybookInterface
```

### MariaDB

```go
func mariadb.NewInstall() playbook.PlaybookInterface
func mariadb.NewSecure() playbook.PlaybookInterface
func mariadb.NewCreateDB() playbook.PlaybookInterface
func mariadb.NewCreateUser() playbook.PlaybookInterface
func mariadb.NewStatus() playbook.PlaybookInterface
func mariadb.NewListDBs() playbook.PlaybookInterface
func mariadb.NewListUsers() playbook.PlaybookInterface
func mariadb.NewBackup() playbook.PlaybookInterface
func mariadb.NewSecurityAudit() playbook.PlaybookInterface
func mariadb.NewChangePort() playbook.PlaybookInterface
func mariadb.NewEnableSSL() playbook.PlaybookInterface
func mariadb.NewEnableEncryption() playbook.PlaybookInterface
func mariadb.NewBackupEncrypt() playbook.PlaybookInterface
```

## Usage Examples

### Basic Node Operations

```go
// Create and configure node
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")

// Get configuration
host := node.GetHost()  // "server.example.com"
port := node.GetPort()  // "2222"

// Set playbook arguments
node.SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

// Run command
results := node.RunCommand("uptime")
result := results.Results["server.example.com"]
```

### Working with Results

```go
results := inv.RunPlaybook(playbooks.NewAptUpdate())

// Get summary
summary := results.Summary()
fmt.Printf("Total: %d, Changed: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Failed)

// Iterate over results
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    } else if result.Changed {
        log.Printf("%s changed: %s", host, result.Message)
    } else {
        log.Printf("%s unchanged", host)
    }
    
    // Access details
    for key, value := range result.Details {
        log.Printf("  %s: %s", key, value)
    }
}
```

### Creating Custom Playbooks

```go
type MyPlaybook struct {
    *playbook.BasePlaybook
}

func (p *MyPlaybook) Check() (bool, error) {
    cfg := p.GetConfig()
    output, _ := ssh.Run(cfg, "check command")
    return !strings.Contains(output, "configured"), nil
}

func (p *MyPlaybook) Run() playbook.Result {
    needsChange, _ := p.Check()
    if !needsChange {
        return playbook.Result{
            Changed: false,
            Message: "Already configured",
        }
    }
    
    cfg := p.GetConfig()
    _, err := ssh.Run(cfg, "apply command")
    if err != nil {
        return playbook.Result{
            Changed: false,
            Error: err,
        }
    }
    
    return playbook.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}

func NewMyPlaybook() playbook.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID("my-playbook")
    pb.SetDescription("Does something useful")
    return &MyPlaybook{BasePlaybook: pb}
}
```

## See Also

- [Overview](overview.md) - High-level introduction
- [Getting Started](getting_started.md) - Step-by-step guide
- [Architecture](architecture.md) - System architecture
- [Data Flow](data_flow.md) - Data flow diagrams
