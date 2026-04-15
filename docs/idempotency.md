# Idempotency

All skills support idempotent execution. Use `Check()` to preview changes before running them.

## Check Before Run

Use `Check()` to determine if changes would be made:

```go
// Check if changes would be made
results := node.Check(skills.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would upgrade packages: %s", result.Message)
    // Now actually run it
    results = node.Run(skills.NewAptUpgrade())
}
```

## Result Structure

Results are returned as `types.Results` with per-node access:

```go
type Results struct {
    Results map[string]Result  // Key is node hostname
}

func (r Results) Summary() Summary

type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}

type Summary struct {
    Total     int
    Changed   int
    Unchanged int
    Failed    int
}
```

## Using Results Summary

When working with multiple nodes, use the summary to get an overview:

```go
results := inv.Run(skills.NewAptUpdate())
summary := results.Summary()

fmt.Printf("Total: %d, Changed: %d, Unchanged: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Unchanged, summary.Failed)
```

## Direct Skill Access (Advanced)

For programmatic skill handling, use the `skill` package directly:

```go
import (
    "github.com/dracory/ork/skill"
    "github.com/dracory/ork/skills"
)

// Execute directly with config
aptUpgrade := skills.NewAptUpgrade()
aptUpgrade.SetNodeConfig(cfg)
result := aptUpgrade.Run()

// Or check before running via Check
needsChange, _ := skills.NewSwapCreate().SetNodeConfig(cfg).Check()
if !needsChange {
    log.Println("Swap already exists, skipping...")
    return
}
```

## Implementing Idempotent Skills

When creating custom skills, implement both `Check()` and `Run()` methods for full idempotency:

```go
type MyCustomSkill struct{}

func (s *MyCustomSkill) GetID() string { return "my-task" }
func (s *MyCustomSkill) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (s *MyCustomSkill) Check(cfg types.NodeConfig) (bool, error) {
    // Check if already configured
    output, _ := ssh.Run(cfg, types.Command{Command: "cat /etc/my-config"})
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (s *MyCustomSkill) Run(cfg types.NodeConfig) skill.Result {
    needsChange, _ := s.Check(cfg)
    if !needsChange {
        return skill.Result{
            Changed: false,
            Message: "Already configured",
        }
    }

    // Apply changes...
    _, err := ssh.Run(cfg, types.Command{Command: "setup-command"})
    if err != nil {
        return skill.Result{Changed: false, Error: err}
    }

    return skill.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}
```
