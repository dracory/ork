# Privilege Escalation (Become)

**Status:** Proposed
**Created:** 2026-04-15
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

## Proposal

Add **privilege escalation** support to Ork, allowing skills and commands to run as different users with configurable escalation methods. This includes:

- **Become user**: Specify which user to become (default: root)
- **Become method**: How to escalate (sudo, su, doas, pbrun, etc.)
- **Become password**: Password for privilege escalation (if required)
- **Per-skill control**: Enable/disable privilege escalation per skill
- **Command wrapping**: Automatically wrap commands with escalation prefix

## Motivation

### Current Limitations

```go
// Must connect as root for privileged operations
node := ork.NewNodeForHost("server.example.com").
    SetUser("root")  // Everything runs as root

node.RunCommand("apt-get update")  // Runs as root
node.Run(skills.NewUserCreate())   // Runs as root
```

### Proposed Solution

```go
// Connect as regular user, escalate when needed
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root").
    SetBecomeMethod("sudo")

// Automatically escalates to root
node.RunCommand("apt-get update")  // Runs as: sudo -u root apt-get update
node.Run(skills.NewUserCreate())   // Escalates to root

// Run specific command as different user
node.SetBecomeUser("postgres").
    RunCommand("psql -c 'SELECT version()'")  // Runs as postgres
```

## Architecture

### Core Types

### BecomeInterface

Define a clean interface for privilege escalation that can be embedded in other interfaces:

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

### Interface Integration

Embed `BecomeInterface` in both `RunnableInterface` and `RunnerInterface`:

```go
// RunnableInterface - skills implement this
type RunnableInterface interface {
    // Existing methods...
    GetID() string
    SetID(id string) RunnableInterface
    GetDescription() string
    SetDescription(description string) RunnableInterface
    GetNodeConfig() NodeConfig
    SetNodeConfig(cfg NodeConfig) RunnableInterface
    GetArg(key string) string
    SetArg(key, value string) RunnableInterface
    GetArgs() map[string]string
    SetArgs(args map[string]string) RunnableInterface
    IsDryRun() bool
    SetDryRun(dryRun bool) RunnableInterface
    GetTimeout() time.Duration
    SetTimeout(timeout time.Duration) RunnableInterface
    Check() (bool, error)
    Run() Result
    
    // Embed BecomeInterface for privilege escalation
    BecomeInterface
}

// RunnerInterface - nodes, groups, inventory implement this
type RunnerInterface interface {
    // Existing methods...
    RunCommand(cmd string) Results
    Run(skill RunnableInterface) Results
    RunByID(id string, opts ...RunnableOptions) Results
    Check(skill RunnableInterface) Results
    GetLogger() *slog.Logger
    SetLogger(logger *slog.Logger) RunnerInterface
    SetDryRunMode(dryRun bool) RunnerInterface
    GetDryRunMode() bool
    
    // Embed BecomeInterface for privilege escalation
    BecomeInterface
}
```

### NodeConfig Extension

```go
// NodeConfig stores the become configuration
type NodeConfig struct {
    // ... existing fields ...
    
    // Privilege escalation
    Become BecomeConfig
}
```

### BaseBecome Implementation

Provide a default implementation that can be embedded:

```go
// BaseBecome provides a default implementation of BecomeInterface
type BaseBecome struct {
    enabled  bool
    method   string
    user     string
    password string
    flags    string
}

func NewBaseBecome() *BaseBecome {
    return &BaseBecome{
        enabled:  false,
        method:   "sudo",
        user:     "root",
        flags:    "-n", // Non-interactive by default
    }
}

func (b *BaseBecome) SetBecome(enabled bool) BecomeInterface {
    b.enabled = enabled
    return b
}

func (b *BaseBecome) GetBecomeEnabled() bool {
    return b.enabled
}

func (b *BaseBecome) SetBecomeUser(user string) BecomeInterface {
    b.user = user
    return b
}

func (b *BaseBecome) GetBecomeUser() string {
    return b.user
}

func (b *BaseBecome) SetBecomeMethod(method string) BecomeInterface {
    b.method = method
    return b
}

func (b *BaseBecome) GetBecomeMethod() string {
    return b.method
}

func (b *BaseBecome) SetBecomePassword(password string) BecomeInterface {
    b.password = password
    return b
}

func (b *BaseBecome) GetBecomePassword() string {
    return b.password
}

func (b *BaseBecome) SetBecomeFlags(flags string) BecomeInterface {
    b.flags = flags
    return b
}

func (b *BaseBecome) GetBecomeFlags() string {
    return b.flags
}
```

### Usage in Implementations

Embed `BaseBecome` in concrete implementations:

```go
// nodeImplementation embeds BaseBecome
type nodeImplementation struct {
    *BaseBecome  // Embedded for become functionality
    cfg       types.NodeConfig
    sshClient *ssh.Client
    connected bool
}

// BaseSkill embeds BaseBecome
type BaseSkill struct {
    *BaseBecome  // Embedded for become functionality
    id          string
    description string
    nodeCfg     NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}

// BasePlaybook embeds BaseBecome
type BasePlaybook struct {
    *BaseBecome  // Embedded for become functionality
    id          string
    description string
    nodeCfg     NodeConfig
    args        map[string]string
    dryRun      bool
    timeout     time.Duration
}
```

### NodeConfig Storage

For passing become settings through SSH execution, store in NodeConfig:

```go
// NodeConfig stores become settings for SSH execution
type NodeConfig struct {
    // ... existing fields ...
    
    // Privilege escalation settings (copied from BecomeInterface at execution time)
    BecomeEnabled  bool
    BecomeMethod   string
    BecomeUser     string
    BecomePassword string
    BecomeFlags    string
}

// Helper to populate NodeConfig from BecomeInterface
func (n *nodeImplementation) GetNodeConfig() types.NodeConfig {
    cfg := n.cfg
    
    // Copy become settings from interface to config for SSH execution
    cfg.BecomeEnabled = n.GetBecomeEnabled()
    cfg.BecomeMethod = n.GetBecomeMethod()
    cfg.BecomeUser = n.GetBecomeUser()
    cfg.BecomePassword = n.GetBecomePassword()
    cfg.BecomeFlags = n.GetBecomeFlags()
    
    return cfg
}

// Skills do the same when they call ssh.Run()
func (s *MySkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    
    // cfg now has become settings from the skill
    // These override any node-level settings (precedence)
    
    output, err := ssh.Run(cfg, types.Command{
        Command: "apt-get update",
    })
    // ...
}
```

**Why NodeConfig needs become fields:**
- `NodeConfig` is passed to `ssh.Run()` which needs to wrap commands
- It's a data transfer object that carries all execution context
- Become settings are resolved at execution time and copied into it
- The SSH layer doesn't know about interfaces, only the config struct

### Become Methods

| Method | Description | Example Command |
|--------|-------------|-----------------|
| `sudo` | Use sudo (default) | `sudo -u root command` |
| `su` | Use su | `su - root -c 'command'` |
| `doas` | Use doas (OpenBSD) | `doas -u root command` |
| `pbrun` | Use PowerBroker | `pbrun -u root command` |
| `pfexec` | Use pfexec (Solaris) | `pfexec -u root command` |
| `runas` | Use runas (Windows) | `runas /user:Administrator command` |

### Command Wrapping

```go
// CommandWrapper wraps commands with privilege escalation
type CommandWrapper struct {
    enabled  bool
    method   string
    user     string
    password string
    flags    string
}

func NewCommandWrapper(b BecomeInterface) *CommandWrapper {
    return &CommandWrapper{
        enabled:  b.GetBecomeEnabled(),
        method:   b.GetBecomeMethod(),
        user:     b.GetBecomeUser(),
        password: b.GetBecomePassword(),
        flags:    b.GetBecomeFlags(),
    }
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
    case "pbrun":
        return w.wrapPbrun(cmd)
    default:
        return w.wrapSudo(cmd)  // Default to sudo
    }
}

func (w *CommandWrapper) wrapSudo(cmd string) string {
    flags := w.flags
    if flags == "" {
        flags = "-n"  // Non-interactive by default
    }
    
    user := w.user
    if user == "" {
        user = "root"
    }
    
    // Build sudo command
    sudoCmd := fmt.Sprintf("sudo %s -u %s", flags, user)
    
    // Handle password if provided
    if w.password != "" {
        // Use sudo -S to read password from stdin
        sudoCmd = fmt.Sprintf("echo '%s' | sudo -S %s -u %s", 
            w.password, flags, user)
    }
    
    return fmt.Sprintf("%s %s", sudoCmd, cmd)
}

func (w *CommandWrapper) wrapSu(cmd string) string {
    user := w.user
    if user == "" {
        user = "root"
    }
    
    // su requires command in quotes
    return fmt.Sprintf("su - %s -c '%s'", user, cmd)
}

func (w *CommandWrapper) wrapDoas(cmd string) string {
    user := w.user
    if user == "" {
        user = "root"
    }
    
    return fmt.Sprintf("doas -u %s %s", user, cmd)
}
```

## API Design

### Precedence Hierarchy

Like Ansible, Ork supports `become` at multiple levels with clear precedence (lowest to highest):

1. **Inventory level** - Default for all groups/nodes
2. **Group level** - Inherited by all nodes in the group
3. **Node level** - Default for all operations on the node
4. **Skill/Runnable level** - Per-skill override (highest precedence)

The most specific setting always wins. This allows you to set sensible defaults at higher levels and override for specific operations.

### Node-Level Configuration

Since `NodeInterface` embeds `BecomeInterface`, all become methods are available:

```go
// Enable privilege escalation at node level (applies to all operations)
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true)  // Default: sudo to root

// Customize become settings using fluent API
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root").
    SetBecomeMethod("sudo").
    SetBecomePassword("secret").
    SetBecomeFlags("-n -H")

// Query become settings
if node.GetBecomeEnabled() {
    fmt.Printf("Will become: %s via %s\n", 
        node.GetBecomeUser(), 
        node.GetBecomeMethod())
}

// Temporarily disable become for specific command
node.SetBecome(false).
    RunCommand("whoami")  // Runs as deploy

// Re-enable for next operation
node.SetBecome(true).
    RunCommand("apt-get update")  // Runs as root
```

### Fluent API Methods

All methods are defined in `BecomeInterface` and available on any type that embeds it:

```go
// BecomeInterface methods (available on Node, Group, Inventory, Skills)
SetBecome(enabled bool) BecomeInterface
GetBecome() BecomeConfig
SetBecomeUser(user string) BecomeInterface
GetBecomeUser() string
SetBecomeMethod(method string) BecomeInterface
GetBecomeMethod() string
SetBecomePassword(password string) BecomeInterface
GetBecomePassword() string
SetBecomeFlags(flags string) BecomeInterface
GetBecomeFlags() string
```

### Return Type Flexibility

Note that `BecomeInterface` methods return `BecomeInterface`, not the concrete type. For fluent chaining with other methods, you may need type assertions or redesign the return types:

**Option 1: Return concrete type (recommended)**
```go
// In NodeInterface, override BecomeInterface methods to return NodeInterface
type NodeInterface interface {
    RunnerInterface
    BecomeInterface  // Embed for become methods
    
    // Override become methods to return NodeInterface for better chaining
    SetBecome(enabled bool) NodeInterface
    SetBecomeUser(user string) NodeInterface
    SetBecomeMethod(method string) NodeInterface
    SetBecomePassword(password string) NodeInterface
    SetBecomeFlags(flags string) NodeInterface
    
    // Other node methods...
    SetPort(port string) NodeInterface
    SetUser(user string) NodeInterface
    // ...
}
```

**Option 2: Use type assertions**
```go
// Less elegant but works
node := ork.NewNodeForHost("server.example.com")
node.SetBecome(true)  // Returns BecomeInterface
node.(NodeInterface).SetUser("deploy")  // Type assertion needed
```

**Option 3: Separate calls**
```go
// Simplest but less fluent
node := ork.NewNodeForHost("server.example.com")
node.SetBecome(true)
node.SetBecomeUser("root")
node.SetUser("deploy")
```

**Recommended approach:** Override become methods in concrete interfaces to return the concrete type for seamless fluent chaining.

### Skill-Level Control (Highest Precedence)

Since skills implement `RunnableInterface` which embeds `BecomeInterface`, they have full control over privilege escalation:

```go
// Skills can force become regardless of node settings
type AptUpdateSkill struct {
    *types.BaseSkill  // Embeds BaseBecome via BaseSkill
}

func NewAptUpdateSkill() types.RunnableInterface {
    skill := &AptUpdateSkill{
        BaseSkill: types.NewBaseSkill(),
    }
    
    // Configure become at skill creation
    skill.SetID("apt-update").
        SetDescription("Update package database").
        SetBecome(true).        // Force become
        SetBecomeUser("root")   // Must run as root
    
    return skill
}

func (s *AptUpdateSkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    
    // Skill's become settings override node settings
    // cfg.Become is already set from skill's configuration
    
    output, err := ssh.Run(cfg, types.Command{
        Command: "apt-get update",
    })
    
    // ... handle result
}

// Or disable become for a specific skill
type ReadConfigSkill struct {
    *types.BaseSkill
}

func NewReadConfigSkill() types.RunnableInterface {
    skill := &ReadConfigSkill{
        BaseSkill: types.NewBaseSkill(),
    }
    
    skill.SetID("read-config").
        SetDescription("Read user config file").
        SetBecome(false)  // Explicitly disable become
    
    return skill
}

func (s *ReadConfigSkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    
    // Runs without privilege escalation
    output, err := ssh.Run(cfg, types.Command{
        Command: "cat /home/deploy/config.yml",
    })
    
    // ... handle result
}
```

### Dynamic Become Control in Skills

Skills can also modify become settings dynamically during execution:

```go
type ConditionalPrivilegeSkill struct {
    *types.BaseSkill
}

func (s *ConditionalPrivilegeSkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    
    // Check if we need root access
    testOutput, _ := ssh.Run(cfg, types.Command{
        Command: "test -w /etc/myapp/config",
    })
    
    if testOutput != "" {
        // File not writable, need to become root
        s.SetBecome(true).SetBecomeUser("root")
        cfg = s.GetNodeConfig()  // Get updated config
    }
    
    // Now run the actual command
    output, err := ssh.Run(cfg, types.Command{
        Command: "echo 'setting=value' >> /etc/myapp/config",
    })
    
    // ... handle result
}
```

### Group and Inventory Support

Groups and inventories also embed `BecomeInterface`, allowing them to set become defaults:

```go
// Set become at group level - all nodes inherit
webGroup := ork.NewGroup("webservers").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root")

webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

// All nodes inherit group's become settings
results := webGroup.RunCommand("apt-get update")

// Individual nodes can override group settings
node := ork.NewNodeForHost("web3.example.com").
    SetBecome(false)  // Override: don't use become
webGroup.AddNode(node)

// Set become at inventory level
inv := ork.NewInventory().
    SetUser("deploy").
    SetBecome(true)  // Default for all groups/nodes

inv.AddGroup(webGroup)

// Query become settings at any level
fmt.Printf("Inventory become: %v\n", inv.GetBecomeEnabled())
fmt.Printf("Group become user: %s\n", webGroup.GetBecomeUser())
fmt.Printf("Node become method: %s\n", node.GetBecomeMethod())
```

### Precedence Example

```go
// Inventory level: become enabled, user=root
inv := ork.NewInventory().
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root")

// Group level: override become user to postgres
dbGroup := ork.NewGroup("databases").
    SetBecomeUser("postgres")  // Inherits enabled=true from inventory
inv.AddGroup(dbGroup)

// Node level: disable become for this specific node
node := ork.NewNodeForHost("db-readonly.example.com").
    SetBecome(false)  // Override: no privilege escalation
dbGroup.AddNode(node)

// Skill level: force become regardless of node setting
type BackupSkill struct {
    *types.BaseSkill
}

func (s *BackupSkill) Run() types.Result {
    cfg := s.GetNodeConfig()
    cfg.Become.Enabled = true  // Force become even if node disabled it
    cfg.Become.User = "postgres"
    // ... backup logic
}

// Final precedence for db-readonly node running BackupSkill:
// - Inventory: become=true, user=root
// - Group: user=postgres (overrides inventory user)
// - Node: become=false (overrides group/inventory)
// - Skill: become=true, user=postgres (WINS - highest precedence)
// Result: Runs as postgres via sudo
```

## Usage Examples

### Basic Privilege Escalation

```go
// Connect as regular user, escalate to root
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true)

// Runs as: sudo -n -u root apt-get update
results := node.RunCommand("apt-get update")
```

### Custom Become User

```go
// Run command as postgres user
node := ork.NewNodeForHost("db.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("postgres")

// Runs as: sudo -n -u postgres psql -c 'SELECT version()'
results := node.RunCommand("psql -c 'SELECT version()'")
```

### Different Become Methods

```go
// Use su instead of sudo
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeMethod("su")

// Runs as: su - root -c 'apt-get update'
results := node.RunCommand("apt-get update")

// Use doas (OpenBSD)
node := ork.NewNodeForHost("openbsd.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeMethod("doas")

// Runs as: doas -u root apt-get update
results := node.RunCommand("apt-get update")
```

### Password-Based Sudo

```go
// Provide sudo password
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomePassword("secret123")

// Runs as: echo 'secret123' | sudo -S -n -u root apt-get update
results := node.RunCommand("apt-get update")
```

### Custom Sudo Flags

```go
// Use custom sudo flags
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomeFlags("-n -H")  // Non-interactive, set HOME

// Runs as: sudo -n -H -u root apt-get update
results := node.RunCommand("apt-get update")
```

### Selective Privilege Escalation

```go
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy")

// Run as regular user
node.SetBecome(false).
    RunCommand("whoami")  // Runs as deploy

// Escalate for privileged operation
node.SetBecome(true).
    RunCommand("apt-get update")  // Runs as root

// Back to regular user
node.SetBecome(false).
    RunCommand("ls ~")  // Runs as deploy
```

### Skills with Become

```go
// Skills automatically use node's become settings
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true)

// Skill runs with privilege escalation
results := node.Run(skills.NewAptUpdate())
// Internally runs: sudo -n -u root apt-get update
```

### Multi-User Operations

```go
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true)

// Install packages as root
node.SetBecomeUser("root").
    Run(skills.NewAptUpdate())

// Configure database as postgres
node.SetBecomeUser("postgres").
    RunCommand("psql -c 'CREATE DATABASE myapp'")

// Deploy application as app user
node.SetBecomeUser("appuser").
    RunCommand("cd /app && ./deploy.sh")
```

### Inventory with Become

```go
inv := ork.NewInventory()

// Production servers: connect as deploy, escalate to root
prodGroup := ork.NewGroup("production")
prodGroup.SetUser("deploy").
    SetBecome(true).
    SetBecomeUser("root")
prodGroup.AddNode(ork.NewNodeForHost("prod1.example.com"))
prodGroup.AddNode(ork.NewNodeForHost("prod2.example.com"))
inv.AddGroup(prodGroup)

// Development servers: connect as root directly
devGroup := ork.NewGroup("development")
devGroup.SetUser("root").
    SetBecome(false)  // No escalation needed
devGroup.AddNode(ork.NewNodeForHost("dev1.example.com"))
inv.AddGroup(devGroup)

// Run on all servers with appropriate privileges
results := inv.Run(skills.NewAptUpdate())
```

### Become with Vault Secrets

```go
// Load sudo password from vault
secrets, err := ork.VaultFileToKeysWithPrompt(".env.vault")
if err != nil {
    log.Fatal(err)
}

node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecome(true).
    SetBecomePassword(secrets["SUDO_PASSWORD"])

results := node.RunCommand("apt-get update")
```

## Integration with SSH Package

### SSH Command Execution

```go
// ssh.Run() needs to wrap commands with become
func Run(cfg types.NodeConfig, cmd types.Command) (string, error) {
    // Wrap command if become is enabled
    if cfg.Become.Enabled {
        wrapper := NewCommandWrapper(cfg.Become)
        cmd.Command = wrapper.Wrap(cmd.Command)
    }
    
    // Execute wrapped command
    return executeCommand(cfg, cmd)
}
```

### SSH Client

```go
// Client.Run() also wraps commands
func (c *Client) Run(cmd string) (string, error) {
    // Wrap command if become is enabled
    if c.cfg.Become.Enabled {
        wrapper := NewCommandWrapper(c.cfg.Become)
        cmd = wrapper.Wrap(cmd)
    }
    
    // Execute wrapped command
    return c.executeCommand(cmd)
}
```

## Security Considerations

### Password Storage

1. **Never hardcode passwords** in source code
2. **Use vault** for password storage
3. **Prompt at runtime** when possible
4. **Clear passwords** from memory after use

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

Properly escape commands to prevent injection:

```go
// Escape single quotes in commands
func escapeCommand(cmd string) string {
    return strings.ReplaceAll(cmd, "'", "'\\''")
}

func (w *CommandWrapper) wrapSu(cmd string) string {
    user := w.cfg.User
    if user == "" {
        user = "root"
    }
    
    // Escape command for su
    escapedCmd := escapeCommand(cmd)
    return fmt.Sprintf("su - %s -c '%s'", user, escapedCmd)
}
```

## Benefits

1. **Clean Interface Design**: `BecomeInterface` provides clear separation of concerns
2. **Composability**: Can be embedded in any interface that needs privilege escalation
3. **Consistency**: Same API across Node, Group, Inventory, and Skills
4. **Type Safety**: Compiler validates become configuration
5. **Flexibility**: Override at any level with clear precedence
6. **Security**: Follow principle of least privilege
7. **Reusability**: `BaseBecome` implementation can be reused
8. **Testability**: Easy to mock `BecomeInterface` for testing
9. **Discoverability**: All become methods grouped in one interface
10. **Extensibility**: Easy to add new become methods without changing existing code

## Comparison with Ansible

| Feature | Ansible | Ork |
|---------|---------|-----|
| Enable become | `become: yes` | `SetBecome(true)` |
| Become user | `become_user: postgres` | `SetBecomeUser("postgres")` |
| Become method | `become_method: su` | `SetBecomeMethod("su")` |
| Become password | `ansible_become_pass` | `SetBecomePassword()` |
| Inventory level | `host_vars`, `group_vars` | `inventory.SetBecome()`, `group.SetBecome()` |
| Play level | `become: yes` in play | `node.SetBecome(true)` |
| Task level | `become: yes` in task | Skill-level override in `Run()` |
| Per-execution | Task vars | `RunnableOptions.Become` |
| Precedence | Task > Block > Play > Inventory > Config | Skill > Node > Group > Inventory |
| Methods | sudo, su, pbrun, etc. | sudo, su, doas, pbrun, etc. |
| Configuration | YAML | Fluent API |

### Precedence Comparison

**Ansible precedence (lowest to highest):**
1. ansible.cfg defaults
2. Inventory variables (host_vars, group_vars)
3. Play level
4. Block level
5. Task level (highest)

**Ork precedence (lowest to highest):**
1. Inventory level
2. Group level
3. Node level
4. Skill/Runnable level (highest)

Both follow the same principle: **more specific settings override more general ones**.

## Implementation Considerations

### 1. Interface Embedding

Embed `BecomeInterface` in the appropriate interfaces:

```go
// types/become_interface.go
type BecomeInterface interface {
    GetBecome() BecomeConfig
    SetBecome(enabled bool) BecomeInterface
    GetBecomeUser() string
    SetBecomeUser(user string) BecomeInterface
    GetBecomeMethod() string
    SetBecomeMethod(method string) BecomeInterface
    GetBecomePassword() string
    SetBecomePassword(password string) BecomeInterface
    GetBecomeFlags() string
    SetBecomeFlags(flags string) BecomeInterface
}

// types/runnable_interface.go
type RunnableInterface interface {
    // ... existing methods ...
    BecomeInterface  // Embed become functionality
}

// types/runner_interface.go
type RunnerInterface interface {
    // ... existing methods ...
    BecomeInterface  // Embed become functionality
}
```

### 2. Method Return Type Override

For fluent chaining, override become methods in concrete interfaces:

```go
// node_interface.go
type NodeInterface interface {
    RunnerInterface
    
    // Override BecomeInterface methods to return NodeInterface
    SetBecome(enabled bool) NodeInterface
    SetBecomeUser(user string) NodeInterface
    SetBecomeMethod(method string) NodeInterface
    SetBecomePassword(password string) NodeInterface
    SetBecomeFlags(flags string) NodeInterface
    
    // Other node-specific methods...
    SetPort(port string) NodeInterface
    SetUser(user string) NodeInterface
    SetKey(key string) NodeInterface
    // ...
}

// node_implementation.go
func (n *nodeImplementation) SetBecome(enabled bool) NodeInterface {
    n.BaseBecome.SetBecome(enabled)
    return n
}

func (n *nodeImplementation) SetBecomeUser(user string) NodeInterface {
    n.BaseBecome.SetBecomeUser(user)
    return n
}

// ... similar for other become methods
```

### 3. Precedence Resolution

Implement precedence resolution when executing commands:

```go
// resolveBecomeSettings resolves become settings with precedence
func resolveBecomeSettings(inventory, group, node, skill BecomeInterface) BecomeInterface {
    // Create a new BaseBecome with resolved settings
    resolved := NewBaseBecome()
    
    // Start with inventory defaults
    if inventory.GetBecomeEnabled() {
        resolved.SetBecome(true)
        resolved.SetBecomeUser(inventory.GetBecomeUser())
        resolved.SetBecomeMethod(inventory.GetBecomeMethod())
        resolved.SetBecomePassword(inventory.GetBecomePassword())
        resolved.SetBecomeFlags(inventory.GetBecomeFlags())
    }
    
    // Override with group settings if set
    if group.GetBecomeEnabled() {
        resolved.SetBecome(true)
        if user := group.GetBecomeUser(); user != "" {
            resolved.SetBecomeUser(user)
        }
        if method := group.GetBecomeMethod(); method != "" {
            resolved.SetBecomeMethod(method)
        }
        if password := group.GetBecomePassword(); password != "" {
            resolved.SetBecomePassword(password)
        }
        if flags := group.GetBecomeFlags(); flags != "" {
            resolved.SetBecomeFlags(flags)
        }
    }
    
    // Override with node settings if set
    if node.GetBecomeEnabled() {
        resolved.SetBecome(true)
        if user := node.GetBecomeUser(); user != "" {
            resolved.SetBecomeUser(user)
        }
        if method := node.GetBecomeMethod(); method != "" {
            resolved.SetBecomeMethod(method)
        }
        if password := node.GetBecomePassword(); password != "" {
            resolved.SetBecomePassword(password)
        }
        if flags := node.GetBecomeFlags(); flags != "" {
            resolved.SetBecomeFlags(flags)
        }
    }
    
    // Override with skill settings if set (highest precedence)
    if skill.GetBecomeEnabled() {
        resolved.SetBecome(true)
        if user := skill.GetBecomeUser(); user != "" {
            resolved.SetBecomeUser(user)
        }
        if method := skill.GetBecomeMethod(); method != "" {
            resolved.SetBecomeMethod(method)
        }
        if password := skill.GetBecomePassword(); password != "" {
            resolved.SetBecomePassword(password)
        }
        if flags := skill.GetBecomeFlags(); flags != "" {
            resolved.SetBecomeFlags(flags)
        }
    }
    
    return resolved
}
```

### 4. SSH Integration

Wrap commands at the SSH execution layer:

```go
// ssh.Run() wraps commands with become
func Run(cfg types.NodeConfig, cmd types.Command) (string, error) {
    // Wrap command if become is enabled
    if cfg.BecomeEnabled {
        // Create wrapper from NodeConfig become settings
        wrapper := &CommandWrapper{
            enabled:  cfg.BecomeEnabled,
            method:   cfg.BecomeMethod,
            user:     cfg.BecomeUser,
            password: cfg.BecomePassword,
            flags:    cfg.BecomeFlags,
        }
        cmd.Command = wrapper.Wrap(cmd.Command)
        
        // Log wrapped command in dry-run mode
        if cfg.IsDryRunMode {
            cfg.GetLoggerOrDefault().Info("dry-run: would run command",
                "original", cmd.Command,
                "wrapped", cmd.Command)
        }
    }
    
    // Execute command
    return executeCommand(cfg, cmd)
}
```

### 2. Password Handling

Use secure password handling:

```go
// Clear password from memory after use
func (n *nodeImplementation) RunCommand(cmd string) types.Results {
    // ... execution logic ...
    
    // Clear password after use
    defer func() {
        if n.GetBecomePassword() != "" {
            n.SetBecomePassword("")
        }
    }()
    
    // ... rest of execution
}
```

### 3. Dry-Run Mode

Show wrapped commands in dry-run mode:

```go
if cfg.IsDryRunMode {
    wrapper := NewCommandWrapper(node)
    wrappedCmd := wrapper.Wrap(cmd)
    log.Printf("[dry-run] Would run: %s", wrappedCmd)
    return "[dry-run]", nil
}
```

### 4. Error Handling

Detect and report privilege escalation failures:

```go
// Check for common sudo errors
if strings.Contains(output, "sudo: no tty present") {
    return "", fmt.Errorf("sudo requires a password but no TTY available")
}
if strings.Contains(output, "sudo: a password is required") {
    return "", fmt.Errorf("sudo password required but not provided")
}
if strings.Contains(output, "Sorry, user") {
    return "", fmt.Errorf("user not authorized to use sudo")
}
```

### 5. Testing

Test with mock implementations:

```go
func TestBecomeWrapping(t *testing.T) {
    become := NewBaseBecome()
    become.SetBecome(true).
        SetBecomeMethod("sudo").
        SetBecomeUser("root")
    
    wrapper := NewCommandWrapper(become)
    wrapped := wrapper.Wrap("apt-get update")
    
    expected := "sudo -n -u root apt-get update"
    if wrapped != expected {
        t.Errorf("Expected %s, got %s", expected, wrapped)
    }
}

func TestBecomePrecedence(t *testing.T) {
    inventory := NewBaseBecome().SetBecome(true).SetBecomeUser("root")
    group := NewBaseBecome().SetBecomeUser("postgres")
    node := NewBaseBecome().SetBecome(false)
    skill := NewBaseBecome().SetBecome(true).SetBecomeUser("app")
    
    resolved := resolveBecomeSettings(inventory, group, node, skill)
    
    // Skill has highest precedence
    if !resolved.GetBecomeEnabled() {
        t.Error("Expected become to be enabled")
    }
    if resolved.GetBecomeUser() != "app" {
        t.Errorf("Expected user 'app', got '%s'", resolved.GetBecomeUser())
    }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Define `BecomeInterface` in types package
2. Implement `BaseBecome` with all getters/setters
3. Embed `BecomeInterface` in `RunnableInterface` and `RunnerInterface`
4. Update `NodeConfig` to store become settings
5. Embed `BaseBecome` in `nodeImplementation`, `BaseSkill`, `BasePlaybook`

### Phase 2: Command Wrapping
6. Implement `CommandWrapper` with sudo support
7. Integrate wrapping into `ssh.Run()`
8. Add precedence resolution logic
9. Update `GetNodeConfig()` to copy become settings

### Phase 3: Additional Methods
10. Implement `su` method
11. Implement `doas` method
12. Implement `pbrun` method
13. Add method detection and validation

### Phase 4: Security & Testing
14. Add password handling and clearing
15. Implement command escaping
16. Add error detection for common failures
17. Write comprehensive tests

### Phase 5: Documentation & Examples
18. Update documentation with examples
19. Add security best practices guide
20. Create example skills using become
21. Add troubleshooting guide

## Success Metrics

1. **Adoption**: Percentage of users using become instead of root
2. **Security**: Reduction in root SSH connections
3. **Flexibility**: Number of different become users used
4. **Reliability**: Success rate of privilege escalation
5. **Performance**: Minimal overhead from command wrapping

## Open Questions

1. Should we support `become_flags` in `RunnableOptions` for per-execution overrides?
2. How to handle interactive sudo prompts in automated environments?
3. Should we cache sudo credentials for multiple commands in a session?
4. How to detect if sudo is configured for passwordless access?
5. Should we support `become_exe` to specify custom sudo path?
6. How to handle Windows runas differently from Unix sudo?
7. Should block-level become be supported (for grouping tasks in playbooks)?
8. How should become settings merge when both node and RunnableOptions specify them?

## Future Enhancements

1. **Become Caching**: Cache sudo credentials for session
2. **Become Detection**: Auto-detect available become methods
3. **Become Validation**: Validate sudo configuration before execution
4. **Become Profiles**: Predefined become configurations
5. **Become Hooks**: Pre/post become execution hooks
6. **Become Logging**: Detailed logging of privilege escalation
7. **Become Metrics**: Track privilege escalation usage
8. **Windows Support**: Full runas implementation

## Related Proposals

- [Dry-Run Mode](2026-04-12-dry-run-mode.md) - Preview commands with become wrapping
- [Structured Logging](2026-04-12-structured-logging.md) - Log privilege escalation events
- [Configuration Management](2026-04-12-configuration-management.md) - Store become settings

## References

- [Ansible Become Documentation](https://docs.ansible.com/ansible/latest/user_guide/become.html)
- [sudo Manual](https://www.sudo.ws/man/1.8.27/sudo.man.html)
- [doas Manual](https://man.openbsd.org/doas)
- [Principle of Least Privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege)
