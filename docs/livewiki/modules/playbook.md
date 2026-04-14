---
path: modules/playbook.md
page-type: module
summary: BasePlaybook implementation and utility functions for automation tasks.
tags: [module, playbook, baseplaybook]
created: 2025-04-14
updated: 2026-04-14
version: 1.1.0
---

# playbook Package

Provides BasePlaybook implementation and utility functions for creating automation tasks.

## Purpose

The `playbook` package provides:
- `BasePlaybook`: Default implementation of types.PlaybookInterface with common functionality
- Utility functions for playbook operations
- Note: `PlaybookInterface`, `Registry`, and `PlaybookOptions` are now defined in the `types` package

## Key Files

| File | Purpose |
|------|---------|
| `base_playbook.go` | `BasePlaybook` default implementation |
| `constants.go` | Playbook ID constants |
| `functions.go` | Utility functions |

## PlaybookInterface (in types package)

The `PlaybookInterface` is now defined in the `types` package to prevent circular dependencies. See [types module](types.md) for the full interface definition.

## Result (in types package)

The `Result` type is now defined in the `types` package. See [types module](types.md) for the full type definition.

## BasePlaybook

Default implementation of `types.PlaybookInterface` providing common functionality.

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

All `types.PlaybookInterface` methods are implemented with sensible defaults:

```go
// Identification
func (b *BasePlaybook) GetID() string
func (b *BasePlaybook) SetID(id string) types.PlaybookInterface
func (b *BasePlaybook) GetDescription() string
func (b *BasePlaybook) SetDescription(desc string) types.PlaybookInterface

// Configuration
func (b *BasePlaybook) GetNodeConfig() config.NodeConfig
func (b *BasePlaybook) SetNodeConfig(cfg config.NodeConfig) types.PlaybookInterface

// Arguments
func (b *BasePlaybook) GetArg(key string) string
func (b *BasePlaybook) SetArg(key, value string) types.PlaybookInterface
func (b *BasePlaybook) GetArgs() map[string]string
func (b *BasePlaybook) SetArgs(args map[string]string) types.PlaybookInterface

// Options
func (b *BasePlaybook) IsDryRun() bool
func (b *BasePlaybook) SetDryRun(dryRun bool) types.PlaybookInterface
func (b *BasePlaybook) GetTimeout() time.Duration
func (b *BasePlaybook) SetTimeout(timeout time.Duration) types.PlaybookInterface
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

func (m *MyPlaybook) Run() types.Result {
    // Implement run logic
}

func NewMyPlaybook() types.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID("my-playbook")
    pb.SetDescription("Does something useful")
    return &MyPlaybook{BasePlaybook: pb}
}
```

## Registry (in types package)

The `Registry` is now defined in the `types` package to prevent circular dependencies. See [types module](types.md) for the full registry definition.

## PlaybookOptions (in types package)

The `PlaybookOptions` type is now defined in the `types` package. See [types module](types.md) for the full type definition.

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
    cfg := m.GetNodeConfig()
    // Check current state
    output, _ := ssh.Run(cfg, "check command")
    return !strings.Contains(output, "configured"), nil
}

// Run executes the playbook
func (m *MyPlaybook) Run() types.Result {
    cfg := m.GetNodeConfig()
    
    // Handle dry-run
    if cfg.IsDryRunMode {
        return types.Result{
            Changed: true,
            Message: "Would apply configuration",
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
    _, err := ssh.Run(cfg, "apply command")
    if err != nil {
        return types.Result{
            Changed: false,
            Error:   err,
        }
    }
    
    return types.Result{
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
    cfg := m.GetNodeConfig()
    output, _ := ssh.Run(cfg, "cat /etc/my-config")
    return !strings.Contains(output, "done"), nil
}

func (m *MyTask) Run() types.Result {
    cfg := m.GetNodeConfig()
    
    if cfg.IsDryRunMode {
        return types.Result{
            Changed: true,
            Message: "Would configure my-task",
        }
    }
    
    needsChange, _ := m.Check()
    if !needsChange {
        return types.Result{
            Changed: false,
            Message: "Already configured",
        }
    }
    
    _, err := ssh.Run(cfg, "echo 'done' > /etc/my-config")
    if err != nil {
        return types.Result{
            Changed: false,
            Error:   err,
        }
    }
    
    return types.Result{
        Changed: true,
        Message: "Configured my-task",
    }
}

func NewMyTask() types.PlaybookInterface {
    pb := playbook.NewBasePlaybook()
    pb.SetID(IDMyTask)
    pb.SetDescription("Configure my custom task")
    return &MyTask{BasePlaybook: pb}
}
```

### Using the Registry

```go
// Register custom playbook
registry, err := ork.GetGlobalPlaybookRegistry()
if err != nil {
    log.Fatal(err)
}
registry.PlaybookRegister(myplaybook.NewMyTask())

// Run by ID
results := node.RunPlaybookByID("my-task")
```

## See Also

- [types](types.md) - PlaybookInterface, Registry, PlaybookOptions
- [ork](ork.md) - Uses playbook package
- [playbooks](playbooks.md) - Built-in playbook implementations
- [config](config.md) - NodeConfig used by playbooks
- [Development](../development.md) - Creating custom playbooks
