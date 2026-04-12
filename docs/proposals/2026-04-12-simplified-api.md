# Proposal: Simplified User-Facing API

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

The current API requires users to import multiple packages and understand the internal structure:

```go
import (
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/playbooks"
)
```

This is not user-friendly because:
- Users need to know the internal package structure
- Multiple imports for simple operations
- Verbose and repetitive code
- Harder to discover functionality
- Not intuitive for newcomers

## Proposed Solution

Create a top-level `ork` package that exposes a clean, intuitive API. Users should only need:

```go
import "github.com/dracory/ork"
```

All common operations should be accessible through simple, discoverable functions.

## API Design

### 1. Top-Level Package Structure

```
ork/
├── ork.go              # Main user-facing API
├── config/             # Internal config (still accessible if needed)
├── ssh/                # Internal SSH (still accessible if needed)
├── playbook/           # Internal playbook interfaces
├── playbooks/          # Internal playbook implementations
└── internal/           # Truly internal, not exported
```

### 2. Simple SSH Operations

```go
// ork.go - Top-level API

package ork

// RunSSH executes a single command on a remote server
func RunSSH(host, cmd string, opts ...Option) (string, error)

// Example usage:
output, err := ork.RunSSH("server.example.com", "uptime")
output, err := ork.RunSSH("server.example.com", "uptime", 
    ork.WithPort("2222"),
    ork.WithUser("deploy"),
    ork.WithKey("id_rsa"),
)
```

### 3. Playbook Execution

```go
// RunPlaybook executes a named playbook
func RunPlaybook(name, host string, opts ...Option) error

// Example usage:
err := ork.RunPlaybook("apt-upgrade", "server.example.com")
err := ork.RunPlaybook("user-create", "server.example.com",
    ork.WithArg("username", "john"),
)
```

### 4. Fluent Builder API

```go
// NewNode creates a new remote node connection with fluent configuration
func NewNode(host string) *Node

type Node struct {
    // internal fields
}

func (n *Node) Port(port string) *Node
func (n *Node) User(user string) *Node
func (n *Node) Key(key string) *Node
func (n *Node) Arg(key, value string) *Node
func (n *Node) DryRun(enabled bool) *Node
func (n *Node) Verbose(enabled bool) *Node

func (n *Node) Run(cmd string) (string, error)
func (n *Node) Playbook(name string) error
func (n *Node) Connect() error
func (n *Node) Close() error

// Example usage:
node := ork.NewNode("server.example.com").
    Port("2222").
    User("root").
    Key("production.prv")

output, err := node.Run("uptime")
err = node.Playbook("apt-upgrade")
```

### 5. Functional Options Pattern

```go
type Option func(*config)

type config struct {
    host     string
    port     string
    user     string
    key      string
    args     map[string]string
    dryRun   bool
    verbose  bool
    timeout  time.Duration
}

func WithPort(port string) Option {
    return func(c *config) {
        c.port = port
    }
}

func WithUser(user string) Option {
    return func(c *config) {
        c.user = user
    }
}

func WithKey(key string) Option {
    return func(c *config) {
        c.key = key
    }
}

func WithArg(key, value string) Option {
    return func(c *config) {
        if c.args == nil {
            c.args = make(map[string]string)
        }
        c.args[key] = value
    }
}

func WithArgs(args map[string]string) Option {
    return func(c *config) {
        c.args = args
    }
}

func WithDryRun(enabled bool) Option {
    return func(c *config) {
        c.dryRun = enabled
    }
}

func WithTimeout(timeout time.Duration) Option {
    return func(c *config) {
        c.timeout = timeout
    }
}
```

## Complete Examples

### Before (Current API)

```go
package main

import (
    "log"
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/playbooks"
)

func main() {
    // Create config
    cfg := config.Config{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }
    
    // Run a command
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
    
    // Run playbook
    ping := playbooks.NewPing()
    if err := ping.Run(cfg); err != nil {
        log.Fatal(err)
    }
    
    // Run with args
    cfg.Args = map[string]string{"username": "john"}
    userCreate := playbooks.NewUserCreate()
    if err := userCreate.Run(cfg); err != nil {
        log.Fatal(err)
    }
}
```

### After (New API - Functional Options)

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Run a command
    output, err := ork.RunSSH("db3.sinevia.com", "uptime",
        ork.WithPort("40022"),
        ork.WithKey("2024_sinevia.prv"),
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
    
    // Run playbook
    err = ork.RunPlaybook("ping", "db3.sinevia.com",
        ork.WithPort("40022"),
        ork.WithKey("2024_sinevia.prv"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Run with args
    err = ork.RunPlaybook("user-create", "db3.sinevia.com",
        ork.WithPort("40022"),
        ork.WithKey("2024_sinevia.prv"),
        ork.WithArg("username", "john"),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### After (New API - Fluent Builder)

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Connect to remote node
    node := ork.NewNode("db3.sinevia.com").
        Port("40022").
        Key("2024_sinevia.prv")
    
    // Run a command
    output, err := node.Run("uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
    
    // Run playbook
    if err := node.Playbook("ping"); err != nil {
        log.Fatal(err)
    }
    
    // Run with args
    err = node.Arg("username", "john").Playbook("user-create")
    if err != nil {
        log.Fatal(err)
    }
}
```

### After (New API - Persistent Connection)

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Connect to node
    node := ork.NewNode("db3.sinevia.com").
        Port("40022").
        Key("2024_sinevia.prv")
    
    if err := node.Connect(); err != nil {
        log.Fatal(err)
    }
    defer node.Close()
    
    // Multiple operations on same connection
    output, _ := node.Run("uptime")
    log.Println(output)
    
    node.Run("apt-get update -y")
    node.Run("apt-get upgrade -y")
    
    // Run playbooks
    node.Playbook("swap-create")
    node.Arg("username", "john").Playbook("user-create")
}
```

## Implementation

### ork.go (Main API)

```go
package ork

import (
    "fmt"
    "time"
    
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbook"
    "github.com/dracory/ork/playbooks"
    "github.com/dracory/ork/ssh"
)

// Global registry
var defaultRegistry = buildDefaultRegistry()

func buildDefaultRegistry() *playbook.Registry {
    r := playbook.NewRegistry()
    
    // Register all built-in playbooks
    r.Register(playbooks.NewPing())
    r.Register(playbooks.NewAptUpdate())
    r.Register(playbooks.NewAptUpgrade())
    r.Register(playbooks.NewAptStatus())
    r.Register(playbooks.NewReboot())
    r.Register(playbooks.NewSwapCreate())
    r.Register(playbooks.NewSwapDelete())
    r.Register(playbooks.NewSwapStatus())
    r.Register(playbooks.NewUserCreate())
    r.Register(playbooks.NewUserDelete())
    r.Register(playbooks.NewUserStatus())
    
    return r
}

// RunSSH executes a single SSH command
func RunSSH(host, cmd string, opts ...Option) (string, error) {
    cfg := applyOptions(host, opts...)
    return ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
}

// RunPlaybook executes a named playbook
func RunPlaybook(name, host string, opts ...Option) error {
    cfg := applyOptions(host, opts...)
    
    pb, ok := defaultRegistry.Get(name)
    if !ok {
        return fmt.Errorf("playbook '%s' not found", name)
    }
    
    return pb.Run(cfg)
}

// ListPlaybooks returns all available playbook names
func ListPlaybooks() []string {
    return defaultRegistry.Names()
}

// GetPlaybook retrieves a playbook by name
func GetPlaybook(name string) (playbook.Playbook, bool) {
    return defaultRegistry.Get(name)
}

// RegisterPlaybook adds a custom playbook to the registry
func RegisterPlaybook(pb playbook.Playbook) {
    defaultRegistry.Register(pb)
}

// applyOptions builds a config from options
func applyOptions(host string, opts ...Option) config.Config {
    cfg := &config{
        host:    host,
        port:    "22",
        user:    "root",
        key:     "id_rsa",
        timeout: 30 * time.Second,
    }
    
    for _, opt := range opts {
        opt(cfg)
    }
    
    return config.Config{
        SSHHost:  cfg.host,
        SSHPort:  cfg.port,
        RootUser: cfg.user,
        SSHKey:   cfg.key,
        Args:     cfg.args,
    }
}
```

### Node Implementation

```go
// Node represents a remote server/node that can be managed
type Node struct {
    cfg        config.Config
    sshClient  *ssh.Client
}

// NewNode creates a new remote node connection
func NewNode(host string) *Node {
    return &Node{
        cfg: config.Config{
            SSHHost:  host,
            SSHPort:  "22",
            RootUser: "root",
            SSHKey:   "id_rsa",
            Args:     make(map[string]string),
        },
    }
}

// Port sets the SSH port
func (n *Node) Port(port string) *Node {
    n.cfg.SSHPort = port
    return n
}

// User sets the SSH user
func (n *Node) User(user string) *Node {
    n.cfg.RootUser = user
    return n
}

// Key sets the SSH key filename
func (n *Node) Key(key string) *Node {
    n.cfg.SSHKey = key
    return n
}

// Arg sets a playbook argument
func (n *Node) Arg(key, value string) *Node {
    if n.cfg.Args == nil {
        n.cfg.Args = make(map[string]string)
    }
    n.cfg.Args[key] = value
    return n
}

// Args sets multiple playbook arguments
func (n *Node) Args(args map[string]string) *Node {
    n.cfg.Args = args
    return n
}

// Connect establishes the SSH connection to the node
func (n *Node) Connect() error {
    client := ssh.NewClient(n.cfg.SSHHost, n.cfg.SSHPort, n.cfg.RootUser, n.cfg.SSHKey)
    if err := client.Connect(); err != nil {
        return err
    }
    n.sshClient = client
    return nil
}

// Close closes the SSH connection to the node
func (n *Node) Close() error {
    if n.sshClient != nil {
        return n.sshClient.Close()
    }
    return nil
}

// Run executes a command on the remote node
func (n *Node) Run(cmd string) (string, error) {
    if n.sshClient != nil {
        // Use persistent connection
        return n.sshClient.Run(cmd)
    }
    // One-off connection
    return ssh.RunOnce(n.cfg.SSHHost, n.cfg.SSHPort, n.cfg.RootUser, n.cfg.SSHKey, cmd)
}

// Playbook executes a named playbook on the remote node
func (n *Node) Playbook(name string) error {
    pb, ok := defaultRegistry.Get(name)
    if !ok {
        return fmt.Errorf("playbook '%s' not found", name)
    }
    return pb.Run(n.cfg)
}

// Config returns the underlying config (for advanced use)
func (n *Node) Config() config.Config {
    return n.cfg
}
```

## Migration Path

### Phase 1: Add New API (Non-Breaking)
- Create `ork.go` with new API
- Keep existing packages unchanged
- Both APIs work simultaneously

### Phase 2: Update Documentation
- Update README with new API examples
- Mark old API as "advanced usage"
- Add migration guide

### Phase 3: Deprecation (Optional)
- Add deprecation notices to old API
- Provide automated migration tool
- Eventually remove in v2.0

## Backward Compatibility

The new API is **completely backward compatible**:

```go
// Old API still works
import (
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
)

cfg := config.Config{...}
ssh.RunOnce(...)

// New API also available
import "github.com/dracory/ork"

ork.RunSSH(...)
```

Users can migrate at their own pace.

## Benefits

### For New Users
- **Discoverability**: IDE autocomplete shows all functions
- **Simplicity**: One import, clear functions
- **Learning Curve**: Intuitive API, less to learn
- **Examples**: Easier to understand and copy

### For Existing Users
- **Backward Compatible**: No breaking changes
- **Optional**: Can continue using old API
- **Gradual Migration**: Migrate code incrementally

### For the Project
- **Professional**: Matches Go best practices
- **Competitive**: Similar to popular libraries (Docker SDK, AWS SDK)
- **Adoption**: Lower barrier to entry
- **Documentation**: Easier to document and explain

## Comparison with Other Libraries

### Docker SDK
```go
import "github.com/docker/docker/client"

cli, err := client.NewClientWithOpts(client.FromEnv)
containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
```

### AWS SDK
```go
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/service/s3"

sess := session.Must(session.NewSession())
svc := s3.New(sess)
```

### Ork (New API)
```go
import "github.com/dracory/ork"

node := ork.NewNode("server.example.com")
output, err := node.Run("uptime")
```

All follow similar patterns: simple import, clear entry points.

## Implementation Plan

### Week 1: Core API
- Create `ork.go` with basic functions
- Implement `RunSSH` and `RunPlaybook`
- Add functional options

### Week 2: Node API
- Implement fluent `Node` type
- Add connection management
- Write tests

### Week 3: Documentation
- Update README with new examples
- Create migration guide
- Add godoc comments

### Week 4: Polish
- Add more convenience functions
- Improve error messages
- Gather feedback

## Success Metrics

- New users can run first command in <5 minutes
- 90% of use cases need only `import "github.com/dracory/ork"`
- Documentation examples are clear and concise
- Positive feedback from early adopters

## Open Questions

1. Should we provide both functional options AND builder patterns?
2. What should the default SSH key be? (currently "id_rsa")
3. Should `Node` auto-connect on first `Run()` call?
4. How to handle global configuration (default user, key, etc.)?

## Examples in README

Update the README Quick Start:

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Simple command execution
    output, err := ork.RunSSH("server.example.com", "uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
    
    // Run a playbook
    err = ork.RunPlaybook("apt-upgrade", "server.example.com",
        ork.WithPort("2222"),
        ork.WithKey("production.prv"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Or use the fluent node API
    node := ork.NewNode("server.example.com").
        Port("2222").
        Key("production.prv")
    
    node.Playbook("ping")
    node.Arg("username", "john").Playbook("user-create")
}
```

Much cleaner and more intuitive!
