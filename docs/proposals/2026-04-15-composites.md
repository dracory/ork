# Composites

**Status:** Proposed
**Created:** 2026-04-15
**Author:** @dracory

## Problem Statement

Ork currently has atomic "skills" that perform single operations like `apt-update`, `user-create`, `ping`. While these are useful, users often need to run multiple related operations in sequence (e.g., update packages → upgrade → reboot → verify). Currently, this requires manual orchestration in user code.

## Proposal

Add a **builder pattern** for creating composites. A composite is a skill that orchestrates other skills in sequence using a clean builder API.

This is not a new concept - it's a builder pattern convenience for creating skills that orchestrate other skills sequentially. The builder pattern provides:
- Clean, readable API for defining sequential skill execution
- Built-in error handling (stop/continue on error)
- Reusability through builder functions
- No parallel execution, no complex branching, no dependencies

This is simpler than full workflow orchestration - just a clean way to create skills that run other skills in sequence.

## Motivation

### Current Limitations

Users must manually sequence skills:

```go
// Manual orchestration - repetitive
result1 := node.RunSkill(skills.NewAptUpdate())
if result1.Error != nil {
    return result1
}
result2 := node.RunSkill(skills.NewAptUpgrade())
if result2.Error != nil {
    return result2
}
result3 := node.RunSkill(skills.NewReboot())
if result3.Error != nil {
    return result3
}
result4 := node.RunSkill(skills.NewPing())
if result4.Error != nil {
    return result4
}
```

### Proposed Solution

Composite orchestration:

```go
composite := NewComposite("package-update").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade()).
    AddSkill(skills.NewReboot()).
    AddSkill(skills.NewPing())

result := node.RunComposite(composite)
```

## Architecture

### Core Types

```go
// Composite is a skill that orchestrates other skills in sequence
type Composite struct {
    ID          string
    Description string
    Skills      []SkillWithCondition
    StopOnError bool  // Stop on first error (default: true)
}

// SkillWithCondition wraps a skill with optional condition
type SkillWithCondition struct {
    Skill     types.SkillInterface
    Condition func(types.Result) bool  // Run skill only if condition returns true
    Loop      *LoopConfig              // Optional loop configuration
}

// LoopConfig defines simple loop behavior
type LoopConfig struct {
    Count     int                       // Number of iterations
    EachFunc  func(int, types.SkillInterface) types.SkillInterface  // Transform skill per iteration
}

// CompositeResult contains results from all skill executions
type CompositeResult struct {
    Steps     []StepResult
    Summary   CompositeSummary
    Completed bool
}

type StepResult struct {
    StepIndex int
    SkillID   string
    Result    types.Result
}

type CompositeSummary struct {
    TotalSteps     int
    CompletedSteps int
    FailedSteps    int
}
```

### API Design

#### Builder Pattern

```go
func NewComposite(id string) *Composite

// AddSkill appends a skill to the composite
func (c *Composite) AddSkill(skill types.SkillInterface) *Composite

// AddConditionalSkill appends a skill that runs only if condition is true
func (c *Composite) AddConditionalSkill(
    skill types.SkillInterface,
    condition func(types.Result) bool,
) *Composite

// AddLoopSkill appends a skill that runs N times with optional transformation
func (c *Composite) AddLoopSkill(
    skill types.SkillInterface,
    count int,
    eachFunc func(int, types.SkillInterface) types.SkillInterface,
) *Composite

// SetDescription sets the composite description
func (c *Composite) SetDescription(desc string) *Composite

// StopOnError sets whether to stop on first error (default: true)
func (c *Composite) StopOnError(stop bool) *Composite
```

#### Execution

```go
// RunnerInterface extension
func (n *nodeImplementation) RunComposite(composite *Composite) CompositeResult

// Also add to Group and Inventory for multi-node execution
func (g *groupImplementation) RunComposite(composite *Composite) CompositeResults
func (i *inventoryImplementation) RunComposite(composite *Composite) CompositeResults
```

### Integration with Node, Group, and Inventory

#### Single Node Execution (Primary Use Case)

```go
node := ork.NewNodeForHost("server.example.com")
composite := NewComposite("deploy").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade())

result := node.RunComposite(composite)
```

#### Group Execution

```go
group := ork.NewGroup("webservers")
group.AddNode(ork.NewNodeForHost("web1.example.com"))
group.AddNode(ork.NewNodeForHost("web2.example.com"))

composite := NewComposite("deploy").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade())

results := group.RunComposite(composite)
// results.Results[host] = CompositeResult for each node
```

#### Inventory Execution

```go
inv := ork.NewInventory()
inv.AddGroup(webGroup)

composite := NewComposite("cluster-update").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade())

results := inv.RunComposite(composite)
summary := results.Summary()
```

### Usage Examples

#### Basic Sequential Execution

```go
composite := NewComposite("package-update").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade())

result := node.RunComposite(composite)
```

#### Continue on Error

```go
composite := NewComposite("best-effort-update").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade()).
    StopOnError(false)  // Continue even if apt-update fails

result := node.RunComposite(composite)
```

#### Conditional Execution

```go
composite := NewComposite("smart-upgrade").
    AddSkill(skills.NewAptStatus()).
    AddConditionalSkill(
        skills.NewAptUpgrade(),
        func(result types.Result) bool {
            // Only upgrade if updates are available
            return result.Changed
        },
    )
```

#### Loop Execution

```go
// Repeat a skill N times
composite := NewComposite("create-users").
    AddLoopSkill(
        skills.NewUserCreate(),
        3,  // Create 3 users
        func(i int, skill types.SkillInterface) types.SkillInterface {
            // Transform skill for each iteration
            skill.SetArg("username", fmt.Sprintf("user%d", i))
            return skill
        },
    )
```

#### Complex Sequential Orchestration

```go
composite := NewComposite("secure-webserver").
    AddSkill(skills.NewAptUpdate()).
    AddSkill(skills.NewAptUpgrade()).
    AddSkill(skills.NewUserCreate().
        SetArg("username", "web").
        SetArg("shell", "/bin/bash")).
    AddSkill(skills.NewSshHarden()).
    AddSkill(skills.NewFail2banInstall()).
    AddSkill(skills.NewUfwInstall()).
    AddSkill(skills.NewUfwAllowMariaDB())

result := node.RunComposite(composite)
```

#### Reusable Composite Functions

```go
// Create reusable composite as a function
func SecureWebserverComposite() *Composite {
    return NewComposite("secure-webserver").
        AddSkill(skills.NewAptUpdate()).
        AddSkill(skills.NewAptUpgrade()).
        AddSkill(skills.NewSshHarden()).
        AddSkill(skills.NewFail2banInstall())
}

// Use it
composite := SecureWebserverComposite()
node.RunComposite(composite)
```

## Implementation Considerations

### 1. Backward Compatibility

- Keep all existing skill APIs unchanged
- Composite builder is additive, not breaking
- Users can still use skills directly if they prefer

### 2. Dry-Run Support

Composite should respect the node's dry-run mode and propagate it to all skills:

```go
func (n *nodeImplementation) RunComposite(composite *Composite) CompositeResult {
    result := CompositeResult{
        Steps: make([]StepResult, 0, len(composite.Skills)),
    }

    for i, skillWithCondition := range composite.Skills {
        // Check condition if present
        if skillWithCondition.Condition != nil {
            // Use result from previous step for condition check
            prevResult := types.Result{Changed: false}
            if i > 0 {
                prevResult = result.Steps[i-1].Result
            }
            if !skillWithCondition.Condition(prevResult) {
                // Skip this skill
                result.Steps = append(result.Steps, StepResult{
                    StepIndex: i,
                    SkillID:   skillWithCondition.Skill.GetID(),
                    Result:    types.Result{Changed: false, Message: "Skipped (condition false)"},
                })
                continue
            }
        }

        skill := skillWithCondition.Skill
        skill.SetNodeConfig(n.cfg)
        skill.SetDryRun(n.cfg.IsDryRunMode)
        stepResult := skill.Run()
        
        result.Steps = append(result.Steps, StepResult{
            StepIndex: i,
            SkillID:   skill.GetID(),
            Result:    stepResult,
        })

        if stepResult.Error != nil && composite.StopOnError {
            break
        }
    }
    return result
}
```

### 3. Result Aggregation

For Group/Inventory execution, aggregate results per-node:

```go
type CompositeResults struct {
    Results map[string]CompositeResult  // Key: node hostname
    Summary CompositeSummary
}

type CompositeSummary struct {
    TotalNodes     int
    CompletedNodes int
    FailedNodes    int
    TotalSteps     int
    CompletedSteps int
    FailedSteps    int
}
```

## Benefits

1. **Declarative Orchestration**: Clear, readable intent
2. **Reduced Boilerplate**: No manual sequencing code
3. **Consistent Error Handling**: Built-in stop/continue on error
4. **Reusability**: Composite functions for common patterns
5. **Type Safety**: Compiler validates skill composition
6. **Testing**: Easy to mock skills for composite tests
7. **Documentation**: Composite names describe intent clearly
8. **Simplicity**: Easy to understand and use - just sequential execution

## Comparison with Ansible, Ork Playbooks, and Ork Workflow

| Feature | Ansible Playbook | Ork Composites | Ork Playbooks | Ork Workflow |
|---------|-----------------|----------------------|-----------------|--------------|
| Language | YAML | Go (builder) | Go (full) | Go |
| Paradigm | Declarative | Declarative | Imperative | Imperative |
| Execution | Sequential + Parallel | Sequential only | Sequential + Parallel | Sequential + Parallel |
| Dependencies | Implicit (ordering) | Implicit (ordering) | None | Explicit dependencies |
| Branching | Conditionals, loops | Simple conditions | Full control flow | If/else conditions |
| Loops | Loop, with_items | Simple repeat N times | Unlimited | Loop over data/collections with transformation |
| State Management | Variables/facts | None | Manual | Pass data between steps |
| Retry | Built-in | None | Custom | Configurable retry policy |
| Type Safety | Runtime validation | Compile-time validation | Compile-time validation | Compile-time validation |
| Complexity | Full-featured | Simple | Unlimited | Complex |
| Learning Curve | YAML + Ansible DSL | Go (simple) | Go | Go |
| Use Case | Full automation | Common sequential operations | Custom orchestration logic | Complex orchestration needs |
| API | Ansible DSL | Builder pattern (convenience) | Same RunSkill() as regular skills | New API needed |

**Progression:**
- **Skills**: Atomic operations (like Ansible modules) - implements SkillInterface
- **Ork Composites**: Builder pattern for sequential skill execution (convenience)
- **Ork Playbooks**: Full Go orchestration power (implements PlaybookInterface → SkillInterface)
- **Ork Workflows**: Complex orchestration with parallel/dependencies/state
- **Ansible Playbooks**: Full-featured declarative automation

Composites are a builder pattern convenience for the common case (run these skills in order). Playbooks give full Go power for custom logic using the same `RunSkill()` method as regular skills (PlaybookInterface extends SkillInterface). Workflows solve the complex orchestration case (parallel execution, branching, etc.).

## Open Questions

1. Should composites have their own `Check()` method to preview execution?
2. Should we add a registry for reusable composites?

## Timeline

This is a **builder pattern convenience** - users can implement sequential skills today with the existing skill interface. The builder pattern can be added as a helper package to make the pattern more discoverable and easier to use.
