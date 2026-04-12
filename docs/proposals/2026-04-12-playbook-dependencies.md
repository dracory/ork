# Proposal: Playbook Dependencies

**Date:** 2026-04-12  
**Status:** Not Implemented  
**Author:** System Review

> **Note:** Allows playbooks to declare dependencies (e.g., apt-upgrade depends on apt-update).

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
    Playbook
    Dependencies() []string // Returns playbook names
}

type ConditionalDependency interface {
    Playbook
    DependenciesFor(cfg Config) []string // Context-aware dependencies
}
```

### 2. Dependency Graph

```go
type DependencyGraph struct {
    nodes map[string]*Node
    edges map[string][]string
}

type Node struct {
    Playbook Playbook
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

func (g *DependencyGraph) AddPlaybook(pb Playbook)
func (g *DependencyGraph) AddDependency(from, to string)
func (g *DependencyGraph) TopologicalSort() ([]Playbook, error)
func (g *DependencyGraph) DetectCycles() error
```

### 3. Dependency Resolver

```go
type DependencyResolver struct {
    registry *playbook.Registry
    cache    map[string]Result // Cache results to avoid re-running
}

func (r *DependencyResolver) Resolve(pb Playbook) ([]Playbook, error) {
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

func (r *DependencyResolver) buildGraph(graph *DependencyGraph, pb Playbook) error {
    graph.AddPlaybook(pb)
    
    // Get dependencies
    var deps []string
    if dpb, ok := pb.(DependentPlaybook); ok {
        deps = dpb.Dependencies()
    }
    
    // Recursively add dependencies
    for _, depName := range deps {
        depPb, ok := r.registry.Get(depName)
        if !ok {
            return fmt.Errorf("dependency '%s' not found", depName)
        }
        
        graph.AddDependency(pb.Name(), depName)
        
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
type AptUpgrade struct{}

func (a *AptUpgrade) Name() string {
    return "apt-upgrade"
}

func (a *AptUpgrade) Description() string {
    return "Install available package updates"
}

func (a *AptUpgrade) Dependencies() []string {
    return []string{"apt-update"} // Must run apt-update first
}

func (a *AptUpgrade) Run(cfg config.Config) error {
    // apt-update already ran, just do the upgrade
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt-get upgrade -y")
    if err != nil {
        return fmt.Errorf("apt upgrade failed: %w", err)
    }
    
    log.Println("Apt upgrade completed successfully")
    return nil
}
```

### Application Deployment with Multiple Dependencies

```go
type DeployWebApp struct{}

func (d *DeployWebApp) Name() string {
    return "deploy-webapp"
}

func (d *DeployWebApp) Description() string {
    return "Deploy web application"
}

func (d *DeployWebApp) Dependencies() []string {
    return []string{
        "user-create",      // Create deploy user
        "install-nginx",    // Install web server
        "install-nodejs",   // Install runtime
        "setup-firewall",   // Configure firewall
    }
}

func (d *DeployWebApp) Run(cfg config.Config) error {
    // All dependencies satisfied, deploy app
    log.Println("Deploying web application...")
    // ... deployment logic ...
    return nil
}
```

### Conditional Dependencies

```go
type InstallDocker struct{}

func (i *InstallDocker) Name() string {
    return "install-docker"
}

func (i *InstallDocker) DependenciesFor(cfg config.Config) []string {
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

func (i *InstallDocker) Run(cfg config.Config) error {
    // Install Docker
    return nil
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
    Playbook
    DependenciesAdvanced() []Dependency
}

type DeployApp struct{}

func (d *DeployApp) DependenciesAdvanced() []Dependency {
    return []Dependency{
        {Name: "install-nodejs", Version: ">=18.0.0"},
        {Name: "install-nginx", Version: ">=1.20.0"},
        {Name: "setup-ssl", Optional: true}, // Nice to have
    }
}
```

## Execution with Dependencies

### Automatic Resolution

```go
func RunWithDependencies(pb Playbook, cfg config.Config, registry *playbook.Registry) error {
    resolver := NewDependencyResolver(registry)
    
    // Resolve execution order
    executionOrder, err := resolver.Resolve(pb)
    if err != nil {
        return fmt.Errorf("failed to resolve dependencies: %w", err)
    }
    
    log.Printf("Execution order: %v", playbookNames(executionOrder))
    
    // Execute in order
    for _, p := range executionOrder {
        // Check cache
        if result, cached := resolver.cache[p.Name()]; cached {
            if result.Error == nil {
                log.Printf("Skipping %s (already completed)", p.Name())
                continue
            }
        }
        
        log.Printf("Running %s...", p.Name())
        err := p.Run(cfg)
        
        // Cache result
        resolver.cache[p.Name()] = Result{Error: err}
        
        if err != nil {
            return fmt.Errorf("dependency '%s' failed: %w", p.Name(), err)
        }
    }
    
    return nil
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
    pb := g.nodes[root].Playbook
    fmt.Printf("%s%s\n", strings.Repeat("  ", indent), pb.Name())
    
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
    Result    Result
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
func (g *DependencyGraph) ExecuteParallel(cfg config.Config) error {
    // Group by dependency level
    levels := g.GetLevels()
    
    for _, level := range levels {
        // Execute all playbooks in this level in parallel
        var wg sync.WaitGroup
        errors := make(chan error, len(level))
        
        for _, pb := range level {
            wg.Add(1)
            go func(p Playbook) {
                defer wg.Done()
                if err := p.Run(cfg); err != nil {
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

func (g *DependencyGraph) GetLevels() [][]Playbook {
    // Return playbooks grouped by dependency level
    // Level 0: No dependencies
    // Level 1: Depends only on level 0
    // Level 2: Depends on level 0 or 1, etc.
}
```

## Implementation Plan

### Phase 1: Core Framework
- Add `DependentPlaybook` interface
- Create `DependencyGraph` with topological sort
- Circular dependency detection

### Phase 2: Execution
- Dependency resolution at runtime
- Caching of completed dependencies
- Parallel execution of independent playbooks

### Phase 3: CLI Integration
- Add `--with-deps` flag
- Add `deps` command for visualization

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
