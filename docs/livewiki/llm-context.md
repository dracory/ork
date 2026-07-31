---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption.
tags: [llm, context, summary, vault, prompts]
created: 2025-04-14
updated: 2026-07-31
version: 2.1.0
---

# LLM Context: Ork

## Changelog
- **v2.1.0** (2026-07-31): Documented ToMap/FromMap cloning for concurrency safety, added missing skills to IDs list, updated RunnableInterface snippet, marked RunByID deprecated, refreshed registry function names
- **v2.0.0** (2026-04-15): Major terminology refactoring - playbooks renamed to skills, PlaybookInterface renamed to RunnableInterface, BasePlaybook moved to types package, NodeConfig moved to types package, config package removed, playbook package removed
- **v1.2.0** (2026-04-14): Added vault support for secure secrets management and prompt functions for interactive user input
- **v1.1.0** (2026-04-14): Updated registry functions and API references
- **v1.0.0** (2025-04-14): Initial creation

## Project Summary

Ork is a Go-based SSH automation framework for server management. It provides a type-safe, idempotent API for managing remote Linux servers over SSH, similar to Ansible but with Go's strong typing and concurrency features. The framework supports single-node operations through `Node`, multi-node operations through `Group` and `Inventory`, and includes 30+ built-in skills for common automation tasks like package management, user management, database setup, and security hardening.

Key differentiators:
- **Type-safe**: Full Go type safety with interfaces
- **Concurrent**: Inventory operations run in parallel across nodes
- **Idempotent**: All operations are safe to run multiple times
- **Dry-run mode**: Preview changes without executing on servers
- **Fluent API**: Chain methods for readable configuration

## Key Technologies

- **Go 1.25+**: Core language
- **golang.org/x/crypto/ssh**: SSH client
- **envenc**: Vault encryption/decryption (github.com/dracory/envenc)
- **testcontainers-go**: Integration testing
- **slog**: Structured logging (standard library)

## Directory Structure

```
ork/
├── ork.go                      # Package documentation
├── node_interface.go             # NodeInterface definition + constructors
├── node_implementation.go        # nodeImplementation struct + methods (includes skill cloning)
├── node_interface_test.go        # Node tests
├── group_implementation.go       # GroupInterface implementation
├── group_implementation_test.go  # Group tests
├── inventory_implementation.go     # InventoryInterface implementation
├── inventory_implementation_test.go
├── inventory_interface.go        # InventoryInterface definition
├── runner_interface.go           # RunnerInterface base for Node/Group/Inventory
├── command_implementation.go     # CommandInterface implementation
├── clone_id_test.go              # Tests for skill cloning
├── race_detector_test.go         # Race detector test for concurrent node.Run()
├── constants.go                  # Skill ID constants (ork package aliases)
├── registry.go                   # Global registry + NewDefaultRegistry factory
├── registry_test.go
├── vault.go                     # Vault functions for secure secrets management
├── prompts.go                    # Interactive prompt functions for user input
├── prompts_test.go
├── ssh/
│   ├── ssh.go                    # SSH Client wrapper
│   ├── functions.go              # Run, PrivateKeyPath
│   └── ssh_test.go
├── skills/
│   ├── doc.go                    # Package documentation
│   ├── constants.go              # Skill ID constants
│   ├── functions.go              # Shared skill utilities (shell/SQL escaping)
│   ├── apt/                      # apt-install, apt-update, apt-upgrade, apt-status
│   ├── ping/                     # ping connectivity check
│   ├── reboot/                   # server reboot
│   ├── swap/                     # swap-create, swap-delete, swap-status
│   ├── user/                     # user-create, user-delete, user-list, user-status, user-add-to-group
│   ├── mariadb/                  # 14 MariaDB skills (incl. mariadb-purge)
│   ├── security/                 # ssh-harden, kernel-harden, aide-install, auditd-install, ssh-change-port
│   ├── ufw/                      # ufw-install, ufw-status, ufw-allow, ufw-deny, ufw-delete, ufw-enable, ufw-default, ufw-disable, ufw-reset, ufw-allow-mariadb
│   └── fail2ban/                 # fail2ban-install, fail2ban-status
├── types/
│   ├── runnable_interface.go     # RunnableInterface (skills/playbooks contract, incl. ToMap/FromMap)
│   ├── runner_interface.go       # RunnerInterface base for Node/Group/Inventory
│   ├── become_interface.go       # BecomeInterface (privilege escalation)
│   ├── registry.go               # Registry, RunnableOptions
│   ├── command.go                # Command struct with description
│   ├── results.go                # Result, Results, Summary types
│   ├── prompt.go                 # PromptConfig, PromptResult types
│   ├── node_config.go            # NodeConfig struct + methods
│   ├── base_playbook.go          # BasePlaybook default implementation
│   ├── base_skill.go             # BaseSkill default implementation
│   └── constants.go              # Property/map key constants for cloning
├── internal/
│   ├── skilltest/                # Test helpers for skill testing
│   ├── sshtest/                  # Mock SSH client for testing
│   └── README.md                 # Testing framework documentation
└── docs/
    └── livewiki/                 # This documentation
```

## Core Concepts

1. **Node**: Represents a single remote server with SSH connection settings
2. **Group**: Collection of nodes that can be operated on together
3. **Inventory**: Manages multiple groups for large-scale operations
4. **Skill**: Reusable automation task implementing RunnableInterface
5. **RunnerInterface**: Base interface for Node, Group, Inventory (RunCommand, Run, etc.)
6. **Dry-run mode**: Safety feature that prevents actual server modifications
7. **Idempotency**: Check() determines if changes needed, Run() applies them
8. **Vault**: Secure secrets management using envenc for encrypted vault files
9. **Prompts**: Interactive user input functions for configuration and secrets collection

## Important Interfaces

```go
// NodeInterface - Single server management
type NodeInterface interface {
    RunnerInterface
    GetHost() string
    SetPort(port string) NodeInterface
    Connect() error
    Close() error
    // ... getters/setters for SSH config
}

// GroupInterface - Server group management
type GroupInterface interface {
    RunnerInterface
    GetName() string
    AddNode(node NodeInterface) GroupInterface
    // ...
}

// InventoryInterface - Multi-group management
type InventoryInterface interface {
    RunnerInterface
    AddGroup(group GroupInterface) InventoryInterface
    SetMaxConcurrency(max int) InventoryInterface
}

// RunnableInterface (in types package) - Automation tasks
type RunnableInterface interface {
    GetID() string
    SetNodeConfig(cfg NodeConfig) RunnableInterface
    Check() (bool, error)
    Run() Result
    ToMap() map[string]any    // serialize state for cloning
    FromMap(m map[string]any) // restore state from a clone
    BecomeInterface
    // ...
}

// RunnerInterface - Common operations
type RunnerInterface interface {
    RunCommand(cmd string) types.Results
    Run(runnable types.RunnableInterface) types.Results
    // Deprecated: Use Run instead.
    RunByID(id string, opts ...types.RunnableOptions) types.Results
    Check(runnable types.RunnableInterface) types.Results
    SetDryRunMode(dryRun bool) RunnerInterface
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

### Skill Execution
```go
// Direct instance (preferred)
results := node.Run(skills.NewAptUpdate())

// By ID (registry lookup) - deprecated, prefer Run() with a direct instance
results := node.RunByID(skills.IDAptUpdate)

// Check mode (dry-run for single skill)
results := node.Check(skills.NewAptUpgrade())
```

### Result Handling
```go
results := inv.Run(skills.NewPing())
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
| `node_implementation.go:477-545` | `cloneFromMap` - per-goroutine skill cloning for concurrency safety |
| `node_implementation.go:549-679` | `Run`/`RunByID`/`Check` - clone skill, set config, execute |
| `types/runnable_interface.go:8-104` | RunnableInterface incl. `ToMap()`/`FromMap()` |
| `types/runner_interface.go:8-41` | RunnerInterface - base for Node/Group/Inventory |
| `inventory_interface.go:5-29` | InventoryInterface definition |
| `group_implementation.go:13-174` | Group implementation with dry-run propagation |
| `types/registry.go:15-97` | Registry, RunnableOptions |
| `types/command.go:13-18` | Command struct with description |
| `types/prompt.go:1-16` | PromptConfig, PromptResult types for user input |
| `types/base_playbook.go:1-297` | BasePlaybook default implementation (Atom-backed, ToMap/FromMap) |
| `types/base_skill.go:1-297` | BaseSkill default implementation (Atom-backed, ToMap/FromMap) |
| `types/constants.go:1-55` | Property/map key constants for state storage and cloning |
| `types/node_config.go:6-80` | NodeConfig with SSHAddr(), GetArgOr(), KexAlgorithms, HostKeyAlgorithms |
| `ssh/functions.go:39-47` | Run() with dry-run safety check |
| `types/results.go:6-52` | Result, Results, Summary types |
| `registry.go:37-46` | GetGlobalSkillRegistry, NewDefaultRegistry |
| `race_detector_test.go` | Verifies concurrent node.Run() with shared skill is race-free |
| `vault.go:1-76` | Vault functions for secure secrets management |
| `prompts.go:1-241` | Interactive prompt functions for user input |
| `internal/skilltest/helpers.go` | Test helpers for skill testing |
| `internal/sshtest/mock_client.go` | Mock SSH client for testing |

## Skill IDs (for registry lookup)

System: `ping`, `apt-install`, `apt-update`, `apt-upgrade`, `apt-status`, `reboot`

Users: `user-create`, `user-delete`, `user-list`, `user-status`, `user-add-to-group`

Swap: `swap-create`, `swap-delete`, `swap-status`

Security: `ssh-harden`, `kernel-harden`, `aide-install`, `auditd-install`, `ssh-change-port`

UFW: `ufw-install`, `ufw-status`, `ufw-allow`, `ufw-deny`, `ufw-delete`, `ufw-enable`, `ufw-default`, `ufw-disable`, `ufw-reset`, `ufw-allow-mariadb`

Fail2ban: `fail2ban-install`, `fail2ban-status`

MariaDB: `mariadb-install`, `mariadb-secure`, `mariadb-create-db`, `mariadb-create-user`, `mariadb-status`, `mariadb-list-dbs`, `mariadb-list-users`, `mariadb-backup`, `mariadb-backup-encrypt`, `mariadb-security-audit`, `mariadb-change-port`, `mariadb-enable-ssl`, `mariadb-enable-encryption`, `mariadb-purge`

## Key Design Decisions

1. **Interface-based design**: All major components use interfaces for testability
2. **Dry-run at execution layer**: Safety in `ssh.Run()`, not in skills (though skills can detect)
3. **Result aggregation**: Results map keyed by hostname for multi-node operations
4. **Concurrent inventory**: Parallel execution with configurable concurrency
5. **Fluent API**: Method chaining for readable configuration
6. **Skill registry**: Global registry (types.Registry) for ID-based skill lookup with GetGlobalSkillRegistry() singleton
7. **Config propagation**: Dry-run mode propagates Inventory -> Group -> Node -> Skill
8. **Registry factory pattern**: NewDefaultRegistry() for isolated registries in testing
9. **Command struct**: types.Command wraps shell commands with descriptions for better dry-run output
10. **Internal testing framework**: skilltest and sshtest packages for comprehensive unit testing
11. **Vault integration**: envenc-based encrypted vault files for secure secrets management with dual loading strategies (keys map or environment variables)
12. **Prompt system**: Comprehensive user input functions with validation, confirmation, and multi-prompt support for interactive configuration
13. **BasePlaybook vs BaseSkill**: BasePlaybook provides fluent API with optional Check(), BaseSkill provides both Check() and Run() stubs that must be implemented
14. **Concurrency-safe skill cloning**: Before mutating a shared skill instance, `nodeImplementation` clones it via `cloneFromMap(skill, skill.ToMap())` so each goroutine gets an isolated copy. `BaseSkill`/`BasePlaybook` back their state with an `omni.AtomInterface` (thread-safe) and implement `ToMap()`/`FromMap()` for the clone. Function-typed fields (`runFunc`, `checkFunc`) are copied directly from the template since they cannot be serialized. Verified by `race_detector_test.go`.
15. **RunByID deprecated**: Prefer `Run(skills.NewXxx())` with a direct skill instance over `RunByID("xxx")`.

## Testing Approach

- **Unit tests**: Mock SSH via `internal/sshtest.MockClient` or `ssh.SetRunFunc()`
- **Test helpers**: `internal/skilltest` provides comprehensive test utilities
- **Integration tests**: Use testcontainers-go with real SSH containers
- **Thread safety**: Group uses `sync.RWMutex` for dry-run mode; `BaseSkill`/`BasePlaybook` use an `omni.AtomInterface` plus a per-struct mutex for `NodeConfig`
- **Race detector**: `race_detector_test.go` runs concurrent `node.Run()` calls sharing a single skill instance to verify the cloning-based isolation prevents data races
- **Mock SSH**: `internal/sshtest` provides expectation-based mock client for testing without SSH servers

## Extension Points

- **Custom skills**: Implement types.RunnableInterface, register in registry
- **SSH mocking**: Use `internal/sshtest.MockClient` or `ssh.SetRunFunc()` in tests
- **Custom logger**: Implement slog.Handler, set via SetLogger()
- **Isolated registries**: Use `NewDefaultRegistry()` for testing without global state
