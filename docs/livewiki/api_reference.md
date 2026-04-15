---
path: api_reference.md
page-type: reference
summary: Complete API reference for all public interfaces, functions, and types.
tags: [reference, api, interfaces, vault, prompts]
created: 2025-04-14
updated: 2026-04-14
version: 1.2.0
---

# API Reference

## Changelog
- **v1.2.0** (2026-04-14): Added vault functions for secure secrets management and prompt functions for interactive user input
- **v1.1.0** (2026-04-14): Updated API references from playbook to playbooks package
- **v1.0.0** (2025-04-14): Initial creation

Complete reference for all public APIs in Ork.

## Package Overview

| Package | Path | Purpose |
|---------|------|---------|
| `ork` | `github.com/dracory/ork` | Main API with Node, Group, Inventory |
| `config` | `github.com/dracory/ork/config` | Configuration types |
| `ssh` | `github.com/dracory/ork/ssh` | SSH client utilities |
| `playbook` | `github.com/dracory/ork/playbook` | BasePlaybook implementation |
| `playbooks` | `github.com/dracory/ork/playbooks` | Built-in playbook implementations |
| `types` | `github.com/dracory/ork/types` | PlaybookInterface, Registry, Command, Result types |

## ork Package

### NodeInterface

The primary interface for managing a single remote server.

```go
type NodeInterface interface {
    RunnerInterface
    
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
    RunPlaybookByID(id string, opts ...types.PlaybookOptions) types.Results
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
    RunnerInterface
    
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
    RunnerInterface
    
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

### RunnerInterface

Base interface for all executable entities.

```go
type RunnerInterface interface {
    RunCommand(cmd string) types.Results
    RunPlaybook(pb types.PlaybookInterface) types.Results
    RunPlaybookByID(id string, opts ...types.PlaybookOptions) types.Results
    CheckPlaybook(pb types.PlaybookInterface) types.Results
    GetLogger() *slog.Logger
    SetLogger(logger *slog.Logger) RunnerInterface
    SetDryRunMode(dryRun bool) RunnerInterface
    GetDryRunMode() bool
}
```

### Registry Functions

```go
// Get the global playbook registry singleton (lazily initialized)
func GetGlobalPlaybookRegistry() (*types.Registry, error)

// Create a new empty registry
func NewPlaybookRegistry() *types.Registry

// Create a new isolated registry with all built-in playbooks registered
func NewDefaultRegistry() (*types.Registry, error)
```

### Vault Functions

Secure vault integration for secrets management using envenc.

```go
// Load keys from vault content string
func VaultContentToKeys(vaultContent, vaultPassword string) (map[string]string, error)

// Load keys from vault file
func VaultFileToKeys(vaultFilePath, vaultPassword string) (map[string]string, error)

// Hydrate environment variables from vault content string
func VaultContentToEnv(vaultContent, vaultPassword string) error

// Decrypt vault file and hydrate environment variables
func VaultFileToEnv(vaultFilePath, vaultPassword string) error

// Prompt for password and load keys from vault file
func VaultFileToKeysWithPrompt(vaultFilePath string) (map[string]string, error)

// Prompt for password and hydrate environment variables from vault file
func VaultFileToEnvWithPrompt(vaultFilePath string) error

// Prompt for password and load keys from vault content string
func VaultContentToKeysWithPrompt(vaultContent string) (map[string]string, error)

// Prompt for password and hydrate environment variables from vault content string
func VaultContentToEnvWithPrompt(vaultContent string) error
```

### Prompt Functions

Interactive user input functions for configuration and secrets.

```go
// Prompt for password (hidden input)
func PromptPassword(prompt string) (string, error)

// Prompt for string value
func PromptForString(prompt string) (string, error)

// Prompt for string value with default
func PromptForStringWithDefault(prompt, defaultValue string) (string, error)

// Prompt for password (hidden input)
func PromptForPassword(prompt string) (string, error)

// Prompt for password with confirmation
func PromptForPasswordWithConfirmation(prompt string) (string, error)

// Prompt for integer value
func PromptForInt(prompt string) (int, error)

// Prompt for integer value with default
func PromptForIntWithDefault(prompt string, defaultValue int) (int, error)

// Prompt for boolean value (yes/no)
func PromptForBool(prompt string) (bool, error)

// Prompt for boolean value with default
func PromptForBoolWithDefault(prompt string, defaultValue bool) (bool, error)

// Prompt user to select from list of options
func PromptWithOptions(prompt string, options []string) (int, error)

// Prompt for multiple variables using configuration
func PromptMultiple(configs []types.PromptConfig, existingValues ...map[string]string) (types.PromptResult, error)
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

### Command

Represents a shell command with its description.

```go
type Command struct {
    Command     string
    Description string
}
```

Used to display and execute shell commands in a structured way, especially useful in dry-run mode to show what commands would be executed.

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

### PlaybookInterface

All playbooks must implement this interface (defined in types package).

```go
type PlaybookInterface interface {
    // Identification
    GetID() string
    SetID(id string) PlaybookInterface
    GetDescription() string
    SetDescription(description string) PlaybookInterface
    
    // Configuration
    GetNodeConfig() config.NodeConfig
    SetNodeConfig(cfg config.NodeConfig) PlaybookInterface
    
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

## playbook Package

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
func ping.NewPing() types.PlaybookInterface

// Apt
func apt.NewAptUpdate() types.PlaybookInterface
func apt.NewAptUpgrade() types.PlaybookInterface
func apt.NewAptStatus() types.PlaybookInterface

// Reboot
func reboot.NewReboot() types.PlaybookInterface
```

### User Management

```go
func user.NewUserCreate() types.PlaybookInterface
func user.NewUserDelete() types.PlaybookInterface
func user.NewUserList() types.PlaybookInterface
func user.NewUserStatus() types.PlaybookInterface
```

### Swap Management

```go
func swap.NewSwapCreate() types.PlaybookInterface
func swap.NewSwapDelete() types.PlaybookInterface
func swap.NewSwapStatus() types.PlaybookInterface
```

### Security

```go
func security.NewSshHarden() types.PlaybookInterface
func security.NewKernelHarden() types.PlaybookInterface
func security.NewAideInstall() types.PlaybookInterface
func security.NewAuditdInstall() types.PlaybookInterface
func security.NewSshChangePort() types.PlaybookInterface
```

### Firewall

```go
func ufw.NewUfwInstall() types.PlaybookInterface
func ufw.NewUfwStatus() types.PlaybookInterface
func ufw.NewAllowMariaDB() types.PlaybookInterface
```

### Fail2ban

```go
func fail2ban.NewFail2banInstall() types.PlaybookInterface
func fail2ban.NewFail2banStatus() types.PlaybookInterface
```

### MariaDB

```go
func mariadb.NewInstall() types.PlaybookInterface
func mariadb.NewSecure() types.PlaybookInterface
func mariadb.NewCreateDB() types.PlaybookInterface
func mariadb.NewCreateUser() types.PlaybookInterface
func mariadb.NewStatus() types.PlaybookInterface
func mariadb.NewListDBs() types.PlaybookInterface
func mariadb.NewListUsers() types.PlaybookInterface
func mariadb.NewBackup() types.PlaybookInterface
func mariadb.NewSecurityAudit() types.PlaybookInterface
func mariadb.NewChangePort() types.PlaybookInterface
func mariadb.NewEnableSSL() types.PlaybookInterface
func mariadb.NewEnableEncryption() types.PlaybookInterface
func mariadb.NewBackupEncrypt() types.PlaybookInterface
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
    cfg := p.GetNodeConfig()
    output, _ := ssh.Run(cfg, "check command")
    return !strings.Contains(output, "configured"), nil
}

func (p *MyPlaybook) Run() types.Result {
    needsChange, _ := p.Check()
    if !needsChange {
        return types.Result{
            Changed: false,
            Message: "Already configured",
        }
    }
    
    cfg := p.GetNodeConfig()
    _, err := ssh.Run(cfg, "apply command")
    if err != nil {
        return types.Result{
            Changed: false,
            Error: err,
        }
    }
    
    return types.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}

func NewMyPlaybook() types.PlaybookInterface {
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
