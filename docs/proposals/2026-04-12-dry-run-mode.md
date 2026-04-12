# Proposal: Dry-Run Mode

**Date:** 2026-04-12  
**Status:** Partially Implemented  
**Author:** System Review

> **Note:** `DryRun` field exists in `PlaybookOptions`. Remaining: Playbook implementations to check `IsDryRun()` and skip changes, plus optional `DryRunnable` interface for detailed previews.

## Problem Statement

Users need to preview what changes a playbook will make before executing it. This is critical for:

- Production environments
- Learning what a playbook does
- Debugging playbook logic
- Compliance and audit requirements

Ansible's `--check` mode is one of its most valuable features.

## Proposed Solution

### 1. Add DryRun Interface

```go
type DryRunnable interface {
    Playbook
    DryRun(config Config) ([]Action, error)
}

type Action struct {
    Type        string            // "create", "modify", "delete", "execute"
    Resource    string            // What's being changed
    Description string            // Human-readable description
    Command     string            // Actual command that would run
    Details     map[string]string // Additional context
}
```

### 2. Add DryRun Flag to Config

```go
type Config struct {
    // ... existing fields ...
    
    DryRun bool // If true, don't make actual changes
}
```

### 3. Create DryRun Executor

```go
type DryRunExecutor struct {
    actions []Action
}

func (e *DryRunExecutor) WouldRun(cmd string, description string) {
    e.actions = append(e.actions, Action{
        Type:        "execute",
        Resource:    "command",
        Description: description,
        Command:     cmd,
    })
}

func (e *DryRunExecutor) WouldCreate(resource, description string) {
    e.actions = append(e.actions, Action{
        Type:        "create",
        Resource:    resource,
        Description: description,
    })
}

func (e *DryRunExecutor) Actions() []Action {
    return e.actions
}
```

## Implementation Examples

### AptUpgrade with DryRun

```go
func (a *AptUpgrade) DryRun(cfg config.Config) ([]Action, error) {
    // Check what would be upgraded
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "apt list --upgradable 2>/dev/null | tail -n +2")
    if err != nil {
        return nil, err
    }
    
    packages := strings.Split(strings.TrimSpace(output), "\n")
    if len(packages) == 0 || packages[0] == "" {
        return []Action{{
            Type:        "skip",
            Resource:    "packages",
            Description: "All packages are up to date",
        }}, nil
    }
    
    actions := []Action{{
        Type:        "execute",
        Resource:    "apt-get",
        Description: fmt.Sprintf("Would upgrade %d packages", len(packages)),
        Command:     "apt-get upgrade -y",
        Details: map[string]string{
            "packages": strings.Join(packages, ", "),
        },
    }}
    
    return actions, nil
}
```

### SwapCreate with DryRun

```go
func (s *SwapCreate) DryRun(cfg config.Config) ([]Action, error) {
    sizeStr := cfg.GetArgOr("size", "1")
    size, _ := strconv.Atoi(sizeStr)
    
    // Check if swap exists
    output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        "swapon --show=NAME --noheadings")
    
    if strings.TrimSpace(output) != "" {
        return []Action{{
            Type:        "skip",
            Resource:    "/swapfile",
            Description: "Swap file already exists",
        }}, nil
    }
    
    return []Action{
        {
            Type:        "create",
            Resource:    "/swapfile",
            Description: fmt.Sprintf("Would create %dGB swap file", size),
            Command:     fmt.Sprintf("fallocate -l %dG /swapfile", size),
        },
        {
            Type:        "modify",
            Resource:    "/swapfile",
            Description: "Would set permissions to 600",
            Command:     "chmod 600 /swapfile",
        },
        {
            Type:        "execute",
            Resource:    "swap",
            Description: "Would initialize and enable swap",
            Command:     "mkswap /swapfile && swapon /swapfile",
        },
        {
            Type:        "modify",
            Resource:    "/etc/fstab",
            Description: "Would add swap to fstab for persistence",
            Command:     "echo '/swapfile none swap sw 0 0' >> /etc/fstab",
        },
    }, nil
}
```

### UserCreate with DryRun

```go
func (u *UserCreate) DryRun(cfg config.Config) ([]Action, error) {
    username := cfg.GetArg("username")
    if username == "" {
        return nil, fmt.Errorf("username is required")
    }
    
    // Check if user exists
    _, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
        fmt.Sprintf("id %s", username))
    
    if err == nil {
        return []Action{{
            Type:        "skip",
            Resource:    username,
            Description: fmt.Sprintf("User '%s' already exists", username),
        }}, nil
    }
    
    return []Action{
        {
            Type:        "create",
            Resource:    username,
            Description: fmt.Sprintf("Would create user '%s'", username),
            Command:     fmt.Sprintf("adduser --disabled-password --gecos '' %s", username),
        },
        {
            Type:        "modify",
            Resource:    username,
            Description: fmt.Sprintf("Would add '%s' to sudo group", username),
            Command:     fmt.Sprintf("usermod -aG sudo %s", username),
        },
    }, nil
}
```

## Usage Examples

### Programmatic Usage

```go
cfg := config.Config{
    SSHHost: "server.example.com",
    SSHPort: "22",
    SSHKey:  "id_rsa",
    RootUser: "root",
    DryRun:  true,
}

playbook := playbooks.NewAptUpgrade()

if dryRunnable, ok := playbook.(DryRunnable); ok {
    actions, err := dryRunnable.DryRun(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Would perform the following actions:")
    for _, action := range actions {
        fmt.Printf("  [%s] %s: %s\n", action.Type, action.Resource, action.Description)
        if action.Command != "" {
            fmt.Printf("    Command: %s\n", action.Command)
        }
    }
}
```

### CLI Usage (if CLI exists)

```bash
# Preview changes
ork run --host server.example.com --playbook apt-upgrade --dry-run

# Output:
# Would perform the following actions:
#   [execute] apt-get: Would upgrade 23 packages
#     Command: apt-get upgrade -y
#     Packages: nginx, postgresql, redis-server, ...
```

## Implementation Plan

### Phase 1: Core Framework
- Add `DryRunnable` interface to `playbook` package
- Add `Action` type
- Add `DryRun` bool field to `config.Config`

### Phase 2: Playbook Implementation
- Add `DryRun()` to `AptUpgrade`, `SwapCreate`, `UserCreate`
- Test accuracy vs actual execution

### Phase 3: CLI Integration
- Add `--dry-run` flag to CLI tool

## Benefits

- **Safety**: Preview changes before applying
- **Learning**: Understand what playbooks do
- **Debugging**: Verify playbook logic without side effects
- **Compliance**: Audit trail of planned changes
- **Confidence**: Reduce fear of running automation

## Challenges & Solutions

**Challenge:** Some operations can't be accurately predicted  
**Solution:** Mark actions as "estimated" vs "certain"

**Challenge:** Dry-run might need to run some read-only commands  
**Solution:** Allow read-only SSH commands during dry-run

**Challenge:** State might change between dry-run and actual run  
**Solution:** Document this limitation, recommend immediate execution after dry-run

## Success Metrics

- All core playbooks implement DryRun
- Dry-run predictions match actual execution >95% of the time
- User feedback indicates increased confidence

## Open Questions

1. Should dry-run execute read-only commands or use cached state?
2. How to handle conditional logic that depends on command output?
3. Should we support "what-if" scenarios with hypothetical state?
4. How to visualize complex multi-step operations?
