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

**Playbooks** are complex orchestration logic that run multiple skills. They implement `PlaybookInterface` which extends `SkillInterface`:

- `node.RunSkill(skill)` - Run skill (implements SkillInterface)
- `node.RunSkill(playbook)` - Run playbook (implements PlaybookInterface → SkillInterface)
- `group.RunSkill(playbook)` - Run playbook on all nodes in a group
- `inventory.RunSkill(playbook)` - Run playbook on all nodes in inventory

Playbooks are not simple atomic operations like regular skills - they are elaborate orchestration logic that may include:
- Complex decision making based on runtime conditions
- Dynamic skill creation and execution
- Custom error handling and retry logic
- Data parsing and transformation between steps
- Multi-node orchestration with custom logic

Since `PlaybookInterface` extends `SkillInterface`, playbooks can use all existing skill methods and execution APIs. The distinction is purely in the complexity of the `Run()` implementation.

## Architecture

### Core Concept

A playbook implements `PlaybookInterface` (which extends `SkillInterface`) and orchestrates other skills in its `Run()` method with custom logic:

```go
type DeployWebserverPlaybook struct {
    ID          string
    Description string
    nodeCfg     config.NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}

func NewDeployWebserverPlaybook() types.PlaybookInterface {
    return &DeployWebserverPlaybook{
        ID:          "deploy-webserver",
        Description: "Deploy web server with custom orchestration",
    }
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
        node.RunSkill(upgradeSkill)
    }
    
    // 3. Loop over data and create skills dynamically
    users := []string{"alice", "bob", "charlie"}
    for _, username := range users {
        userSkill := skills.NewUserCreate().
            SetArg("username", username)
        userSkill.SetNodeConfig(cfg)
        node.RunSkill(userSkill)
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
        node.RunSkill(backupSkill)
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
node.RunSkill(NewDeployWebserverPlaybook())
```

### Helper Utilities

Provide helper utilities for common patterns:

```go
// Helper package: ork/playbook

// RunSequential runs skills in sequence
func RunSequential(node NodeInterface, skills []types.SkillInterface) types.Results

// RunParallel runs skills concurrently
func RunParallel(node NodeInterface, skills []types.SkillInterface) types.Results

// RunWithRetry runs skill with retry logic
func RunWithRetry(node NodeInterface, skill types.SkillInterface, maxAttempts int, backoff time.Duration) types.Result

// ParseResult extracts structured data from skill result
func ParseResult(result types.Result, parser func(string) (interface{}, error)) (interface{}, error)

// SkillBuilder creates skills dynamically
func SkillBuilder(skillType string) func() types.SkillInterface
```

### Usage Examples

#### Basic Playbook

```go
type DeployWebserverPlaybook struct {
    *skills.BaseSkill
}

func NewDeployWebserverPlaybook() types.SkillInterface {
    skill := skills.NewBaseSkill()
    skill.SetID("deploy-webserver")
    skill.SetDescription("Deploy web server")
    return &DeployWebserverPlaybook{BaseSkill: skill}
}

func (d *DeployWebserverPlaybook) Run() types.Result {
    cfg := d.GetNodeConfig()
    
    // Sequential execution
    updateSkill := skills.NewAptUpdate()
    updateSkill.SetNodeConfig(cfg)
    node.RunSkill(updateSkill)
    
    upgradeSkill := skills.NewAptUpgrade()
    upgradeSkill.SetNodeConfig(cfg)
    node.RunSkill(upgradeSkill)
    
    nginxSkill := skills.NewNginxInstall()
    nginxSkill.SetNodeConfig(cfg)
    node.RunSkill(nginxSkill)
    
    pingSkill := skills.NewPing()
    pingSkill.SetNodeConfig(cfg)
    return node.RunSkill(pingSkill)
}

// Usage
node.RunSkill(NewDeployWebserverPlaybook())
```

#### Dynamic Decision Making

```go
type SmartUpdatePlaybook struct {
    *skills.BaseSkill
}

func NewSmartUpdatePlaybook() types.SkillInterface {
    skill := skills.NewBaseSkill()
    skill.SetID("smart-update")
    skill.SetDescription("Update only if needed")
    return &SmartUpdatePlaybook{BaseSkill: skill}
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
        node.RunSkill(upgradeSkill)
        
        // Only reboot if critical security update
        if strings.Contains(statusResult.Message, "security") {
            rebootSkill := skills.NewReboot()
            rebootSkill.SetNodeConfig(cfg)
            node.RunSkill(rebootSkill)
        }
    } else {
        // No updates - just verify connectivity
        pingSkill := skills.NewPing()
        pingSkill.SetNodeConfig(cfg)
        return node.RunSkill(pingSkill)
    }
    
    return statusResult
}
```

#### Loop and Dynamic Skill Creation

```go
func SetupUsers() *Playbook {
    return &Playbook{
        Name: "setup-users",
        Description: "Create multiple users",
        Execute: func(node NodeInterface) types.Results {
            // User list could come from config file, API, database, etc.
            users := []string{"alice", "bob", "charlie", "dave"}
            
            for _, username := range users {
                // Create skill dynamically
                userSkill := skills.NewUserCreate().
                    SetArg("username", username).
                    SetArg("shell", "/bin/bash")
                
                result := node.RunSkill(userSkill)
                if result.Error != nil {
                    log.Printf("Failed to create user %s: %v", username, result.Error)
                }
            }
            
            return types.Results{Results: map[string]types.Result{
                node.GetHost(): types.Result{Changed: true, Message: "Users setup complete"},
            }}
        },
    }
}
```

#### Parse Results and Spin Up Dynamic Skills

```go
func BackupAllDatabases() *Playbook {
    return &Playbook{
        Name: "backup-all-databases",
        Description: "Backup all databases",
        Execute: func(node NodeInterface) types.Results {
            // List all databases
            listResult := node.RunSkill(skills.NewMariadbListDBs())
            
            // Parse the result to get database names
            dbs, err := parseDBList(listResult.Message)
            if err != nil {
                return types.Results{Results: map[string]types.Result{
                    node.GetHost(): types.Result{Error: err},
                }}
            }
            
            // Dynamically create backup skills for each database
            for _, db := range dbs {
                backupSkill := skills.NewMariadbBackup().
                    SetArg("database", db)
                
                result := node.RunSkill(backupSkill)
                if result.Error != nil {
                    log.Printf("Failed to backup %s: %v", db, result.Error)
                }
            }
            
            return types.Results{Results: map[string]types.Result{
                node.GetHost(): types.Result{Changed: true, Message: "Backup complete"},
            }}
        },
    }
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
```

#### Custom Error Handling and Retry

```go
func ResilientDownload() *Playbook {
    return &Playbook{
        Name: "resilient-download",
        Description: "Download with retry logic",
        Execute: func(node NodeInterface) types.Results {
            // Custom retry logic with exponential backoff
            var lastError error
            for attempt := 0; attempt < 5; attempt++ {
                result := node.RunSkill(skills.NewDownload())
                if result.Error == nil {
                    return types.Results{Results: map[string]types.Result{
                        node.GetHost(): result,
                    }}
                }
                lastError = result.Error
                
                // Exponential backoff
                backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
                log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, backoff, result.Error)
                time.Sleep(backoff)
            }
            
            return types.Results{Results: map[string]types.Result{
                node.GetHost(): types.Result{Error: fmt.Errorf("failed after 5 attempts: %w", lastError)},
            }}
        },
    }
}
```

#### Complex Orchestration with State

```go
func ComplexDeploy() *Playbook {
    return &Playbook{
        Name: "complex-deploy",
        Description: "Deploy with state and rollback",
        Execute: func(node NodeInterface) types.Results {
            state := make(map[string]interface{})
            
            // Stage 1: Pre-flight checks
            check1 := node.RunSkill(skills.NewPing())
            if check1.Error != nil {
                return types.Results{Results: map[string]types.Result{
                    node.GetHost(): check1,
                }}
            }
            
            check2 := node.RunSkill(skills.NewDiskCheck())
            state["disk-space"] = parseDiskSpace(check2.Message)
            
            // Stage 2: Conditional actions based on state
            if state["disk-space"].(float64) < 10.0 {
                // Low disk space - clean up before deploy
                node.RunSkill(skills.NewCleanOldLogs())
            }
            
            // Stage 3: Deploy
            node.RunSkill(skills.NewAptUpdate())
            deployResult := node.RunSkill(skills.NewDeployApp())
            
            // Stage 4: Post-deploy validation
            if deployResult.Changed {
                // Only verify if something changed
                healthCheck := node.RunSkill(skills.NewHealthCheck())
                if healthCheck.Error != nil {
                    // Rollback
                    node.RunSkill(skills.NewRollback())
                    return types.Results{Results: map[string]types.Result{
                        node.GetHost(): types.Result{Error: healthCheck.Error, Message: "Deploy failed, rolled back"},
                    }}
                }
            }
            
            return types.Results{Results: map[string]types.Result{
                node.GetHost(): deployResult,
            }}
        },
    }
}
```

#### Multi-Node Programmatic Playbook

```go
func ClusterRollout() *Playbook {
    return &Playbook{
        Name: "cluster-rollout",
        Description: "Rollout across cluster with custom logic",
        Execute: func(node NodeInterface) types.Results {
            // This playbook is designed to run on inventory
            // It accesses the inventory through the node's context
            // For multi-node execution, use inv.RunPlaybook(playbook)
            return types.Results{Results: map[string]types.Result{
                node.GetHost(): types.Result{Message: "Use inventory.RunPlaybook() for multi-node"},
            }}
        },
    }
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
        
        node.RunSkill(skills.NewAptUpdate())
        result := node.RunSkill(skills.NewDeployApp())
        
        if result.Error != nil {
            log.Printf("Failed on %s, stopping rollout", node.GetHost())
            break
        }
        
        // Health check between nodes
        health := node.RunSkill(skills.NewHealthCheck())
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
            n.RunSkill(skills.NewAptUpdate())
            n.RunSkill(skills.NewDeployApp())
        }(node)
    }
    wg.Wait()
    
    return inv.RunSkill(skills.NewPing())
}
```

## Helper Utilities

### Sequential Execution Helper

```go
package playbook

// RunSequential runs skills in sequence
func RunSequential(node NodeInterface, skills []types.SkillInterface) types.Results {
    results := types.Results{
        Results: make(map[string]types.Result),
    }
    
    for _, skill := range skills {
        skill.SetNodeConfig(node.GetNodeConfig())
        result := skill.Run()
        results.Results[node.GetHost()] = result
        
        if result.Error != nil {
            return results  // Stop on first error
        }
    }
    
    return results
}
```

### Parallel Execution Helper

```go
// RunParallel runs skills concurrently
func RunParallel(node NodeInterface, skills []types.SkillInterface) types.Results {
    results := types.Results{
        Results: make(map[string]types.Result),
    }
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, skill := range skills {
        wg.Add(1)
        go func(s types.SkillInterface) {
            defer wg.Done()
            s.SetNodeConfig(node.GetNodeConfig())
            result := s.Run()
            
            mu.Lock()
            results.Results[node.GetHost()] = result
            mu.Unlock()
        }(skill)
    }
    wg.Wait()
    
    return results
}
```

### Retry Helper

```go
// RunWithRetry runs skill with exponential backoff retry
func RunWithRetry(node NodeInterface, skill types.SkillInterface, maxAttempts int, backoff time.Duration) types.Result {
    var lastError error
    
    for attempt := 0; attempt < maxAttempts; attempt++ {
        skill.SetNodeConfig(node.GetNodeConfig())
        result := skill.Run()
        
        if result.Error == nil {
            return result
        }
        
        lastError = result.Error
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
| Structure | Builder pattern | Implements PlaybookInterface → SkillInterface | Stages with parallel/sequential |
| Flexibility | Limited to builder methods | Unlimited (full Go) | Configurable |
| Decision Making | Simple conditions only | Full control flow | If/else conditions |
| Result Parsing | Manual in code | Manual in code | Manual in code |
| Dynamic Skills | Limited (loop with transform) | Unlimited | Loop over data/collections |
| Learning Curve | Simple builder API | Go knowledge required | Go knowledge required |
| Type Safety | Compile-time | Compile-time | Compile-time |
| Testing | Mock skills | Mock skills + unit tests | Mock skills + unit tests |
| Use Case | Simple sequential operations | Complex orchestration logic | Complex orchestration needs |
| API | Builder pattern (convenience) | Same RunSkill() as regular skills | New API needed |

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

### 1. Interface Hierarchy

The interface hierarchy would be:

```go
// SkillInterface - existing interface for simple atomic operations
type SkillInterface interface {
    GetID() string
    SetID(id string) SkillInterface
    GetDescription() string
    SetDescription(description string) SkillInterface
    GetNodeConfig() config.NodeConfig
    SetNodeConfig(cfg config.NodeConfig) SkillInterface
    GetArg(key string) string
    SetArg(key, value string) SkillInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) SkillInterface
    IsDryRun() bool
    SetDryRun(dryRun bool) SkillInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) SkillInterface
    Check() (bool, error)
    Run() Result
}

// PlaybookInterface - extends SkillInterface for complex orchestration
type PlaybookInterface interface {
    SkillInterface
    // Playbook-specific methods can be added here if needed
}
```

### 2. No New Execution Methods Needed

Since `PlaybookInterface` extends `SkillInterface`, playbooks can use the existing `node.RunSkill()` method. No new execution API required - the distinction is purely in the complexity of the `Run()` implementation.

### 3. Documentation

Provide examples and patterns for common orchestration patterns:
- Sequential execution in Run()
- Parallel execution with goroutines
- Retry logic
- Result parsing
- Dynamic skill creation
- Multi-node orchestration

### 4. Helper Package

Create `ork/playbook` package with helper utilities:
- `RunSequential()`
- `RunParallel()`
- `RunWithRetry()`
- `ParseResult()`
- Common patterns and examples

## Timeline

This requires no new API - playbooks can be implemented today with the existing skill interface. The helper package and documentation can be added as future enhancements to make the pattern more discoverable.
