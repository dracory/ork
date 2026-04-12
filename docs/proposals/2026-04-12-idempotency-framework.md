# Proposal: Idempotency Framework

**Date:** 2026-04-12  
**Status:** Implemented  
**Author:** System Review

> **Note:** All core playbooks now implement `CheckablePlaybook`. Use `playbook.Execute()` for automatic idempotency handling.

## What's Implemented

- `Result` type with `Changed`, `Message`, `Details`, `Error` fields
- `CheckablePlaybook` interface with `Check()` and `RunWithResult()` methods
- `CheckExists()` and `EnsureState()` helper functions
- `Execute()` wrapper that auto-detects `CheckablePlaybook`
- All 7 core playbooks implement `CheckablePlaybook`:
  - `Ping`, `AptUpdate`, `AptUpgrade`, `AptStatus`
  - `SwapCreate`, `SwapDelete`, `SwapStatus`, `Reboot`

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

## Implementation Complete

### Core Types (Done)
- ✅ `Result` type added to `playbook` package
- ✅ `CheckablePlaybook` interface added
- ✅ `CheckExists()` and `EnsureState()` helpers implemented
- ✅ `Execute()` wrapper function added

### Playbook Migration (Done)
- ✅ `Ping` - Read-only, always returns `Changed: false`
- ✅ `AptUpdate` - Cache refresh, always returns `Changed: true`
- ✅ `AptUpgrade` - Checks for upgradable packages, `Changed` only when upgrades installed
- ✅ `AptStatus` - Read-only, always returns `Changed: false`
- ✅ `SwapCreate` - Checks if swap exists, `Changed` only if created
- ✅ `SwapDelete` - Checks if swap exists, `Changed` only if removed
- ✅ `SwapStatus` - Read-only, always returns `Changed: false`
- ✅ `Reboot` - Explicit action, always returns `Changed: true`
- ✅ All playbooks maintain backward compatibility with `Run()` delegating to `RunWithResult()`

## Benefits

- **Safety**: Running playbooks multiple times is safe
- **Efficiency**: Skip unnecessary operations
- **Visibility**: Know what changed vs what was already correct
- **Ansible-like**: Familiar behavior for Ansible users
- **Debugging**: Easier to understand what happened

## Backward Compatibility

All playbooks maintain backward compatibility:

```go
// Run() delegates to RunWithResult() for idempotency support
func (p *Ping) Run(cfg Config) error {
    result := p.RunWithResult(cfg)
    return result.Error
}
```

Legacy playbooks (without `CheckablePlaybook`) are handled by `Execute()`:

```go
func Execute(pb Playbook, cfg Config) Result {
    if checkable, ok := pb.(CheckablePlaybook); ok {
        return checkable.RunWithResult(cfg)
    }
    // Fallback for legacy playbooks
    err := pb.Run(cfg)
    return Result{
        Changed: true, // Assume changed if using old interface
        Message: "Executed playbook (idempotency not available)",
        Error:   err,
    }
}
```

## Usage Examples

### Basic Usage

```go
// Using Execute() - automatically handles idempotency
result := playbook.Execute(playbooks.NewAptUpgrade(), cfg)
if result.Error != nil {
    log.Fatal(result.Error)
}
log.Printf("Changed: %v - %s", result.Changed, result.Message)
// Output: Changed: false - All packages are up to date
```

### Check Before Run

```go
pb := playbooks.NewSwapCreate()
if checkable, ok := pb.(playbook.CheckablePlaybook); ok {
    needsCreate, _ := checkable.Check(cfg)
    if !needsCreate {
        log.Println("Swap already exists, skipping...")
        return
    }
}
```

### Direct RunWithResult

```go
pb := playbooks.NewAptUpgrade()
result := pb.RunWithResult(cfg)
if result.Changed {
    log.Printf("Upgraded packages: %s", result.Details["output"])
}
```
