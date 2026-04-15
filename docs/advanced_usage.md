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

Register your custom skills to use them via `node.RunSkillByID("custom-id")`:

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
if err := registry.SkillRegister(customSkill); err != nil {
    log.Fatalf("Failed to register skill: %v", err)
}

// Now use it like any built-in skill
node := ork.NewNodeForHost("server.example.com")
result := node.RunSkillByID("install-docker")
```

### Custom Skills with Full Idempotency

For full idempotency support, implement all methods:

```go
type MyCustomSkill struct{}

func (s *MyCustomSkill) GetID() string { return "my-task" }
func (s *MyCustomSkill) Description() string { return "Does something" }

// Check() - returns true if changes needed
func (s *MyCustomSkill) Check(cfg config.NodeConfig) (bool, error) {
    // Check if already configured
    output, _ := ssh.Run(cfg, types.Command{Command: "cat /etc/my-config"})
    return !strings.Contains(output, "configured"), nil
}

// Run() - execute and return Result
func (s *MyCustomSkill) Run(cfg config.NodeConfig) skill.Result {
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

## Internal Packages

For advanced use cases or when you need fine-grained control, you can use the internal packages directly.

### Package Overview

- `ork` - Main API: `NodeInterface`, `InventoryInterface`, `GroupInterface`, `RunnerInterface`
- `types` - Shared types: `Result`, `Results`, `Summary`
- `runnable` - `RunnerInterface` for Node, Group, and Inventory
- `config` - Configuration types
- `ssh` - SSH client with connection management
- `skill` - Skill interface and registry
- `skills` - Built-in skill implementations

### Using Internal Packages Directly

```go
package main

import (
    "log"

    "github.com/dracory/ork/config"
    "github.com/dracory/ork/skills"
)

func main() {
    cfg := config.NodeConfig{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := skills.NewPing()
    ping.SetConfig(cfg)
    result := ping.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Update packages
    aptUpdate := skills.NewAptUpdate()
    aptUpdate.SetConfig(cfg)
    result = aptUpdate.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := skills.NewSwapCreate()
    swapCreate.SetConfig(cfg)
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
for i := 0; i < maxRetries; i++ {
    results := node.RunSkill(skills.NewAptUpdate())
    result := results.Results["server.example.com"]
    
    if result.Error == nil {
        break
    }
    
    if i == maxRetries-1 {
        log.Fatalf("Failed after %d retries: %v", maxRetries, result.Error)
    }
    
    log.Printf("Attempt %d failed, retrying...", i+1)
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
        results := node.RunSkill(skills.NewPing())
        // Process results...
    }(host)
}
wg.Wait()
```
