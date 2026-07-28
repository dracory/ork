# Privilege Escalation (Become)

**Status:** Completed
**Created:** 2026-04-15
**Completed:** 2026-05-05
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

Embedded in `RunnerInterface` and `RunnableInterface`:

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
    // Command-level BecomeUser takes precedence over config-level
    becomeUser := cmd.BecomeUser
    if becomeUser == "" {
        becomeUser = cfg.BecomeUser
    }

    commandToRun := cmd.Command
    if becomeUser != "" {
        commandToRun = fmt.Sprintf("sudo -u %s %s", becomeUser, cmd.Command)
    }
    // ...
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

## Implementation Status

All steps completed:

- [x] `BecomeInterface` defined in `types/become_interface.go` with `SetBecomeUser` / `GetBecomeUser`
- [x] `BaseBecome` struct provides default implementation, embedded in `types/BaseSkill` and `types/BasePlaybook`
- [x] `BecomeInterface` embedded in `types/RunnerInterface` and `types/RunnableInterface`
- [x] `BecomeUser` field added to `types/NodeConfig`
- [x] `WithBecomeUser` fluent method added to `types/NodeConfig`
- [x] Command wrapping implemented in `ssh/functions.go` — command-level `BecomeUser` takes precedence over config-level
- [x] `BecomeUser` propagated from node config to skill in `node_implementation.go` (`Run`, `RunByID`)
- [x] `SetBecomeUser` / `GetBecomeUser` implemented on `nodeImplementation`
- [x] `WithBecomeUser` fluent helper added to `CommandInterface` and `command_implementation.go`
- [x] `BecomeUser` field added to `types/Command` for per-command override
- [x] Tests added: `TestRun_CommandBecomeUser`, `TestRun_ConfigBecomeUser`, `TestRun_CombinedChdirAndBecomeUser` in `ssh/ssh_test.go`
- [x] Tests added: `TestNewSkill_WithBecomeUser` in `skill_test.go`

## Notes on Scope

The implementation follows the simple design from this proposal rather than the expanded design in [Privilege Escalation (Expanded)](2026-04-15-privilege-escalation-expanded.md). Specifically:

- Only `sudo -u <user>` is supported (no `su`, `doas`, etc.)
- No become password support
- No `SetBecome(bool)` enable/disable flag — an empty `BecomeUser` means no escalation
- No custom flags

These were intentional omissions per the "start simple" principle. See the expanded proposal if those features are needed.

## Future Enhancements

- Support `su`, `doas` (if needed)
- Custom flags (if needed)
- Password support (if needed)

## See Also

- [Privilege Escalation (Expanded)](2026-04-15-privilege-escalation-expanded.md) - Detailed design with multiple methods, passwords, flags, etc.
