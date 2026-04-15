---
path: modules/types.md
page-type: module
summary: Shared types including RunnableInterface, BasePlaybook, BaseSkill, Registry, NodeConfig, Command, PromptConfig, PromptResult, and result types for operation outcomes across all Ork packages.
tags: [module, types, results, prompts]
created: 2025-04-14
updated: 2026-04-15
version: 2.0.0
---

# types Package

## Changelog
- **v2.0.0** (2026-04-15): Major terminology refactoring - PlaybookInterface renamed to RunnableInterface, PlaybookOptions renamed to RunnableOptions, BasePlaybook and BaseSkill moved to types package, NodeConfig moved from config to types package
- **v1.2.0** (2026-04-14): Added PromptConfig and PromptResult types for interactive user input
- **v1.1.0** (2026-04-14): Updated PlaybookInterface and Registry documentation
- **v1.0.0** (2025-04-14): Initial creation

Shared types for skills, registries, commands, configuration, and operation results across all Ork packages.

## Purpose

The `types` package provides:
- `RunnableInterface`: The interface all skills must implement
- `BasePlaybook`: Default implementation with fluent API and optional Check()
- `BaseSkill`: Default implementation with Check() and Run() stubs
- `NodeConfig`: Central configuration structure for remote operations
- `Registry`: For registering and looking up skills by ID
- `Command`: Struct for shell commands with descriptions
- `RunnableOptions`: Configuration options for skill execution
- `PromptConfig`, `PromptResult`: Types for interactive user input
- `Result`, `Results`, `Summary`: Operation outcome types

## Key Files

| File | Purpose |
|------|---------|
| `runner_interface.go` | RunnerInterface - base for all executables |
| `base_playbook.go` | BasePlaybook default implementation |
| `base_skill.go` | BaseSkill default implementation |
| `node_config.go` | NodeConfig struct and methods |
| `registry.go` | RunnableInterface, RunnableOptions, Registry |
| `command.go` | Command struct with description |
| `prompt.go` | PromptConfig, PromptResult types |
| `results.go` | Result, Results, and Summary types |

## RunnableInterface

All automation skills must implement this interface.

```go
type RunnableInterface interface {
    // Identification
    GetID() string
    SetID(id string) RunnableInterface
    GetDescription() string
    SetDescription(description string) RunnableInterface
    
    // Configuration
    GetNodeConfig() NodeConfig
    SetNodeConfig(cfg NodeConfig) RunnableInterface
    
    // Arguments
    GetArg(key string) string
    SetArg(key, value string) RunnableInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) RunnableInterface
    
    // Execution options
    IsDryRun() bool
    SetDryRun(dryRun bool) RunnableInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) RunnableInterface
    
    // Core operations
    Check() (bool, error)
    Run() Result
}
```

### Check

Determines if the skill needs to make changes.

```go
func (p RunnableInterface) Check() (bool, error)
```

- Returns `true` if changes are needed
- Returns `false` if system is already in desired state
- Returns error if the check itself fails

### Run

Executes the skill and returns the result.

```go
func (p RunnableInterface) Run() Result
```

The `Result.Changed` field indicates whether any modifications were made.

## BasePlaybook

Provides a default implementation of RunnableInterface with fluent API. Use this when you want optional Check() with a default implementation.

```go
type BasePlaybook struct {
    id          string
    description string
    nodeCfg     NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}
```

### Constructor

```go
func NewBasePlaybook() *BasePlaybook
```

### Features

- Fluent API for method chaining
- Optional Check() (returns false by default)
- Must implement Run() yourself
- Useful for simple skills where Check() is not critical

### Example

```go
type MySkill struct {
    *types.BasePlaybook
}

func NewMySkill() types.RunnableInterface {
    return &MySkill{
        BasePlaybook: types.NewBasePlaybook().
            SetID("my-skill").
            SetDescription("Does something useful"),
    }
}

func (m *MySkill) Run() types.Result {
    // Your implementation here
    return types.Result{Changed: true, Message: "Done"}
}
```

## BaseSkill

Provides a default implementation of RunnableInterface with Check() and Run() stubs that must be implemented. Use this when you want to enforce implementation of both Check() and Run().

```go
type BaseSkill struct {
    id          string
    description string
    nodeCfg     NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}
```

### Constructor

```go
func NewBaseSkill() *BaseSkill
```

### Features

- Fluent API for method chaining
- Check() stub that returns error (must be implemented)
- Run() stub that returns error (must be implemented)
- Enforces idempotency pattern

### Example

```go
type MySkill struct {
    *types.BaseSkill
}

func NewMySkill() types.RunnableInterface {
    return &MySkill{
        BaseSkill: types.NewBaseSkill().
            SetID("my-skill").
            SetDescription("Does something useful"),
    }
}

func (m *MySkill) Check() (bool, error) {
    // Your check implementation here
    return true, nil
}

func (m *MySkill) Run() types.Result {
    // Your run implementation here
    return types.Result{Changed: true, Message: "Done"}
}
```

## RunnableOptions

Configuration options for skill execution.

```go
type RunnableOptions struct {
    Args    map[string]string  // Override node-level args
    DryRun  bool               // Override dry-run mode
    Timeout time.Duration      // Execution timeout
}
```

Used with `RunByID()` for per-execution overrides.

## NodeConfig

Central configuration structure for all remote operations.

```go
type NodeConfig struct {
    // SSH connection settings
    SSHHost  string            // Hostname or IP address
    SSHPort  string            // SSH port (default: "22")
    SSHLogin string            // SSH login user
    SSHKey   string            // Private key filename (resolved to ~/.ssh/)
    
    // User settings
    RootUser    string         // Root/admin user
    NonRootUser string         // Non-root user
    
    // Database settings
    DBPort         string      // Database port
    DBRootPassword string      // Database root password
    
    // Extra arguments for skills
    Args map[string]string
    
    // Logger for structured logging
    Logger *slog.Logger
    
    // Dry-run mode flag
    IsDryRunMode bool
}
```

### Methods

```go
// SSHAddr returns host:port
func (c NodeConfig) SSHAddr() string

// GetArg retrieves from Args map
func (c NodeConfig) GetArg(key string) string

// GetArgOr retrieves with default value
func (c NodeConfig) GetArgOr(key, defaultValue string) string

// GetLoggerOrDefault returns logger or slog.Default()
func (c NodeConfig) GetLoggerOrDefault() *slog.Logger
```

### SSHAddr

Returns the full SSH address as `host:port`. Port defaults to "22" if not set.

### GetArg

Retrieves an argument from the `Args` map. Returns empty string if not found.

### GetArgOr

Retrieves an argument with a default value fallback.

### GetLoggerOrDefault

Returns the configured logger or `slog.Default()` if nil.

## PromptConfig

Configuration for a single user prompt.

```go
type PromptConfig struct {
    Name     string             // Variable name
    Prompt   string             // Prompt message to display
    Private  bool               // Hide input (true) or show it (false)
    Default  string             // Default value if user provides no input
    Confirm  bool               // Require confirmation (for passwords)
    Validate func(string) error // Validation function
    Required bool               // Whether the field is required
}
```

### Fields

- **Name**: The variable name for storing the result
- **Prompt**: The message displayed to the user
- **Private**: If true, input is hidden (like a password)
- **Default**: Default value used if user provides no input
- **Confirm**: If true, requires confirmation (for passwords)
- **Validate**: Optional validation function that returns an error if invalid
- **Required**: If true, empty input is rejected

### Example

```go
cfg := types.PromptConfig{
    Name:     "email",
    Prompt:   "Email address",
    Private:  false,
    Default:  "user@example.com",
    Confirm:  false,
    Required: true,
    Validate: func(value string) error {
        if !strings.Contains(value, "@") {
            return fmt.Errorf("invalid email format")
        }
        return nil
    },
}
```

## PromptResult

Contains the results of a prompt session.

```go
type PromptResult map[string]string
```

A map of variable names to user-provided values.

### Example

```go
results := types.PromptResult{
    "username": "admin",
    "password": "secret123",
    "port":     "8080",
}

// Access values
username := results["username"]
password := results["password"]
```

### With PromptMultiple

```go
prompts := []types.PromptConfig{
    {Name: "username", Prompt: "Username", Required: true},
    {Name: "password", Prompt: "Password", Private: true, Confirm: true, Required: true},
}

results, err := ork.PromptMultiple(prompts)
if err != nil {
    log.Fatal(err)
}

username := results["username"]
password := results["password"]
```

## Registry

Skill registry for ID-based lookup.

```go
type Registry struct {
    playbooks map[string]RunnableInterface
    mu        sync.RWMutex
}
```

### Constructor

```go
func NewRegistry() *Registry
```

### Methods

```go
// Register a skill
func (r *Registry) PlaybookRegister(p RunnableInterface) error

// Find skill by ID
func (r *Registry) PlaybookFindByID(id string) (RunnableInterface, bool)

// List all registered skills
func (r *Registry) PlaybookList() []RunnableInterface

// Get all skill IDs
func (r *Registry) GetPlaybookIDs() []string
```

### Usage

```go
// Create registry
registry := types.NewRegistry()

// Register skills
registry.PlaybookRegister(skills.NewPing())
registry.PlaybookRegister(skills.NewAptUpdate())

// Lookup by ID
skill, ok := registry.PlaybookFindByID("ping")
if ok {
    result := skill.Run()
}
```

## Command

Represents a shell command with its description.

```go
type Command struct {
    Command     string
    Description string
}
```

Used to display and execute shell commands in a structured way, especially useful in dry-run mode to show what commands would be executed.

### Example

```go
cmd := types.Command{
    Command: "ls -la",
    Description: "List all files in long format",
}
```

## Result

Represents the outcome of a single operation (command or skill execution).

```go
type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}
```

### Changed

Indicates whether the operation modified the system.

- `true`: Changes were made (e.g., packages updated, user created)
- `false`: System was already in the desired state (idempotent operation)

```go
if result.Changed {
    log.Println("System was modified")
} else {
    log.Println("No changes needed")
}
```

### Message

A human-readable description of what happened.

```go
// Examples:
result.Message = "Package database updated"
result.Message = "User 'alice' created"
result.Message = "2GB swap file created"
result.Message = "Already configured - no changes made"
```

### Details

Additional key-value information about the operation.

```go
result.Details = map[string]string{
    "size":        "2GB",
    "file":        "/swapfile",
    "swappiness":  "10",
    "output":      "... command output ...",
}

// Access details
for key, value := range result.Details {
    log.Printf("%s: %s", key, value)
}
```

Common detail keys by skill:

| Skill | Detail Keys |
|----------|-------------|
| ping | `uptime` |
| apt-update | `output` |
| swap-create | `size`, `file`, `swappiness`, `status` |
| user-create | `username`, `home`, `shell` |
| mariadb-status | Various status fields |

### Error

Non-nil if the operation failed. When `Error` is non-nil, `Changed` may still be `true` if some changes occurred before the failure.

```go
if result.Error != nil {
    log.Fatalf("Operation failed: %v", result.Error)
}
```

## Results

Contains per-node results from any operation (command or skill) on multiple nodes.

```go
type Results struct {
    Results map[string]Result  // Key is node hostname
}
```

The `Results` map keys are hostnames (as configured when creating nodes).

### Summary

Returns aggregated statistics about all results.

```go
func (r Results) Summary() Summary
```

```go
results := inv.Run(skills.NewPing())
summary := results.Summary()

fmt.Printf("Total: %d\n", summary.Total)
fmt.Printf("Changed: %d\n", summary.Changed)
fmt.Printf("Unchanged: %d\n", summary.Unchanged)
fmt.Printf("Failed: %d\n", summary.Failed)
```

### Iterating Results

```go
results := group.Run(skills.NewAptUpdate())

for hostname, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s: FAILED - %v", hostname, result.Error)
    } else if result.Changed {
        log.Printf("%s: CHANGED - %s", hostname, result.Message)
    } else {
        log.Printf("%s: OK (no changes)", hostname)
    }
}
```

## Summary

Aggregated statistics from a `Results` collection.

```go
type Summary struct {
    Total     int  // Total number of nodes
    Changed   int  // Nodes where changes were made
    Unchanged int  // Nodes with no changes needed
    Failed    int  // Nodes where execution failed
}
```

### Usage

```go
results := inv.Run(skills.NewAptUpgrade())
summary := results.Summary()

// Quick status check
if summary.Failed > 0 {
    log.Printf("WARNING: %d nodes failed", summary.Failed)
}

if summary.Changed == summary.Total {
    log.Println("All nodes were updated")
}
```

## Type Relationships

```mermaid
graph TD
    A[Node/Group/Inventory] -->|RunCommand| B[types.Results]
    A -->|Run| B
    B -->|Summary| C[types.Summary]
    B -->|Results map| D[types.Result]
    E[Skill] -->|Run| F[types.Result]
    F -->|Converted| D
```

Note: Skills return `types.Result` directly, no conversion needed.

## Conversion

Skills return `types.Result` directly, no conversion needed:

```go
// Skill returns types.Result
result := skill.Run()
// result is already types.Result
```

## Examples

### Handling Single Node Result

```go
node := ork.NewNodeForHost("server.example.com")
results := node.RunCommand("uptime")

// Single node - key is the hostname
result := results.Results["server.example.com"]

if result.Error != nil {
    log.Fatalf("Command failed: %v", result.Error)
}

log.Println(result.Message)
```

### Handling Multiple Node Results

```go
inv := ork.NewInventory()
// ... add groups with nodes ...

results := inv.RunPlaybook(playbooks.NewPing())

// Get summary first
summary := results.Summary()
log.Printf("Ping results: %d total, %d failed", 
    summary.Total, summary.Failed)

// Process individual results
for hostname, result := range results.Results {
    if result.Error != nil {
        log.Printf("[%s] Connection failed: %v", hostname, result.Error)
        continue
    }
    
    log.Printf("[%s] Connected: %s", hostname, result.Message)
    
    if uptime, ok := result.Details["uptime"]; ok {
        log.Printf("[%s] Uptime: %s", hostname, uptime)
    }
}
```

### Error Handling Patterns

```go
results := group.Run(skills.NewAptUpgrade())

// Pattern 1: Fail on any error
for hostname, result := range results.Results {
    if result.Error != nil {
        log.Fatalf("%s failed: %v", hostname, result.Error)
    }
}

// Pattern 2: Collect errors and report
var failures []string
for hostname, result := range results.Results {
    if result.Error != nil {
        failures = append(failures, fmt.Sprintf("%s: %v", hostname, result.Error))
    }
}
if len(failures) > 0 {
    log.Printf("Completed with %d failures:\n%s", 
        len(failures), strings.Join(failures, "\n"))
}

// Pattern 3: Continue on errors, report at end
summary := results.Summary()
if summary.Failed > 0 {
    log.Printf("WARNING: %d/%d nodes failed", summary.Failed, summary.Total)
}
```

### Check Mode Results

```go
// Preview changes
results := node.Check(skills.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would upgrade: %s", result.Message)
    // Now actually run it
    results = node.Run(skills.NewAptUpgrade())
} else {
    log.Println("No upgrades needed")
}
```

## Design Notes

### Why RunnableInterface in types Package?

Package isolation prevents circular dependencies:
- `types` package defines `RunnableInterface`, `BasePlaybook`, `BaseSkill`, and `Registry`
- `skills` package implements RunnableInterface
- `ork` package uses types.RunnableInterface for type safety
- This allows the registry to be in a lower-level package

### Why BasePlaybook vs BaseSkill?

- **BasePlaybook**: Use when you want optional Check() with a default implementation (returns false). Check() is not enforced.
- **BaseSkill**: Use when you want to enforce implementation of both Check() and Run(). Both methods return errors by default and must be overridden.

### Map Key Choice

Results use hostname as the map key because:
1. Hostname is the primary identifier for nodes
2. IP addresses may change
3. Natural fit for `NewNodeForHost(host)` pattern
4. Human-readable in logs and output

### Empty Results

Always check if hostname exists in the map:

```go
result, ok := results.Results["server.example.com"]
if !ok {
    log.Fatal("No result for host")
}
```

## See Also

- [ork](ork.md) - Uses types.Results for all operations
- [skills](skills.md) - Implements RunnableInterface
- [API Reference](../api_reference.md) - Complete API
