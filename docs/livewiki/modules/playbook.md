---
path: modules/playbook.md
page-type: module
summary: Playbook interface, base implementation, and registry for automation tasks.
tags: [module, playbook, interface, registry]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# playbook Package

Defines the playbook interface and provides base implementations for creating automation tasks.

## Purpose

The `playbook` package defines:
- `PlaybookInterface`: The interface all playbooks must implement
- `BasePlaybook`: Default implementation with common functionality
- `Result`: The outcome of playbook execution
- `Registry`: For registering and looking up playbooks by ID
- Constants: Playbook ID constants for built-in playbooks

## Key Files

| File | Purpose |
|------|---------|
| `playbook.go` | `PlaybookInterface`, `Result`, `PlaybookOptions` |
| `base_playbook.go` | `BasePlaybook` default implementation |
| `constants.go` | Playbook ID constants |
| `functions.go` | Utility functions |
| `registry.go` | `Registry` for playbook lookup |

## PlaybookInterface

All automation playbooks must implement this interface.

```go
type PlaybookInterface interface {
    // Identification
    GetID() string
    SetID(id string) PlaybookInterface
    GetDescription() string
    SetDescription(description string) PlaybookInterface
    
    // Configuration
    GetConfig() config.NodeConfig
    SetConfig(cfg config.NodeConfig) PlaybookInterface
    
    // Arguments
    GetArg(key string) string
    SetArg(key, value string) PlaybookInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) PlaybookInterface
    
    // Execution options
    IsDryRun() bool
    SetDryRun(dryRun bool) PlaybookInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) PlaybookInterface
    
    // Core operations
    Check() (bool, error)
    Run() Result
}
```

### Check

Determines if the playbook needs to make changes.

```go
func (p PlaybookInterface) Check() (bool, error)
```

- Returns `true` if changes are needed
- Returns `false` if system is already in desired state
- Returns error if the check itself fails

### Run

Executes the playbook and returns the result.

```go
func (p PlaybookInterface) Run() Result
```

The `Result.Changed` field indicates whether any modifications were made.

## Result

Represents the outcome of playbook execution.

```go
type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}
```

### Changed

- `true`: The playbook modified the system
- `false`: The system was already in the desired state

### Message

Human-readable description of what happened.

```go
Result{
    Changed: true,
    Message: "Created 2GB swap file",
}
```

### Details

Additional information as key-value strings.

```go
Result{
    Changed: true,
    Message: "Created swap file",
    Details: map[string]string{
        "size": "2GB",
        "file": "/swapfile",
        "swappiness": "10",
    },
}
```

### Error

Non-nil if execution failed. When `Error` is non-nil, `Changed` may be `true` if some changes occurred before the failure.

## BasePlaybook

Default implementation of `PlaybookInterface` providing common functionality.

```go
type BasePlaybook struct {
    id          string
    description string
    config      config.NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}
```

### Constructor

```go
func NewBasePlaybook() *BasePlaybook
```

### Methods

All `PlaybookInterface` methods are implemented with sensible defaults:

```go
// Identification
func (b *BasePlaybook) GetID() string
func (b *BasePlaybook) SetID(id string) PlaybookInterface
func (b *BasePlaybook) GetDescription() string
func (b *BasePlaybook) SetDescription(desc string) PlaybookInterface

// Configuration
func (b *BasePlaybook) GetConfig() config.NodeConfig
func (b *BasePlaybook) SetConfig(cfg config.NodeConfig) PlaybookInterface

// Arguments
func (b *BasePlaybook) GetArg(key string) string
func (b *BasePlaybook) SetArg(key, value string) PlaybookInterface
func (b *BasePlaybook) GetArgs() map[string]string
func (b *BasePlaybook) SetArgs(args map[string]string) PlaybookInterface

// Options
func (b *BasePlaybook) IsDryRun() bool
func (b *BasePlaybook) SetDryRun(dryRun bool) PlaybookInterface
func (b *BasePlaybook) GetTimeout() time.Duration
func (b *BasePlaybook) SetTimeout(timeout time.Duration) PlaybookInterface
```

Note: `Check()` and `Run()` must be implemented by concrete types.

### Embedding BasePlaybook

Create custom playbooks by embedding `BasePlaybook`:

```go
type MyPlaybook struct {
    *playbook.BasePlaybook
}

func (m *MyPlaybook) Check() (bool, error) {
    // Implement check logic
}

func (m *MyPlaybook) Run() playbook.Result {
    // Implement run logic
}

func NewMyPlaybook() playbook.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID("my-playbook")
    pb.SetDescription("Does something useful")
    return &MyPlaybook{BasePlaybook: pb}
}
```

## PlaybookOptions

Configuration options for playbook execution.

```go
type PlaybookOptions struct {
    Args    map[string]string  // Override node-level args
    DryRun  bool               // Override dry-run mode
    Timeout time.Duration      // Execution timeout
}
```

Used with `RunPlaybookByID()` for per-execution overrides.

## Registry

Global registry for playbook lookup by ID.

```go
type Registry struct {
    playbooks map[string]PlaybookInterface
    mu        sync.RWMutex
}
```

### Constructor

```go
func NewRegistry() *Registry
```

### Methods

```go
// Register a playbook
func (r *Registry) PlaybookRegister(pb PlaybookInterface) error

// Find playbook by ID
func (r *Registry) PlaybookFindByID(id string) (PlaybookInterface, bool)

// List all registered playbooks
func (r *Registry) PlaybookList() []PlaybookInterface
```

### Usage

```go
// Create registry
registry := playbook.NewRegistry()

// Register playbooks
registry.PlaybookRegister(playbooks.NewPing())
registry.PlaybookRegister(playbooks.NewAptUpdate())

// Lookup by ID
pb, ok := registry.PlaybookFindByID("ping")
if ok {
    result := pb.Run()
}
```

## Playbook ID Constants

Built-in playbook IDs defined in `constants.go`:

```go
const (
    // System
    IDPing              = "ping"
    IDAptUpdate         = "apt-update"
    IDAptUpgrade        = "apt-upgrade"
    IDAptStatus         = "apt-status"
    IDReboot            = "reboot"
    
    // Swap
    IDSwapCreate        = "swap-create"
    IDSwapDelete        = "swap-delete"
    IDSwapStatus        = "swap-status"
    
    // Users
    IDUserCreate        = "user-create"
    IDUserDelete        = "user-delete"
    IDUserStatus        = "user-status"
    
    // Security
    IDSshHarden         = "ssh-harden"
    IDKernelHarden      = "kernel-harden"
    IDAideInstall       = "aide-install"
    IDAuditdInstall     = "auditd-install"
    IDSshChangePort     = "ssh-change-port"
    
    // UFW
    IDUfwInstall        = "ufw-install"
    IDUfwStatus         = "ufw-status"
    IDUfwAllowMariaDB   = "ufw-allow-mariadb"
    
    // Fail2ban
    IDFail2banInstall   = "fail2ban-install"
    IDFail2banStatus    = "fail2ban-status"
    
    // MariaDB
    IDMariadbInstall            = "mariadb-install"
    IDMariadbSecure             = "mariadb-secure"
    IDMariadbCreateDB           = "mariadb-create-db"
    IDMariadbCreateUser         = "mariadb-create-user"
    IDMariadbStatus             = "mariadb-status"
    IDMariadbListDBs            = "mariadb-list-dbs"
    IDMariadbListUsers          = "mariadb-list-users"
    IDMariadbBackup             = "mariadb-backup"
    IDMariadbSecurityAudit      = "mariadb-security-audit"
    IDMariadbChangePort         = "mariadb-change-port"
    IDMariadbEnableSSL          = "mariadb-enable-ssl"
    IDMariadbEnableEncryption   = "mariadb-enable-encryption"
    IDMariadbBackupEncrypt      = "mariadb-backup-encrypt"
)
```

## Idempotency Pattern

Standard playbook implementation pattern:

```go
type MyPlaybook struct {
    *playbook.BasePlaybook
}

// Check determines if changes are needed
func (m *MyPlaybook) Check() (bool, error) {
    cfg := m.GetConfig()
    // Check current state
    output, _ := ssh.Run(cfg, "check command")
    return !strings.Contains(output, "configured"), nil
}

// Run executes the playbook
func (m *MyPlaybook) Run() playbook.Result {
    cfg := m.GetConfig()
    
    // Handle dry-run
    if cfg.IsDryRunMode {
        return playbook.Result{
            Changed: true,
            Message: "Would apply configuration",
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
    _, err := ssh.Run(cfg, "apply command")
    if err != nil {
        return playbook.Result{
            Changed: false,
            Error:   err,
        }
    }
    
    return playbook.Result{
        Changed: true,
        Message: "Configuration applied",
    }
}
```

## Examples

### Creating a Custom Playbook

```go
package myplaybook

import (
    "fmt"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/ssh"
)

const IDMyTask = "my-task"

type MyTask struct {
    *playbook.BasePlaybook
}

func (m *MyTask) Check() (bool, error) {
    cfg := m.GetConfig()
    output, _ := ssh.Run(cfg, "cat /etc/my-config")
    return !strings.Contains(output, "done"), nil
}

func (m *MyTask) Run() playbook.Result {
    cfg := m.GetConfig()
    
    if cfg.IsDryRunMode {
        return playbook.Result{
            Changed: true,
            Message: "Would configure my-task",
        }
    }
    
    needsChange, _ := m.Check()
    if !needsChange {
        return playbook.Result{
            Changed: false,
            Message: "Already configured",
        }
    }
    
    _, err := ssh.Run(cfg, "echo 'done' > /etc/my-config")
    if err != nil {
        return playbook.Result{
            Changed: false,
            Error:   err,
        }
    }
    
    return playbook.Result{
        Changed: true,
        Message: "Configured my-task",
    }
}

func NewMyTask() playbook.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID(IDMyTask)
    pb.SetDescription("Configure my custom task")
    return &MyTask{BasePlaybook: pb}
}
```

### Using the Registry

```go
// Register custom playbook
ork.GetDefaultRegistry().PlaybookRegister(myplaybook.NewMyTask())

// Run by ID
results := node.RunPlaybookByID("my-task")
```

## See Also

- [ork](ork.md) - Uses playbook package
- [playbooks](playbooks.md) - Built-in playbook implementations
- [config](config.md) - NodeConfig used by playbooks
- [Development](../development.md) - Creating custom playbooks
