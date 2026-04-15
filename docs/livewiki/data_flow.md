---
path: data_flow.md
page-type: reference
summary: Detailed data flow diagrams showing how information moves through the Ork system.
tags: [data-flow, diagrams, internals]
created: 2025-04-14
updated: 2026-04-15
version: 2.1.0
---

## Changelog
- **v2.1.0** (2026-04-15): Updated terminology from playbooks to skills, RunPlaybook to Run, config package to types package
- **v1.0.0** (2025-04-14): Initial creation

# Data Flow

This document describes how data flows through the Ork system during various operations.

## Node Creation Flow

```mermaid
sequenceDiagram
    User->>+NewNodeForHost: host string
    NewNodeForHost->>NewNodeForHost: Set defaults
    Note over NewNodeForHost: SSHPort: "22"<br/>RootUser: "root"<br/>SSHKey: "id_rsa"
    NewNodeForHost->>+nodeImplementation: Create struct
    nodeImplementation-->>-NewNodeForHost: &nodeImplementation
    NewNodeForHost-->>-User: NodeInterface
```

### Configuration Flow

```mermaid
graph LR
    A[User] -->|NewNodeForHost| B[Node with Defaults]
    A -->|SetPort| C[Update Port]
    A -->|SetUser| D[Update User]
    A -->|SetKey| E[Update Key]
    A -->|SetArg| F[Update Args]
    B --> G[types.NodeConfig]
    C --> G
    D --> G
    E --> G
    F --> G
```

## Command Execution Flow

### RunCommand (One-Time Connection)

```mermaid
sequenceDiagram
    User->>+Node: RunCommand("uptime")
    
    alt IsDryRunMode
        Node->>Logger: Info "dry-run: would run command"
        Node-->>User: Results{[dry-run]}
    else Normal Execution
        alt Has Persistent Connection
            Node->>sshClient: Run("uptime")
            sshClient->>simplessh: Exec(cmd)
            simplessh->>Remote: SSH Execute
            Remote-->>simplessh: Output
            simplessh-->>sshClient: Output
            sshClient-->>Node: Output
        else No Persistent Connection
            Node->>ssh: Run(cfg, "uptime")
            ssh->>ssh: NewClient
            ssh->>ssh: Connect()
            ssh->>simplessh: ConnectWithKeyFile
            simplessh->>Remote: SSH Handshake
            Remote-->>simplessh: Connected
            simplessh-->>ssh: Client
            ssh->>simplessh: Exec("uptime")
            simplessh->>Remote: Execute
            Remote-->>simplessh: Output
            simplessh-->>ssh: Output
            ssh->>ssh: Close()
            ssh-->>Node: Output
        end
        
        Node->>Node: Build Result
        Node-->>User: Results{host: Result}
    end
```

### Persistent Connection Flow

```mermaid
sequenceDiagram
    User->>+Node: Connect()
    Node->>ssh: NewClient(host, port, user, key)
    ssh-->>Node: Client
    Node->>ssh: Connect()
    ssh->>simplessh: ConnectWithKeyFile(addr, user, keyPath)
    simplessh->>Remote: SSH Connection
    Remote-->>simplessh: Established
    simplessh-->>ssh: Client
    ssh-->>Node: nil (success)
    Node-->>-User: nil (success)
    Note over Node: connected = true
    
    User->>Node: RunCommand("uptime")
    Node->>ssh: Run("uptime")
    ssh->>simplessh: Exec
    simplessh->>Remote: Execute
    Remote-->>simplessh: Output
    simplessh-->>ssh: Output
    ssh-->>Node: Output
    Node-->>User: Results
    
    User->>+Node: Close()
    Node->>ssh: Close()
    ssh->>simplessh: Close()
    simplessh->>Remote: Disconnect
    simplessh-->>ssh: nil
    ssh-->>Node: nil
    Node-->>-User: nil
    Note over Node: connected = false
```

## Skill Execution Flow

### Direct Skill Execution

```mermaid
sequenceDiagram
    User->>+Node: Run(skill)
    
    Node->>Skill: SetConfig(node.cfg)
    Note over Node,Skill: Copy NodeConfig to skill
    
    Node->>Skill: SetDryRun(node.IsDryRunMode)
    
    alt Dry Run Mode
        Skill->>Skill: Check if changes needed
        Skill->>Logger: Log planned actions
        Skill->>Skill: Return Result{Changed: true/false}
    else Normal Execution
        Skill->>Skill: Check()
        Skill->>ssh: Run(check command)
        ssh->>Remote: Execute
        Remote-->>ssh: Output
        ssh-->>Skill: Output
        
        alt Changes Needed
            Skill->>ssh: Run(apply command)
            ssh->>Remote: Execute
            Remote-->>ssh: Output
            ssh-->>Skill: Output
            Skill->>Skill: Build Result{Changed: true}
        else No Changes Needed
            Skill->>Skill: Build Result{Changed: false}
        end
    end
    
    Skill-->>Node: Result
    Node->>Node: Wrap in Results map
    Node-->>-User: Results{host: Result}
```

### Registry-Based Skill Execution

```mermaid
sequenceDiagram
    User->>+Node: RunByID("apt-update")
    
    Node->>Registry: PlaybookFindByID("apt-update")
    Registry->>Registry: Lookup in map
    
    alt Found
        Registry-->>Node: RunnableInterface, true
        Node->>Skill: SetConfig(node.cfg)
        Node->>Skill: SetDryRun(node.IsDryRunMode)
        Node->>Skill: Run()
        Skill-->>Node: Result
        Node-->>User: Results{host: Result}
    else Not Found
        Registry-->>Node: nil, false
        Node-->>User: Results{Error: "not found"}
    end
```

## Group Execution Flow

```mermaid
sequenceDiagram
    User->>+Group: Run(skill)
    
    Group->>Group: propagateDryRun()
    loop For each node
        Group->>Node: SetDryRunMode(group.dryRunMode)
    end
    
    Group->>Group: Create empty Results
    
    loop For each node
        Group->>Node: Run(skill)
        Node->>Node: Execute
        Node-->>Group: NodeResults
        Group->>Group: Merge into Results
    end
    
    Group-->>User: Aggregated Results
```

## Inventory Execution Flow

```mermaid
sequenceDiagram
    User->>+Inventory: Run(skill)
    
    Inventory->>Inventory: Collect all nodes from groups
    Inventory->>Inventory: Apply maxConcurrency limit
    
    par Concurrent Execution
        loop For each node (up to maxConcurrency)
            Inventory->>Node: Run(skill)
            Node->>Node: Execute
            Node-->>Inventory: Results
            Inventory->>Inventory: Aggregate
        end
    end
    
    Inventory-->>User: All Results
```

## Dry-Run Mode Propagation

```mermaid
graph TD
    A[User sets DryRun] --> B{Level}
    B -->|Inventory| C[inv.SetDryRunMode]
    B -->|Group| D[group.SetDryRunMode]
    B -->|Node| E[node.SetDryRunMode]
    
    C --> F[Store in Inventory]
    C --> G[Propagate to Groups]
    
    G --> H[group.SetDryRunMode]
    H --> I[Store in Group]
    H --> J[Propagate to Nodes]
    
    J --> K[node.SetDryRunMode]
    K --> L[Store in Node]
    
    D --> I
    E --> L
    
    L --> M[types.NodeConfig.IsDryRunMode]
    
    M --> N{Execution Time}
    N -->|RunCommand| O{IsDryRunMode?}
    N -->|Run| P{IsDryRunMode?}
    
    O -->|Yes| Q[Log & Return marker]
    O -->|No| R[Execute on Server]
    
    P -->|Yes| S[Log & Return dry-run Result]
    P -->|No| T[Execute Skill]
```

## Check Mode Flow

```mermaid
sequenceDiagram
    User->>+Node: Check(skill)
    
    Node->>Skill: SetDryRun(node.IsDryRunMode)
    
    Note over Node,Skill: Check mode does NOT<br/>automatically enable dry-run
    
    Skill->>Skill: Run()
    
    alt Skill implements Check properly
        Skill->>Skill: Check() internally
        Skill->>ssh: Run(check command)
        ssh->>Remote: Execute
        Remote-->>ssh: Output
        ssh-->>Skill: Output
        Skill->>Skill: Determine if changes needed
        Skill-->>Node: Result{Changed: true/false}
    else Skill just runs
        Skill->>ssh: Run(apply command)
        ssh->>Remote: Execute
        Remote-->>ssh: Output
        ssh-->>Skill: Output
        Skill-->>Node: Result
    end
    
    Node-->>User: Results{host: Result}
```

## Results Aggregation Flow

```mermaid
graph TD
    A[Operation on Multiple Nodes] --> B[Create Results Map]
    B --> C[Execute on Node 1]
    B --> D[Execute on Node 2]
    B --> E[Execute on Node 3]
    
    C --> F[Result 1]
    D --> G[Result 2]
    E --> H[Result 3]
    
    F --> I[results.Results["node1"] = Result 1]
    G --> J[results.Results["node2"] = Result 2]
    H --> K[results.Results["node3"] = Result 3]
    
    I --> L[Results Struct]
    J --> L
    K --> L
    
    L --> M[Summary]
    M --> N[Total Count]
    M --> O[Changed Count]
    M --> P[Unchanged Count]
    M --> Q[Failed Count]
```

## SSH Connection Flow (Detailed)

```mermaid
sequenceDiagram
    participant User
    participant ssh.Client
    participant simplessh
    participant Remote
    
    User->>ssh.Client: NewClient(host, port, user, key)
    ssh.Client->>ssh.Client: Resolve key path
    ssh.Client->>ssh.Client: Set defaults (port 22)
    ssh.Client-->>User: &Client
    
    User->>ssh.Client: Connect()
    ssh.Client->>ssh.Client: Validate host not empty
    ssh.Client->>ssh.Client: Build addr (host:port)
    ssh.Client->>simplessh: ConnectWithKeyFile(addr, user, keyPath)
    
    simplessh->>Remote: TCP Connection
    Remote-->>simplessh: TCP Established
    
    simplessh->>Remote: SSH Handshake
    Remote-->>simplessh: SSH Session
    
    simplessh->>Remote: Key Authentication
    Remote-->>simplessh: Auth Success
    
    simplessh-->>ssh.Client: *simplessh.Client
    ssh.Client-->>User: nil
    
    User->>ssh.Client: Run(cmd)
    ssh.Client->>simplessh: Exec(cmd)
    simplessh->>Remote: Execute command
    Remote-->>simplessh: stdout/stderr
    simplessh-->>ssh.Client: []byte output
    ssh.Client-->>User: string output
    
    User->>ssh.Client: Close()
    ssh.Client->>simplessh: Close()
    simplessh->>Remote: Disconnect
    simplessh-->>ssh.Client: nil
    ssh.Client-->>User: nil
```

## Configuration Inheritance Flow

```mermaid
graph TD
    A[User sets Arg] --> B{Level}
    
    B -->|Inventory| C[inv.SetArg?]
    Note over C: Inventory doesn't store args<br/>directly - uses groups
    
    B -->|Group| D[group.SetArg]
    D --> E[group.args map]
    E --> F[Inherited by nodes?]
    Note over F: No - args are NOT<br/>automatically propagated
    
    B -->|Node| G[node.SetArg]
    G --> H[node.cfg.Args map]
    
    I[Skill Execution] --> J{Arg Source}
    J -->|Node Level| K[GetArg from node.cfg.Args]
    J -->|Skill Level| L[GetArg from skill.cfg.Args]
    
    M[Node to Skill] --> N[skill.SetConfig node.cfg]
    N --> O[Skill copies Args map]
    O --> P[Skill can override args]
```

## Error Handling Flow

```mermaid
sequenceDiagram
    User->>Node: RunCommand("invalid")
    
    alt Connection Error
        Node->>ssh: Connect
        ssh->>simplessh: ConnectWithKeyFile
        simplessh->>Remote: TCP Connect
        Remote-->>simplessh: Connection Refused
        simplessh-->>ssh: error
        ssh-->>Node: error
        Node->>Node: Build Result
        Node-->>User: Result{Error: wrapped error}
    else Command Error
        Node->>ssh: Run
        ssh->>Remote: Execute
        Remote-->>ssh: Exit code 1
        ssh-->>Node: error
        Node->>Node: Build Result
        Node-->>User: Result{Error: command failed}
    else Success
        Node->>ssh: Run
        ssh->>Remote: Execute
        Remote-->>ssh: Output
        ssh-->>Node: Output
        Node->>Node: Build Result
        Node-->>User: Result{Message: output}
    end
```

## See Also

- [Architecture](architecture.md) - High-level architecture
- [API Reference](api_reference.md) - API documentation
- [Configuration](configuration.md) - Configuration options
