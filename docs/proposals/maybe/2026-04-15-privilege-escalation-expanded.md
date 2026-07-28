# Privilege Escalation (Become) — Expanded

**Status:** Partially Implemented, Out of Scope at the Moment
**Created:** 2026-04-15
**Updated:** 2026-05-05
**Author:** Kiro AI

## Problem Statement

Ork currently connects to remote servers as a single user (typically `root`) and executes all commands with that user's privileges. This creates several problems:

1. **Security Risk**: Running everything as root violates the principle of least privilege
2. **No Privilege Separation**: Cannot connect as a regular user and escalate only when needed
3. **Audit Trail**: All actions appear to be performed by the same user
4. **Compliance**: Many organizations require non-root SSH access with sudo for specific operations
5. **Limited Flexibility**: Cannot run different commands as different users in the same session

Ansible solves this with the `become` directive, which allows:
- Connecting as a regular user
- Escalating privileges for specific tasks using sudo, su, or other methods
- Running commands as different users
- Fine-grained control over privilege escalation

## Implementation Status

The [simple proposal](2026-04-15-privilege-escalation.md) was implemented. This expanded proposal covers the full feature set, most of which is **not yet implemented**.

### What Is Implemented

- [x] `BecomeInterface` with `SetBecomeUser` / `GetBecomeUser` only (`types/become_interface.go`)
- [x] `BaseBecome` struct embedded in `BaseSkill` and `BasePlaybook`
- [x] `BecomeInterface` embedded in `RunnableInterface` (`types/runnable_interface.go`)
- [x] `BecomeInterface` embedded in the ork-level `RunnerInterface` (`runner_interface.go`)
- [x] `BecomeUser` field in `NodeConfig` with `WithBecomeUser` fluent method
- [x] `BecomeUser` field in `types.Command` for per-command override
- [x] Command wrapping with `sudo -u <user>` in `ssh/functions.go`
- [x] Command-level `BecomeUser` takes precedence over config-level
- [x] `SetBecomeUser` / `GetBecomeUser` on `nodeImplementation`, `groupImplementation`, `inventoryImplementation`
- [x] Propagation: inventory → group → node when `SetBecomeUser` is called
- [x] Propagation: node → skill in `node_implementation.go` (`Run`, `RunByID`) — skill's own value wins if set
- [x] `WithBecomeUser` fluent helper on `CommandInterface`

### What Is NOT Implemented (This Proposal)

- [ ] `SetBecome(bool)` / `GetBecomeEnabled()` — enable/disable flag (empty user means disabled today)
- [ ] `SetBecomeMethod` / `GetBecomeMethod` — only `sudo -u` is supported
- [ ] `SetBecomePassword` / `GetBecomePassword` — no password support
- [ ] `SetBecomeFlags` / `GetBecomeFlags` — no custom flags
- [ ] `CommandWrapper` with multiple method implementations (`su`, `doas`, `pbrun`, `pfexec`, `runas`)
- [ ] `resolveBecomeSettings()` precedence resolution function
- [ ] `BecomeInterface` embedded in `types.RunnerInterface` (only in the ork-level wrapper)
- [ ] `BecomeInterface` exposed on `InventoryInterface` and `GroupInterface` interfaces (implementations have it, interfaces don't declare it)
- [ ] Per-execution become via `RunnableOptions`
- [ ] Become caching, detection, validation, profiles, hooks, metrics

---

## Proposal (Remaining Work)

### Simplified BecomeInterface (Current)

```go
// What exists today in types/become_interface.go
type BecomeInterface interface {
    SetBecomeUser(user string) BecomeInterface
    GetBecomeUser() string
}
```

### Full BecomeInterface (This Proposal)

```go
// BecomeInterface defines privilege escalation methods
type BecomeInterface interface {
    // SetBecome enables/disables privilege escalation
    SetBecome(enabled bool) BecomeInterface

    // GetBecomeEnabled returns whether privilege escalation is enabled
    GetBecomeEnabled() bool

    // SetBecomeUser sets the user to become (default: root)
    SetBecomeUser(user string) BecomeInterface

    // GetBecomeUser returns the user to become
    GetBecomeUser() string

    // SetBecomeMethod sets the escalation method (default: sudo)
    // Supported: sudo, su, doas, pbrun, pfexec, runas
    SetBecomeMethod(method string) BecomeInterface

    // GetBecomeMethod returns the escalation method
    GetBecomeMethod() string

    // SetBecomePassword sets the password for escalation
    SetBecomePassword(password string) BecomeInterface

    // GetBecomePassword returns the password for escalation
    GetBecomePassword() string

    // SetBecomeFlags sets additional flags for the become method
    SetBecomeFlags(flags string) BecomeInterface

    // GetBecomeFlags returns additional flags for the become method
    GetBecomeFlags() string
}
```

**Migration note:** The current `BecomeInterface` only has `SetBecomeUser`/`GetBecomeUser`. Expanding it is a breaking change for any code that implements the interface. The `BaseBecome` struct would need to be updated and all implementations recompiled.

### BaseBecome Implementation

```go
type BaseBecome struct {
    enabled  bool
    method   string
    user     string
    password string
    flags    string
}

func NewBaseBecome() *BaseBecome {
    return &BaseBecome{
        method: "sudo",
        user:   "root",
        flags:  "-n", // Non-interactive by default
    }
}
```

### Become Methods

| Method | Description | Example Command |
|--------|-------------|-----------------|
| `sudo` | Use sudo (default, **implemented**) | `sudo -u root command` |
| `su` | Use su | `su - root -c 'command'` |
| `doas` | Use doas (OpenBSD) | `doas -u root command` |
| `pbrun` | Use PowerBroker | `pbrun -u root command` |
| `pfexec` | Use pfexec (Solaris) | `pfexec -u root command` |
| `runas` | Use runas (Windows) | `runas /user:Administrator command` |

### CommandWrapper

```go
type CommandWrapper struct {
    enabled  bool
    method   string
    user     string
    password string
    flags    string
}

func (w *CommandWrapper) Wrap(cmd string) string {
    if !w.enabled {
        return cmd
    }
    switch w.method {
    case "sudo":
        return w.wrapSudo(cmd)
    case "su":
        return w.wrapSu(cmd)
    case "doas":
        return w.wrapDoas(cmd)
    default:
        return w.wrapSudo(cmd)
    }
}

func (w *CommandWrapper) wrapSudo(cmd string) string {
    flags := w.flags
    if flags == "" {
        flags = "-n"
    }
    user := w.user
    if user == "" {
        user = "root"
    }
    if w.password != "" {
        return fmt.Sprintf("echo '%s' | sudo -S %s -u %s %s", w.password, flags, user, cmd)
    }
    return fmt.Sprintf("sudo %s -u %s %s", flags, user, cmd)
}

func (w *CommandWrapper) wrapSu(cmd string) string {
    user := w.user
    if user == "" {
        user = "root"
    }
    escapedCmd := strings.ReplaceAll(cmd, "'", "'\\''")
    return fmt.Sprintf("su - %s -c '%s'", user, escapedCmd)
}

func (w *CommandWrapper) wrapDoas(cmd string) string {
    user := w.user
    if user == "" {
        user = "root"
    }
    return fmt.Sprintf("doas -u %s %s", user, cmd)
}
```

### Interface Exposure Gap

Today `InventoryInterface` and `GroupInterface` do not declare `BecomeInterface` in their interface definitions, even though the implementations support it. This means callers must type-assert to access become methods:

```go
// Current: requires type assertion
inv := ork.NewInventory()
inv.(types.BecomeInterface).SetBecomeUser("root")  // awkward

// Desired: declared on the interface
inv.SetBecomeUser("root")  // clean
```

Fix: embed `types.BecomeInterface` in `InventoryInterface` and `GroupInterface`.

### NodeConfig Extension

```go
// NodeConfig additions needed for full become support
type NodeConfig struct {
    // ... existing fields ...

    BecomeUser     string  // already exists
    BecomeEnabled  bool    // not yet added
    BecomeMethod   string  // not yet added
    BecomePassword string  // not yet added
    BecomeFlags    string  // not yet added
}
```

### Precedence Hierarchy

Lowest to highest (higher wins):

1. Inventory level
2. Group level
3. Node level
4. Skill/Runnable level (highest)

The current implementation handles inventory→group→node propagation via `propagateBecomeUser()`. Skill-level override is handled in `node_implementation.Run()` — the skill's `BecomeUser` wins if non-empty. A formal `resolveBecomeSettings()` function is not yet implemented.

## Usage (Current — Partial)

```go
// Connect as deploy, run as root (works today)
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecomeUser("root")

node.RunCommand("apt-get update")  // Runs: sudo -u root apt-get update

// Per-command become (works today via types.Command directly in skills)
ssh.Run(cfg, types.Command{
    Command:    "psql -l",
    BecomeUser: "postgres",  // overrides cfg.BecomeUser
})
```

## Usage (Proposed — Not Yet Implemented)

```go
// Enable/disable flag
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root").
    SetBecomeMethod("sudo").
    SetBecomePassword("secret").
    SetBecomeFlags("-n -H")

// Use su instead of sudo
node.SetBecomeMethod("su")
node.RunCommand("apt-get update")  // Runs: su - root -c 'apt-get update'

// Use doas (OpenBSD)
node.SetBecomeMethod("doas")
node.RunCommand("apt-get update")  // Runs: doas -u root apt-get update

// Temporarily disable become
node.SetBecome(false).RunCommand("whoami")  // Runs as deploy
node.SetBecome(true).RunCommand("apt-get update")  // Runs as root
```

## Security Considerations

### Password Storage

1. **Never hardcode passwords** in source code
2. **Use vault** for password storage
3. **Prompt at runtime** when possible

```go
// Good: Load from vault
secrets, _ := ork.VaultFileToKeysWithPrompt(".env.vault")
node.SetBecomePassword(secrets["SUDO_PASSWORD"])

// Good: Prompt at runtime
password, _ := ork.PromptPassword("Sudo password: ")
node.SetBecomePassword(password)

// Bad: Hardcoded password
node.SetBecomePassword("secret123")  // DON'T DO THIS
```

### Sudo Configuration

For passwordless sudo, configure `/etc/sudoers`:

```bash
# Allow deploy user to run all commands without password
deploy ALL=(ALL) NOPASSWD: ALL

# Or restrict to specific commands
deploy ALL=(ALL) NOPASSWD: /usr/bin/apt-get, /usr/bin/systemctl
```

### Command Injection

When implementing `su` wrapping, escape single quotes in commands:

```go
func escapeCommand(cmd string) string {
    return strings.ReplaceAll(cmd, "'", "'\\''")
}
```

## Implementation Plan

### Phase 1: Interface Exposure (Low Risk)
1. Embed `types.BecomeInterface` in `InventoryInterface` and `GroupInterface`
2. Embed `types.BecomeInterface` in `types.RunnerInterface` (currently only in ork-level wrapper)

### Phase 2: Full BecomeInterface
3. Expand `BecomeInterface` with `SetBecome`, `SetBecomeMethod`, `SetBecomePassword`, `SetBecomeFlags`
4. Update `BaseBecome` with all fields and defaults
5. Update `NodeConfig` with `BecomeEnabled`, `BecomeMethod`, `BecomePassword`, `BecomeFlags`
6. Update all implementations (`nodeImplementation`, `groupImplementation`, `inventoryImplementation`)

### Phase 3: CommandWrapper
7. Implement `CommandWrapper` with `sudo`, `su`, `doas` support
8. Integrate into `ssh/functions.go` replacing the current inline wrapping
9. Add command escaping for `su` method

### Phase 4: Security & Testing
10. Add password handling
11. Add error detection for common sudo failures
12. Write comprehensive tests for all methods and precedence

### Phase 5: Documentation
13. Update docs/privilege_escalation.md with full examples
14. Add troubleshooting guide for common sudo/su errors

## Open Questions

1. Should `SetBecome(false)` on a node override a skill that calls `SetBecome(true)`? (Currently skill always wins.)
2. How to handle interactive sudo prompts in automated environments?
3. Should we cache sudo credentials for multiple commands in a session?
4. Should `become_exe` be supported to specify a custom sudo path?
5. How should `BecomePassword` interact with `RunnableOptions` for per-execution overrides?

## Future Enhancements

- Become caching: cache sudo credentials for a session
- Become detection: auto-detect available become methods on the target
- Become validation: verify sudo configuration before execution
- Windows `runas` support

## Comparison with Ansible

| Feature | Ansible | Ork (current) | Ork (proposed) |
|---------|---------|---------------|----------------|
| Enable become | `become: yes` | n/a (empty user = disabled) | `SetBecome(true)` |
| Become user | `become_user: postgres` | `SetBecomeUser("postgres")` ✅ | same |
| Become method | `become_method: su` | sudo only | `SetBecomeMethod("su")` |
| Become password | `ansible_become_pass` | not supported | `SetBecomePassword()` |
| Inventory level | `group_vars` | `inventory.SetBecomeUser()` ✅ | same + full interface |
| Node level | host vars | `node.SetBecomeUser()` ✅ | same + full interface |
| Skill/task level | task `become:` | `skill.SetBecomeUser()` ✅ | same + enable/disable |
| Precedence | task > play > inventory | skill > node > group > inventory ✅ | same |

## Related Proposals

- [Privilege Escalation (Simple)](2026-04-15-privilege-escalation.md) — implemented subset
- [Dry-Run Mode](../dry_run.md) — works with become today

## References

- [Ansible Become Documentation](https://docs.ansible.com/ansible/latest/user_guide/become.html)
- [sudo Manual](https://www.sudo.ws/man/1.8.27/sudo.man.html)
- [doas Manual](https://man.openbsd.org/doas)
- [Principle of Least Privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege)
