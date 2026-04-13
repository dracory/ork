# Ork Enhancement Proposal: Inventory System

**Status:** Implemented  
**Date:** 2026-04-13  
**Author:** Ork Team

## Summary

Introduce an **Inventory** system to Ork, allowing users to define and manage groups of nodes (remote servers). `Node`, `Group`, and `Inventory` all implement a shared `RunnableInterface` enabling operations to run against single nodes, groups, or entire inventories uniformly.

## Motivation

Currently, Ork only supports operating on individual nodes:

```go
node := ork.NewNodeForHost("server1.example.com")
result := node.RunPlaybook(playbooks.NewPing())
```

For managing multiple servers, users must manually iterate:

```go
hosts := []string{"server1", "server2", "server3"}
for _, host := range hosts {
    node := ork.NewNodeForHost(host)
    node.RunPlaybook(playbooks.NewPing())
}
```

An inventory system would provide:
- **Group management** - Organize nodes into logical groups (webservers, dbservers)
- **Group variables** - Define shared configuration per group
- **Parallel execution** - Run playbooks across multiple nodes concurrently
- **Unified interface** - Same API for Node, Group, and Inventory

## Proposal

### 1. Shared Types (types package)

All shared types live in `github.com/dracory/ork/types`:

```go
import "github.com/dracory/ork/types"

// RunnableInterface - implemented by Node and Inventory
type RunnableInterface = types.RunnableInterface

// Results - unified result collection
type Results = types.Results

// Summary - aggregated statistics
type Summary = types.Summary
```

### 2. Interfaces

```go
// InventoryInterface for managing collections of nodes
type InventoryInterface interface {
    RunnableInterface
    AddGroup(group GroupInterface) InventoryInterface
    GetGroupByName(name string) GroupInterface
    AddNode(node NodeInterface) InventoryInterface
    GetNodes() []NodeInterface
    SetMaxConcurrency(max int) InventoryInterface
}

// GroupInterface for managing groups of nodes
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

### 3. Inventory Creation Patterns

**Programmatic creation:**
```go
// Create inventory programmatically
inv := ork.NewInventory()

// Add nodes directly
inv.AddNode("web1.example.com").
    SetPort("2222").
    SetUser("deploy")

// Or add to groups
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))
webGroup.SetArg("env", "production")
inv.AddGroup(webGroup)

dbGroup := ork.NewGroup("dbservers")
dbGroup.AddNode(ork.NewNodeForHost("db1.example.com"))
dbGroup.SetArg("db_role", "primary")
inv.AddGroup(dbGroup)
```

### 4. Running Operations on Inventory

```go
// Run playbook on entire inventory
inv := ork.NewInventory()
webGroup := inv.AddGroup("webservers")
webGroup.AddNode("web1.example.com")

result := inv.RunPlaybook(playbooks.NewPing())

// Check summary
summary := result.Summary()
fmt.Printf("Changed: %d, Unchanged: %d, Failed: %d\n",
    summary.Changed,
    summary.Unchanged,
    summary.Failed)

// Check individual results
for nodeID, nodeResult := range result.Results {
    if nodeResult.Error != nil {
        log.Printf("%s failed: %v", nodeID, nodeResult.Error)
    }
}

// Run on specific group only
webServers := inv.GetGroupByName("webservers")
result := webServers.RunPlaybook(playbooks.NewAptUpgrade())
```

### 5. Variable Precedence

When running playbooks on inventory, variables resolve with this precedence (highest first):

1. Playbook-level args (set via `SetArg()`)
2. Host-level variables
3. Group-level variables
4. Parent group variables
5. Inventory-level variables
6. Node defaults

### 6. Parallel Execution

Inventory operations run concurrently by default:

```go
// Configure concurrency (default: 10)
inv.SetMaxConcurrency(20)

// Run with context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result := inv.RunPlaybookWithContext(ctx, playbooks.NewAptUpgrade())
```

### 7. Node Interface Alignment

Update `NodeInterface` to match `RunnableInterface`:

```go
type NodeInterface interface {
    // ... existing configuration methods ...

    // RunnableInterface is embedded for unified API
    RunnableInterface

    // RunPlaybookByID executes a playbook by ID from the registry
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
}

// RunnableInterface defines operations for Node, Group, and Inventory
type RunnableInterface interface {
    RunCommand(cmd string) types.Results
    RunPlaybook(pb playbook.PlaybookInterface) types.Results
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
    CheckPlaybook(pb playbook.PlaybookInterface) types.Results
}
```

For single nodes, the results collection contains exactly one entry.

## Implementation Status

**COMPLETED** - All core functionality implemented:

- ✅ `types` package with `Result`, `Results`, `Summary`, `RunnableInterface`
- ✅ `InventoryInterface` and `inventoryImplementation`
- ✅ `GroupInterface` and `groupImplementation` 
- ✅ `RunnableInterface` implemented by Node, Group, and Inventory
- ✅ All methods return `types.Results` for unified API
- ✅ Tests updated for new return types

## Future Enhancements

### Phase 4: Parallel Execution
- Worker pool for concurrent operations
- Context support for cancellation
- Error handling strategies

### Phase 5: Advanced Features
- Dynamic inventory plugins
- Host patterns (e.g., `web*.example.com`)
- Limit execution to subsets

## Benefits

1. **Scalability** - Manage 1 or 1000 nodes with the same API
2. **Organization** - Group nodes logically with shared configuration
3. **Performance** - Parallel execution reduces total runtime
4. **Familiarity** - Ansible users will recognize the inventory concept
5. **Flexibility** - Choose single node or inventory based on use case

## Compatibility

- **BREAKING CHANGE**: `RunCommand()` now returns `types.Results` instead of `(string, error)`
- **BREAKING CHANGE**: `RunPlaybook()` now returns `types.Results` instead of `playbook.Result`
- **BREAKING CHANGE**: `RunPlaybookByID()` now returns `types.Results` instead of `playbook.Result`
- `RunPlaybookByID` remains deprecated across all interfaces

## Open Questions

1. Should Inventory support nested groups (groups containing groups)?
2. How should partial failures be handled - fail fast or continue?
3. Should there be a `Limit()` method to restrict execution to subset?
4. How to handle connection pooling across inventory nodes?
