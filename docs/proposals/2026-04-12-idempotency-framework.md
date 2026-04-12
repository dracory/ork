# Proposal: Idempotency Framework

**Date:** 2026-04-12  
**Status:** Not Implemented  
**Author:** System Review

## Problem Statement

Playbooks currently have ad-hoc idempotency checks (e.g., checking if swap exists before creating). We need:

- Standard `Result` type reporting `Changed` status
- `CheckablePlaybook` interface for pre-flight checks
- Helper functions for common patterns

Ansible's strength is idempotent operations - running the same playbook multiple times should be safe and only make necessary changes.

## Implementation

### 1. Add Result Type

```go
package playbook

type Result struct {
    Changed bool              // Whether any changes were made
    Message string            // Human-readable result
    Details map[string]string // Additional information
    Error   error             // Any error that occurred
}
```

### 2. Add CheckablePlaybook Interface

```go
type CheckablePlaybook interface {
    Playbook
    Check(config Config) (bool, error) // Returns true if changes needed
    RunWithResult(config Config) Result
}
```

### 3. Idempotency Helpers

```go
// CheckExists runs a command and returns true if it succeeds
func CheckExists(client *ssh.Client, checkCmd string) bool

// EnsureState ensures a desired state, returns whether changes were made
func EnsureState(client *ssh.Client, checkCmd, applyCmd string) (bool, error)
```

## Example Implementation

### AptUpgrade with Idempotency

```go
func (a *AptUpgrade) Check(cfg config.Config) (bool, error) {
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt list --upgradable 2>/dev/null | tail -n +2 | wc -l")
    count := strings.TrimSpace(output)
    return count != "0" && count != "", err
}

func (a *AptUpgrade) RunWithResult(cfg config.Config) Result {
    needsUpgrade, err := a.Check(cfg)
    if err != nil {
        return Result{Error: err}
    }
    
    if !needsUpgrade {
        return Result{Changed: false, Message: "All packages up to date"}
    }
    
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt-get upgrade -y")
    
    if err != nil {
        return Result{Error: err}
    }
    return Result{Changed: true, Message: "Packages upgraded", Details: map[string]string{"output": output}}
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
- Add `Result` type to `playbook` package
- Add `CheckablePlaybook` interface
- Create `CheckExists()` and `EnsureState()` helpers

### Phase 2: Playbook Migration
- Update `AptUpgrade`, `SwapCreate`, `UserCreate` to use new interfaces
- Maintain backward compatibility with existing `Run()`

### Phase 3: Documentation
- Document idempotency patterns
- Add usage examples

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
