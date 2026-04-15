# Proposal: Parallel Execution

**Date:** 2026-04-12  
**Status:** Implemented  
**Author:** System Review

> **Note:** Parallel execution implemented in `InventoryInterface` using semaphore-based concurrency control with panic recovery.

## Problem Statement

Currently, playbooks run sequentially against one host at a time. For managing multiple servers, this is inefficient:

- Updating 10 servers takes 10x the time of updating 1 server
- No built-in support for running the same playbook across a fleet
- Manual goroutine management required for parallel operations

Ansible's parallel execution (`-f` flag) is essential for managing infrastructure at scale.

## Proposed Solution

### 1. Multi-Host Executor

```go
type Executor struct {
    MaxParallel int           // Max concurrent executions
    StopOnError bool          // Stop all if one fails
    Timeout     time.Duration // Per-host timeout
}

type HostResult struct {
    Host   string
    Config config.Config
    Result types.Result
    Error  error
    Duration time.Duration
}

func (e *Executor) RunOnHosts(p types.PlaybookInterface, configs []config.Config) []HostResult
```

### 2. Inventory Management (IMPLEMENTED)

See `2026-04-13-inventory.md` for Inventory system design.

Inventory now provides:
- `InventoryInterface` with `RunnableInterface`
- `GroupInterface` with `RunnableInterface`
- `SetMaxConcurrency()` for parallel execution control
- **Worker pool with semaphore-based concurrency** in all operations:
  - `RunCommand()` - executes shell commands across all nodes concurrently
  - `RunPlaybook()` - runs playbooks across all nodes concurrently
  - `RunPlaybookByID()` - executes playbooks by ID with concurrency
  - `CheckPlaybook()` - runs check mode across all nodes concurrently
- **Panic recovery** in all goroutines to prevent deadlocks
- **Thread-safe operations** with mutex protection for inventory state

### 3. Progress Tracking

```go
type ProgressTracker struct {
    Total     int
    Completed int
    Failed    int
    Running   int
    mu        sync.RWMutex
}

func (pt *ProgressTracker) OnStart(host string)
func (pt *ProgressTracker) OnComplete(host string, err error)
func (pt *ProgressTracker) Summary() string
```

## Implementation

### Actual Implementation Pattern

The parallel execution is implemented directly in `InventoryInterface` methods using Go's semaphore pattern:

```go
func (i *inventoryImplementation) RunPlaybook(pb types.PlaybookInterface) types.Results {
    results := types.Results{Results: make(map[string]types.Result)}
    nodes := i.GetNodes()

    // Determine concurrency limit (0 = unlimited)
    concurrency := i.maxConcurrency
    if concurrency == 0 {
        concurrency = len(nodes)
    }

    // Semaphore for concurrency control
    sem := make(chan struct{}, concurrency)
    var wg sync.WaitGroup

    for _, node := range nodes {
        wg.Add(1)
        go func(n NodeInterface) {
            defer func() {
                // Panic recovery prevents deadlocks
                if r := recover(); r != nil {
                    i.GetLogger().Error("panic in goroutine", "error", r)
                    i.mu.Lock()
                    results.Results[n.GetHost()] = types.Result{
                        Changed: false,
                        Message: fmt.Sprintf("panic: %v", r),
                    }
                    i.mu.Unlock()
                }
                wg.Done()
            }()

            sem <- struct{}{}        // Acquire
            defer func() { <-sem }() // Release

            nodeResults := n.RunPlaybook(pb)

            i.mu.Lock()
            maps.Copy(results.Results, nodeResults.Results)
            i.mu.Unlock()
        }(node)
    }

    wg.Wait()
    return results
}
```

### Inventory-Based Execution

```go
func main() {
    // Load inventory
    inv, err := LoadInventory("inventory.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get web servers
    webServers := inv.GetGroup("webservers")
    
    // Convert to configs
    configs := make([]config.Config, len(webServers))
    for i, host := range webServers {
        configs[i] = host.ToConfig()
    }
    
    // Run playbook
    executor := NewExecutor(10, false)
    results := executor.RunOnHosts(playbooks.NewNginxReload(), configs)
    
    // Summary
    fmt.Printf("Completed: %d, Failed: %d\n", 
        countSuccess(results), countFailures(results))
}
```

### With Progress Tracking

```go
func RunWithProgress(p types.PlaybookInterface, configs []config.Config) {
    tracker := &ProgressTracker{Total: len(configs)}
    
    executor := &Executor{
        MaxParallel: 5,
        OnStart: func(host string) {
            tracker.OnStart(host)
            fmt.Printf("\r%s", tracker.Summary())
        },
        OnComplete: func(host string, err error) {
            tracker.OnComplete(host, err)
            fmt.Printf("\r%s", tracker.Summary())
        },
    }
    
    results := executor.RunOnHosts(p, configs)
    fmt.Printf("\n%s\n", tracker.Summary())
}
```

## Executor Implementation

```go
package executor

import (
    "context"
    "sync"
    "time"
)

type Executor struct {
    MaxParallel int
    StopOnError bool
    Timeout     time.Duration
    OnStart     func(host string)
    OnComplete  func(host string, err error)
}

func NewExecutor(maxParallel int, stopOnError bool) *Executor {
    return &Executor{
        MaxParallel: maxParallel,
        StopOnError: stopOnError,
        Timeout:     5 * time.Minute,
    }
}

func (e *Executor) RunOnHosts(p types.PlaybookInterface, configs []config.Config) []HostResult {
    results := make([]HostResult, len(configs))
    var wg sync.WaitGroup
    
    // Create semaphore for concurrency control
    sem := make(chan struct{}, e.MaxParallel)
    
    // Context for cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    for i, cfg := range configs {
        wg.Add(1)
        go func(idx int, config config.Config) {
            defer wg.Done()
            
            // Acquire semaphore
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                results[idx] = HostResult{
                    Host:  config.SSHHost,
                    Error: ctx.Err(),
                }
                return
            }
            
            // Check if we should stop
            if e.StopOnError {
                select {
                case <-ctx.Done():
                    results[idx] = HostResult{
                        Host:  config.SSHHost,
                        Error: ctx.Err(),
                    }
                    return
                default:
                }
            }
            
            // Execute with timeout
            result := e.executeWithTimeout(ctx, p, config)
            results[idx] = result
            
            // Cancel all if error and StopOnError
            if e.StopOnError && result.Error != nil {
                cancel()
            }
            
            // Callback
            if e.OnComplete != nil {
                e.OnComplete(config.SSHHost, result.Error)
            }
        }(i, cfg)
    }
    
    wg.Wait()
    return results
}

func (e *Executor) executeWithTimeout(ctx context.Context, p types.PlaybookInterface, cfg config.Config) HostResult {
    result := HostResult{
        Host:   cfg.SSHHost,
        Config: cfg,
    }
    
    start := time.Now()
    
    // Create timeout context
    timeoutCtx, cancel := context.WithTimeout(ctx, e.Timeout)
    defer cancel()
    
    // Run in goroutine
    done := make(chan types.Result, 1)
    go func() {
        done <- p.Run(cfg)
    }()
    
    // Wait for completion or timeout
    select {
    case runResult := <-done:
        result = runResult
    case <-timeoutCtx.Done():
        result.Error = fmt.Errorf("timeout after %v", e.Timeout)
    }
    
    result.Duration = time.Since(start)
    return result
}
```

## Inventory Format

### JSON Format
```json
{
  "hosts": [
    {
      "name": "web1",
      "address": "web1.example.com",
      "port": "22",
      "user": "root",
      "ssh_key": "id_rsa",
      "groups": ["webservers", "production"],
      "vars": {
        "nginx_version": "1.20"
      }
    }
  ],
  "groups": {
    "webservers": ["web1", "web2", "web3"],
    "databases": ["db1", "db2"]
  }
}
```

### YAML Format (Alternative)
```yaml
hosts:
  - name: web1
    address: web1.example.com
    port: 22
    user: root
    ssh_key: id_rsa
    groups:
      - webservers
      - production
    vars:
      nginx_version: "1.20"

groups:
  webservers:
    - web1
    - web2
  databases:
    - db1
    - db2
```

## Implementation Plan

### Phase 1: Core Executor (IMPLEMENTED)
- ✅ Semaphore-based worker pool in `InventoryInterface` methods
- ✅ Panic recovery in all goroutines
- ✅ Thread-safe result collection with mutex protection

### Phase 2: Inventory Integration (IMPLEMENTED)
- ✅ `InventoryInterface` with `RunnableInterface`
- ✅ `SetMaxConcurrency()` method (default: 10, 0 = unlimited)
- ✅ `types.Results` for unified result collection
- ✅ Thread-safe `AddNode()`, `AddGroup()`, `GetNodes()` operations

### Phase 3: Advanced Features (FUTURE)
- Progress tracking callbacks
- Rolling updates
- Failure threshold handling
- Per-operation timeout control

## Benefits

- **Speed**: Update 100 servers in the time it takes to update 10
- **Scalability**: Manage large fleets efficiently
- **Flexibility**: Control parallelism based on infrastructure
- **Visibility**: Track progress across all hosts
- **Reliability**: Continue on failure or stop based on policy

## Safety Considerations

- **Rate Limiting**: Don't overwhelm infrastructure
- **Failure Handling**: Decide when to stop vs continue
- **Resource Limits**: Respect system resources (file descriptors, memory)
- **Timeouts**: Prevent hanging on unresponsive hosts

## Success Metrics

- Execute playbook on 100 hosts in <2 minutes (vs 100+ minutes sequential)
- Zero resource exhaustion issues
- Clear progress visibility
- Configurable parallelism works correctly

## Open Questions

1. Should inventory support dynamic sources (cloud APIs)?
2. Should we support serial execution within parallel (e.g., rolling updates)?
3. How to handle dependencies between hosts (e.g., update DB before app servers)?
4. Should Executor be a separate package or part of Inventory?
