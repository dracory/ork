# Proposal: Connection Pooling and Reuse

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

Currently, every SSH command execution creates a new connection via `ssh.RunOnce()`. For playbooks that execute multiple commands sequentially, this creates unnecessary overhead:

- Multiple TCP handshakes
- Repeated SSH authentication
- Increased latency (especially for remote servers)
- Higher resource usage

Example from `apt-status` playbook:
```go
// Two separate connections for related operations
ssh.RunOnce(..., "apt-get update -qq")
ssh.RunOnce(..., "apt list --upgradable")
```

## Proposed Solution

### 1. Enhance Playbook Interface (Optional)

Add optional connection lifecycle methods:

```go
type ConnectionAwarePlaybook interface {
    Playbook
    UsePersistentConnection() bool
}
```

### 2. Add Connection Context

```go
type ExecutionContext struct {
    Config Config
    Client *ssh.Client // Reusable connection
}

func (ctx *ExecutionContext) Run(cmd string) (string, error) {
    if ctx.Client == nil {
        return "", fmt.Errorf("no active connection")
    }
    return ctx.Client.Run(cmd)
}
```

### 3. Update Playbook Interface

```go
type Playbook interface {
    Name() string
    Description() string
    Run(config Config) error
    RunWithContext(ctx *ExecutionContext) error // New method
}
```

### 4. Connection Pool Manager

```go
type ConnectionPool struct {
    connections map[string]*ssh.Client
    mu          sync.RWMutex
}

func (p *ConnectionPool) Get(cfg Config) (*ssh.Client, error)
func (p *ConnectionPool) Release(cfg Config)
func (p *ConnectionPool) CloseAll()
```

## Implementation Plan

### Phase 1: Backward Compatible Enhancement
- Keep existing `RunOnce` for simple use cases
- Add `Client.Run()` method for multiple commands
- Update documentation with examples

### Phase 2: Playbook Migration
- Update existing playbooks to use persistent connections where beneficial
- Add benchmarks showing performance improvements

### Phase 3: Connection Pool (Optional)
- Implement connection pooling for multi-host scenarios
- Add connection timeout and health checks

## Example Usage

### Before:
```go
func (a *AptStatus) Run(cfg config.Config) error {
    _, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get update -qq")
    if err != nil {
        return err
    }
    
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt list --upgradable")
    return err
}
```

### After:
```go
func (a *AptStatus) Run(cfg config.Config) error {
    client := ssh.NewClient(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey)
    if err := client.Connect(); err != nil {
        return err
    }
    defer client.Close()
    
    if _, err := client.Run("apt-get update -qq"); err != nil {
        return err
    }
    
    output, err := client.Run("apt list --upgradable")
    return err
}
```

## Benefits

- **Performance**: Reduce connection overhead by 50-80% for multi-command playbooks
- **Reliability**: Fewer connection attempts = fewer failure points
- **Resource Efficiency**: Lower CPU and memory usage
- **Backward Compatible**: Existing code continues to work

## Risks & Mitigation

**Risk:** Connection timeouts on long-running operations  
**Mitigation:** Add configurable keepalive and timeout settings

**Risk:** Connection state issues between commands  
**Mitigation:** Add connection health checks before each command

**Risk:** Breaking changes for existing playbooks  
**Mitigation:** Keep `RunOnce` as default, make connection reuse opt-in

## Success Metrics

- Reduce total execution time for multi-command playbooks by >40%
- Maintain 100% backward compatibility
- Zero increase in connection-related errors

## Open Questions

1. Should connection pooling be part of core or a separate package?
2. What's the default connection timeout?
3. Should we support connection multiplexing (SSH ControlMaster)?
