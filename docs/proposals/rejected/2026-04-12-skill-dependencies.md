# Proposal: Playbook Dependencies

**Date:** 2026-04-12
**Status:** Rejected. Superseded by Playbooks; built on deprecated APIs.
**Rejected:** 2026-07-31
**Author:** System Review

> **Why rejected:** The core problem this proposal targets — ensuring prerequisites run
> before dependent operations (e.g., `apt-update` before `apt-upgrade`, user creation
> before app deployment) — is already handled by Playbooks, which shipped in the same
> release cycle (see [../implemented/2026-04-15-playbooks.md](../implemented/2026-04-15-playbooks.md)
> and [../playbooks.html](../../playbooks.html)). A playbook's `Run()` method sequences
> skills in order today; this proposal's own "What Exists Today" section even
> acknowledges that "skills can be chained manually in playbooks."
>
> Beyond coverage overlap, the proposal as written is no longer buildable in its
> current form:
>
> 1. **Built on a deprecated API.** The resolver design leans on `registry.FindByID()`
>    plus `RunByID(id)`, but `types.RunnerInterface.RunByID` is now marked
>    `// Deprecated: Use Run() instead.` (see `types/runner_interface.go`). Building a
>    new subsystem on top of a deprecated entry point is the wrong direction; the
>    proposal would need to be reworked against `node.Run(runnable)`.
>
> 2. **CLI integration is unreachable.** Phase 3 (`--with-deps` flag, `ork deps`
>    command) depends on the CLI tool proposal, which was itself rejected
>    (see [rejected/2026-04-12-cli-tool.md](rejected/2026-04-12-cli-tool.md) — Ork
>    remains a Go library). A third of the implementation plan is therefore moot.
>
> 3. **Title vs. body conflates skills and playbooks.** Titled "Playbook Dependencies"
>    but the examples declare dependencies on atomic *skills* (`apt-update`,
>    `apt-upgrade`, `user-create`). With playbooks now shipped as `RunnableInterface`
>    orchestrators, the realistic dependency unit would be a playbook depending on
>    other playbooks, not a skill depending on a skill — skills are already atomic and
>    idempotent. The proposal never resolves this ambiguity.
>
> 4. **Code sketches do not compile against the shipped interfaces.** The examples
>    declare `func (a *AptUpgrade) Run() types.Result` (returning `Result`), but
>    `ExecuteParallel` then calls `p.Run()` and treats the return as `error`
>    (`if err := p.Run(); err != nil`). The real `RunnableInterface.Run()` returns
>    `Result`, not `error`. The `DependencyGraph` / `DependencyResolver` sketches would
>    need to be rewritten against the current interfaces before they could even be
>    evaluated.
>
> 5. **The hard question is unanswered.** Open Question #4 — "How to handle
>    dependencies across different hosts?" — is the actually-difficult part, and it
>    has no answer. The resolver design assumes a single `NodeConfig`, but Ork's whole
>    purpose is multi-node inventory execution. A dependency system that only works
>    on a single node is a significant scope limitation that the proposal does not
>    address.
>
> **What is still genuinely valuable** (and worth a fresh, smaller proposal if ever
> needed): (a) cycle detection, if a multi-playbook orchestration layer is ever built;
> (b) within-execution caching ("don't re-run `apt-update` on this host if it already
> ran this session") — but this is better modeled as a property of the runner/registry
> than as a dependency-resolver feature.
>
> **Path forward:** If automatic dependency resolution ever becomes wanted, re-propose
> it against the current `RunnableInterface` / `node.Run(runnable)` APIs, drop the CLI
> integration, and scope it to playbook-on-playbook dependencies with a real answer for
> cross-host execution.

## Current State

### What Exists Today

The Ork framework has a mature skill/registry system but **no dependency resolution**:

- ✅ `types.RunnableInterface` with `GetID()` for skill identification
- ✅ `types.Registry` with `FindByID()` for skill lookup
- ✅ `RunByID(id string)` on `RunnerInterface` (nodes, groups, inventory)
- ✅ Built-in skills with IDs like `"apt-update"`, `"apt-upgrade"`, `"user-create"`
- ✅ Skills can be chained manually in playbooks (see `examples/example_playbook.go`)

### What Does NOT Exist

- ❌ No `DependentPlaybook` interface for declaring dependencies
- ❌ No `DependencyGraph` for topological sorting
- ❌ No circular dependency detection
- ❌ No automatic dependency resolution
- ❌ No dependency caching
- ❌ No `--with-deps` CLI flag
- ❌ No `deps` command for visualization

## Problem Statement

Some playbooks require other playbooks to run first:

- `apt-upgrade` should run `apt-update` first
- Application deployment needs user creation first
- Service restart needs service installation first

Currently, users must manually chain playbooks in the correct order, which is error-prone.

## Proposed Solution

Implement a dependency system that:

1. **Declares dependencies** in playbook metadata
2. **Automatically resolves** dependency order
3. **Executes prerequisites** before main playbook
4. **Detects circular dependencies**
5. **Caches results** to avoid redundant execution

## Core Concepts

### 1. Dependency Interface

```go
type DependentPlaybook interface {
    RunnableInterface
    Dependencies() []string // Returns playbook IDs
}

type ConditionalDependency interface {
    RunnableInterface
    DependenciesFor(cfg NodeConfig) []string // Context-aware dependencies
}
```

### 2. Dependency Graph

```go
type DependencyGraph struct {
    nodes map[string]*Node
    edges map[string][]string
}

type Node struct {
    Runnable RunnableInterface
    State    NodeState
    Result   Result
}

type NodeState string

const (
    StatePending   NodeState = "pending"
    StateRunning   NodeState = "running"
    StateCompleted NodeState = "completed"
    StateFailed    NodeState = "failed"
    StateSkipped   NodeState = "skipped"
)

func (g *DependencyGraph) AddPlaybook(pb RunnableInterface)
func (g *DependencyGraph) AddDependency(from, to string)
func (g *DependencyGraph) TopologicalSort() ([]RunnableInterface, error)
func (g *DependencyGraph) DetectCycles() error
```

### 3. Dependency Resolver

```go
type DependencyResolver struct {
    registry *types.Registry
    cache    map[string]Result // Cache results to avoid re-running
}

func (r *DependencyResolver) Resolve(pb RunnableInterface) ([]RunnableInterface, error) {
    graph := NewDependencyGraph()
    
    // Build dependency graph
    if err := r.buildGraph(graph, pb); err != nil {
        return nil, err
    }
    
    // Check for cycles
    if err := graph.DetectCycles(); err != nil {
        return nil, err
    }
    
    // Return execution order
    return graph.TopologicalSort()
}

func (r *DependencyResolver) buildGraph(graph *DependencyGraph, pb RunnableInterface) error {
    graph.AddPlaybook(pb)
    
    // Get dependencies
    var deps []string
    if dpb, ok := pb.(DependentPlaybook); ok {
        deps = dpb.Dependencies()
    }
    
    // Recursively add dependencies
    for _, depName := range deps {
        depPb, ok := r.registry.FindByID(depName)
        if !ok {
            return fmt.Errorf("dependency '%s' not found", depName)
        }
        
        graph.AddDependency(pb.GetID(), depName)
        
        if err := r.buildGraph(graph, depPb); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Implementation Examples

### AptUpgrade with Dependencies

```go
type AptUpgrade struct {
    *types.BaseSkill
}

func (a *AptUpgrade) Dependencies() []string {
    return []string{"apt-update"} // Must run apt-update first
}

func (a *AptUpgrade) Run() types.Result {
    // apt-update already ran, just do the upgrade
    output, err := ssh.Run(cfg, types.Command{Command: "apt-get upgrade -y"})
    // ...
}
```

### Application Deployment with Multiple Dependencies

```go
type DeployWebApp struct {
    *types.BaseSkill
}

func (d *DeployWebApp) Dependencies() []string {
    return []string{
        "user-create",      // Create deploy user
        "install-nginx",    // Install web server
        "install-nodejs",   // Install runtime
        "setup-firewall",   // Configure firewall
    }
}

func (d *DeployWebApp) Run() types.Result {
    // All dependencies satisfied, deploy app
    // ... deployment logic ...
}
```

### Conditional Dependencies

```go
type InstallDocker struct {
    *types.BaseSkill
}

func (i *InstallDocker) DependenciesFor(cfg NodeConfig) []string {
    // Check OS type
    osType := cfg.GetArg("os_type")
    
    switch osType {
    case "ubuntu", "debian":
        return []string{"apt-update"}
    case "centos", "rhel":
        return []string{"yum-update"}
    default:
        return []string{}
    }
}
```

### Dependency with Version Requirements

```go
type Dependency struct {
    Name    string
    Version string // Optional version constraint
    Optional bool  // If true, continue even if dependency fails
}

type AdvancedDependentPlaybook interface {
    RunnableInterface
    DependenciesAdvanced() []Dependency
}
```

## Execution with Dependencies

### Automatic Resolution

```go
func RunWithDependencies(pb RunnableInterface, cfg NodeConfig, registry *types.Registry) (Result, error) {
    resolver := NewDependencyResolver(registry)
    
    // Resolve execution order
    executionOrder, err := resolver.Resolve(pb)
    if err != nil {
        return Result{}, fmt.Errorf("failed to resolve dependencies: %w", err)
    }
    
    log.Printf("Execution order: %v", playbookNames(executionOrder))
    
    // Execute in order
    for _, p := range executionOrder {
        // Check cache
        if result, cached := resolver.cache[p.GetID()]; cached {
            if result.Error == nil {
                log.Printf("Skipping %s (already completed)", p.GetID())
                continue
            }
        }
        
        log.Printf("Running %s...", p.GetID())
        result := p.Run()
        
        // Cache result
        resolver.cache[p.GetID()] = result
        
        if result.Error != nil {
            return Result{}, fmt.Errorf("dependency '%s' failed: %w", p.GetID(), result.Error)
        }
    }
    
    return Result{Changed: true, Message: "All dependencies completed"}, nil
}
```

### CLI Usage

```bash
# Automatically run dependencies
ork run deploy-webapp --host server.example.com --with-deps

# Show dependency tree without running
ork deps deploy-webapp

# Output:
# deploy-webapp
# ├── user-create
# ├── install-nginx
# │   └── apt-update
# ├── install-nodejs
# │   └── apt-update
# └── setup-firewall
```

### Dependency Visualization

```go
func (g *DependencyGraph) PrintTree(root string, indent int) {
    pb := g.nodes[root].Runnable
    fmt.Printf("%s%s\n", strings.Repeat("  ", indent), pb.GetID())
    
    for _, dep := range g.edges[root] {
        g.PrintTree(dep, indent+1)
    }
}
```

## Dependency Caching

### Cache Strategy

```go
type CacheStrategy string

const (
    CacheNone    CacheStrategy = "none"    // Always re-run
    CacheSession CacheStrategy = "session" // Cache within execution
    CachePersist CacheStrategy = "persist" // Cache across executions
)

type CachedResult struct {
    Playbook  string
    Host      string
    Timestamp time.Time
    Result    types.Results // Note: Results now use types.Results with per-node results map
    TTL       time.Duration
}

func (r *DependencyResolver) SetCacheStrategy(strategy CacheStrategy)
func (r *DependencyResolver) ClearCache()
func (r *DependencyResolver) InvalidateCache(playbook string)
```

### Cache Example

```go
// Session cache (default)
resolver := NewDependencyResolver(registry)
resolver.SetCacheStrategy(CacheSession)

// Run multiple playbooks - shared dependencies run once
RunWithDependencies(playbook1, cfg, registry) // apt-update runs
RunWithDependencies(playbook2, cfg, registry) // apt-update cached

// Persistent cache
resolver.SetCacheStrategy(CachePersist)
resolver.SetCacheTTL(1 * time.Hour)

// apt-update won't run again for 1 hour
RunWithDependencies(playbook1, cfg, registry)
```

## Circular Dependency Detection

```go
func (g *DependencyGraph) DetectCycles() error {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    
    for node := range g.nodes {
        if g.hasCycle(node, visited, recStack) {
            return fmt.Errorf("circular dependency detected involving '%s'", node)
        }
    }
    
    return nil
}

func (g *DependencyGraph) hasCycle(node string, visited, recStack map[string]bool) bool {
    visited[node] = true
    recStack[node] = true
    
    for _, dep := range g.edges[node] {
        if !visited[dep] {
            if g.hasCycle(dep, visited, recStack) {
                return true
            }
        } else if recStack[dep] {
            return true
        }
    }
    
    recStack[node] = false
    return false
}
```

## Parallel Dependency Execution

```go
func (g *DependencyGraph) ExecuteParallel(cfg NodeConfig) error {
    // Group by dependency level
    levels := g.GetLevels()
    
    for _, level := range levels {
        // Execute all playbooks in this level in parallel
        var wg sync.WaitGroup
        errors := make(chan error, len(level))
        
        for _, pb := range level {
            wg.Add(1)
            go func(p RunnableInterface) {
                defer wg.Done()
                if err := p.Run(); err != nil {
                    errors <- err
                }
            }(pb)
        }
        
        wg.Wait()
        close(errors)
        
        // Check for errors
        for err := range errors {
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}

func (g *DependencyGraph) GetLevels() [][]RunnableInterface {
    // Return playbooks grouped by dependency level
    // Level 0: No dependencies
    // Level 1: Depends only on level 0
    // Level 2: Depends on level 0 or 1, etc.
}
```

## Implementation Plan

### Phase 1: Core Framework
- [ ] Add `DependentPlaybook` interface
- [ ] Create `DependencyGraph` with topological sort
- [ ] Circular dependency detection
- [ ] Add `DependencyResolver` with caching

### Phase 2: Execution
- [ ] Dependency resolution at runtime
- [ ] Caching of completed dependencies
- [ ] Parallel execution of independent playbooks

### Phase 3: CLI Integration
- [ ] Add `--with-deps` flag to `ork run`
- [ ] Add `deps` command for visualization

## Benefits

- **Correctness**: Ensure prerequisites are met
- **Convenience**: Automatic dependency resolution
- **Efficiency**: Cache results to avoid redundant work
- **Safety**: Detect circular dependencies
- **Clarity**: Visualize dependency relationships

## Success Metrics

- All complex playbooks declare dependencies
- Zero manual dependency management needed
- Dependency resolution time <100ms
- Clear error messages for missing dependencies

## Open Questions

1. Should dependencies be strict or optional by default?
2. How to handle version conflicts between dependencies?
3. Should we support "soft" dependencies (recommended but not required)?
4. How to handle dependencies across different hosts?

## Related Proposals

- [Privilege Escalation](2026-04-15-privilege-escalation.md) — `RunByID` is used for skill execution

## References

- [Ansible Dependencies](https://docs.ansible.com/ansible/latest/user_guide/playbooks_intro.html#includes-and-imports)
- [Topological Sorting](https://en.wikipedia.org/wiki/Topological_sorting)
- [Dependency Injection](https://en.wikipedia.org/wiki/Dependency_injection)
