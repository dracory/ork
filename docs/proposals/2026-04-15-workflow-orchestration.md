# Workflow Orchestration Layer

**Status:** Proposed
**Created:** 2026-04-15
**Author:** @dracory

## Problem Statement

Ork currently has atomic "skills" (formerly playbooks) that perform single operations like `apt-update`, `user-create`, `ping`. While these are useful, real-world operations require:

1. **Parallel execution** - multiple independent tasks simultaneously
2. **Conditional branching** - if A succeeds, do B, else do C
3. **Dependencies** - task B depends on task A's output
4. **Error handling** - retry, fallback, rollback on failure
5. **State management** - pass data between steps

Currently, this requires complex manual orchestration in user code.

## Proposal

Add a **workflow orchestration layer** on top of the existing skills system. This would provide:

- **Skills**: Atomic, idempotent operations (current implementation)
- **Workflows**: Go-based orchestration with parallel execution, branching, dependencies, and error handling

This mirrors real-world workflow engines (like GitHub Actions, GitLab CI, Airflow) where:
- Steps/Jobs = atomic operations
- Workflows = orchestration with parallel execution, dependencies, and branching

## Motivation

### Current Limitations

Users must manually orchestrate complex operations:

```go
// Manual orchestration - complex and error-prone
// Parallel execution requires goroutines and sync.WaitGroup
// Dependencies require manual checking of results
// Error handling is repetitive
// No way to pass state between steps
```

### Proposed Solution

Declarative workflow orchestration with parallel execution, dependencies, and branching:

```go
workflow := NewWorkflow("webserver-deploy").
    Parallel(
        NewStep("update-packages").AddSkill(skills.NewAptUpdate()),
        NewStep("upgrade-packages").AddSkill(skills.NewAptUpgrade()),
    ).
    Then(
        NewStep("create-user").AddSkill(skills.NewUserCreate()),
    ).
    Then(
        NewStep("harden-ssh").AddSkill(skills.NewSshHarden()),
    )

result := node.RunWorkflow(workflow)
```

## Architecture

### Core Types

```go
// Workflow orchestrates skills with parallel execution, dependencies, and branching
type Workflow struct {
    ID          string
    Description string
    Stages      []*WorkflowStage
}

// WorkflowStage contains steps that can run in parallel
type WorkflowStage struct {
    Steps       []*WorkflowStep
    Condition   func(WorkflowContext) bool
    OnFailure   WorkflowFailureAction
}

// WorkflowStep is a single skill execution
type WorkflowStep struct {
    ID          string
    Skills      []types.SkillInterface
    DependsOn   []string  // Step IDs this step depends on
    Condition   func(WorkflowContext) bool
    RetryPolicy RetryPolicy
}

// WorkflowContext provides state and results between steps
type WorkflowContext struct {
    Node       NodeInterface
    Results    map[string]StepResult
    State      map[string]interface{}  // User-defined state
    DryRun     bool
}

// WorkflowFailureAction defines error handling strategy
type WorkflowFailureAction string

const (
    StopOnFailure      WorkflowFailureAction = "stop"
    ContinueOnFailure  WorkflowFailureAction = "continue"
    RetryOnFailure     WorkflowFailureAction = "retry"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
    MaxAttempts int
    Backoff      time.Duration
}

// WorkflowResult contains execution results
type WorkflowResult struct {
    Stages     []StageResult
    Summary    WorkflowSummary
    Completed  bool
    Context    *WorkflowContext
}
```

### API Design

#### Builder Pattern

```go
func NewWorkflow(id string) *Workflow

// Parallel adds a stage with parallel step execution
func (w *Workflow) Parallel(steps ...*WorkflowStep) *Workflow

// Then adds a sequential stage that runs after previous stages complete
func (w *Workflow) Then(steps ...*WorkflowStep) *Workflow

// If adds a conditional branch
func (w *Workflow) If(condition func(WorkflowContext) bool) *Workflow

// Else adds the else branch
func (w *Workflow) Else(steps ...*WorkflowStep) *Workflow

// SetState allows setting initial workflow state
func (w *Workflow) SetState(key string, value interface{}) *Workflow

// SetFailureAction sets the default failure behavior
func (w *Workflow) SetFailureAction(action WorkflowFailureAction) *Workflow
```

#### Step Builder

```go
func NewStep(id string) *WorkflowStep

// AddSkill adds a skill to the step
func (s *WorkflowStep) AddSkill(skill types.SkillInterface) *WorkflowStep

// DependsOn adds dependency on another step
func (s *WorkflowStep) DependsOn(stepID string) *WorkflowStep

// WithRetry adds retry policy
func (s *WorkflowStep) WithRetry(maxAttempts int, backoff time.Duration) *WorkflowStep

// When adds conditional execution
func (s *WorkflowStep) When(condition func(WorkflowContext) bool) *WorkflowStep

// Loop adds loop over data collection
func (s *WorkflowStep) Loop(data []interface{}, eachFunc func(interface{}) *WorkflowStep) *WorkflowStep
```

#### Execution

```go
// RunnerInterface extension
func (n *nodeImplementation) RunWorkflow(workflow *Workflow) WorkflowResult

// Also add to Group and Inventory for multi-node execution
func (g *groupImplementation) RunWorkflow(workflow *Workflow) WorkflowResults
func (i *inventoryImplementation) RunWorkflow(workflow *Workflow) WorkflowResults
```

### Integration with Node, Group, and Inventory

#### Single Node Execution (Primary Use Case)

Workflows can run directly on a single node, just like skills:

```go
node := ork.NewNodeForHost("server.example.com")
workflow := NewWorkflow("deploy").
    Then(
        NewStep("update").AddSkill(skills.NewAptUpdate()),
        NewStep("upgrade").AddSkill(skills.NewAptUpgrade()),
    )

result := node.RunWorkflow(workflow)
// result.Context.Node == node
// result.Context.State is per-node state
```

This is the primary use case - workflows work exactly like skills but with orchestration capabilities.

#### Group Execution

Workflows run on all nodes in a group with configurable concurrency:

```go
group := ork.NewGroup("webservers")
group.AddNode(ork.NewNodeForHost("web1.example.com"))
group.AddNode(ork.NewNodeForHost("web2.example.com"))
group.AddNode(ork.NewNodeForHost("web3.example.com"))

workflow := NewWorkflow("deploy").
    Then(
        NewStep("update").AddSkill(skills.NewAptUpdate()),
        NewStep("upgrade").AddSkill(skills.NewAptUpgrade()),
    )

results := group.RunWorkflow(workflow)
// results.Results[host] = WorkflowResult for each node
// Each node gets its own WorkflowContext
// Concurrency controlled by group's max concurrency setting
```

#### Inventory Execution

Workflows run on all nodes across all groups:

```go
inv := ork.NewInventory()
inv.AddGroup(webGroup)
inv.AddGroup(dbGroup)

workflow := NewWorkflow("cluster-update").
    Then(
        NewStep("update").AddSkill(skills.NewAptUpdate()),
        NewStep("upgrade").AddSkill(skills.NewAptUpgrade()),
    )

results := inv.RunWorkflow(workflow)
// results.Results[host] = WorkflowResult for each node
// Concurrency controlled by inventory's max concurrency setting
summary := results.Summary()
```

#### Execution Semantics

**Per-Node Workflow Context:**
- Each node gets its own `WorkflowContext`
- `WorkflowContext.Node` points to the specific node
- `WorkflowContext.State` is isolated per-node
- `WorkflowContext.Results` contains that node's step results

**Multi-Node Concurrency:**
- Group/Inventory execute workflow on nodes in parallel
- Within each node, workflow stages execute as defined (parallel/sequential)
- Two levels of concurrency: node-level (across nodes) and step-level (within workflow)

**Result Aggregation:**
```go
type WorkflowResults struct {
    Results map[string]WorkflowResult  // Key: node hostname
    Summary WorkflowSummary
}

type WorkflowSummary struct {
    TotalNodes     int
    CompletedNodes int
    FailedNodes    int
    TotalSteps     int
    CompletedSteps int
    FailedSteps    int
}
```

**Failure Propagation:**
- Node-level failure doesn't stop other nodes (configurable)
- Step-level failure follows workflow's `OnFailure` policy per-node
- Group/Inventory can set overall failure policy

### Usage Examples

#### Parallel Execution

```go
workflow := NewWorkflow("parallel-tasks").
    Parallel(
        NewStep("ping").AddSkill(skills.NewPing()),
        NewStep("check-apt").AddSkill(skills.NewAptStatus()),
        NewStep("check-swap").AddSkill(skills.NewSwapStatus()),
    )

result := node.RunWorkflow(workflow)
```

#### Sequential with Dependencies

```go
workflow := NewWorkflow("deploy-app").
    Then(
        NewStep("update-packages").AddSkill(skills.NewAptUpdate()),
    ).
    Then(
        NewStep("upgrade-packages").
            AddSkill(skills.NewAptUpgrade()).
            DependsOn("update-packages"),  // Only runs if update succeeds
    ).
    Then(
        NewStep("restart-service").
            AddSkill(skills.NewReboot()).
            DependsOn("upgrade-packages"),
    )

result := node.RunWorkflow(workflow)
```

#### Conditional Branching

```go
workflow := NewWorkflow("smart-deploy").
    Then(
        NewStep("check-updates").
            AddSkill(skills.NewAptStatus()),
    ).
    If(func(ctx *WorkflowContext) bool {
        // Branch based on whether updates are available
        result := ctx.Results["check-updates"]
        return result.Changed
    }).
    Then(
        NewStep("upgrade").
            AddSkill(skills.NewAptUpgrade()),
    ).
    Else(
        NewStep("skip-message").
            AddSkill(skills.NewPing()),  // Just ping to verify connectivity
    )
```

#### State Management

```go
workflow := NewWorkflow("deploy-with-state").
    SetState("app-version", "2.0.0").
    Then(
        NewStep("backup").
            AddSkill(skills.NewCustomBackup()).
            When(func(ctx *WorkflowContext) bool {
                return ctx.State["app-version"] != "1.0.0"  // Backup if not initial deploy
            }),
    ).
    Then(
        NewStep("deploy").
            AddSkill(skills.NewDeployApp()),
    )
```

#### Error Handling with Retry

```go
workflow := NewWorkflow("resilient-deploy").
    Then(
        NewStep("download").
            AddSkill(skills.NewDownload()).
            WithRetry(3, time.Second*5),  // Retry 3 times with 5s backoff
    ).
    SetFailureAction(RetryOnFailure)
```

#### Complex Multi-Stage Workflow

```go
workflow := NewWorkflow("production-deploy").
    // Stage 1: Parallel checks
    Parallel(
        NewStep("health-check").AddSkill(skills.NewPing()),
        NewStep("disk-space").AddSkill(skills.NewDiskCheck()),
    ).
    // Stage 2: Sequential deployment
    Then(
        NewStep("update").AddSkill(skills.NewAptUpdate()).
        NewStep("upgrade").AddSkill(skills.NewAptUpgrade()).
    ).
    // Stage 3: Conditional hardening
    If(func(ctx *WorkflowContext) bool {
        return ctx.State["environment"] == "production"
    }).
    Then(
        NewStep("harden").AddSkill(skills.NewSshHarden()),
    )
```

#### Multi-Node Workflow Execution

```go
workflow := NewWorkflow("cluster-rollout").
    // Stage 1: Parallel checks across each node
    Parallel(
        NewStep("ping").AddSkill(skills.NewPing()),
        NewStep("disk").AddSkill(skills.NewDiskCheck()),
    ).
    // Stage 2: Sequential update per node
    Then(
        NewStep("update").AddSkill(skills.NewAptUpdate()),
        NewStep("upgrade").AddSkill(skills.NewAptUpgrade()),
    )

// Run across all nodes in inventory
// - Nodes execute in parallel (node-level concurrency)
// - Within each node, stages execute sequentially (workflow-level concurrency)
inv.SetMaxConcurrency(5)  // Max 5 nodes in parallel
results := inv.RunWorkflow(workflow)
summary := results.Summary()
```

#### Loop Over Data Collection

```go
workflow := NewWorkflow("database-backup").
    Then(
        NewStep("list-dbs").
            AddSkill(skills.NewMariadbListDBs()).
            Loop(
                []string{"db1", "db2", "db3"},
                func(dbName interface{}) *WorkflowStep {
                    return NewStep(fmt.Sprintf("backup-%s", dbName)).
                        AddSkill(skills.NewMariadbBackup().
                            SetArg("database", dbName.(string)))
                },
            ),
    )
```

## Implementation Considerations

### 1. Backward Compatibility

- Keep all existing skill APIs unchanged
- Workflow layer is additive, not breaking
- Users can still use skills directly if they prefer

### 2. Concurrency Model

Parallel execution within stages using goroutines and sync.WaitGroup:

```go
func (w *Workflow) executeStage(stage *WorkflowStage, ctx *WorkflowContext) {
    var wg sync.WaitGroup
    for _, step := range stage.Steps {
        wg.Add(1)
        go func(s *WorkflowStep) {
            defer wg.Done()
            w.executeStep(s, ctx)
        }(step)
    }
    wg.Wait()
}
```

### 3. Dependency Resolution

Steps with `DependsOn` should wait for dependency completion:

```go
func (w *Workflow) executeStep(step *WorkflowStep, ctx *WorkflowContext) {
    // Wait for dependencies
    for _, depID := range step.DependsOn {
        depResult := ctx.Results[depID]
        if depResult.Error != nil {
            return // Skip if dependency failed
        }
    }
    // Execute step
    // ...
}
```

### 4. Dry-Run Support

Workflow should respect the node's dry-run mode and propagate it to all skills:

```go
func (n *nodeImplementation) RunWorkflow(workflow *Workflow) WorkflowResult {
    ctx := &WorkflowContext{
        Node:   n,
        DryRun: n.GetDryRunMode(),
    }
    // Execute workflow with context
}
```

### 5. State Management

State map should be safe for concurrent access:

```go
type WorkflowContext struct {
    State     map[string]interface{}
    stateMu   sync.RWMutex
}

func (c *WorkflowContext) SetState(key string, value interface{}) {
    c.stateMu.Lock()
    defer c.stateMu.Unlock()
    c.State[key] = value
}
```

### 6. Retry Logic

Exponential backoff for retries:

```go
func (s *WorkflowStep) executeWithRetry(skill types.SkillInterface, ctx *WorkflowContext) types.Result {
    for attempt := 0; attempt < s.RetryPolicy.MaxAttempts; attempt++ {
        result := skill.Run()
        if result.Error == nil {
            return result
        }
        if attempt < s.RetryPolicy.MaxAttempts-1 {
            time.Sleep(s.RetryPolicy.Backoff * time.Duration(attempt+1))
        }
    }
    return result
}
```

## Benefits

1. **Parallel Execution**: Run independent tasks simultaneously for faster execution
2. **Declarative Orchestration**: Clear, readable intent with builder pattern
3. **Dependency Management**: Automatic dependency resolution and ordering
4. **Conditional Branching**: If/else logic based on execution results
5. **State Management**: Pass data between steps for complex workflows
6. **Error Handling**: Built-in retry, continue, and stop strategies
7. **Reusability**: Workflow functions for common deployment patterns
8. **Type Safety**: Compiler validates workflow composition
9. **Testing**: Easy to mock skills for workflow tests
10. **Documentation**: Workflow names and structure describe intent clearly

## Open Questions

1. Should workflows have their own `Check()` method to preview execution?
2. Should workflows support rollback/cleanup on failure?
3. Should we add "Roles" as reusable workflow templates?
4. Should workflow results include execution timing and metrics?
5. Should we support workflow composition (workflows that call other workflows)?
6. Should we add workflow persistence (save/restore workflow state)?

## Timeline

This is a **future enhancement** and should be implemented as a separate feature after the current skills refactoring is complete and stable. The workflow orchestration layer is more complex than simple sequential execution and requires careful design and testing.
