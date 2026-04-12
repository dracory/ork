# Proposal: Structured Logging

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

Currently, playbooks use `log.Printf()` directly for output. This creates several issues:

- No log levels (debug, info, warn, error)
- Difficult to parse logs programmatically
- No structured data (JSON, key-value pairs)
- Can't filter or search logs effectively
- No context propagation
- Hard to integrate with log aggregation systems

## Proposed Solution

Implement structured logging with:

1. **Log levels** for filtering
2. **Structured fields** for context
3. **Multiple outputs** (console, file, remote)
4. **Contextual logging** with request IDs
5. **Performance** with minimal overhead

## Logging Library Choice

Use `slog` (Go 1.21+ standard library) for:
- Zero dependencies
- High performance
- Structured by default
- Standard library support

Alternative: `zap` or `zerolog` for more features

## Implementation

### 1. Logger Interface

```go
package logging

import (
    "log/slog"
    "os"
)

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
}

type SlogLogger struct {
    logger *slog.Logger
}

func New(level slog.Level, format string) *SlogLogger {
    var handler slog.Handler
    
    opts := &slog.HandlerOptions{Level: level}
    
    switch format {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, opts)
    default:
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return &SlogLogger{
        logger: slog.New(handler),
    }
}

func (l *SlogLogger) Debug(msg string, args ...any) {
    l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
    l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
    l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
    l.logger.Error(msg, args...)
}

func (l *SlogLogger) With(args ...any) Logger {
    return &SlogLogger{
        logger: l.logger.With(args...),
    }
}
```

### 2. Context-Aware Logging

```go
type ExecutionContext struct {
    Config Config
    Client *ssh.Client
    Logger Logger
    
    executionID string
    startTime   time.Time
}

func NewExecutionContext(cfg Config, logger Logger) *ExecutionContext {
    execID := generateExecutionID()
    
    return &ExecutionContext{
        Config:      cfg,
        Logger:      logger.With("execution_id", execID, "host", cfg.SSHHost),
        executionID: execID,
        startTime:   time.Now(),
    }
}

func (ctx *ExecutionContext) Run(cmd string) (string, error) {
    ctx.Logger.Debug("executing command",
        "command", cmd,
    )
    
    start := time.Now()
    output, err := ctx.Client.Run(cmd)
    duration := time.Since(start)
    
    if err != nil {
        ctx.Logger.Error("command failed",
            "command", cmd,
            "error", err,
            "duration", duration,
        )
        return output, err
    }
    
    ctx.Logger.Debug("command completed",
        "command", cmd,
        "duration", duration,
        "output_length", len(output),
    )
    
    return output, nil
}
```

### 3. Update Playbook Interface

```go
type Playbook interface {
    Name() string
    Description() string
    Run(config Config) error
    RunWithContext(ctx *ExecutionContext) error
}
```

## Usage Examples

### Text Output (Human-Readable)

```
time=2026-04-12T10:30:00.000Z level=INFO msg="starting playbook" playbook=apt-upgrade host=db3.example.com execution_id=abc123
time=2026-04-12T10:30:01.234Z level=DEBUG msg="executing command" command="apt-get update -y" execution_id=abc123
time=2026-04-12T10:30:05.678Z level=DEBUG msg="command completed" command="apt-get update -y" duration=4.444s execution_id=abc123
time=2026-04-12T10:30:05.679Z level=INFO msg="apt update completed" execution_id=abc123
time=2026-04-12T10:30:05.680Z level=DEBUG msg="executing command" command="apt-get upgrade -y" execution_id=abc123
time=2026-04-12T10:30:45.123Z level=DEBUG msg="command completed" command="apt-get upgrade -y" duration=39.443s execution_id=abc123
time=2026-04-12T10:30:45.124Z level=INFO msg="playbook completed" playbook=apt-upgrade duration=45.124s changed=true execution_id=abc123
```

### JSON Output (Machine-Readable)

```json
{"time":"2026-04-12T10:30:00.000Z","level":"INFO","msg":"starting playbook","playbook":"apt-upgrade","host":"db3.example.com","execution_id":"abc123"}
{"time":"2026-04-12T10:30:01.234Z","level":"DEBUG","msg":"executing command","command":"apt-get update -y","execution_id":"abc123"}
{"time":"2026-04-12T10:30:05.678Z","level":"DEBUG","msg":"command completed","command":"apt-get update -y","duration":"4.444s","execution_id":"abc123"}
{"time":"2026-04-12T10:30:45.124Z","level":"INFO","msg":"playbook completed","playbook":"apt-upgrade","duration":"45.124s","changed":true,"execution_id":"abc123"}
```

### Updated Playbook Example

```go
// Before
func (a *AptUpgrade) Run(cfg config.Config) error {
    log.Println("Running apt upgrade...")
    
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get upgrade -y")
    if err != nil {
        return fmt.Errorf("apt upgrade failed: %w\nOutput: %s", err, output)
    }
    
    log.Println("Apt upgrade completed successfully")
    return nil
}

// After
func (a *AptUpgrade) RunWithContext(ctx *ExecutionContext) error {
    ctx.Logger.Info("starting apt upgrade")
    
    output, err := ctx.Run("apt-get upgrade -y")
    if err != nil {
        ctx.Logger.Error("apt upgrade failed",
            "error", err,
            "output", output,
        )
        return fmt.Errorf("apt upgrade failed: %w", err)
    }
    
    ctx.Logger.Info("apt upgrade completed",
        "changed", true,
    )
    return nil
}
```

### Ping Playbook with Logging

```go
func (p *Ping) RunWithContext(ctx *ExecutionContext) error {
    ctx.Logger.Info("pinging server")
    
    output, err := ctx.Run("uptime")
    if err != nil {
        ctx.Logger.Error("ping failed",
            "error", err,
        )
        return fmt.Errorf("failed to ping: %w", err)
    }
    
    ctx.Logger.Info("server is alive",
        "uptime", strings.TrimSpace(output),
    )
    return nil
}
```

### UserCreate with Detailed Logging

```go
func (u *UserCreate) RunWithContext(ctx *ExecutionContext) error {
    username := ctx.Config.GetArg("username")
    if username == "" {
        ctx.Logger.Error("missing required argument", "arg", "username")
        return fmt.Errorf("username is required")
    }
    
    ctx.Logger.Info("creating user", "username", username)
    
    // Check if user exists
    _, err := ctx.Run(fmt.Sprintf("id %s", username))
    if err == nil {
        ctx.Logger.Warn("user already exists", "username", username)
        return nil
    }
    
    // Create user
    cmd := fmt.Sprintf("adduser --disabled-password --gecos '' %s", username)
    _, err = ctx.Run(cmd)
    if err != nil {
        ctx.Logger.Error("failed to create user",
            "username", username,
            "error", err,
        )
        return err
    }
    
    // Add to sudo group
    _, err = ctx.Run(fmt.Sprintf("usermod -aG sudo %s", username))
    if err != nil {
        ctx.Logger.Warn("failed to add sudo access",
            "username", username,
            "error", err,
        )
    }
    
    ctx.Logger.Info("user created successfully",
        "username", username,
        "sudo", err == nil,
    )
    return nil
}
```

## Configuration

### Via Config File

```yaml
logging:
  level: info        # debug, info, warn, error
  format: text       # text, json
  output: stdout     # stdout, stderr, file
  file: /var/log/ork.log
  
  # Additional options
  add_source: true   # Add file:line to logs
  time_format: rfc3339
```

### Via Environment Variables

```bash
export ORK_LOG_LEVEL=debug
export ORK_LOG_FORMAT=json
export ORK_LOG_FILE=/var/log/ork.log
```

### Via CLI Flags

```bash
ork run apt-upgrade --host server.example.com --log-level debug --log-format json
```

## Advanced Features

### Multiple Outputs

```go
func NewMultiLogger(outputs ...slog.Handler) *SlogLogger {
    handler := &MultiHandler{handlers: outputs}
    return &SlogLogger{
        logger: slog.New(handler),
    }
}

// Log to both console and file
logger := NewMultiLogger(
    slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
    slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
)
```

### Remote Logging

```go
// Send logs to remote syslog, Loki, etc.
type RemoteHandler struct {
    endpoint string
    client   *http.Client
}

func (h *RemoteHandler) Handle(ctx context.Context, r slog.Record) error {
    // Send log to remote endpoint
    return h.client.Post(h.endpoint, "application/json", logData)
}
```

### Sensitive Data Redaction

```go
type RedactingHandler struct {
    handler slog.Handler
    secrets []string
}

func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
    // Redact sensitive data before logging
    r.Attrs(func(a slog.Attr) bool {
        if contains(h.secrets, a.Key) {
            a.Value = slog.StringValue("***REDACTED***")
        }
        return true
    })
    return h.handler.Handle(ctx, r)
}
```

## Implementation Plan

### Phase 1: Core Logging
- Implement Logger interface with slog
- Add ExecutionContext
- Update 1-2 playbooks as examples

### Phase 2: Playbook Migration
- Update all playbooks to use structured logging
- Add context propagation
- Maintain backward compatibility

### Phase 3: Advanced Features
- Add multiple outputs
- Implement log rotation
- Add remote logging support

## Benefits

- **Searchability**: Query logs by fields
- **Debugging**: Trace execution with IDs
- **Monitoring**: Integrate with observability tools
- **Performance**: Structured logging is faster
- **Compliance**: Audit trails with detailed context

## Success Metrics

- All playbooks use structured logging
- Logs are parseable by standard tools (jq, grep)
- Zero performance degradation
- Clear documentation with examples

## Open Questions

1. Should we support log sampling for high-volume scenarios?
2. How to handle multi-line output (command results)?
3. Should we add metrics alongside logs?
4. How to handle log rotation for file outputs?
