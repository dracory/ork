# Playbooks

**Status:** Proposed
**Created:** 2026-04-15
**Author:** @dracory

## Problem Statement

Composites (builder pattern with `AddSkill()`) are simple but limiting. Users need the full power of Go to:
- Dynamically decide step order based on runtime conditions
- Parse results and make complex decisions
- Spin up new skills dynamically based on previous results
- Use Go's full control flow (loops, conditionals, error handling)
- Transform data between steps
- Implement custom orchestration logic

## Proposal

**Playbooks** are complex orchestration logic that run multiple skills. They implement `RunnableInterface`:

- `node.Run(skill)` - Run skill (implements RunnableInterface)
- `node.Run(playbook)` - Run playbook (implements RunnableInterface)
- `group.Run(playbook)` - Run playbook on all nodes in a group
- `inventory.Run(playbook)` - Run playbook on all nodes in inventory

Playbooks are not simple atomic operations like regular skills - they are elaborate orchestration logic that may include:
- Complex decision making based on runtime conditions
- Dynamic skill creation and execution
- Custom error handling and retry logic
- Data parsing and transformation between steps
- Multi-node orchestration with custom logic

Since playbooks implement `RunnableInterface`, they can use all existing runnable methods and execution APIs. The distinction is purely in the complexity of the `Run()` implementation.

## Architecture

### Core Concept

A playbook implements `RunnableInterface` and orchestrates other skills in its `Run()` method with custom logic. Playbooks can embed `types.BasePlaybook` which provides a foundation for playbook development:

```go
type DeployWebserverPlaybook struct {
    *types.BasePlaybook
}

func NewDeployWebserverPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("deploy-webserver")
    playbook.SetDescription("Deploy web server with custom orchestration")
    return &DeployWebserverPlaybook{BasePlaybook: playbook}
}

func (d *DeployWebserverPlaybook) Run() types.Result {
    cfg := d.GetNodeConfig()
    
    // 1. Run initial skill
    result1 := skills.NewAptStatus()
    result1.SetNodeConfig(cfg)
    statusResult := result1.Run()
    
    // 2. Parse result and decide what to do
    if statusResult.Changed {
        // Updates available - upgrade them
        upgradeSkill := skills.NewAptUpgrade()
        upgradeSkill.SetNodeConfig(cfg)
        upgradeSkill.Run()
    }
    
    // 3. Loop over data and create skills dynamically
    users := []string{"alice", "bob", "charlie"}
    for _, username := range users {
        userSkill := skills.NewUserCreate().
            SetArg("username", username)
        userSkill.SetNodeConfig(cfg)
        userSkill.Run()
    }
    
    // 4. Parse complex results
    dbListResult := skills.NewMariadbListDBs()
    dbListResult.SetNodeConfig(cfg)
    listResult := dbListResult.Run()
    dbs := parseDBList(listResult.Message)
    
    // 5. Spin up skills dynamically based on results
    for _, db := range dbs {
        backupSkill := skills.NewMariadbBackup().
            SetArg("database", db)
        backupSkill.SetNodeConfig(cfg)
        backupSkill.Run()
    }
    
    // 6. Custom error handling and retry logic
    for attempt := 0; attempt < 3; attempt++ {
        downloadSkill := skills.NewDownload()
        downloadSkill.SetNodeConfig(cfg)
        result := downloadSkill.Run()
        if result.Error == nil {
            break
        }
        time.Sleep(time.Second * 5)
    }
    
    return types.Result{Changed: true, Message: "Complete"}
}

// Usage
node.Run(NewDeployWebserverPlaybook())
```

### Helper Utilities

Provide helper utilities for common patterns:

```go
// Helper package: ork/playbook

// RunSequential runs skills in sequence
func RunSequential(runner RunnerInterface, runnables []types.RunnableInterface) types.Results

// RunParallel runs skills concurrently
func RunParallel(runner RunnerInterface, runnables []types.RunnableInterface) types.Results

// RunWithRetry runs skill with retry logic
func RunWithRetry(runner RunnerInterface, runnable types.RunnableInterface, maxAttempts int, backoff time.Duration) types.Result

// ParseResult extracts structured data from skill result
func ParseResult(result types.Result, parser func(string) (interface{}, error)) (interface{}, error)

// SkillBuilder creates skills dynamically
func SkillBuilder(skillType string) func() types.RunnableInterface
```

### Usage Examples

#### Basic Playbook

```go
type DeployWebserverPlaybook struct {
    *types.BasePlaybook
}

func NewDeployWebserverPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("deploy-webserver")
    playbook.SetDescription("Deploy web server")
    return &DeployWebserverPlaybook{BasePlaybook: playbook}
}

func (d *DeployWebserverPlaybook) Run() types.Result {
    cfg := d.GetNodeConfig()

    // Sequential execution
    updateSkill := skills.NewAptUpdate()
    updateSkill.SetNodeConfig(cfg)
    updateSkill.Run()

    upgradeSkill := skills.NewAptUpgrade()
    upgradeSkill.SetNodeConfig(cfg)
    upgradeSkill.Run()

    nginxSkill := skills.NewNginxInstall()
    nginxSkill.SetNodeConfig(cfg)
    nginxSkill.Run()

    pingSkill := skills.NewPing()
    pingSkill.SetNodeConfig(cfg)
    return pingSkill.Run()
}

// Usage
node.Run(NewDeployWebserverPlaybook())
```

#### Dynamic Decision Making

```go
type SmartUpdatePlaybook struct {
    *types.BasePlaybook
}

func NewSmartUpdatePlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("smart-update")
    playbook.SetDescription("Update only if needed")
    return &SmartUpdatePlaybook{BasePlaybook: playbook}
}

func (s *SmartUpdatePlaybook) Run() types.Result {
    cfg := s.GetNodeConfig()

    // Check if updates available
    statusSkill := skills.NewAptStatus()
    statusSkill.SetNodeConfig(cfg)
    statusResult := statusSkill.Run()

    // Parse result and decide
    if statusResult.Changed {
        // Updates available - upgrade them
        upgradeSkill := skills.NewAptUpgrade()
        upgradeSkill.SetNodeConfig(cfg)
        upgradeSkill.Run()

        // Only reboot if critical security update
        if strings.Contains(statusResult.Message, "security") {
            rebootSkill := skills.NewReboot()
            rebootSkill.SetNodeConfig(cfg)
            rebootSkill.Run()
        }
    } else {
        // No updates - just verify connectivity
        pingSkill := skills.NewPing()
        pingSkill.SetNodeConfig(cfg)
        return pingSkill.Run()
    }

    return statusResult
}
```

#### Loop and Dynamic Skill Creation

```go
type SetupUsersPlaybook struct {
    *types.BasePlaybook
}

func NewSetupUsersPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("setup-users")
    playbook.SetDescription("Create multiple users")
    return &SetupUsersPlaybook{BasePlaybook: playbook}
}

func (s *SetupUsersPlaybook) Run() types.Result {
    cfg := s.GetNodeConfig()

    // User list could come from config file, API, database, etc.
    users := []string{"alice", "bob", "charlie", "dave"}

    for _, username := range users {
        // Create skill dynamically
        userSkill := skills.NewUserCreate().
            SetArg("username", username).
            SetArg("shell", "/bin/bash")
        userSkill.SetNodeConfig(cfg)

        result := userSkill.Run()
        if result.Error != nil {
            log.Printf("Failed to create user %s: %v", username, result.Error)
        }
    }

    return types.Result{Changed: true, Message: "Users setup complete"}
}

// Usage
node.Run(NewSetupUsersPlaybook())
```

#### Parse Results and Spin Up Dynamic Skills

```go
type BackupAllDatabasesPlaybook struct {
    *types.BasePlaybook
}

func NewBackupAllDatabasesPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("backup-all-databases")
    playbook.SetDescription("Backup all databases")
    return &BackupAllDatabasesPlaybook{BasePlaybook: playbook}
}

func (b *BackupAllDatabasesPlaybook) Run() types.Result {
    cfg := b.GetNodeConfig()

    // List all databases
    listSkill := skills.NewMariadbListDBs()
    listSkill.SetNodeConfig(cfg)
    listResult := listSkill.Run()

    // Parse the result to get database names
    dbs, err := parseDBList(listResult.Message)
    if err != nil {
        return types.Result{Error: err}
    }

    // Dynamically create backup skills for each database
    for _, db := range dbs {
        backupSkill := skills.NewMariadbBackup().
            SetArg("database", db)
        backupSkill.SetNodeConfig(cfg)

        result := backupSkill.Run()
        if result.Error != nil {
            log.Printf("Failed to backup %s: %v", db, result.Error)
        }
    }

    return types.Result{Changed: true, Message: "Backup complete"}
}

func parseDBList(output string) ([]string, error) {
    // Custom parsing logic
    lines := strings.Split(output, "\n")
    var dbs []string
    for _, line := range lines {
        if strings.Contains(line, "database") {
            // Extract DB name using custom logic
            // ...
            dbs = append(dbs, extractedName)
        }
    }
    return dbs, nil
}

// Usage
node.Run(NewBackupAllDatabasesPlaybook())
```

#### Custom Error Handling and Retry

```go
type ResilientDownloadPlaybook struct {
    *types.BasePlaybook
}

func NewResilientDownloadPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("resilient-download")
    playbook.SetDescription("Download with retry logic")
    return &ResilientDownloadPlaybook{BasePlaybook: playbook}
}

func (r *ResilientDownloadPlaybook) Run() types.Result {
    cfg := r.GetNodeConfig()

    // Custom retry logic with exponential backoff
    var lastError error
    for attempt := 0; attempt < 5; attempt++ {
        downloadSkill := skills.NewDownload()
        downloadSkill.SetNodeConfig(cfg)
        result := downloadSkill.Run()

        if result.Error == nil {
            return result
        }
        lastError = result.Error

        // Exponential backoff
        backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, backoff, result.Error)
        time.Sleep(backoff)
    }

    return types.Result{Error: fmt.Errorf("failed after 5 attempts: %w", lastError)}
}

// Usage
node.Run(NewResilientDownloadPlaybook())
```

#### Complex Orchestration with State

```go
type ComplexDeployPlaybook struct {
    *types.BasePlaybook
}

func NewComplexDeployPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("complex-deploy")
    playbook.SetDescription("Deploy with state and rollback")
    return &ComplexDeployPlaybook{BasePlaybook: playbook}
}

func (c *ComplexDeployPlaybook) Run() types.Result {
    cfg := c.GetNodeConfig()
    state := make(map[string]interface{})

    // Stage 1: Pre-flight checks
    pingSkill := skills.NewPing()
    pingSkill.SetNodeConfig(cfg)
    check1 := pingSkill.Run()
    if check1.Error != nil {
        return check1
    }

    diskCheckSkill := skills.NewDiskCheck()
    diskCheckSkill.SetNodeConfig(cfg)
    check2 := diskCheckSkill.Run()
    state["disk-space"] = parseDiskSpace(check2.Message)

    // Stage 2: Conditional actions based on state
    if state["disk-space"].(float64) < 10.0 {
        // Low disk space - clean up before deploy
        cleanSkill := skills.NewCleanOldLogs()
        cleanSkill.SetNodeConfig(cfg)
        cleanSkill.Run()
    }

    // Stage 3: Deploy
    updateSkill := skills.NewAptUpdate()
    updateSkill.SetNodeConfig(cfg)
    updateSkill.Run()

    deploySkill := skills.NewDeployApp()
    deploySkill.SetNodeConfig(cfg)
    deployResult := deploySkill.Run()

    // Stage 4: Post-deploy validation
    if deployResult.Changed {
        // Only verify if something changed
        healthCheckSkill := skills.NewHealthCheck()
        healthCheckSkill.SetNodeConfig(cfg)
        healthCheck := healthCheckSkill.Run()
        if healthCheck.Error != nil {
            // Rollback
            rollbackSkill := skills.NewRollback()
            rollbackSkill.SetNodeConfig(cfg)
            rollbackSkill.Run()
            return types.Result{Error: healthCheck.Error, Message: "Deploy failed, rolled back"}
        }
    }

    return deployResult
}

// Usage
node.Run(NewComplexDeployPlaybook())
```

#### Multi-Node Programmatic Playbook

```go
type ClusterRolloutPlaybook struct {
    *types.BasePlaybook
}

func NewClusterRolloutPlaybook() types.RunnableInterface {
    playbook := types.NewBasePlaybook()
    playbook.SetID("cluster-rollout")
    playbook.SetDescription("Rollout across cluster with custom logic")
    return &ClusterRolloutPlaybook{BasePlaybook: playbook}
}

func (c *ClusterRolloutPlaybook) Run() types.Result {
    cfg := c.GetNodeConfig()

    // This playbook is designed to run on inventory
    // For multi-node execution, use inv.Run(playbook)
    return types.Result{Message: "Use inventory.Run() for multi-node"}
}

// For multi-node execution, use inventory
func ClusterRolloutMultiNode(inv InventoryInterface) types.Results {
    nodes := inv.GetNodes()

    // Custom node selection logic
    var productionNodes []NodeInterface
    var stagingNodes []NodeInterface

    for _, node := range nodes {
        cfg := node.GetNodeConfig()
        if cfg.Args["environment"] == "production" {
            productionNodes = append(productionNodes, node)
        } else {
            stagingNodes = append(stagingNodes, node)
        }
    }

    // Different logic for different node groups
    // Production: Sequential rollout with health checks
    for i, node := range productionNodes {
        log.Printf("Deploying to production node %d/%d: %s", i+1, len(productionNodes), node.GetHost())

        updateSkill := skills.NewAptUpdate()
        updateSkill.SetNodeConfig(node.GetNodeConfig())
        node.Run(updateSkill)

        deploySkill := skills.NewDeployApp()
        deploySkill.SetNodeConfig(node.GetNodeConfig())
        result := node.Run(deploySkill)

        if result.Error != nil {
            log.Printf("Failed on %s, stopping rollout", node.GetHost())
            break
        }

        // Health check between nodes
        healthSkill := skills.NewHealthCheck()
        healthSkill.SetNodeConfig(node.GetNodeConfig())
        health := node.Run(healthSkill)
        if health.Error != nil {
            log.Printf("Health check failed on %s", node.GetHost())
        }
    }

    // Staging: Parallel deployment
    var wg sync.WaitGroup
    for _, node := range stagingNodes {
        wg.Add(1)
        go func(n NodeInterface) {
            defer wg.Done()
            updateSkill := skills.NewAptUpdate()
            updateSkill.SetNodeConfig(n.GetNodeConfig())
            n.Run(updateSkill)

            deploySkill := skills.NewDeployApp()
            deploySkill.SetNodeConfig(n.GetNodeConfig())
            n.Run(deploySkill)
        }(node)
    }
    wg.Wait()

    pingSkill := skills.NewPing()
    return inv.Run(pingSkill)
}

// Usage
inv.Run(NewClusterRolloutPlaybook())
```

## Helper Utilities

### Sequential Execution Helper

```go
package playbook

// RunSequential runs skills in sequence
func RunSequential(runner RunnerInterface, runnables []types.RunnableInterface) types.Results {
    results := types.Results{
        Results: make(map[string]types.Result),
    }

    for _, runnable := range runnables {
        result := runner.Run(runnable)
        // Merge results
        for host, res := range result.Results {
            results.Results[host] = res
        }

        // Check if any error occurred
        for _, res := range result.Results {
            if res.Error != nil {
                return results  // Stop on first error
            }
        }
    }

    return results
}
```

### Parallel Execution Helper

```go
// RunParallel runs skills concurrently
func RunParallel(runner RunnerInterface, runnables []types.RunnableInterface) types.Results {
    results := types.Results{
        Results: make(map[string]types.Result),
    }

    var wg sync.WaitGroup
    var mu sync.Mutex

    for _, runnable := range runnables {
        wg.Add(1)
        go func(r types.RunnableInterface) {
            defer wg.Done()
            result := runner.Run(r)

            mu.Lock()
            // Merge results
            for host, res := range result.Results {
                results.Results[host] = res
            }
            mu.Unlock()
        }(runnable)
    }
    wg.Wait()

    return results
}
```

### Retry Helper

```go
// RunWithRetry runs skill with exponential backoff retry
func RunWithRetry(runner RunnerInterface, runnable types.RunnableInterface, maxAttempts int, backoff time.Duration) types.Result {
    var lastError error

    for attempt := 0; attempt < maxAttempts; attempt++ {
        results := runner.Run(runnable)
        // Extract first result (for single node execution)
        for _, result := range results.Results {
            if result.Error == nil {
                return result
            }
            lastError = result.Error
        }

        if attempt < maxAttempts-1 {
            time.Sleep(backoff * time.Duration(attempt+1))
        }
    }

    return types.Result{Error: fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastError)}
}
```

### Result Parsing Helper

```go
// ParseResult extracts structured data from skill result
func ParseResult(result types.Result, parser func(string) (interface{}, error)) (interface{}, error) {
    if result.Error != nil {
        return nil, result.Error
    }
    return parser(result.Message)
}
```

## Benefits

1. **Full Go Power**: Complete language capabilities (loops, conditionals, error handling, etc.)
2. **Dynamic Decision Making**: Make choices based on runtime conditions
3. **Complex Parsing**: Parse results and transform data between steps
4. **Dynamic Skill Creation**: Spin up skills based on previous results
5. **Custom Logic**: Implement any orchestration pattern imaginable
6. **Type Safety**: Compiler validates all code
7. **Testing**: Full Go testing framework support
8. **Flexibility**: No constraints on what you can do

## Comparison

| Aspect | Ork Composites | Ork Playbooks | Ork Workflow |
|--------|------------------------------|-------------|--------------|
| Structure | Builder pattern | Implements RunnableInterface | Stages with parallel/sequential |
| Flexibility | Limited to builder methods | Unlimited (full Go) | Configurable |
| Decision Making | Simple conditions only | Full control flow | If/else conditions |
| Result Parsing | Manual in code | Manual in code | Manual in code |
| Dynamic Skills | Limited (loop with transform) | Unlimited | Loop over data/collections |
| Learning Curve | Simple builder API | Go knowledge required | Go knowledge required |
| Type Safety | Compile-time | Compile-time | Compile-time |
| Testing | Mock skills | Mock skills + unit tests | Mock skills + unit tests |
| Use Case | Simple sequential operations | Complex orchestration logic | Complex orchestration needs |
| API | Builder pattern (convenience) | Same Run() as regular skills | New API needed |
| Options | SkillOptions | RunnableOptions | WorkflowOptions |

## When to Use Each

**Use Composites for:**
- Simple sequential operations
- Common patterns that repeat across projects
- Team members who prefer declarative syntax
- 80% of common use cases

**Use Playbooks for:**
- Complex decision logic
- Dynamic skill creation based on results
- Custom error handling and retry logic
- Parsing and transforming data between steps
- Multi-node orchestration with custom logic
- 20% of complex use cases

**Use Workflows for:**
- Complex orchestration with parallel execution
- Explicit dependencies between steps
- State management between steps
- DAG-like execution graphs

## Implementation Considerations

### 1. BasePlaybook

Create `types.BasePlaybook` as a foundation for playbook development. Similar to `types.BaseSkill`, it should:

- Implement `RunnableInterface`
- Provide a default empty `Run()` implementation that returns an error (to force override)
- Provide a default empty `Check()` implementation that returns `(false, nil)` (no changes needed by default)
- Store ID, description, node config, args, dry-run mode, and timeout
- Provide fluent setter methods for chaining

**Package Structure:**
- `types.BaseSkill` is in the `types` package (foundational type for skills)
- `types.BasePlaybook` is in the `types` package (foundational type for playbooks)
- Both ork and skills already import types, so no circular dependency
- Consistent API: both foundational types in the same package

```go
package types

// BasePlaybook provides a foundation for playbook development.
type BasePlaybook struct {
    id          string
    description string
    nodeCfg     config.NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}

// NewBasePlaybook creates a new BasePlaybook.
func NewBasePlaybook() *BasePlaybook {
    return &BasePlaybook{
        args:   make(map[string]string),
        dryRun: false,
    }
}

// Run must be overridden by playbook implementations.
func (b *BasePlaybook) Run() Result {
    return Result{Error: fmt.Errorf("Run() must be implemented by playbook")}
}

// Check returns false (no changes) by default. Can be overridden.
func (b *BasePlaybook) Check() (bool, error) {
    return false, nil
}

// Implement all other RunnableInterface methods...
```

### 2. Interface Hierarchy

The interface hierarchy is:

```go
// RunnableInterface - unified interface for both simple skills and complex playbooks
type RunnableInterface interface {
    GetID() string
    SetID(id string) RunnableInterface
    GetDescription() string
    SetDescription(description string) RunnableInterface
    GetNodeConfig() config.NodeConfig
    SetNodeConfig(cfg config.NodeConfig) RunnableInterface
    GetArg(key string) string
    SetArg(key, value string) RunnableInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) RunnableInterface
    IsDryRun() bool
    SetDryRun(dryRun bool) RunnableInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) RunnableInterface
    Check() (bool, error)
    Run() Result
}
```

Both simple skills and complex playbooks implement `RunnableInterface`. The distinction is purely in the complexity of the `Run()` implementation.

### 3. Execution API

Playbooks use the same execution API as skills through `RunnerInterface.Run()`:

- `node.Run(runnable)` - Run a skill or playbook on a single node
- `group.Run(runnable)` - Run a skill or playbook on all nodes in a group
- `inventory.Run(runnable)` - Run a skill or playbook on all nodes in inventory

The `Run()` method accepts any `RunnableInterface`, which includes both simple skills and complex playbooks. No separate execution API is needed.

### 4. Documentation

Provide examples and patterns for common orchestration patterns:
- Sequential execution in Run()
- Parallel execution with goroutines
- Retry logic
- Result parsing
- Dynamic skill creation
- Multi-node orchestration

### 5. Helper Package

Create `ork/playbook` package with helper utilities:
- `RunSequential()`
- `RunParallel()`
- `RunWithRetry()`
- `ParseResult()`
- Common patterns and examples

## Timeline

**Available Foundation:**
- RunnableInterface implementation (unified interface for skills and playbooks)
- RunnerInterface.Run() method (unified execution API)

**Completed:**
- `types.BaseSkill` foundation struct with `NewBaseSkill()` (in types package)
- `types.BasePlaybook` foundation struct with `NewBasePlaybook()` (in types package)
- Example playbook demonstrating the pattern (`examples/ExamplePlaybook`)

**To Implement:**
- Helper package (`ork/playbook`) with common patterns
- Additional documentation and examples
- Best practices guide for playbook development
