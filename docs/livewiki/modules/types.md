---
path: modules/types.md
page-type: module
summary: Shared result types for operation outcomes across all Ork packages.
tags: [module, types, results]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# types Package

Shared types for operation results across all Ork packages.

## Purpose

The `types` package provides common result types used by the `ork`, `playbook`, and other packages. It defines the data structures for representing operation outcomes from commands and playbooks.

## Key Files

| File | Purpose |
|------|---------|
| `results.go` | Result, Results, and Summary types |

## Result

Represents the outcome of a single operation (command or playbook execution).

```go
type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}
```

### Changed

Indicates whether the operation modified the system.

- `true`: Changes were made (e.g., packages updated, user created)
- `false`: System was already in the desired state (idempotent operation)

```go
if result.Changed {
    log.Println("System was modified")
} else {
    log.Println("No changes needed")
}
```

### Message

A human-readable description of what happened.

```go
// Examples:
result.Message = "Package database updated"
result.Message = "User 'alice' created"
result.Message = "2GB swap file created"
result.Message = "Already configured - no changes made"
```

### Details

Additional key-value information about the operation.

```go
result.Details = map[string]string{
    "size":        "2GB",
    "file":        "/swapfile",
    "swappiness":  "10",
    "output":      "... command output ...",
}

// Access details
for key, value := range result.Details {
    log.Printf("%s: %s", key, value)
}
```

Common detail keys by playbook:

| Playbook | Detail Keys |
|----------|-------------|
| ping | `uptime` |
| apt-update | `output` |
| swap-create | `size`, `file`, `swappiness`, `status` |
| user-create | `username`, `home`, `shell` |
| mariadb-status | Various status fields |

### Error

Non-nil if the operation failed. When `Error` is non-nil, `Changed` may still be `true` if some changes occurred before the failure.

```go
if result.Error != nil {
    log.Fatalf("Operation failed: %v", result.Error)
}
```

## Results

Contains per-node results from any operation (command or playbook) on multiple nodes.

```go
type Results struct {
    Results map[string]Result  // Key is node hostname
}
```

The `Results` map keys are hostnames (as configured when creating nodes).

### Summary

Returns aggregated statistics about all results.

```go
func (r Results) Summary() Summary
```

```go
results := inv.RunPlaybook(playbooks.NewPing())
summary := results.Summary()

fmt.Printf("Total: %d\n", summary.Total)
fmt.Printf("Changed: %d\n", summary.Changed)
fmt.Printf("Unchanged: %d\n", summary.Unchanged)
fmt.Printf("Failed: %d\n", summary.Failed)
```

### Iterating Results

```go
results := group.RunPlaybook(playbooks.NewAptUpdate())

for hostname, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s: FAILED - %v", hostname, result.Error)
    } else if result.Changed {
        log.Printf("%s: CHANGED - %s", hostname, result.Message)
    } else {
        log.Printf("%s: OK (no changes)", hostname)
    }
}
```

## Summary

Aggregated statistics from a `Results` collection.

```go
type Summary struct {
    Total     int  // Total number of nodes
    Changed   int  // Nodes where changes were made
    Unchanged int  // Nodes with no changes needed
    Failed    int  // Nodes where execution failed
}
```

### Usage

```go
results := inv.RunPlaybook(playbooks.NewAptUpgrade())
summary := results.Summary()

// Quick status check
if summary.Failed > 0 {
    log.Printf("WARNING: %d nodes failed", summary.Failed)
}

if summary.Changed == summary.Total {
    log.Println("All nodes were updated")
}
```

## Type Relationships

```mermaid
graph TD
    A[Node/Group/Inventory] -->|RunCommand| B[types.Results]
    A -->|RunPlaybook| B
    B -->|Summary| C[types.Summary]
    B -->|Results map| D[types.Result]
    E[Playbook] -->|Run| F[playbook.Result]
    F -->|Converted| D
```

Note: `playbook.Result` and `types.Result` have identical structures but are defined separately for package isolation.

## Conversion

The `ork` package converts `playbook.Result` to `types.Result`:

```go
// From playbook.Result
pbResult := playbook.Result{
    Changed: true,
    Message: "Success",
    Details: map[string]string{"key": "value"},
    Error:   nil,
}

// To types.Result
typesResult := types.Result{
    Changed: pbResult.Changed,
    Message: pbResult.Message,
    Details: pbResult.Details,
    Error:   pbResult.Error,
}
```

## Examples

### Handling Single Node Result

```go
node := ork.NewNodeForHost("server.example.com")
results := node.RunCommand("uptime")

// Single node - key is the hostname
result := results.Results["server.example.com"]

if result.Error != nil {
    log.Fatalf("Command failed: %v", result.Error)
}

log.Println(result.Message)
```

### Handling Multiple Node Results

```go
inv := ork.NewInventory()
// ... add groups with nodes ...

results := inv.RunPlaybook(playbooks.NewPing())

// Get summary first
summary := results.Summary()
log.Printf("Ping results: %d total, %d failed", 
    summary.Total, summary.Failed)

// Process individual results
for hostname, result := range results.Results {
    if result.Error != nil {
        log.Printf("[%s] Connection failed: %v", hostname, result.Error)
        continue
    }
    
    log.Printf("[%s] Connected: %s", hostname, result.Message)
    
    if uptime, ok := result.Details["uptime"]; ok {
        log.Printf("[%s] Uptime: %s", hostname, uptime)
    }
}
```

### Error Handling Patterns

```go
results := group.RunPlaybook(playbooks.NewAptUpgrade())

// Pattern 1: Fail on any error
for hostname, result := range results.Results {
    if result.Error != nil {
        log.Fatalf("%s failed: %v", hostname, result.Error)
    }
}

// Pattern 2: Collect errors and report
var failures []string
for hostname, result := range results.Results {
    if result.Error != nil {
        failures = append(failures, fmt.Sprintf("%s: %v", hostname, result.Error))
    }
}
if len(failures) > 0 {
    log.Printf("Completed with %d failures:\n%s", 
        len(failures), strings.Join(failures, "\n"))
}

// Pattern 3: Continue on errors, report at end
summary := results.Summary()
if summary.Failed > 0 {
    log.Printf("WARNING: %d/%d nodes failed", summary.Failed, summary.Total)
}
```

### Check Mode Results

```go
// Preview changes
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would upgrade: %s", result.Message)
    // Now actually run it
    results = node.RunPlaybook(playbooks.NewAptUpgrade())
} else {
    log.Println("No upgrades needed")
}
```

## Design Notes

### Why Separate types.Result and playbook.Result?

Package isolation prevents circular dependencies:
- `playbook` package defines `playbook.Result`
- `types` package defines `types.Result` 
- `ork` package converts between them
- Both have identical structure for consistency

### Map Key Choice

Results use hostname as the map key because:
1. Hostname is the primary identifier for nodes
2. IP addresses may change
3. Natural fit for `NewNodeForHost(host)` pattern
4. Human-readable in logs and output

### Empty Results

Always check if hostname exists in the map:

```go
result, ok := results.Results["server.example.com"]
if !ok {
    log.Fatal("No result for host")
}
```

## See Also

- [ork](ork.md) - Uses types.Results for all operations
- [playbook](playbook.md) - Defines playbook.Result
- [API Reference](../api_reference.md) - Complete API
