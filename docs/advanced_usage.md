# Advanced Usage

## Inspecting Configuration

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy")

fmt.Printf("Host: %s\n", node.GetHost())
fmt.Printf("Port: %s\n", node.GetPort())
fmt.Printf("User: %s\n", node.GetUser())

// Get full config for integration with internal packages
cfg := node.GetNodeConfig()
```

## Custom Skills

Extend Ork with custom automation tasks by implementing the `Skill` interface.

### Registering Custom Skills

Register your custom skills to use them via `node.RunByID("custom-id")`:

```go
import (
    "github.com/dracory/ork"
    "github.com/dracory/ork/skill"
)

// Create a custom skill
customSkill := skill.NewBaseSkill()
customSkill.SetID("install-docker")
customSkill.SetDescription("Install Docker on the server")

// Register it globally
registry, err := ork.GetGlobalSkillRegistry()
if err != nil {
    log.Fatalf("Failed to get registry: %v", err)
}
if err := registry.Register(customSkill); err != nil {
    log.Fatalf("Failed to register skill: %v", err)
}

// Now use it like any built-in skill
node := ork.NewNodeForHost("server.example.com")
result := node.RunByID("install-docker")
```

**Note**: For complex orchestration logic with decision making, loops, and custom error handling, see [Playbooks](playbooks.md).

## Privilege Escalation in Custom Skills

When creating custom skills that require elevated privileges, you can configure the become user at multiple levels:

### Skill-Level Configuration

```go
type MyCustomSkill struct {
    types.BaseSkill
}

func (s *MyCustomSkill) Run() types.Result {
    // The skill inherits SetBecomeUser and GetBecomeUser from BaseBecome
    // Set it before running:
    s.SetBecomeUser("root")

    // Commands will run as root
    output, err := ssh.Run(s.GetNodeConfig(), types.Command{
        Command: "apt-get update",
    })
    // ...
}
```

### Dynamic User Selection

```go
func (s *MyCustomSkill) Run() types.Result {
    cfg := s.GetNodeConfig()

    // Choose become user based on node configuration
    becomeUser := "root"
    if cfg.GetArg("environment") == "production" {
        becomeUser = "admin"
    }

    s.SetBecomeUser(becomeUser)
    // ...
}
```

### Combining with Check/Run Pattern

```go
func (s *MyCustomSkill) Check() (bool, error) {
    // Run check as the become user
    s.SetBecomeUser("postgres")
    output, err := ssh.Run(s.GetNodeConfig(), types.Command{
        Command: "psql -c 'SELECT 1'",
    })
    return err == nil, nil
}

func (s *MyCustomSkill) Run() types.Result {
    s.SetBecomeUser("postgres")
    // Execute as postgres user
    output, err := ssh.Run(s.GetNodeConfig(), types.Command{
        Command: "psql -f /path/to/migration.sql",
    })
    // ...
}
```

### Custom Playbooks with Privilege Escalation

```go
type MyCustomPlaybook struct {
    types.BasePlaybook
}

func (p *MyCustomPlaybook) Run() types.Result {
    // Set become user for this playbook
    p.SetBecomeUser("root")

    cfg := p.GetNodeConfig()
    // All commands in this playbook will run as root
    _, err := ssh.Run(cfg, types.Command{
        Command: "systemctl restart nginx",
    })
    // ...
}
```

### Custom Playbooks with Full Idempotency

For comprehensive documentation on creating playbooks with complex orchestration logic, see [Playbooks Documentation](playbooks.md).

For simple playbook implementation with full idempotency support:

```go
type MyCustomPlaybook struct{}

func (p *MyCustomPlaybook) GetID() string { return "my-task" }
func (p *MyCustomPlaybook) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (p *MyCustomPlaybook) Check(cfg types.NodeConfig) (bool, error) {
    // Check if already configured
    output, _ := ssh.Run(cfg, types.Command{Command: "cat /etc/my-config"})
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (p *MyCustomPlaybook) Run(cfg types.NodeConfig) skill.Result {
    needsChange, _ := p.Check(cfg)
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

## Internal Packages

For advanced use cases or when you need fine-grained control, you can use the internal packages directly.

### Package Overview

- `ork` - Main API: `NodeInterface`, `InventoryInterface`, `GroupInterface`, `RunnerInterface`
- `types` - Shared types: `Result`, `Results`, `Summary`, `NodeConfig`
- `runnable` - `RunnerInterface` for Node, Group, and Inventory
- `ssh` - SSH client with connection management
- `skill` - Skill interface and registry
- `skills` - Built-in skill implementations

### Using Internal Packages Directly

```go
package main

import (
    "log"

    "github.com/dracory/ork/skills"
    "github.com/dracory/ork/types"
)

func main() {
    cfg := types.NodeConfig{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := skills.NewPing()
    ping.SetNodeConfig(cfg)
    result := ping.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Update packages
    aptUpdate := skills.NewAptUpdate()
    aptUpdate.SetNodeConfig(cfg)
    result = aptUpdate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := skills.NewSwapCreate()
    swapCreate.SetNodeConfig(cfg)
    result = swapCreate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }
}
```

## Advanced Configuration Patterns

### Conditional Execution Based on Node Properties

```go
node := ork.NewNodeForHost("server.example.com")
node.SetArg("environment", "production")

// In your skill, check the environment
if node.GetArg("environment") == "production" {
    // Apply production-specific configuration
}
```

### Error Handling and Retry Logic

```go
maxRetries := 3
for attempt := 0; attempt < maxRetries; attempt++ {
    results := node.Run(skills.NewAptUpdate())
    result := results.Results["server.example.com"]
    
    if result.Error == nil {
        break
    }
    
    if attempt == maxRetries-1 {
        log.Fatalf("Failed after %d retries: %v", maxRetries, result.Error)
    }
    
    log.Printf("Attempt %d failed, retrying...", attempt+1)
    time.Sleep(time.Second * 5)
}
```

### Concurrent Operations with Rate Limiting

```go
// Create a semaphore to limit concurrent operations
sem := make(chan struct{}, 5) // Max 5 concurrent operations

var wg sync.WaitGroup
for _, host := range hosts {
    wg.Add(1)
    go func(h string) {
        defer wg.Done()
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        
        node := ork.NewNodeForHost(h)
        results := node.Run(skills.NewPing())
        // Process results...
    }(host)
}
wg.Wait()
```
