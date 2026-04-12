# Proposal: Rollback Support

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

When playbooks fail or produce unexpected results, there's no built-in way to undo changes:

- Failed deployments leave systems in inconsistent states
- No automatic cleanup on errors
- Manual intervention required to restore previous state
- Risky to experiment with playbooks in production

Ansible has limited rollback support, but it's a common pain point. Ork can do better.

## Proposed Solution

Implement a rollback framework that:

1. **Tracks changes** made by playbooks
2. **Generates rollback operations** automatically
3. **Executes rollbacks** on failure or manually
4. **Supports transactions** for atomic operations

## Core Concepts

### 1. Reversible Operations

```go
type ReversiblePlaybook interface {
    Playbook
    Rollback(ctx *ExecutionContext) error
}

type Operation struct {
    Type        string            // "create", "modify", "delete", "execute"
    Resource    string            // What was changed
    Forward     func() error      // Apply the change
    Backward    func() error      // Undo the change
    State       OperationState    // pending, applied, rolled_back, failed
}

type OperationState string

const (
    StatePending    OperationState = "pending"
    StateApplied    OperationState = "applied"
    StateRolledBack OperationState = "rolled_back"
    StateFailed     OperationState = "failed"
)
```

### 2. Transaction Manager

```go
type Transaction struct {
    ID         string
    Operations []Operation
    State      TransactionState
    StartTime  time.Time
    EndTime    time.Time
}

type TransactionState string

const (
    TxPending    TransactionState = "pending"
    TxCommitted  TransactionState = "committed"
    TxRolledBack TransactionState = "rolled_back"
    TxFailed     TransactionState = "failed"
)

type TransactionManager struct {
    current *Transaction
    history []Transaction
}

func (tm *TransactionManager) Begin() *Transaction
func (tm *TransactionManager) AddOperation(op Operation)
func (tm *TransactionManager) Commit() error
func (tm *TransactionManager) Rollback() error
```

### 3. Snapshot System

```go
type Snapshot struct {
    ID        string
    Timestamp time.Time
    Resource  string
    State     interface{} // Captured state before change
}

type SnapshotManager struct {
    snapshots map[string]Snapshot
}

func (sm *SnapshotManager) Capture(resource string, state interface{}) string
func (sm *SnapshotManager) Restore(snapshotID string) error
func (sm *SnapshotManager) List() []Snapshot
```

## Implementation Examples

### UserCreate with Rollback

```go
func (u *UserCreate) RunWithContext(ctx *ExecutionContext) error {
    username := ctx.Config.GetArg("username")
    
    // Start transaction
    tx := ctx.BeginTransaction()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    // Check if user exists
    _, err := ctx.Run(fmt.Sprintf("id %s", username))
    userExists := (err == nil)
    
    if !userExists {
        // Add create operation with rollback
        tx.AddOperation(Operation{
            Type:     "create",
            Resource: fmt.Sprintf("user:%s", username),
            Forward: func() error {
                _, err := ctx.Run(fmt.Sprintf("adduser --disabled-password --gecos '' %s", username))
                return err
            },
            Backward: func() error {
                _, err := ctx.Run(fmt.Sprintf("deluser %s", username))
                return err
            },
        })
    }
    
    // Add sudo access
    tx.AddOperation(Operation{
        Type:     "modify",
        Resource: fmt.Sprintf("user:%s:groups", username),
        Forward: func() error {
            _, err := ctx.Run(fmt.Sprintf("usermod -aG sudo %s", username))
            return err
        },
        Backward: func() error {
            _, err := ctx.Run(fmt.Sprintf("gpasswd -d %s sudo", username))
            return err
        },
    })
    
    // Execute all operations
    if err := tx.Commit(); err != nil {
        ctx.Logger.Error("transaction failed, rolling back", "error", err)
        tx.Rollback()
        return err
    }
    
    return nil
}
```

### SwapCreate with Rollback

```go
func (s *SwapCreate) RunWithContext(ctx *ExecutionContext) error {
    size := ctx.Config.GetArgOr("size", "1")
    
    tx := ctx.BeginTransaction()
    
    // Create swap file
    tx.AddOperation(Operation{
        Type:     "create",
        Resource: "/swapfile",
        Forward: func() error {
            _, err := ctx.Run(fmt.Sprintf("fallocate -l %sG /swapfile", size))
            return err
        },
        Backward: func() error {
            _, err := ctx.Run("rm -f /swapfile")
            return err
        },
    })
    
    // Set permissions
    tx.AddOperation(Operation{
        Type:     "modify",
        Resource: "/swapfile:permissions",
        Forward: func() error {
            _, err := ctx.Run("chmod 600 /swapfile")
            return err
        },
        Backward: func() error {
            return nil // Will be deleted anyway
        },
    })
    
    // Initialize swap
    tx.AddOperation(Operation{
        Type:     "execute",
        Resource: "swap:mkswap",
        Forward: func() error {
            _, err := ctx.Run("mkswap /swapfile")
            return err
        },
        Backward: func() error {
            return nil // No undo needed
        },
    })
    
    // Enable swap
    tx.AddOperation(Operation{
        Type:     "execute",
        Resource: "swap:swapon",
        Forward: func() error {
            _, err := ctx.Run("swapon /swapfile")
            return err
        },
        Backward: func() error {
            _, err := ctx.Run("swapoff /swapfile")
            return err
        },
    })
    
    // Add to fstab
    tx.AddOperation(Operation{
        Type:     "modify",
        Resource: "/etc/fstab",
        Forward: func() error {
            // Capture current fstab
            snapshot, _ := ctx.Run("cat /etc/fstab")
            ctx.CaptureSnapshot("/etc/fstab", snapshot)
            
            _, err := ctx.Run("echo '/swapfile none swap sw 0 0' >> /etc/fstab")
            return err
        },
        Backward: func() error {
            // Restore from snapshot
            return ctx.RestoreSnapshot("/etc/fstab")
        },
    })
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        ctx.Logger.Error("swap creation failed, rolling back", "error", err)
        tx.Rollback()
        return err
    }
    
    return nil
}
```

### File Modification with Snapshot

```go
type FileEdit struct {
    Path    string
    Content string
}

func (f *FileEdit) RunWithContext(ctx *ExecutionContext) error {
    tx := ctx.BeginTransaction()
    
    // Capture current file state
    currentContent, err := ctx.Run(fmt.Sprintf("cat %s", f.Path))
    if err != nil {
        currentContent = "" // File doesn't exist
    }
    
    tx.AddOperation(Operation{
        Type:     "modify",
        Resource: f.Path,
        Forward: func() error {
            // Write new content
            cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", f.Path, f.Content)
            _, err := ctx.Run(cmd)
            return err
        },
        Backward: func() error {
            if currentContent == "" {
                // File didn't exist, delete it
                _, err := ctx.Run(fmt.Sprintf("rm -f %s", f.Path))
                return err
            }
            // Restore original content
            cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", f.Path, currentContent)
            _, err := ctx.Run(cmd)
            return err
        },
    })
    
    if err := tx.Commit(); err != nil {
        tx.Rollback()
        return err
    }
    
    return nil
}
```

## Manual Rollback

### CLI Usage

```bash
# Run with automatic rollback on failure
ork run user-create --host server.example.com --arg username=john --rollback-on-error

# Show transaction history
ork transactions list --host server.example.com

# Manually rollback last transaction
ork transactions rollback --host server.example.com --last

# Rollback specific transaction
ork transactions rollback --host server.example.com --id tx-abc123
```

### Programmatic Usage

```go
// Run with automatic rollback
ctx := NewExecutionContext(cfg, logger)
ctx.EnableAutoRollback(true)

playbook := playbooks.NewUserCreate()
err := playbook.RunWithContext(ctx)
if err != nil {
    // Rollback already happened automatically
    log.Printf("Playbook failed and was rolled back: %v", err)
}

// Manual rollback
tx := ctx.GetLastTransaction()
if err := tx.Rollback(); err != nil {
    log.Printf("Rollback failed: %v", err)
}
```

## Transaction History

### Storage Format (JSON)

```json
{
  "id": "tx-abc123",
  "playbook": "user-create",
  "host": "server.example.com",
  "start_time": "2026-04-12T10:30:00Z",
  "end_time": "2026-04-12T10:30:05Z",
  "state": "committed",
  "operations": [
    {
      "type": "create",
      "resource": "user:john",
      "state": "applied",
      "timestamp": "2026-04-12T10:30:01Z"
    },
    {
      "type": "modify",
      "resource": "user:john:groups",
      "state": "applied",
      "timestamp": "2026-04-12T10:30:02Z"
    }
  ]
}
```

### Storage Location

```
~/.ork/transactions/
├── server.example.com/
│   ├── tx-abc123.json
│   ├── tx-def456.json
│   └── snapshots/
│       ├── snapshot-001.tar.gz
│       └── snapshot-002.tar.gz
```

## Advanced Features

### Checkpoint System

```go
// Create checkpoint before major changes
checkpoint := ctx.CreateCheckpoint("before-upgrade")

// Run risky operations
err := aptUpgrade.Run(cfg)

if err != nil {
    // Restore to checkpoint
    ctx.RestoreCheckpoint(checkpoint)
}
```

### Partial Rollback

```go
// Rollback only specific operations
tx.RollbackOperation("user:john:groups") // Remove from sudo, keep user

// Rollback from specific point
tx.RollbackFrom(operationIndex)
```

### Dry-Run Rollback

```go
// Preview what rollback would do
actions := tx.DryRunRollback()
for _, action := range actions {
    fmt.Printf("Would execute: %s\n", action.Description)
}
```

## Safety Considerations

### Rollback Validation

```go
type RollbackValidator struct{}

func (v *RollbackValidator) CanRollback(op Operation) error {
    // Check if rollback is safe
    switch op.Type {
    case "delete":
        // Can't rollback deletion if we didn't capture state
        if op.Snapshot == nil {
            return fmt.Errorf("cannot rollback deletion without snapshot")
        }
    case "execute":
        // Some commands can't be undone
        if !op.Reversible {
            return fmt.Errorf("operation is not reversible")
        }
    }
    return nil
}
```

### Rollback Timeout

```go
// Set timeout for rollback operations
tx.SetRollbackTimeout(2 * time.Minute)

// Rollback with context
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
tx.RollbackWithContext(ctx)
```

## Implementation Plan

### Phase 1: Core Framework
- Implement Operation and Transaction types
- Add basic rollback support
- Update 1-2 playbooks as examples

### Phase 2: Snapshot System
- Implement snapshot capture/restore
- Add file state tracking
- Add transaction history storage

### Phase 3: Advanced Features
- Add checkpoint system
- Implement partial rollback
- Add rollback validation

### Phase 4: CLI Integration
- Add transaction management commands
- Add rollback flags
- Add transaction history viewer

## Benefits

- **Safety**: Undo mistakes automatically
- **Confidence**: Experiment without fear
- **Reliability**: Atomic operations
- **Debugging**: Understand what changed
- **Compliance**: Audit trail of changes

## Limitations

- Some operations are inherently irreversible (e.g., data deletion without backup)
- Rollback may fail if system state changed externally
- Snapshots consume disk space
- Complex operations may have incomplete rollback

## Success Metrics

- 80% of playbooks support rollback
- Rollback success rate >95%
- Zero data loss from rollback operations
- Clear documentation of non-reversible operations

## Open Questions

1. Should rollback be opt-in or opt-out?
2. How long to keep transaction history?
3. Should we support distributed transactions (multiple hosts)?
4. How to handle rollback of database operations?
