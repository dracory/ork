# Proposal: Connection Pooling (Multi-Host)

**Date:** 2026-04-12  
**Status:** Rejected. Out of scope.  
**Author:** System Review

> **Note:** Single-host connection reuse via `Node.Connect()`. Multi-host pooling needed for Inventory operations.

## Problem Statement

When managing multiple hosts simultaneously, each `Node` creates its own SSH connection. For operations across fleets (10s or 100s of hosts), this can exhaust local resources (file descriptors, memory). A true connection pool would:

- Limit concurrent connections to a manageable number
- Reuse connections across multiple playbook runs
- Queue operations when connection limit reached

## Solution

For fleet management scenarios, implement a true connection pool:

```go
type ConnectionPool struct {
    maxConnections int
    connections    map[string]*pooledClient
    mu             sync.RWMutex
    semaphore      chan struct{}
}

type pooledClient struct {
    client      *ssh.Client
    lastUsed    time.Time
    inUse       bool
}

func NewConnectionPool(maxConnections int) *ConnectionPool

func (p *ConnectionPool) Acquire(host string, cfg Config) (*ssh.Client, error)
func (p *ConnectionPool) Release(host string)
func (p *ConnectionPool) CloseAll()
```

## Implementation Plan

### Phase 1: Connection Pool Core
- Implement `ConnectionPool` with semaphore-based limiting
- Add connection health checks (reuse connections < 5 min old)
- Add timeout and idle connection cleanup

### Phase 2: Integration with Parallel Execution
- Use pool in `Executor` for multi-host operations
- Queue hosts when pool is at capacity
- Track connection metrics

### Phase 3: Optimization
- Connection multiplexing via SSH ControlMaster
- Benchmark improvements

## Example Usage

```go
pool := NewConnectionPool(10) // Max 10 concurrent connections

// Run playbook on 50 hosts with only 10 concurrent connections
for _, cfg := range hostConfigs {
    client, err := pool.Acquire(cfg.SSHHost, cfg)
    if err != nil {
        return err
    }
    
    // Run commands
    client.Run("uptime")
    
    pool.Release(cfg.SSHHost)
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
**Mitigation:** Keep `Run` as default, make connection reuse opt-in

## Success Metrics

- Reduce total execution time for multi-command playbooks by >40%
- Maintain 100% backward compatibility
- Zero increase in connection-related errors

## Open Questions

1. Should connection pooling be part of core or a separate package?
2. What's the default connection timeout?
3. Should we support connection multiplexing (SSH ControlMaster)?
