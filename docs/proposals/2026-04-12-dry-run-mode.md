# Proposal: Dry-Run Mode

**Date:** 2026-04-12  
**Status:** Implemented  
**Author:** System Review

## Problem Statement

Users need to preview what changes a playbook will make before executing it. This is critical for:

- Production environments
- Learning what a playbook does
- Debugging playbook logic
- Compliance and audit requirements

## Solution: Safe Mode with Dry-Run Logging

The implementation ensures **safety is guaranteed at the execution layer**, not dependent on playbook cooperation.

### Core Design

**Principle:** No command executes on the server in dry-run mode. Safety is enforced in `ssh.Run()`, not in playbooks.

### Implementation

**1. NodeConfig with IsDryRunMode flag**

```go
type NodeConfig struct {
    // ... existing fields ...
    IsDryRunMode bool
}
```

**2. ssh.Run enforces safety - never executes commands in dry-run mode**

```go
func Run(cfg config.NodeConfig, cmd string) (string, error) {
    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would run", "command", cmd)
        // Return marker that playbook can detect
        return "[dry-run]", nil
    }
    // Normal execution
    return RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.SSHLogin, cfg.SSHKey, cmd)
}
```

**3. RunnableInterface with SetDryRunMode and GetDryRunMode**

```go
type RunnableInterface interface {
    // ... other methods ...
    
    // SetDryRunMode sets whether to simulate execution without making changes.
    // Stores the flag locally; propagation to nodes happens at execution time.
    SetDryRunMode(dryRun bool) RunnableInterface
    
    // GetDryRunMode returns true if dry-run mode is enabled.
    GetDryRunMode() bool
}
```

**4. Runtime Propagation with propagateDryRun() Helper**

Dry-run mode is applied to child nodes/groups at execution time via a consolidated helper method. This approach avoids code duplication and ensures consistency:

```go
// Group implementation with mutex-protected dryRunMode
type groupImplementation struct {
    // ... other fields ...
    dryRunMode bool
    mu         sync.RWMutex
}

// propagateDryRun applies the group's dry-run mode to all nodes
func (g *groupImplementation) propagateDryRun() {
    g.mu.RLock()
    mode := g.dryRunMode
    g.mu.RUnlock()
    for _, node := range g.nodes {
        if node.GetDryRunMode() != mode {
            node.SetDryRunMode(mode)
        }
    }
}

// RunPlaybook uses propagateDryRun before execution
func (g *groupImplementation) RunPlaybook(pb playbook.PlaybookInterface) types.Results {
    g.propagateDryRun()  // Apply dry-run to all nodes before execution
    for _, node := range g.nodes {
        nodeResults := node.RunPlaybook(pb)
        // ...
    }
}

// Inventory implementation with same pattern
type inventoryImplementation struct {
    // ... other fields ...
    dryRunMode bool
    mu         sync.RWMutex
}

func (i *inventoryImplementation) propagateDryRun() {
    i.mu.RLock()
    mode := i.dryRunMode
    i.mu.RUnlock()
    for _, group := range i.groups {
        if group.GetDryRunMode() != mode {
            group.SetDryRunMode(mode)
        }
    }
    for _, node := range i.nodes {
        if node.GetDryRunMode() != mode {
            node.SetDryRunMode(mode)
        }
    }
}
```

**5. Consistent Propagation on Add/Set Operations**

For double safety, propagateDryRun is also called when adding nodes/groups or setting the mode:

```go
// AddNode propagates dry-run mode to new nodes
func (g *groupImplementation) AddNode(node NodeInterface) GroupInterface {
    g.nodes = append(g.nodes, node)
    // Ensure new node inherits current dry-run mode
    if node.GetDryRunMode() != g.GetDryRunMode() {
        node.SetDryRunMode(g.GetDryRunMode())
    }
    return g
}

// SetDryRunMode updates mode and propagates immediately
func (g *groupImplementation) SetDryRunMode(dryRun bool) RunnableInterface {
    g.mu.Lock()
    g.dryRunMode = dryRun
    g.mu.Unlock()
    g.propagateDryRun()  // Propagate immediately for consistency
    return g
}
```

**6. Node-Level Dry-Run Enforcement**

Nodes enforce dry-run mode in both command execution and playbook running:

```go
// RunCommand checks dry-run before SSH execution
func (n *nodeImplementation) RunCommand(cmd string) types.Results {
    if n.cfg.IsDryRunMode {
        n.cfg.GetLoggerOrDefault().Info("dry-run: would run command", 
            "host", n.GetHost(), "command", cmd)
        return types.Results{
            Results: map[string]types.Result{
                n.GetHost(): {Changed: true, Message: "[dry-run]"},
            },
        }
    }
    // Normal SSH execution...
}

// RunPlaybook propagates dry-run to playbook
func (n *nodeImplementation) RunPlaybook(pb playbook.PlaybookInterface) types.Results {
    pb.SetConfig(n.cfg)
    pb.SetDryRun(n.cfg.IsDryRunMode)  // Propagate dry-run to playbook
    result := pb.Run()
    // ...
}
```

### Usage

```go
// Enable dry-run at node level
node.SetDryRunMode(true)
results := node.RunPlaybook(pb)

// Enable dry-run at group level
group.SetDryRunMode(true)
results := group.RunPlaybook(pb)

// Enable dry-run at inventory level
inventory.SetDryRunMode(true)
results := inventory.RunPlaybook(pb)

// Check if dry-run is enabled
if node.GetDryRunMode() {
    log.Println("Running in dry-run mode")
}
```

### Playbook Implementation

Playbooks can optionally detect dry-run mode for better UX:

```go
func (a *AptUpgrade) Run() playbook.Result {
    // In dry-run: ssh.Run returns "[dry-run]" marker, logs the command
    // In real-run: ssh.Run executes command on server
    output, _ := ssh.Run(a.cfg, "apt-get upgrade -y")
    
    if output == "[dry-run]" {
        return playbook.Result{
            Changed: true,
            Message: "Would run: apt-get upgrade -y",
        }
    }
    // Normal execution handling...
}
```

**Key Safety Feature:** Even if a playbook forgets to check for the "[dry-run]" marker, **no command executes on the server**. The playbook might just show confusing output, but the system remains safe. Safety is enforced at the `ssh.Run()` level, not dependent on playbook cooperation.

### Benefits

- **Guaranteed safety** - Zero commands execute on server in dry-run mode
- **Audit logging** - All "would run" commands are logged
- **Simple playbooks** - Optional dry-run awareness, safety is enforced
- **Production ready** - Safe for use in any environment

### Limitations

- Cannot predict if changes are actually needed (no state inspection)
- Output is "what commands would run" not "what would change"

## Implementation Examples

### AptUpgrade in Safe Mode

```go
func (a *AptUpgrade) Run() playbook.Result {
    cfg := a.GetConfig()
    
    // In dry-run: ssh.Run returns "[dry-run]" marker
    // In real-run: ssh.Run executes command
    output, _ := ssh.Run(cfg, "apt-get upgrade -y")
    
    if output == "[dry-run]" {
        return playbook.Result{
            Changed: true,
            Message: "Would run: apt-get upgrade -y",
        }
    }
    
    // Normal execution: parse actual output
    if strings.Contains(output, "0 upgraded") {
        return playbook.Result{Changed: false, Message: "All packages up to date"}
    }
    return playbook.Result{Changed: true, Message: "Packages upgraded successfully"}
}
```

### SwapCreate in Safe Mode

```go
func (s *SwapCreate) Run() playbook.Result {
    sizeStr := s.GetArgOr("size", "1")
    
    // Commands that would run (logged in dry-run, executed otherwise)
    commands := []string{
        fmt.Sprintf("fallocate -l %sG /swapfile", sizeStr),
        "chmod 600 /swapfile",
        "mkswap /swapfile && swapon /swapfile",
        "echo '/swapfile none swap sw 0 0' >> /etc/fstab",
    }
    
    // Check dry-run mode via ssh.Run output marker
    output, _ := ssh.Run(s.cfg, "echo test")
    if output == "[dry-run]" {
        return playbook.Result{
            Changed: true,
            Message: fmt.Sprintf("Would run %d commands to create swap", len(commands)),
        }
    }
    
    // Execute commands normally...
}
```

## Usage Examples

### Programmatic Usage

```go
// Simple dry-run execution
node := ork.NewNode("server.example.com").
    SetPort("22").
    SetKey("id_rsa").
    SetDryRunMode(true)

results := node.RunPlaybook(playbooks.NewAptUpgrade())

// Process results
for host, result := range results.Results {
    if result.Changed {
        fmt.Printf("%s: %s\n", host, result.Message)
    }
}

// Check dry-run status
if node.GetDryRunMode() {
    log.Println("Commands were logged but not executed")
}
```

### CLI Usage (if CLI exists)

```bash
# Preview changes
ork run --host server.example.com --playbook apt-upgrade --dry-run

# Output:
# Would perform the following actions:
#   [execute] apt-get: Would upgrade 23 packages
#     Command: apt-get upgrade -y
#     Packages: nginx, postgresql, redis-server, ...
```

## Implementation Plan

### Phase 1: Core Framework (COMPLETE)
- ✅ `IsDryRunMode` field added to `NodeConfig`
- ✅ `ssh.Run()` checks `IsDryRunMode` and returns `[dry-run]` marker
- ✅ `SetDryRunMode()` and `GetDryRunMode()` added to `RunnableInterface`
- ✅ Implemented on `Node`, `Group`, and `Inventory`
- ✅ Safety enforced at execution layer - no commands execute on server in dry-run mode

### Phase 2: Future Work
- CLI integration with `--dry-run` flag
- Optional: Playbook-level dry-run awareness for better UX

## Benefits

- **Safety**: Preview changes before applying
- **Learning**: Understand what playbooks do
- **Debugging**: Verify playbook logic without side effects
- **Compliance**: Audit trail of planned changes
- **Confidence**: Reduce fear of running automation

## Challenges & Solutions

**Challenge:** Some operations can't be accurately predicted  
**Solution:** Mark actions as "estimated" vs "certain"

**Challenge:** Dry-run might need to run some read-only commands  
**Solution:** Allow read-only SSH commands during dry-run

**Challenge:** State might change between dry-run and actual run  
**Solution:** Document this limitation, recommend immediate execution after dry-run

## Success Metrics

- All core playbooks handle `[dry-run]` marker from ssh.Run()
- Zero commands execute on server when dry-run mode is enabled
- User feedback indicates increased confidence

## Open Questions

1. Should playbooks report detailed actions during dry-run?
2. How to handle conditional logic that depends on command output?
