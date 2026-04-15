# Privilege Escalation (Become)

**Status:** Proposed
**Created:** 2026-04-15
**Author:** Kiro AI

## Problem Statement

Ork currently connects as a single user and runs all commands with that user's privileges. This violates the principle of least privilege and prevents running commands as different users (e.g., connect as `deploy`, run privileged commands as `root`).

## Proposal

Add simple privilege escalation: if a "become user" is set, wrap commands with `sudo -u <user>`.

## Core Design

### BecomeInterface

```go
// BecomeInterface defines privilege escalation
type BecomeInterface interface {
    SetBecomeUser(user string) BecomeInterface
    GetBecomeUser() string
}
```

### Implementation

```go
// BaseBecome provides default implementation
type BaseBecome struct {
    becomeUser string
}

func (b *BaseBecome) SetBecomeUser(user string) BecomeInterface {
    b.becomeUser = user
    return b
}

func (b *BaseBecome) GetBecomeUser() string {
    return b.becomeUser
}
```

### Integration

Embed in `RunnerInterface` and `RunnableInterface`:

```go
type RunnerInterface interface {
    // ... existing methods ...
    BecomeInterface
}

type RunnableInterface interface {
    // ... existing methods ...
    BecomeInterface
}
```

### Command Wrapping

```go
// In ssh.Run()
func Run(cfg NodeConfig, cmd Command) (string, error) {
    if cfg.BecomeUser != "" {
        cmd.Command = fmt.Sprintf("sudo -u %s %s", cfg.BecomeUser, cmd.Command)
    }
    return executeCommand(cfg, cmd)
}
```

## Usage

```go
// Connect as deploy, run as root
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecomeUser("root")

node.RunCommand("apt-get update")  // Runs: sudo -u root apt-get update

// Run as postgres
node.SetBecomeUser("postgres").
    RunCommand("psql -c 'SELECT version()'")

// Run as deploy (no escalation)
node.SetBecomeUser("").
    RunCommand("ls ~")
```

## Precedence

Lower levels override higher levels:
1. Inventory
2. Group  
3. Node
4. Skill (highest)

```go
// Inventory: become root
inv.SetBecomeUser("root")

// Node: override to postgres
node.SetBecomeUser("postgres")

// Skill: override to app
skill.SetBecomeUser("app")  // Wins
```

## Security

**Configure passwordless sudo:**
```bash
# /etc/sudoers
deploy ALL=(ALL) NOPASSWD: ALL
```

**Never hardcode passwords.** Use vault or prompts if passwords are required.

## Implementation

1. Add `BecomeInterface` to types
2. Embed in `RunnerInterface` and `RunnableInterface`
3. Add `BecomeUser` field to `NodeConfig`
4. Wrap commands in `ssh.Run()`
5. Copy become user from interface to config in `GetNodeConfig()`

## Future Enhancements

- Support `su`, `doas` (if needed)
- Custom flags (if needed)
- Password support (if needed)

Start simple. Add complexity only when actually needed.

## See Also

- [Privilege Escalation (Expanded)](2026-04-15-privilege-escalation-expanded.md) - Detailed design with multiple methods, passwords, flags, etc.
