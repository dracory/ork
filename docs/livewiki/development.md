---
path: development.md
page-type: tutorial
summary: Development workflow, testing guidelines, and contributing to Ork.
tags: [development, testing, contributing]
created: 2025-04-14
updated: 2026-04-14
version: 1.1.0
---

# Development Guide

This document covers the development workflow for contributing to Ork.

## Project Structure

```
ork/
├── ork.go                  # Main package entry point
├── node_interface.go       # NodeInterface definition
├── node_implementation.go  # Node implementation
├── node_interface_test.go  # Node tests
├── group_implementation.go # Group implementation
├── group_implementation_test.go
├── inventory_implementation.go
├── inventory_implementation_test.go
├── inventory_interface.go
├── runnable_interface.go   # Base runnable interface
├── constants.go            # Playbook ID constants
├── registry.go             # Global registry + NewDefaultRegistry factory
├── registry_test.go
├── config/
│   └── node_config.go      # Configuration types
├── ssh/
│   ├── ssh.go             # SSH client wrapper
│   ├── functions.go       # SSH utility functions
│   └── ssh_test.go
├── playbook/
│   ├── playbook.go        # BasePlaybook implementation
│   ├── base_playbook.go   # Base implementation
│   ├── constants.go       # Playbook IDs
│   └── functions.go       # Utility functions
├── playbooks/
│   ├── doc.go             # Package documentation
│   ├── apt/               # Apt playbooks
│   ├── ping/              # Ping playbook
│   ├── reboot/            # Reboot playbook
│   ├── swap/              # Swap management
│   ├── user/              # User management
│   ├── mariadb/           # MariaDB playbooks
│   ├── security/          # Security hardening
│   ├── ufw/               # UFW firewall
│   └── fail2ban/          # Fail2ban playbooks
├── types/
│   ├── registry.go        # PlaybookInterface, Registry, PlaybookOptions
│   ├── command.go         # Command struct
│   └── results.go         # Result types
├── internal/
│   ├── playbooktest/      # Test helpers for playbook testing
│   ├── sshtest/           # Mock SSH client for testing
│   └── README.md          # Testing framework documentation
└── docs/
    └── livewiki/          # This documentation
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
go test ./playbook/

# Run with coverage
go test -cover ./...
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
    ssh.SetRunFunc(func(cfg config.NodeConfig, cmd types.Command) (string, error) {
        return "mocked output", nil
    })
    defer ssh.SetRunFunc(nil)

    // Test with mocked SSH
    node := NewNodeForHost("test.example.com")
    results := node.RunCommand("uptime")
    // ... assertions
}
```

## Creating a New Playbook

### 1. Create Package Structure

```bash
mkdir -p playbooks/myplaybook
touch playbooks/myplaybook/constants.go
touch playbooks/myplaybook/myplaybook.go
```

### 2. Define Constants

```go
// playbooks/myplaybook/constants.go
package myplaybook

const (
    ArgParameter = "parameter"
    DefaultValue = "default"
)
```

### 3. Implement Playbook

```go
// playbooks/myplaybook/myplaybook.go
package myplaybook

import (
    "fmt"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/ssh"
)

// MyPlaybook does something useful.
type MyPlaybook struct {
    *playbook.BasePlaybook
}

// Check determines if the playbook needs to run.
func (m *MyPlaybook) Check() (bool, error) {
    cfg := m.GetConfig()
    parameter := m.GetArg(ArgParameter)
    
    // Check current state
    output, _ := ssh.Run(cfg, fmt.Sprintf("check %s", parameter))
    return output == "", nil
}

// Run executes the playbook.
func (m *MyPlaybook) Run() playbook.Result {
    cfg := m.GetConfig()
    parameter := m.GetArg(ArgParameter)
    
    if parameter == "" {
        parameter = DefaultValue
    }
    
    // Check dry-run
    if cfg.IsDryRunMode {
        return playbook.Result{
            Changed: true,
            Message: fmt.Sprintf("Would run myplaybook with %s", parameter),
        }
    }
    
    // Check if needed
    needsChange, _ := m.Check()
    if !needsChange {
        return playbook.Result{
            Changed: false,
            Message: "Already configured",
        }
    }
    
    // Apply changes
    _, err := ssh.Run(cfg, fmt.Sprintf("apply %s", parameter))
    if err != nil {
        return playbook.Result{
            Changed: false,
            Message: "Failed to apply",
            Error:   err,
        }
    }
    
    return playbook.Result{
        Changed: true,
        Message: fmt.Sprintf("Applied %s", parameter),
    }
}

// NewMyPlaybook creates a new instance.
func NewMyPlaybook() types.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID(playbooks.IDMyPlaybook)  // Add to playbook/constants.go
    pb.SetDescription("Does something useful")
    return &MyPlaybook{BasePlaybook: pb}
}
```

### 4. Add ID to playbook/constants.go

```go
const (
    // ... existing constants
    IDMyPlaybook = "my-playbook"
)
```

### 5. Add to ork/constants.go (Optional)

```go
const (
    // ... existing constants
    PlaybookMyPlaybook = playbooks.IDMyPlaybook
)
```

### 6. Register in registry.go

```go
import "github.com/dracory/ork/playbooks/myplaybook"

// Add to the playbooks slice in NewDefaultRegistry()
playbooks := []types.PlaybookInterface{
    // ... existing playbooks
    myplaybook.NewMyPlaybook(),
}
```

### 7. Write Tests

Using the internal/playbooktest helper (recommended):

```go
// playbooks/myplaybook/myplaybook_test.go
package myplaybook

import (
    "testing"
    "github.com/dracory/ork/internal/playbooktest"
)

func TestMyPlaybook_Check(t *testing.T) {
    test := playbooktest.New(t)
    defer test.Cleanup()
    test.Setup()
    
    test.SetArg(ArgParameter, "test")
    test.ExpectCommand("check parameter", "not configured")
    
    pb := NewMyPlaybook()
    pb.SetNodeConfig(test.Config())
    
    needsChange, err := pb.Check()
    test.AssertNoError(err)
    if !needsChange {
        t.Error("expected changes needed")
    }
}

func TestMyPlaybook_Run(t *testing.T) {
    test := playbooktest.New(t)
    defer test.Cleanup()
    test.Setup()
    
    pb := NewMyPlaybook()
    pb.SetNodeConfig(test.Config())
    
    result := pb.Run()
    test.AssertResultChanged(result)
}
```

Or using traditional mocking:

```go
// playbooks/myplaybook/myplaybook_test.go
package myplaybook

import (
    "testing"
    "github.com/dracory/ork/config"
)

func TestMyPlaybook_Check(t *testing.T) {
    pb := NewMyPlaybook()
    pb.SetNodeConfig(config.NodeConfig{
        SSHHost: "test.example.com",
        Args: map[string]string{
            ArgParameter: "test",
        },
    })
    
    needsChange, err := pb.Check()
    // Add assertions
}

func TestMyPlaybook_Run(t *testing.T) {
    pb := NewMyPlaybook()
    pb.SetNodeConfig(config.NodeConfig{
        SSHHost: "test.example.com",
        IsDryRunMode: true,
    })
    
    result := pb.Run()
    if !result.Changed {
        t.Error("expected Changed=true in dry-run mode")
    }
}
```

## Code Style Guidelines

### Naming Conventions

- **Interfaces**: `NodeInterface`, `PlaybookInterface`
- **Implementations**: `nodeImplementation`, `groupImplementation`
- **Constructors**: `NewNodeForHost()`, `NewPing()`, `NewAptUpdate()`
- **Constants**: `IDAptUpdate`, `ArgUsername`, `DefaultShell`

### Documentation

All public types and functions must have documentation comments:

```go
// MyPlaybook does something useful.
// It provides detailed functionality for X.
type MyPlaybook struct {
    *playbook.BasePlaybook
}

// Check determines if changes are needed.
// Returns true if the system needs modification.
func (m *MyPlaybook) Check() (bool, error) {
    // ...
}
```

### Error Handling

Always wrap errors with context:

```go
output, err := ssh.Run(cfg, cmd)
if err != nil {
    return playbook.Result{
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
feat: add mysql backup playbook

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
