# Proposal: Structured Logging

**Date:** 2026-04-12  
**Status:** Not Implemented  
**Author:** System Review

> **Note:** Replace `log.Printf()` with `slog` throughout the codebase.

## Problem Statement

Currently, playbooks use `log.Printf()` directly for output. This creates several issues:

- No log levels (debug, info, warn, error)
- Difficult to parse logs programmatically
- No structured data (JSON, key-value pairs)
- Can't filter or search logs effectively

## Proposed Solution

Use `slog` (Go 1.21+ standard library) directly. No custom wrapper.

## Implementation

### 1. RunnableInterface Logger

All runnable types (Node, Group, Inventory) have logging via `RunnableInterface`:

```go
type RunnableInterface interface {
    RunCommand(cmd string) types.Results
    RunPlaybook(pb playbook.PlaybookInterface) types.Results
    RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results
    CheckPlaybook(pb playbook.PlaybookInterface) types.Results
    
    // GetLogger returns the logger. Returns slog.Default() if not set.
    GetLogger() *slog.Logger
    
    // SetLogger sets a custom logger. Returns self for chaining.
    SetLogger(logger *slog.Logger) RunnableInterface
}
```

### 2. Default Behavior

- Default: `slog.Default()` (outputs to stderr, text format)
- No configuration needed for basic usage
- Playbooks use `node.GetLogger().Info()`, etc.

### 3. Custom Logger Example

```go
// JSON logging to file
logFile, _ := os.Create("/var/log/ork.log")
logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Single node
node := ork.NewNodeForHost("server.example.com").
    SetLogger(logger)

// Group
webGroup := ork.NewGroup("webservers").
    SetLogger(logger)

// Inventory
inv := ork.NewInventory().
    SetLogger(logger)

// Or use default everywhere
node := ork.NewNodeForHost("server.example.com")
// Uses slog.Default()
```

### 4. Playbook Usage

Logger is accessed via `NodeConfig.Logger`:

```go
func (a *AptUpgrade) Run(cfg config.NodeConfig) error {
    cfg.Logger.Info("starting apt upgrade", "host", cfg.SSHHost)
    
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get upgrade -y")
    if err != nil {
        cfg.Logger.Error("apt upgrade failed", "error", err)
        return fmt.Errorf("apt upgrade failed: %w", err)
    }
    
    cfg.Logger.Info("apt upgrade completed", "output", len(output))
    return nil
}
```

### 5. NodeConfig Update

Add `Logger` field to `config.NodeConfig`:

```go
type NodeConfig struct {
    // SSH connection settings
    SSHHost  string
    SSHPort  string
    SSHLogin string
    SSHKey   string
    
    // ... other fields ...
    
    // Logger for structured logging. Defaults to slog.Default().
    Logger *slog.Logger
}
```

## Implementation Plan

### Phase 1: Add Logger to RunnableInterface and NodeConfig
- Add `Logger *slog.Logger` to `config.NodeConfig`
- Add `GetLogger()` and `SetLogger()` to `RunnableInterface`
- Implement methods on `nodeImplementation`, `groupImplementation`, `inventoryImplementation`
- Default to `slog.Default()`

### Phase 2: Replace log.Printf
- Find all `log.Printf` calls in playbooks
- Replace with `cfg.Logger.Info()`, `cfg.Logger.Error()`, etc.

## Benefits

- **Simple**: Just use `slog` directly
- **Standard**: No custom interfaces to learn
- **Flexible**: Full power of `slog` (JSON, levels, handlers)
- **Zero overhead**: No wrapper abstraction

## Migration Example

```go
// Before
import "log"
log.Printf("Running apt upgrade on %s", cfg.SSHHost)

// After
import "log/slog"
slog.Info("running apt upgrade", "host", cfg.SSHHost)
```
