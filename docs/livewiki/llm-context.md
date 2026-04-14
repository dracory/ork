---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption.
tags: [llm, context, summary]
created: 2025-04-14
updated: 2026-04-14
version: 1.1.0
---

# LLM Context: Ork

## Project Summary

Ork is a Go-based SSH automation framework for server management. It provides a type-safe, idempotent API for managing remote Linux servers over SSH, similar to Ansible but with Go's strong typing and concurrency features. The framework supports single-node operations through `Node`, multi-node operations through `Group` and `Inventory`, and includes 30+ built-in playbooks for common automation tasks like package management, user management, database setup, and security hardening.

Key differentiators:
- **Type-safe**: Full Go type safety with interfaces
- **Concurrent**: Inventory operations run in parallel across nodes
- **Idempotent**: All operations are safe to run multiple times
- **Dry-run mode**: Preview changes without executing on servers
- **Fluent API**: Chain methods for readable configuration

## Key Technologies

- **Go 1.25+**: Core language
- **simplessh**: SSH client wrapper (github.com/sfreiberg/simplessh)
- **testcontainers-go**: Integration testing
- **slog**: Structured logging (standard library)

## Directory Structure

```
ork/
├── ork.go                      # Package documentation
├── node_interface.go             # NodeInterface definition + constructors
├── node_implementation.go        # nodeImplementation struct + methods
├── node_interface_test.go        # Node tests
├── group_implementation.go       # GroupInterface implementation
├── group_implementation_test.go  # Group tests
├── inventory_implementation.go     # InventoryInterface implementation
├── inventory_implementation_test.go
├── inventory_interface.go        # InventoryInterface definition
├── runnable_interface.go         # RunnableInterface base
├── constants.go                  # Playbook ID constants (ork package)
├── registry.go                   # Global registry + NewDefaultRegistry factory
├── registry_test.go
├── config/
│   └── node_config.go            # NodeConfig struct + methods
├── ssh/
│   ├── ssh.go                    # SSH Client wrapper
│   ├── functions.go              # Run, PrivateKeyPath
│   └── ssh_test.go
├── playbook/
│   ├── playbook.go               # BasePlaybook implementation
│   ├── base_playbook.go          # BasePlaybook default implementation
│   ├── constants.go              # Playbook ID constants
│   └── functions.go              # Utility functions
├── playbooks/
│   ├── doc.go                    # Package documentation
│   ├── apt/                      # apt-update, apt-upgrade, apt-status
│   ├── ping/                     # ping connectivity check
│   ├── reboot/                   # server reboot
│   ├── swap/                     # swap-create, swap-delete, swap-status
│   ├── user/                     # user-create, user-delete, user-list, user-status
│   ├── mariadb/                  # 13 MariaDB playbooks
│   ├── security/                 # ssh-harden, kernel-harden, aide-install, auditd-install, ssh-change-port
│   ├── ufw/                      # ufw-install, ufw-status, ufw-allow-mariadb
│   └── fail2ban/                 # fail2ban-install, fail2ban-status
├── types/
│   ├── registry.go               # Registry, PlaybookInterface, PlaybookOptions
│   ├── command.go                # Command struct with description
│   └── results.go                # Result, Results, Summary types
├── internal/
│   ├── playbooktest/             # Test helpers for playbook testing
│   ├── sshtest/                  # Mock SSH client for testing
│   └── README.md                 # Testing framework documentation
└── docs/
    └── livewiki/                 # This documentation
```

## Core Concepts

1. **Node**: Represents a single remote server with SSH connection settings
2. **Group**: Collection of nodes that can be operated on together
3. **Inventory**: Manages multiple groups for large-scale operations
4. **Playbook**: Reusable automation task implementing PlaybookInterface
5. **RunnableInterface**: Base interface for Node, Group, Inventory (RunCommand, RunPlaybook, etc.)
6. **Dry-run mode**: Safety feature that prevents actual server modifications
7. **Idempotency**: Check() determines if changes needed, Run() applies them

## Important Interfaces

```go
// NodeInterface - Single server management
type NodeInterface interface {
    RunnableInterface
    GetHost() string
    SetPort(port string) NodeInterface
    Connect() error
    Close() error
    // ... getters/setters for SSH config
}

// GroupInterface - Server group management
type GroupInterface interface {
    RunnableInterface
    GetName() string
    AddNode(node NodeInterface) GroupInterface
    // ...
}

// InventoryInterface - Multi-group management
type InventoryInterface interface {
    RunnableInterface
    AddGroup(group GroupInterface) InventoryInterface
    SetMaxConcurrency(max int) InventoryInterface
}

// PlaybookInterface (in types package) - Automation tasks
type PlaybookInterface interface {
    GetID() string
    SetNodeConfig(cfg config.NodeConfig) PlaybookInterface
    Check() (bool, error)
    Run() Result
    // ...
}

// RunnableInterface - Common operations
type RunnableInterface interface {
    RunCommand(cmd string) types.Results
    RunPlaybook(pb types.PlaybookInterface) types.Results
    SetDryRunMode(dryRun bool) RunnableInterface
}
```

## Common Patterns

### Fluent Configuration
```go
node := ork.NewNodeForHost("server.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")
```

### Playbook Execution
```go
// Direct instance (preferred)
results := node.RunPlaybook(playbooks.NewAptUpdate())

// By ID (registry lookup)
results := node.RunPlaybookByID(playbooks.IDAptUpdate)

// Check mode (dry-run for single playbook)
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
```

### Result Handling
```go
results := inv.RunPlaybook(playbooks.NewPing())
summary := results.Summary()

for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    } else if result.Changed {
        log.Printf("%s changed: %s", host, result.Message)
    }
}
```

### Dry-Run Safety
```go
// Set at any level, propagates down
inv.SetDryRunMode(true)
group.SetDryRunMode(true)
node.SetDryRunMode(true)

// Safety enforced at ssh.Run() - returns "[dry-run]" marker
```

## Important Files

| File | Purpose |
|------|---------|
| `node_interface.go:17-244` | NodeInterface definition with full documentation |
| `node_implementation.go:28-435` | Node implementation, connection management |
| `runnable_interface.go:11-45` | RunnableInterface - base for all executables |
| `inventory_interface.go:5-29` | InventoryInterface definition |
| `group_implementation.go:13-174` | Group implementation with dry-run propagation |
| `types/registry.go:27-97` | PlaybookInterface, PlaybookOptions, Registry |
| `types/command.go:13-18` | Command struct with description |
| `playbook/base_playbook.go` | BasePlaybook default implementation |
| `config/node_config.go:6-67` | NodeConfig with SSHAddr(), GetArgOr() |
| `ssh/functions.go:39-47` | Run() with dry-run safety check |
| `types/results.go:6-52` | Result, Results, Summary types |
| `registry.go:37-46` | GetGlobalPlaybookRegistry, NewDefaultRegistry |
| `internal/playbooktest/helpers.go` | Test helpers for playbook testing |
| `internal/sshtest/mock_client.go` | Mock SSH client for testing |

## Playbook IDs (for registry lookup)

System: `ping`, `apt-update`, `apt-upgrade`, `apt-status`, `reboot`

Users: `user-create`, `user-delete`, `user-list`, `user-status`

Swap: `swap-create`, `swap-delete`, `swap-status`

Security: `ssh-harden`, `kernel-harden`, `aide-install`, `auditd-install`, `ssh-change-port`

UFW: `ufw-install`, `ufw-status`, `ufw-allow-mariadb`

Fail2ban: `fail2ban-install`, `fail2ban-status`

MariaDB: `mariadb-install`, `mariadb-secure`, `mariadb-create-db`, `mariadb-create-user`, `mariadb-status`, `mariadb-list-dbs`, `mariadb-list-users`, `mariadb-backup`, `mariadb-security-audit`, `mariadb-change-port`, `mariadb-enable-ssl`, `mariadb-enable-encryption`, `mariadb-backup-encrypt`

## Key Design Decisions

1. **Interface-based design**: All major components use interfaces for testability
2. **Dry-run at execution layer**: Safety in `ssh.Run()`, not in playbooks (though playbooks can detect)
3. **Result aggregation**: Results map keyed by hostname for multi-node operations
4. **Concurrent inventory**: Parallel execution with configurable concurrency
5. **Fluent API**: Method chaining for readable configuration
6. **Playbook registry**: Global registry (types.Registry) for ID-based playbook lookup with GetGlobalPlaybookRegistry() singleton
7. **Config propagation**: Dry-run mode propagates Inventory -> Group -> Node -> Playbook
8. **Registry factory pattern**: NewDefaultRegistry() for isolated registries in testing
9. **Command struct**: types.Command wraps shell commands with descriptions for better dry-run output
10. **Internal testing framework**: playbooktest and sshtest packages for comprehensive unit testing

## Testing Approach

- **Unit tests**: Mock SSH via `internal/sshtest.MockClient` or `ssh.SetRunFunc()`
- **Test helpers**: `internal/playbooktest` provides comprehensive test utilities
- **Integration tests**: Use testcontainers-go with real SSH containers
- **Thread safety**: Group uses `sync.RWMutex` for dry-run mode
- **Mock SSH**: `internal/sshtest` provides expectation-based mock client for testing without SSH servers

## Extension Points

- **Custom playbooks**: Implement types.PlaybookInterface, register in registry
- **SSH mocking**: Use `internal/sshtest.MockClient` or `ssh.SetRunFunc()` in tests
- **Custom logger**: Implement slog.Handler, set via SetLogger()
- **Isolated registries**: Use `NewDefaultRegistry()` for testing without global state
