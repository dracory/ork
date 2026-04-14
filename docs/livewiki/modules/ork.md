---
path: modules/ork.md
page-type: module
summary: Main ork package providing Node, Group, and Inventory interfaces for SSH-based server automation.
tags: [module, ork, node, group, inventory]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# ork Package

The main package providing the public API for Ork. This package defines and implements `NodeInterface`, `GroupInterface`, and `InventoryInterface` for SSH-based server automation.

## Purpose

The `ork` package is the primary entry point for users of the framework. It provides:

- **Node management**: Single server operations via `NodeInterface`
- **Group management**: Multi-server operations via `GroupInterface`
- **Inventory management**: Large-scale operations via `InventoryInterface`
- **Playbook execution**: Running automation tasks across nodes

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
| `runnable_interface.go` | `RunnableInterface` base interface |
| `constants.go` | Playbook ID aliases |
| `registry.go` | Global playbook registry initialization |

## NodeInterface

Represents a single remote server.

```go
type NodeInterface interface {
    RunnableInterface
    
    // Configuration getters
    GetHost() string
    GetPort() string
    GetUser() string
    GetKey() string
    GetArg(key string) string
    GetArgs() map[string]string
    GetNodeConfig() config.NodeConfig
    
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
    
    // Deprecated: Use RunPlaybook instead
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
}
```

### Constructor Functions

```go
// Create node from hostname (recommended)
func NewNodeForHost(host string) NodeInterface

// Create empty node (configure manually)
func NewNode() NodeInterface

// Create from existing config
func NewNodeFromConfig(cfg config.NodeConfig) NodeInterface
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

// Run playbook
results = node.RunPlaybook(playbooks.NewAptUpdate())

// Persistent connection
node.Connect()
defer node.Close()
results = node.RunCommand("df -h")
```

## GroupInterface

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

// Run playbook on all nodes
results := group.RunPlaybook(playbooks.NewPing())
```

## InventoryInterface

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
results := inv.RunPlaybook(playbooks.NewAptUpdate())
```

## RunnableInterface

Base interface for all executable entities (Node, Group, Inventory).

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

## Playbook Constants

Convenient aliases for playbook IDs:

```go
const (
    PlaybookPing              = playbook.IDPing
    PlaybookAptUpdate         = playbook.IDAptUpdate
    PlaybookAptUpgrade        = playbook.IDAptUpgrade
    PlaybookUserCreate        = playbook.IDUserCreate
    PlaybookUserDelete        = playbook.IDUserDelete
    PlaybookSwapCreate        = playbook.IDSwapCreate
    // ... see constants.go for full list
)
```

## Registry

Global playbook registry for ID-based playbook lookup:

```go
// Get the global registry
registry := ork.GetDefaultRegistry()

// Find playbook by ID
pb, ok := registry.PlaybookFindByID("apt-update")

// Register custom playbook
registry.PlaybookRegister(myPlaybook)
```

## Dependencies

| Package | Usage |
|---------|-------|
| `config` | `NodeConfig` configuration |
| `playbook` | `PlaybookInterface`, `Result` |
| `ssh` | SSH command execution |
| `types` | `Result`, `Results`, `Summary` |

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
results = node.RunPlaybook(playbooks.NewUserCreate())

// Check mode
results = node.CheckPlaybook(playbooks.NewAptUpgrade())
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
results := group.RunPlaybook(playbooks.NewPing())

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
results := inv.RunPlaybook(playbooks.NewAptUpdate())
```

## See Also

- [config](config.md) - Configuration types
- [playbook](playbook.md) - Playbook interface
- [playbooks](playbooks.md) - Built-in playbooks
- [types](types.md) - Result types
- [Getting Started](../getting_started.md) - Tutorial
