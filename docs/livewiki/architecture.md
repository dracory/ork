---
path: architecture.md
page-type: reference
summary: System architecture, design patterns, and key architectural decisions in Ork.
tags: [architecture, design, patterns]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# Ork Architecture

This document describes the architecture of Ork, including design patterns, component relationships, and key architectural decisions.

## System Overview

```mermaid
graph TB
    subgraph "User Application"
        A[User Code]
    end
    
    subgraph "Ork Framework"
        B[ork Package]
        C[NodeInterface]
        D[GroupInterface]
        E[InventoryInterface]
        F[RunnableInterface]
    end
    
    subgraph "Core Components"
        G[config Package]
        H[ssh Package]
        I[playbook Package]
        J[types Package]
    end
    
    subgraph "Playbook Implementations"
        K[playbooks/apt]
        L[playbooks/user]
        M[playbooks/swap]
        N[playbooks/mariadb]
        O[playbooks/security]
        P[playbooks/ufw]
        Q[playbooks/fail2ban]
        R[playbooks/ping]
        S[playbooks/reboot]
    end
    
    subgraph "External"
        T[SSH Server]
        U[Remote Server]
    end
    
    A --> B
    B --> C
    B --> D
    B --> E
    C --> F
    D --> F
    E --> F
    C --> G
    C --> H
    C --> I
    I --> K
    I --> L
    I --> M
    I --> N
    I --> O
    I --> P
    I --> Q
    I --> R
    I --> S
    H --> T
    T --> U
```

## Layered Architecture

Ork follows a layered architecture pattern:

### 1. API Layer (ork Package)

The public API that users interact with:

```go
// NodeInterface - Single server management
type NodeInterface interface {
    RunnableInterface
    GetHost() string
    SetPort(port string) NodeInterface
    Connect() error
    Close() error
    // ...
}

// GroupInterface - Server group management  
type GroupInterface interface {
    RunnableInterface
    GetName() string
    AddNode(node NodeInterface) GroupInterface
    // ...
}

// InventoryInterface - Multi-group management
type InventoryInterface interface {
    RunnableInterface
    AddGroup(group GroupInterface) InventoryInterface
    SetMaxConcurrency(max int) InventoryInterface
    // ...
}
```

### 2. Core Services Layer

#### SSH Package

Handles all SSH connectivity:

```mermaid
graph LR
    A[ssh.Client] --> B[Connect]
    A --> C[Run]
    A --> D[Close]
    E[ssh.RunOnce] --> F[simplessh]
    A --> F
```

Key features:
- Connection pooling via persistent connections
- Dry-run mode support
- Key-based authentication

#### Config Package

Central configuration management:

```go
type NodeConfig struct {
    SSHHost      string
    SSHPort      string
    SSHLogin     string
    SSHKey       string
    RootUser     string
    NonRootUser  string
    DBPort       string
    DBRootPassword string
    Args         map[string]string
    Logger       *slog.Logger
    IsDryRunMode bool
}
```

### 3. Playbook System Layer

#### Playbook Interface

All automation tasks implement this interface:

```go
type PlaybookInterface interface {
    GetID() string
    GetDescription() string
    SetConfig(cfg config.NodeConfig) PlaybookInterface
    GetArg(key string) string
    SetArg(key, value string) PlaybookInterface
    Check() (bool, error)
    Run() Result
}
```

#### Base Playbook

Provides common functionality:

```mermaid
classDiagram
    class BasePlaybook {
        -id string
        -description string
        -config NodeConfig
        -args map[string]string
        -dryRun bool
        -timeout Duration
        +GetID() string
        +SetID(id string) PlaybookInterface
        +GetDescription() string
        +SetDescription(desc string) PlaybookInterface
        +GetConfig() NodeConfig
        +SetConfig(cfg NodeConfig) PlaybookInterface
        +GetArg(key string) string
        +SetArg(key, value string) PlaybookInterface
    }
    
    class PlaybookInterface {
        <<interface>>
        +Check() (bool, error)
        +Run() Result
    }
    
    BasePlaybook ..|> PlaybookInterface
```

## Design Patterns

### 1. Fluent Interface (Method Chaining)

Configuration uses fluent API for readability:

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv").
    SetArg("env", "production")
```

### 2. Repository Pattern (Registry)

Playbook registry for ID-based lookup:

```mermaid
graph LR
    A[Registry] --> B[Register Playbook]
    C[Node] --> D[Run by ID]
    D --> A
    A --> E[Find by ID]
```

### 3. Strategy Pattern

Different playbooks implement the same interface:

```mermaid
graph TB
    A[PlaybookInterface] --> B[AptUpdate]
    A --> C[UserCreate]
    A --> D[SwapCreate]
    A --> E[MariaDBInstall]
    A --> F[SecurityHarden]
```

### 4. Composite Pattern

Inventory/Group/Node hierarchy:

```mermaid
graph TD
    A[Inventory] --> B[Web Group]
    A --> C[DB Group]
    B --> D[Node 1]
    B --> E[Node 2]
    C --> F[Node 3]
    C --> G[Node 4]
```

All implement `RunnableInterface` with unified execution.

### 5. Factory Pattern

Node creation methods:

```go
// Factory methods
func NewNodeForHost(host string) NodeInterface
func NewNode() NodeInterface
func NewNodeFromConfig(cfg config.NodeConfig) NodeInterface
func NewGroup(name string) GroupInterface
func NewInventory() InventoryInterface
```

## Concurrency Model

### Inventory-Level Concurrency

```mermaid
sequenceDiagram
    User->>Inventory: RunPlaybook(pb)
    Inventory->>Inventory: Collect all nodes
    par Concurrent execution
        Inventory->>Node1: RunPlaybook(pb)
        Inventory->>Node2: RunPlaybook(pb)
        Inventory->>Node3: RunPlaybook(pb)
    end
    Inventory->>User: Aggregate Results
```

Configurable via `SetMaxConcurrency()`.

### Thread Safety

Key thread-safe mechanisms:

```go
// Group uses mutex for dry-run mode
type groupImplementation struct {
    // ...
    dryRunMode bool
    mu         sync.RWMutex
}

func (g *groupImplementation) SetDryRunMode(dryRun bool) RunnableInterface {
    g.mu.Lock()
    g.dryRunMode = dryRun
    g.mu.Unlock()
    // Propagate to nodes
    g.propagateDryRun()
    return g
}
```

## Data Flow

### Command Execution Flow

```mermaid
sequenceDiagram
    User->>Node: RunCommand("uptime")
    Node->>NodeConfig: Check IsDryRunMode
    alt Dry Run
        Node->>Logger: Log "would run command"
        Node->>User: Return [dry-run] marker
    else Normal Execution
        alt Has Persistent Connection
            Node->>SSH Client: Run(cmd)
        else One-Time Connection
            Node->>SSH: RunOnce(host, port, user, key, cmd)
        end
        SSH->>Remote: Execute command
        Remote->>SSH: Return output
        SSH->>Node: Return output
        Node->>User: Return Results
    end
```

### Playbook Execution Flow

```mermaid
sequenceDiagram
    User->>Node: RunPlaybook(pb)
    Node->>Playbook: SetConfig(node.cfg)
    Node->>Playbook: SetDryRun(node.IsDryRunMode)
    
    alt Dry Run
        Playbook->>Logger: Log planned actions
        Playbook->>Playbook: Return [dry-run] Result
    else Normal Execution
        Playbook->>SSH: Run(command)
        SSH->>Remote: Execute
        Remote->>SSH: Return output
        SSH->>Playbook: Return output
        Playbook->>Playbook: Build Result
    end
    
    Playbook->>Node: Return Result
    Node->>Node: Wrap in Results
    Node->>User: Return Results
```

## Idempotency Design

All playbooks follow the Check-Run pattern:

```mermaid
graph TD
    A[Playbook Run] --> B{Check}
    B -->|Changes Needed| C[Execute Changes]
    B -->|No Changes| D[Return Unchanged]
    C --> E[Return Changed]
    D --> F[Result]
    E --> F
```

Example implementation:

```go
func (p *MyPlaybook) Check() (bool, error) {
    // Check current state
    output, _ := ssh.Run(cfg, "check command")
    return !isAlreadyConfigured(output), nil
}

func (p *MyPlaybook) Run() Result {
    needsChange, _ := p.Check()
    if !needsChange {
        return Result{Changed: false, Message: "Already configured"}
    }
    
    // Apply changes
    _, err := ssh.Run(cfg, "apply command")
    if err != nil {
        return Result{Changed: false, Error: err}
    }
    
    return Result{Changed: true, Message: "Changes applied"}
}
```

## Error Handling Strategy

### Result-Based Error Handling

```go
type Result struct {
    Changed bool
    Message string
    Details map[string]string
    Error   error  // Non-nil if execution failed
}
```

Errors bubble up with context:

```go
output, err := ssh.Run(cfg, cmd)
if err != nil {
    return playbook.Result{
        Changed: false,
        Message: "Operation failed",
        Error:   fmt.Errorf("failed to execute '%s': %w", cmd, err),
    }
}
```

## Security Architecture

### SSH Security

- Key-based authentication only
- Private keys stored in ~/.ssh/
- No password authentication in framework

### Dry-Run Safety

```mermaid
graph TD
    A[SetDryRunMode] --> B[Propagate to Nodes]
    B --> C[Propagate to Groups]
    C --> D[Propagate to Inventory]
    E[Run Command] --> F{IsDryRunMode?}
    F -->|Yes| G[Log & Return marker]
    F -->|No| H[Execute on Server]
```

Safety enforced at execution layer - no way to bypass.

## Extension Points

### Custom Playbooks

```go
type MyPlaybook struct {
    *playbook.BasePlaybook
}

func (p *MyPlaybook) Check() (bool, error) { ... }
func (p *MyPlaybook) Run() playbook.Result { ... }

// Register globally
registry, err := ork.GetGlobalPlaybookRegistry()
if err != nil {
    log.Fatal(err)
}
registry.PlaybookRegister(myPlaybook)
```

### Custom SSH Logic

```go
// sshRunOnce is a variable that can be mocked
type sshRunOnce = func(host, port, user, key, cmd string) (string, error)
```

## See Also

- [Data Flow](data_flow.md) - Detailed data flow diagrams
- [Configuration](configuration.md) - Configuration options
- [API Reference](api_reference.md) - Complete API documentation
- [Modules](modules/ork.md) - Package-level documentation
