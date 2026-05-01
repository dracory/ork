---
path: api_reference.md
page-type: reference
summary: Complete API reference for all public interfaces, functions, and types.
tags: [reference, api, interfaces, vault, prompts, privilege-escalation]
created: 2025-04-14
updated: 2026-05-01
version: 2.2.0
---

# API Reference

## Changelog
- **v2.2.0** (2026-05-01): Added CommandInterface for shell command execution, added Chdir field to NodeConfig for working directory support
- **v2.1.0** (2026-04-15): Added privilege escalation (become) feature with BecomeInterface, BaseBecome, and BecomeUser field in NodeConfig
- **v2.0.0** (2026-04-15): Major terminology refactoring - playbooks renamed to skills, PlaybookInterface renamed to RunnableInterface, BasePlaybook moved to types package, NodeConfig moved to types package, config package removed, playbook package removed
- **v1.2.0** (2026-04-14): Added vault functions for secure secrets management and prompt functions for interactive user input
- **v1.1.0** (2026-04-14): Updated API references from playbook to playbooks package
- **v1.0.0** (2025-04-14): Initial creation

Complete reference for all public APIs in Ork.

## Package Overview

| Package | Path | Purpose |
|---------|------|---------|
| `ork` | `github.com/dracory/ork` | Main API with Node, Group, Inventory |
| `ssh` | `github.com/dracory/ork/ssh` | SSH client utilities |
| `skills` | `github.com/dracory/ork/skills` | Built-in skill implementations |
| `types` | `github.com/dracory/ork/types` | RunnableInterface, BasePlaybook, BaseSkill, Registry, Command, Result, NodeConfig types |

## ork Package

### NodeInterface

The primary interface for managing a single remote server.

```go
type NodeInterface interface {
    RunnerInterface
    
    // Configuration getters
    GetArg(key string) string
    GetArgs() map[string]string
    GetNodeConfig() types.NodeConfig
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
}
```

#### Constructor Functions

```go
// Create node from hostname with defaults
func NewNodeForHost(host string) NodeInterface

// Create empty node for manual configuration
func NewNode() NodeInterface

// Create node from existing config
func NewNodeFromConfig(cfg types.NodeConfig) NodeInterface
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
    Run(runnable types.RunnableInterface) types.Results
    RunByID(id string, opts ...types.RunnableOptions) types.Results
    Check(runnable types.RunnableInterface) types.Results
    GetLogger() *slog.Logger
    SetLogger(logger *slog.Logger) RunnerInterface
    SetDryRunMode(dryRun bool) RunnerInterface
    GetDryRunMode() bool
}
```

### Registry Functions

```go
// Get the global skill registry singleton (lazily initialized)
func GetGlobalPlaybookRegistry() (*types.Registry, error)

// Create a new empty registry
func NewPlaybookRegistry() *types.Registry

// Create a new isolated registry with all built-in skills registered
func NewDefaultRegistry() (*types.Registry, error)
```

### CommandInterface

Fluent interface for executing shell commands on nodes and inventories.

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

    // Fluent chaining methods for RunnableInterface
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

#### Constructor

```go
func NewCommand() CommandInterface
```

#### Methods

**Command Configuration:**
- `SetCommand(cmd string) CommandInterface` - Sets the shell command to execute
- `SetRequired(required bool) CommandInterface` - Sets whether the command must succeed
- `SetChdir(dir string) CommandInterface` - Sets the working directory for command execution
- `WithCommand(cmd string) CommandInterface` - Fluent alternative to SetCommand
- `WithRequired(required bool) CommandInterface` - Fluent alternative to SetRequired
- `WithChdir(dir string) CommandInterface` - Fluent alternative to SetChdir

**Fluent Chaining Methods:**
- `WithDescription(description string) CommandInterface` - Sets description with fluent chaining
- `WithID(id string) CommandInterface` - Sets ID with fluent chaining
- `WithArg(key, value string) CommandInterface` - Sets single argument with fluent chaining
- `WithArgs(args map[string]string) CommandInterface` - Sets arguments map with fluent chaining
- `WithNodeConfig(cfg types.NodeConfig) CommandInterface` - Sets node config with fluent chaining
- `WithDryRun(dryRun bool) CommandInterface` - Sets dry-run mode with fluent chaining
- `WithTimeout(timeout interface{}) CommandInterface` - Sets timeout with fluent chaining
- `WithBecomeUser(user string) CommandInterface` - Sets become user with fluent chaining

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

### Skill Constants

```go
const (
    IDPing              = "ping"
    IDAptUpdate         = "apt-update"
    IDAptUpgrade        = "apt-upgrade"
    IDAptStatus         = "apt-status"
    IDReboot            = "reboot"
    IDSwapCreate        = "swap-create"
    IDSwapDelete        = "swap-delete"
    IDSwapStatus        = "swap-status"
    IDUserCreate        = "user-create"
    IDUserDelete        = "user-delete"
    IDUserStatus        = "user-status"
    // ... see skills/constants.go for full list
)
```

## types Package

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

    // BecomeUser is the user to become when executing commands via sudo
    BecomeUser string

    // Chdir is the working directory for command execution
    Chdir string
}

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

// SetChdir sets the working directory for command execution
func (c *NodeConfig) SetChdir(dir string)
```

### RunnableInterface

All skills must implement this interface (defined in types package).

```go
type RunnableInterface interface {
    // Identification
    GetID() string
    SetID(id string) RunnableInterface
    GetDescription() string
    SetDescription(description string) RunnableInterface

    // Configuration
    GetNodeConfig() NodeConfig
    SetNodeConfig(cfg NodeConfig) RunnableInterface

    // Arguments
    GetArg(key string) string
    SetArg(key, value string) RunnableInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) RunnableInterface

    // Execution options
    IsDryRun() bool
    SetDryRun(dryRun bool) RunnableInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) RunnableInterface

    // Core operations
    Check() (bool, error)
    Run() types.Result

    BecomeInterface
}
```

### BasePlaybook

Provides default implementation of RunnableInterface with fluent API.

```go
type BasePlaybook struct {
    // ... internal fields
}

// Constructor
func NewBasePlaybook() *BasePlaybook
```

### BaseSkill

Provides default implementation of RunnableInterface with Check() and Run() stubs.

```go
type BaseSkill struct {
    // ... internal fields
}

// Constructor
func NewBaseSkill() *BaseSkill
```

### Command

Represents a shell command with its description.

```go
type Command struct {
    Command     string
    Description string
}
```

Used to display and execute shell commands in a structured way, especially useful in dry-run mode to show what commands would be executed.

### BecomeInterface

Defines privilege escalation capabilities for running commands as a different user via sudo.

```go
type BecomeInterface interface {
    SetBecomeUser(user string) BecomeInterface
    GetBecomeUser() string
}
```

#### Methods

```go
// SetBecomeUser sets the user to become when executing commands via sudo
// Returns BecomeInterface for fluent method chaining
SetBecomeUser(user string) BecomeInterface

// GetBecomeUser returns the configured become user
// Returns empty string if not set
GetBecomeUser() string
```

### BaseBecome

Provides a default implementation of BecomeInterface that can be embedded in other structs.

```go
type BaseBecome struct {
    becomeUser string
}
```

Embed `BaseBecome` in your custom skills or playbooks to automatically get privilege escalation support:

```go
type MyCustomSkill struct {
    types.BaseSkill  // Already embeds BaseBecome
}
```

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
func Run(cfg types.NodeConfig, cmd string) (string, error)
```

## skills Package

### System Skills

```go
// Ping
func ping.NewPing() types.RunnableInterface

// Apt
func apt.NewAptUpdate() types.RunnableInterface
func apt.NewAptUpgrade() types.RunnableInterface
func apt.NewAptStatus() types.RunnableInterface

// Reboot
func reboot.NewReboot() types.RunnableInterface
```

### User Management

```go
func user.NewUserCreate() types.RunnableInterface
func user.NewUserDelete() types.RunnableInterface
func user.NewUserList() types.RunnableInterface
func user.NewUserStatus() types.RunnableInterface
```

### Swap Management

```go
func swap.NewSwapCreate() types.RunnableInterface
func swap.NewSwapDelete() types.RunnableInterface
func swap.NewSwapStatus() types.RunnableInterface
```

### Security

```go
func security.NewSshHarden() types.RunnableInterface
func security.NewKernelHarden() types.RunnableInterface
func security.NewAideInstall() types.RunnableInterface
func security.NewAuditdInstall() types.RunnableInterface
func security.NewSshChangePort() types.RunnableInterface
```

### Firewall

```go
func ufw.NewUfwInstall() types.RunnableInterface
func ufw.NewUfwStatus() types.RunnableInterface
func ufw.NewAllowMariaDB() types.RunnableInterface
```

### Fail2ban

```go
func fail2ban.NewFail2banInstall() types.RunnableInterface
func fail2ban.NewFail2banStatus() types.RunnableInterface
```

### MariaDB

```go
func mariadb.NewInstall() types.RunnableInterface
func mariadb.NewSecure() types.RunnableInterface
func mariadb.NewCreateDB() types.RunnableInterface
func mariadb.NewCreateUser() types.RunnableInterface
func mariadb.NewStatus() types.RunnableInterface
func mariadb.NewListDBs() types.RunnableInterface
func mariadb.NewListUsers() types.RunnableInterface
func mariadb.NewBackup() types.RunnableInterface
func mariadb.NewSecurityAudit() types.RunnableInterface
func mariadb.NewChangePort() types.RunnableInterface
func mariadb.NewEnableSSL() types.RunnableInterface
func mariadb.NewEnableEncryption() types.RunnableInterface
func mariadb.NewBackupEncrypt() types.RunnableInterface
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
results := inv.Run(skills.NewAptUpdate())

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

### Creating Custom Skills

```go
type MySkill struct {
    *types.BaseSkill
}

func (p *MySkill) Check() (bool, error) {
    cfg := p.GetNodeConfig()
    output, _ := ssh.Run(cfg, "check command")
    return !strings.Contains(output, "configured"), nil
}

func (p *MySkill) Run() types.Result {
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

func NewMySkill() types.RunnableInterface {
    skill := types.NewBaseSkill()
    skill.SetID("my-skill")
    skill.SetDescription("Does something useful")
    return &MySkill{BaseSkill: skill}
}
```

## See Also

- [Overview](overview.md) - High-level introduction
- [Getting Started](getting_started.md) - Step-by-step guide
- [Architecture](architecture.md) - System architecture
- [Data Flow](data_flow.md) - Data flow diagrams
