# Proposal: Idempotency Framework

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

Some playbooks lack proper idempotency checks, meaning they may perform unnecessary operations or fail when run multiple times:

- `AptUpgrade` always runs even if no updates available
- Operations don't report whether changes were made
- No standard pattern for "check before change"

Ansible's strength is idempotent operations - running the same playbook multiple times should be safe and only make necessary changes.

## Proposed Solution

### 1. Add Result Type

```go
type Result struct {
    Changed bool              // Whether any changes were made
    Message string            // Human-readable result
    Details map[string]string // Additional information
    Error   error             // Any error that occurred
}
```

### 2. Update Playbook Interface

```go
type Playbook interface {
    Name() string
    Description() string
    Run(config Config) error
    RunWithResult(config Config) Result // New method
}

// Optional interface for check mode
type CheckablePlaybook interface {
    Playbook
    Check(config Config) (bool, error) // Returns true if changes needed
}
```

### 3. Standard Idempotency Helpers

```go
package playbook

// CheckExists runs a command and returns true if it succeeds
func CheckExists(client *ssh.Client, checkCmd string) bool

// EnsureState ensures a desired state, returns whether changes were made
func EnsureState(client *ssh.Client, checkCmd, applyCmd string) (bool, error)

// CompareState checks if current state matches desired state
func CompareState(current, desired string) bool
```

## Implementation Examples

### AptUpgrade with Idempotency

```go
func (a *AptUpgrade) Check(cfg config.Config) (bool, error) {
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt list --upgradable 2>/dev/null | tail -n +2 | wc -l")
    if err != nil {
        return false, err
    }
    
    count := strings.TrimSpace(output)
    return count != "0" && count != "", nil
}

func (a *AptUpgrade) RunWithResult(cfg config.Config) Result {
    needsUpgrade, err := a.Check(cfg)
    if err != nil {
        return Result{Error: err}
    }
    
    if !needsUpgrade {
        return Result{
            Changed: false,
            Message: "All packages are up to date",
        }
    }
    
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt-get upgrade -y")
    if err != nil {
        return Result{Error: err}
    }
    
    return Result{
        Changed: true,
        Message: "Packages upgraded successfully",
        Details: map[string]string{"output": output},
    }
}
```

### SwapCreate with Better Idempotency

```go
func (s *SwapCreate) RunWithResult(cfg config.Config) Result {
    // Check current state
    output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "swapon --show=NAME --noheadings")
    
    if strings.Contains(output, "/swapfile") {
        return Result{
            Changed: false,
            Message: "Swap file already exists",
            Details: map[string]string{"current": strings.TrimSpace(output)},
        }
    }
    
    // Create swap...
    err := s.createSwap(cfg)
    if err != nil {
        return Result{Error: err}
    }
    
    return Result{
        Changed: true,
        Message: fmt.Sprintf("Created %s swap file", cfg.GetArgOr("size", "1")),
    }
}
```

### UserCreate with State Verification

```go
func (u *UserCreate) RunWithResult(cfg config.Config) Result {
    username := cfg.GetArg("username")
    
    // Check if user exists
    _, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        fmt.Sprintf("id %s", username))
    
    if err == nil {
        // User exists, check sudo access
        output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
            fmt.Sprintf("groups %s", username))
        
        hasSudo := strings.Contains(output, "sudo")
        
        if hasSudo {
            return Result{
                Changed: false,
                Message: fmt.Sprintf("User '%s' already exists with sudo access", username),
            }
        }
        
        // Add sudo access
        ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
            fmt.Sprintf("usermod -aG sudo %s", username))
        
        return Result{
            Changed: true,
            Message: fmt.Sprintf("Added sudo access to existing user '%s'", username),
        }
    }
    
    // Create user...
    return Result{
        Changed: true,
        Message: fmt.Sprintf("Created user '%s' with sudo access", username),
    }
}
```

## Implementation Plan

### Phase 1: Core Types
- Add `Result` type
- Add `CheckablePlaybook` interface
- Create helper functions

### Phase 2: Migrate Existing Playbooks
- Update all playbooks to implement `RunWithResult`
- Add idempotency checks where missing
- Maintain backward compatibility with `Run()`

### Phase 3: Documentation & Testing
- Document idempotency patterns
- Add tests verifying idempotent behavior
- Create examples

## Benefits

- **Safety**: Running playbooks multiple times is safe
- **Efficiency**: Skip unnecessary operations
- **Visibility**: Know what changed vs what was already correct
- **Ansible-like**: Familiar behavior for Ansible users
- **Debugging**: Easier to understand what happened

## Backward Compatibility

Keep existing `Run()` method, add new `RunWithResult()`:

```go
// Default implementation for backward compatibility
func (p *BasePlaybook) RunWithResult(cfg Config) Result {
    err := p.Run(cfg)
    return Result{
        Changed: true, // Assume changed if using old interface
        Error:   err,
    }
}
```

## Success Metrics

- All core playbooks implement idempotency checks
- Running playbooks twice shows "Changed: false" on second run
- Zero false positives (claiming no change when change occurred)

## Open Questions

1. Should `Run()` be deprecated in favor of `RunWithResult()`?
2. How to handle partial failures (some changes succeeded, some failed)?
3. Should we track detailed change logs (what files modified, etc.)?
