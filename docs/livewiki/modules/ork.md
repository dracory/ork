---
path: modules/ork.md
page-type: module
summary: Main ork package providing Node, Group, and Inventory interfaces for SSH-based server automation, with vault and prompts support.
tags: [module, ork, node, group, inventory, vault, prompts]
created: 2025-04-14
updated: 2026-04-15
version: 2.0.0
---

# ork Package

## Changelog
- **v2.0.0** (2026-04-15): Major terminology refactoring - playbooks renamed to skills, PlaybookInterface renamed to RunnableInterface, config package moved to types
- **v1.2.0** (2026-04-14): Added vault functions for secure secrets management and prompt functions for interactive user input
- **v1.1.0** (2026-04-14): Updated registry functions with GetGlobalRegistry and NewDefaultRegistry
- **v1.0.0** (2025-04-14): Initial creation

The main package providing the public API for Ork. This package defines and implements `NodeInterface`, `GroupInterface`, and `InventoryInterface` for SSH-based server automation.

## Purpose

The `ork` package is the primary entry point for users of the framework. It provides:

- **Node management**: Single server operations via `NodeInterface`
- **Group management**: Multi-server operations via `GroupInterface`
- **Inventory management**: Large-scale operations via `InventoryInterface`
- **Skill execution**: Running automation tasks across nodes

## Key Files

| File | Purpose |
|------|---------|
| `ork.go` | Package documentation and entry point |
| `node_interface.go` | `NodeInterface` definition and constructors |
| `node_implementation.go` | `nodeImplementation` struct |
| `node_interface_test.go` | Node tests |
| `group_implementation.go` | `GroupInterface` implementation |
| `group_implementation_test.go` | Group tests |
| `inventory_interface.go` | `InventoryInterface` definition |
| `inventory_implementation.go` | `InventoryInterface` implementation |
| `inventory_implementation_test.go` | Inventory tests |
| `runner_interface.go` | `RunnerInterface` base interface |
| `constants.go` | Skill ID aliases |
| `registry.go` | Global registry + NewDefaultRegistry factory |
| `vault.go` | Vault functions for secure secrets management |
| `prompts.go` | Interactive prompt functions for user input |

## NodeInterface

Represents a single remote server.

```go
type NodeInterface interface {
    RunnerInterface
    
    // Configuration getters
    GetHost() string
    GetPort() string
    GetUser() string
    GetKey() string
    GetArg(key string) string
    GetArgs() map[string]string
    GetNodeConfig() types.NodeConfig
    
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
    
    // Deprecated: Use Run instead
    RunByID(id string, opts ...types.RunnableOptions) types.Results
}
```

### Constructor Functions

```go
// Create node from hostname (recommended)
func NewNodeForHost(host string) NodeInterface

// Create empty node (configure manually)
func NewNode() NodeInterface

// Create from existing config
func NewNodeFromConfig(cfg types.NodeConfig) NodeInterface
```

### Usage Example

```go
// Create and configure node
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")

// Run command
results := node.RunCommand("uptime")

// Run skill
results = node.Run(skills.NewAptUpdate())

// Persistent connection
node.Connect()
defer node.Close()
results = node.RunCommand("df -h")
```

## GroupInterface

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

### Constructor

```go
func NewGroup(name string) GroupInterface
```

### Usage Example

```go
// Create group
group := ork.NewGroup("webservers")

// Add nodes
group.AddNode(node1)
group.AddNode(node2)

// Set group arguments
group.SetArg("env", "production")

// Run skill on all nodes
results := group.Run(skills.NewPing())
```

## InventoryInterface

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

### Constructor

```go
func NewInventory() InventoryInterface
```

### Usage Example

```go
// Create inventory
inv := ork.NewInventory()

// Add groups
inv.AddGroup(webGroup)
inv.AddGroup(dbGroup)

// Configure concurrency
inv.SetMaxConcurrency(20)

// Run on all nodes
results := inv.Run(skills.NewAptUpdate())
```

## RunnerInterface

Base interface for all executable entities (Node, Group, Inventory).

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

## Skill Constants

Convenient aliases for skill IDs:

```go
const (
    SkillPing              = skills.IDPing
    SkillAptUpdate         = skills.IDAptUpdate
    SkillAptUpgrade        = skills.IDAptUpgrade
    SkillUserCreate        = skills.IDUserCreate
    SkillUserDelete        = skills.IDUserDelete
    SkillSwapCreate        = skills.IDSwapCreate
    // ... see constants.go for full list
)
```

## Registry

Global skill registry for ID-based skill lookup:

```go
// Get the global registry singleton (lazily initialized)
registry, err := ork.GetGlobalRegistry()
if err != nil {
    log.Fatal(err)
}

// Find skill by ID
skill, ok := registry.PlaybookFindByID("apt-update")

// Register custom skill
registry.PlaybookRegister(mySkill)

// Create empty registry for custom configuration
emptyRegistry := ork.NewRegistry()

// Create isolated registry for testing
isolatedRegistry, err := ork.NewDefaultRegistry()
```

## Vault Functions

Secure vault integration for secrets management using envenc. These functions allow you to decrypt vault files and load secrets into environment variables or as key-value maps.

### Loading Keys

```go
// Load keys from vault file with password
keys, err := ork.VaultFileToKeys("vault.envenc", "my-password")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded %d keys\n", len(keys))

// Load keys from vault content string
keys, err := ork.VaultContentToKeys(vaultContent, "my-password")
```

### Hydrating Environment Variables

```go
// Decrypt vault file and set environment variables
err := ork.VaultFileToEnv("vault.envenc", "my-password")
if err != nil {
    log.Fatal(err)
}

// Decrypt vault content string and set environment variables
err := ork.VaultContentToEnv(vaultContent, "my-password")
```

### Interactive Prompts

```go
// Prompt for password and load keys
keys, err := ork.VaultFileToKeysWithPrompt("vault.envenc")
if err != nil {
    log.Fatal(err)
}

// Prompt for password and hydrate environment variables
err := ork.VaultFileToEnvWithPrompt("vault.envenc")
```

## Prompt Functions

Interactive user input functions for configuration and secrets collection. These provide a consistent interface for collecting various types of input from users.

### Basic Prompts

```go
// Prompt for string value
name, err := ork.PromptForString("Enter your name")

// Prompt for string with default
email, err := ork.PromptForStringWithDefault("Email", "user@example.com")

// Prompt for password (hidden)
password, err := ork.PromptForPassword("Password")

// Prompt for password with confirmation
password, err := ork.PromptForPasswordWithConfirmation("Password")
```

### Type-Specific Prompts

```go
// Prompt for integer
port, err := ork.PromptForInt("Port number")

// Prompt for integer with default
port, err := ork.PromptForIntWithDefault("Port number", 8080)

// Prompt for boolean
enabled, err := ork.PromptForBool("Enable feature")

// Prompt for boolean with default
enabled, err := ork.PromptForBoolWithDefault("Enable feature", true)
```

### Selection Prompts

```go
// Prompt user to select from options
options := []string{"Production", "Staging", "Development"}
selection, err := ork.PromptWithOptions("Select environment", options)
fmt.Printf("Selected: %s\n", options[selection])
```

### Multiple Prompts

```go
// Prompt for multiple variables at once
prompts := []types.PromptConfig{
    {Name: "username", Prompt: "Username", Default: "admin", Required: true},
    {Name: "password", Prompt: "Password", Private: true, Confirm: true, Required: true},
    {Name: "port", Prompt: "Port", Default: "8080", Required: false},
}

results, err := ork.PromptMultiple(prompts)
if err != nil {
    log.Fatal(err)
}

username := results["username"]
password := results["password"]
port := results["port"]
```

### With Validation

```go
prompts := []types.PromptConfig{
    {
        Name: "email",
        Prompt: "Email address",
        Required: true,
        Validate: func(value string) error {
            if !strings.Contains(value, "@") {
                return fmt.Errorf("invalid email format")
            }
            return nil
        },
    },
}

results, err := ork.PromptMultiple(prompts)
```

## Dependencies

| Package | Usage |
|---------|-------|
| `types` | `NodeConfig`, `RunnableInterface`, `Registry`, `BasePlaybook`, `BaseSkill`, `Command`, `Result`, `Results`, `Summary` |
| `skills` | Built-in skill implementations |
| `ssh` | SSH command execution |
| `github.com/dracory/envenc` | Vault encryption/decryption for secrets management |

## Thread Safety

- **Node**: Not thread-safe for configuration changes
- **Group**: Thread-safe for dry-run mode (uses `sync.RWMutex`)
- **Inventory**: Thread-safe for concurrent operations on multiple nodes

## Examples

### Single Node Operations

```go
node := ork.NewNodeForHost("server.example.com")

// Simple command
results := node.RunCommand("uptime")

// With arguments
node.SetArg("username", "alice")
results = node.Run(skills.NewUserCreate())

// Check mode
results = node.Check(skills.NewAptUpgrade())
```

### Multi-Node Operations

```go
// Create multiple nodes
nodes := []ork.NodeInterface{
    ork.NewNodeForHost("web1.example.com"),
    ork.NewNodeForHost("web2.example.com"),
    ork.NewNodeForHost("web3.example.com"),
}

// Add to group
group := ork.NewGroup("webservers")
for _, node := range nodes {
    group.AddNode(node)
}

// Run on all
results := group.Run(skills.NewPing())

// Check summary
summary := results.Summary()
fmt.Printf("Total: %d, Changed: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Failed)
```

### Large-Scale Operations

```go
// Create inventory with multiple groups
inv := ork.NewInventory()

// Web servers
web := ork.NewGroup("web")
web.AddNode(ork.NewNodeForHost("web1.example.com"))
web.AddNode(ork.NewNodeForHost("web2.example.com"))
inv.AddGroup(web)

// Database servers
db := ork.NewGroup("db")
db.AddNode(ork.NewNodeForHost("db1.example.com"))
db.AddNode(ork.NewNodeForHost("db2.example.com"))
inv.AddGroup(db)

// Run across all with concurrency control
inv.SetMaxConcurrency(10)
results := inv.Run(skills.NewAptUpdate())
```

## See Also

- [types](types.md) - Configuration types, RunnableInterface, and result types
- [skill](skill.md) - BasePlaybook and BaseSkill in types package
- [skills](skills.md) - Built-in skill implementations
- [API Reference](../api_reference.md) - Complete API
- [Getting Started](../getting_started.md) - Tutorial
