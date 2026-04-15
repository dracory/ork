# Playbooks

Playbooks are complex orchestration logic that run multiple skills with full Go programming capabilities. They implement `RunnableInterface` and can use all existing execution APIs.

## Overview

Playbooks are not simple atomic operations like regular skills - they are elaborate orchestration logic that may include:
- Complex decision making based on runtime conditions
- Dynamic skill creation and execution
- Custom error handling and retry logic
- Data parsing and transformation between steps
- Multi-node orchestration with custom logic

Since playbooks implement `RunnableInterface`, they can use all existing runnable methods and execution APIs. The distinction is purely in the complexity of the `Run()` implementation.

## Creating a Playbook

A playbook implements `RunnableInterface` and orchestrates other skills in its `Run()` method. Playbooks can embed `types.BasePlaybook` which provides a foundation for playbook development:

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
    
    return types.Result{Changed: true, Message: "Complete"}
}

// Usage
node.Run(NewDeployWebserverPlaybook())
```

## Usage Patterns

### Dynamic Decision Making

```go
type SmartUpdatePlaybook struct {
    *types.BasePlaybook
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

### Loop and Dynamic Skill Creation

```go
type SetupUsersPlaybook struct {
    *types.BasePlaybook
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
```

### Parse Results and Spin Up Dynamic Skills

```go
type BackupAllDatabasesPlaybook struct {
    *types.BasePlaybook
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
```

### Custom Error Handling and Retry

```go
type ResilientDownloadPlaybook struct {
    *types.BasePlaybook
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
```

## Execution API

Playbooks use the same execution API as skills through `RunnerInterface.Run()`:

- `node.Run(runnable)` - Run a skill or playbook on a single node
- `group.Run(runnable)` - Run a skill or playbook on all nodes in a group
- `inventory.Run(runnable)` - Run a skill or playbook on all nodes in inventory

The `Run()` method accepts any `RunnableInterface`, which includes both simple skills and complex playbooks.

## Benefits

1. **Full Go Power**: Complete language capabilities (loops, conditionals, error handling, etc.)
2. **Dynamic Decision Making**: Make choices based on runtime conditions
3. **Complex Parsing**: Parse results and transform data between steps
4. **Dynamic Skill Creation**: Spin up skills based on previous results
5. **Custom Logic**: Implement any orchestration pattern imaginable
6. **Type Safety**: Compiler validates all code
7. **Testing**: Full Go testing framework support
8. **Flexibility**: No constraints on what you can do

## When to Use Playbooks

**Use Playbooks for:**
- Complex decision logic
- Dynamic skill creation based on results
- Custom error handling and retry logic
- Parsing and transforming data between steps
- Multi-node orchestration with custom logic
- 20% of complex use cases

**Use Simple Skills for:**
- Simple atomic operations
- Common patterns that repeat across projects
- 80% of common use cases

## Implementation

### BasePlaybook

Create `types.BasePlaybook` as a foundation for playbook development:

```go
type BasePlaybook struct {
    id          string
    description string
    nodeCfg     types.NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}

func NewBasePlaybook() *BasePlaybook {
    return &BasePlaybook{
        args:   make(map[string]string),
        dryRun: false,
    }
}

// Run must be overridden by playbook implementations
func (b *BasePlaybook) Run() Result {
    return Result{Error: fmt.Errorf("Run() must be implemented by playbook")}
}

// Check returns false (no changes) by default. Can be overridden
func (b *BasePlaybook) Check() (bool, error) {
    return false, nil
}
```

### Interface Hierarchy

Both simple skills and complex playbooks implement `RunnableInterface`:

```go
type RunnableInterface interface {
    GetID() string
    SetID(id string) RunnableInterface
    GetDescription() string
    SetDescription(description string) RunnableInterface
    GetNodeConfig() types.NodeConfig
    SetNodeConfig(cfg types.NodeConfig) RunnableInterface
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

The distinction between skills and playbooks is purely in the complexity of the `Run()` implementation.
