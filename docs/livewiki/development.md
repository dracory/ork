---
path: development.md
page-type: tutorial
summary: Development workflow, testing guidelines, and contributing to Ork.
tags: [development, testing, contributing]
created: 2025-04-14
updated: 2026-07-31
version: 1.2.0
---

# Development Guide

This document covers the development workflow for contributing to Ork.

## Project Structure

```
ork/
├── ork.go                  # Main package entry point
├── node_interface.go       # NodeInterface definition
├── node_implementation.go  # Node implementation (includes skill cloning logic)
├── node_interface_test.go  # Node tests
├── group_implementation.go # Group implementation
├── group_implementation_test.go
├── inventory_implementation.go
├── inventory_implementation_test.go
├── inventory_interface.go
├── runner_interface.go     # RunnerInterface base for Node/Group/Inventory
├── command_implementation.go # CommandInterface implementation
├── clone_id_test.go        # Tests for skill cloning
├── race_detector_test.go   # Race detector test for concurrent node.Run()
├── constants.go            # Skill ID constants (ork package aliases)
├── registry.go             # Global registry + NewDefaultRegistry factory
├── registry_test.go
├── vault.go                # Vault functions for secure secrets
├── prompts.go              # Interactive prompt functions
├── prompts_test.go
├── ssh/
│   ├── ssh.go              # SSH client wrapper
│   ├── functions.go        # SSH utility functions
│   └── ssh_test.go
├── skills/
│   ├── doc.go              # Package documentation
│   ├── constants.go        # Skill ID constants
│   ├── functions.go        # Shared skill utilities (shell/SQL escaping, etc.)
│   ├── apt/                # apt-install, apt-update, apt-upgrade, apt-status
│   ├── ping/               # ping connectivity check
│   ├── reboot/             # server reboot
│   ├── swap/               # swap-create, swap-delete, swap-status
│   ├── user/               # user-create, user-delete, user-list, user-status, user-add-to-group
│   ├── mariadb/            # 13 MariaDB skills + mariadb-purge
│   ├── security/           # ssh-harden, kernel-harden, aide-install, auditd-install, ssh-change-port
│   ├── ufw/                # ufw-install, ufw-status, ufw-allow, ufw-deny, ufw-delete, ufw-enable, ufw-default, ufw-disable, ufw-reset, ufw-allow-mariadb
│   └── fail2ban/           # fail2ban-install, fail2ban-status
├── types/
│   ├── runnable_interface.go # RunnableInterface (skills/playbooks contract)
│   ├── runner_interface.go   # RunnerInterface (Node/Group/Inventory base)
│   ├── become_interface.go   # BecomeInterface (privilege escalation)
│   ├── base_playbook.go      # BasePlaybook default implementation
│   ├── base_skill.go         # BaseSkill default implementation
│   ├── constants.go          # Property/map key constants for cloning
│   ├── registry.go           # Registry, RunnableOptions
│   ├── command.go            # Command struct
│   ├── node_config.go        # NodeConfig struct + methods
│   ├── prompt.go             # PromptConfig, PromptResult
│   └── results.go            # Result, Results, Summary types
├── internal/
│   ├── skilltest/          # Test helpers for skill testing
│   ├── sshtest/            # Mock SSH client for testing
│   └── README.md           # Testing framework documentation
└── docs/
    └── livewiki/           # This documentation
```

## Setting Up Development Environment

### Prerequisites

- Go 1.25 or later
- SSH key pair for testing
- Access to a test server (or use integration tests with containers)

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/dracory/ork.git
cd ork

# Download dependencies
go mod download

# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

## Testing

### Unit Tests

Unit tests mock SSH connections and test logic without real servers:

```bash
# Run all unit tests
go test ./...

# Run specific package tests
go test ./ssh/
go test ./types/

# Run with coverage
go test -cover ./...

# Run the race detector (recommended when touching node_implementation.go
# or anything that runs skills concurrently)
go test -race ./...
```

### Integration Tests

Integration tests use testcontainers for real SSH connections:

```bash
# Run integration tests
go test -tags=integration ./...

# Run specific integration test
go test -v -run TestIntegration ./...
```

**Note**: Integration tests require Docker.

### Test Structure

```go
// Example test from node_interface_test.go
func TestNode_NewNodeForHost(t *testing.T) {
    node := NewNodeForHost("test.example.com")
    
    if node.GetHost() != "test.example.com" {
        t.Errorf("expected host 'test.example.com', got '%s'", node.GetHost())
    }
    
    if node.GetPort() != "22" {
        t.Errorf("expected default port '22', got '%s'", node.GetPort())
    }
    
    if node.GetUser() != "root" {
        t.Errorf("expected default user 'root', got '%s'", node.GetUser())
    }
}
```

### Mocking SSH

Ork provides two approaches for mocking SSH:

#### Option 1: Using internal/sshtest (Recommended)

```go
import "github.com/dracory/ork/internal/sshtest"

func TestNode_RunCommand(t *testing.T) {
    mock := sshtest.NewMockClient()
    mock.ExpectCommand("uptime", "up 5 days")
    mock.Connect()
    defer mock.Close()

    output, err := mock.Run("uptime")
    // ... assertions
    mock.AssertCommandRun("uptime")
}
```

#### Option 2: Using SetRunFunc

```go
func TestNode_RunCommand(t *testing.T) {
    // Mock SSH via SetRunFunc
    ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
        return "mocked output", nil
    })
    defer ssh.SetRunFunc(nil)

    // Test with mocked SSH
    node := NewNodeForHost("test.example.com")
    results := node.RunCommand("uptime")
    // ... assertions
}
```

## Creating a New Skill

### 1. Create Package Structure

```bash
mkdir -p skills/myskill
touch skills/myskill/constants.go
touch skills/myskill/myskill.go
```

### 2. Define Constants

```go
// skills/myskill/constants.go
package myskill

const (
    ArgParameter = "parameter"
    DefaultValue = "default"
)
```

### 3. Implement Skill

```go
// skills/myskill/myskill.go
package myskill

import (
    "fmt"

    "github.com/dracory/ork/skills"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/types"
)

// MySkill does something useful.
type MySkill struct {
    *types.BaseSkill
}

// Check determines if the skill needs to run.
func (m *MySkill) Check() (bool, error) {
    cfg := m.GetNodeConfig()
    parameter := m.GetArg(ArgParameter)

    // Check current state
    output, _ := ssh.Run(cfg, fmt.Sprintf("check %s", parameter))
    return output == "", nil
}

// Run executes the skill.
func (m *MySkill) Run() types.Result {
    cfg := m.GetNodeConfig()
    parameter := m.GetArg(ArgParameter)

    if parameter == "" {
        parameter = DefaultValue
    }

    // Check dry-run
    if cfg.IsDryRunMode {
        return types.Result{
            Changed: true,
            Message: fmt.Sprintf("Would run myskill with %s", parameter),
        }
    }

    // Check if needed
    needsChange, _ := m.Check()
    if !needsChange {
        return types.Result{
            Changed: false,
            Message: "Already configured",
        }
    }

    // Apply changes
    _, err := ssh.Run(cfg, fmt.Sprintf("apply %s", parameter))
    if err != nil {
        return types.Result{
            Changed: false,
            Message: "Failed to apply",
            Error:   err,
        }
    }

    return types.Result{
        Changed: true,
        Message: fmt.Sprintf("Applied %s", parameter),
    }
}

// NewMySkill creates a new instance.
func NewMySkill() types.RunnableInterface {
    return &MySkill{
        BaseSkill: types.NewBaseSkill().
            WithID(skills.IDMySkill).  // Add to skills/constants.go
            WithDescription("Does something useful"),
    }
}
```

### 4. Add ID to skills/constants.go

```go
const (
    // ... existing constants
    IDMySkill = "my-skill"
)
```

### 5. Add to ork/constants.go (Optional)

```go
const (
    // ... existing constants
    SkillMySkill = skills.IDMySkill
)
```

### 6. Register in registry.go

```go
import "github.com/dracory/ork/skills/myskill"

// Add to the skills slice in NewDefaultRegistry()
skills := []types.RunnableInterface{
    // ... existing skills
    myskill.NewMySkill(),
}
```

### 7. Write Tests

Using the internal/skilltest helper (recommended):

```go
// skills/myskill/myskill_test.go
package myskill

import (
    "testing"
    "github.com/dracory/ork/internal/skilltest"
)

func TestMySkill_Check(t *testing.T) {
    test := skilltest.New(t)
    defer test.Cleanup()
    test.Setup()

    test.SetArg(ArgParameter, "test")
    test.ExpectCommand("check parameter", "not configured")

    s := NewMySkill()
    s.SetNodeConfig(test.Config())

    needsChange, err := s.Check()
    test.AssertNoError(err)
    if !needsChange {
        t.Error("expected changes needed")
    }
}

func TestMySkill_Run(t *testing.T) {
    test := skilltest.New(t)
    defer test.Cleanup()
    test.Setup()

    s := NewMySkill()
    s.SetNodeConfig(test.Config())

    result := s.Run()
    test.AssertResultChanged(result)
}
```

Or using traditional mocking:

```go
// skills/myskill/myskill_test.go
package myskill

import (
    "testing"
    "github.com/dracory/ork/types"
)

func TestMySkill_Check(t *testing.T) {
    s := NewMySkill()
    s.SetNodeConfig(types.NodeConfig{
        SSHHost: "test.example.com",
        Args: map[string]string{
            ArgParameter: "test",
        },
    })

    needsChange, err := s.Check()
    // Add assertions
}

func TestMySkill_Run(t *testing.T) {
    s := NewMySkill()
    s.SetNodeConfig(types.NodeConfig{
        SSHHost: "test.example.com",
        IsDryRunMode: true,
    })

    result := s.Run()
    if !result.Changed {
        t.Error("expected Changed=true in dry-run mode")
    }
}
```

## Code Style Guidelines

### Naming Conventions

- **Interfaces**: `NodeInterface`, `RunnableInterface`, `RunnerInterface`
- **Implementations**: `nodeImplementation`, `groupImplementation`
- **Constructors**: `NewNodeForHost()`, `NewPing()`, `NewAptUpdate()`
- **Constants**: `IDAptUpdate`, `ArgUsername`, `DefaultShell`

### Documentation

All public types and functions must have documentation comments:

```go
// MySkill does something useful.
// It provides detailed functionality for X.
type MySkill struct {
    *types.BaseSkill
}

// Check determines if changes are needed.
// Returns true if the system needs modification.
func (m *MySkill) Check() (bool, error) {
    // ...
}
```

### Error Handling

Always wrap errors with context:

```go
output, err := ssh.Run(cfg, cmd)
if err != nil {
    return types.Result{
        Error: fmt.Errorf("failed to execute '%s': %w", cmd, err),
    }
}
```

## Commit Message Format

```
<type>: <short summary>

<optional body>

<optional footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Example:
```
feat: add mysql backup skill

- Implements mysqldump-based backup
- Supports compression
- Includes encryption option
```

## Running Linting

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run static analysis (if installed)
staticcheck ./...
```

## Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create git tag
4. Push to trigger CI/CD

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Debugging

### Enable Verbose Logging

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

node.SetLogger(logger)
```

### Common Debug Patterns

```go
// Check configuration
log.Printf("Config: %+v", node.GetNodeConfig())

// Check results
t.Logf("Results: %+v", results)

// Dry-run mode for testing
node.SetDryRunMode(true)
```

## See Also

- [Conventions](conventions.md) - Coding conventions
- [Troubleshooting](troubleshooting.md) - Common issues
- [Architecture](architecture.md) - System architecture
