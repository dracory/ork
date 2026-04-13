# Ork Enhancement Proposal: Inventory System

**Status:** Draft  
**Date:** 2026-04-13  
**Author:** Ork Team

## Summary

Introduce an **Inventory** system to Ork, allowing users to define and manage groups of nodes (remote servers) similar to Ansible inventory. Both `Node` and `Inventory` will implement a shared `RunnableInterface` enabling operations to run against single nodes or entire inventories uniformly.

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
- **Unified interface** - Same API for single node or multiple nodes

## Proposal

### 1. RunnableInterface

A shared interface implemented by both `Node` and `Inventory`:

```go
// RunnableInterface defines operations that can be performed on either
// a single Node or an Inventory of nodes.
type RunnableInterface interface {
    // RunCommand executes a shell command and returns the output.
    // For Inventory, runs concurrently across all nodes.
    RunCommand(cmd string) CommandResults

    // RunPlaybook executes a playbook instance.
    // For Inventory, runs concurrently across all nodes.
    RunPlaybook(pb playbook.PlaybookInterface) PlaybookResults

    // RunPlaybookByID executes a playbook by ID from the registry.
    // Deprecated: Use RunPlaybook() instead.
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) PlaybookResults

    // Check runs the playbook's check mode to determine if changes would be made.
    // Returns true if changes are needed, false if already in desired state.
    Check(pb playbook.PlaybookInterface) CheckResults
}
```

### 2. Result Types

Collections of results from multi-node operations:

```go
// CommandResults contains results from running a command on multiple targets.
type CommandResults struct {
    Results map[string]CommandResult // keyed by node identifier
    Errors  []error                  // aggregated errors
}

type CommandResult struct {
    NodeID string
    Output string
    Error  error
}

// PlaybookResults contains results from running a playbook on multiple targets.
type PlaybookResults struct {
    Results map[string]playbook.Result // keyed by node identifier
    Summary PlaybookSummary
}

type PlaybookSummary struct {
    Total    int // total nodes
    Changed  int // nodes where changes were made
    Unchanged int // nodes already in desired state
    Failed   int // nodes with errors
}

// CheckResults contains check results from multiple targets.
type CheckResults struct {
    Results map[string]CheckResult // keyed by node identifier
}

type CheckResult struct {
    NodeID       string
    NeedsChange  bool
    Error        error
}
```

### 3. Inventory Structure

```go
// Inventory manages a collection of nodes organized into groups.
type Inventory struct {
    groups map[string]*Group
    mu     sync.RWMutex
}

// Group represents a collection of nodes with shared variables.
type Group struct {
    Name     string
    Nodes    []NodeInterface
    Vars     map[string]string // group-level variables
    Children []string          // names of child groups
}
```

### 4. Inventory Creation Patterns

**Programmatic creation:**
```go
// Create inventory programmatically
inv := ork.NewInventory()

// Add nodes directly
inv.AddNode("web1.example.com").
    SetPort("2222").
    SetUser("deploy")

// Or add to groups
webGroup := inv.AddGroup("webservers")
webGroup.AddNode("web1.example.com")
webGroup.AddNode("web2.example.com")
webGroup.SetVar("env", "production")

dbGroup := inv.AddGroup("dbservers")
dbGroup.AddNode("db1.example.com")
dbGroup.SetVar("db_role", "primary")
```

**From YAML (similar to Ansible):**
```go
// Load from YAML file
inv, err := ork.NewInventoryFromYAML("inventory.yaml")
if err != nil {
    log.Fatal(err)
}
```

**inventory.yaml format:**
```yaml
all:
  children:
    webservers:
      hosts:
        web1.example.com:
          ansible_port: 2222
        web2.example.com:
      vars:
        env: production
        
    dbservers:
      hosts:
        db1.example.com:
          db_role: primary
        db2.example.com:
          db_role: replica
```

### 5. Running Operations on Inventory

```go
// Run playbook on entire inventory
inv := ork.NewInventoryFromYAML("inventory.yaml")

result := inv.RunPlaybook(playbooks.NewPing())

// Check summary
fmt.Printf("Changed: %d, Unchanged: %d, Failed: %d\n",
    result.Summary.Changed,
    result.Summary.Unchanged,
    result.Summary.Failed)

// Check individual results
for nodeID, nodeResult := range result.Results {
    if nodeResult.Error != nil {
        log.Printf("%s failed: %v", nodeID, nodeResult.Error)
    }
}

// Run on specific group only
webServers := inv.GetGroup("webservers")
result := webServers.RunPlaybook(playbooks.NewAptUpgrade())
```

### 6. Variable Precedence

When running playbooks on inventory, variables resolve with this precedence (highest first):

1. Playbook-level args (set via `SetArg()`)
2. Host-level variables
3. Group-level variables
4. Parent group variables
5. Inventory-level variables
6. Node defaults

### 7. Parallel Execution

Inventory operations run concurrently by default:

```go
// Configure concurrency (default: 10)
inv.SetMaxConcurrency(20)

// Run with context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result := inv.RunPlaybookWithContext(ctx, playbooks.NewAptUpgrade())
```

### 8. Node Interface Alignment

Update `NodeInterface` to match `RunnableInterface`:

```go
type NodeInterface interface {
    // ... existing configuration methods ...

    // RunnableInterface methods return single-result collections
    // for API consistency with Inventory
    RunCommand(cmd string) CommandResults
    RunPlaybook(pb playbook.PlaybookInterface) PlaybookResults
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) PlaybookResults
    Check(pb playbook.PlaybookInterface) CheckResults
}
```

For single nodes, the results collection contains exactly one entry.

## Implementation Phases

### Phase 1: Core Types
- Define `RunnableInterface`
- Create result collection types
- Add to `playbook` package

### Phase 2: Inventory Implementation
- `Inventory` struct
- `Group` struct
- Variable resolution logic

### Phase 3: Node Interface Update
- Modify `NodeInterface` to return result collections
- Update `nodeImplementation`
- Maintain backward compatibility

### Phase 4: Parallel Execution
- Worker pool for concurrent operations
- Context support for cancellation
- Error handling strategies

### Phase 5: YAML Loading
- Inventory YAML parser
- Ansible-compatible format
- Validation

### Phase 6: Advanced Features
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

- Existing single-node code continues to work
- Result type changes are breaking but provide better multi-node support
- `RunPlaybookByID` remains deprecated across both interfaces

## Open Questions

1. Should Inventory support nested groups (groups containing groups)?
2. How should partial failures be handled - fail fast or continue?
3. Should there be a `Limit()` method to restrict execution to subset?
4. How to handle connection pooling across inventory nodes?
