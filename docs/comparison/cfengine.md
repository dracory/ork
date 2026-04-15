# Ork vs CFEngine Comparison

## Quick Comparison

| Aspect | CFEngine | Ork |
|--------|----------|-----|
| **Language** | C (agent) + CFEngine DSL | Go |
| **Architecture** | Agent-based (C binaries) | Agentless (SSH) |
| **Execution** | Pull | Push |
| **Speed** | Very fast (C agent) | Standard SSH speed |
| **Resource Usage** | Extremely lightweight | Connection-based |
| **Model** | Declarative (Promise Theory) | Procedural |
| **Server Required** | Optional (CFEngine Hub) | No |
| **Learning Curve** | Steep (unique concepts) | Low |
| **History** | First config management tool (1993) | Modern Go approach |

## Architecture

### CFEngine (C-based Agent)
```
┌──────────────────┐     ┌──────────────────┐
│ CFEngine Hub     │◄────│  CFEngine Agent  │
│ (Policy Hub -    │     │  (C binary, runs  │
│  optional)       │     │   every 5 min)   │
└──────────────────┘     └──────────────────┘
         │                        │
         │                        │
    ┌────┴────────────────────────┘
    │         Pull policies
    │         Verify promises
    │         Converge if needed
    │
┌───┴─────────────────────────────┐
│ CFEngine Agent (Node 2)       │
│ - Lightweight C binary (~1MB) │
│ - No dependencies               │
│ - Self-healing                  │
└─────────────────────────────────┘
```

- **CFEngine Agent** - Written in C, extremely lightweight (~1MB)
- **Promise Theory** - Unique declarative model based on "promises"
- **Autonomous** - Agents run independently, converge continuously
- **No dependencies** - Single static binary

### Ork (SSH-based, Agentless)
```
┌─────────────┐     SSH      ┌─────────────┐
│   Your Go   │─────────────►│ Target Node │
│   Program   │              │  (no agent) │
└─────────────┘              └─────────────┘
```

- No agent installation
- SSH connections on demand
- Go-based implementation

## Unique Philosophy: Promise Theory

CFEngine is based on **Promise Theory** (invented by Mark Burgess), a fundamentally different approach:

```cfengine3
# promises.cf - CFEngine's declarative language
body common control {
    bundlesequence => { "main" };
}

bundle agent main {
    files:
        "/etc/nginx/nginx.conf"
            create => "true",
            edit_template => "/var/cfengine/templates/nginx.conf.tmpl",
            template_method => "mustache",
            classes => if_repaired("restart_nginx");

    services:
        "nginx"
            service_policy => "start",
            service_enable => "true",
            if => classmatch("restart_nginx");

    packages:
        "nginx"
            policy => "present";
}
```

**Key Promise Theory concepts:**
- **Promises** - Agent makes promises about desired state
- **Convergence** - System naturally converges to promised state
- **Autonomy** - Each agent acts independently
- **No central coordination needed** - Agents self-heal

Compare to Ork's procedural approach:

```go
// Ork: Explicit execution
node := ork.NewNodeForHost("server.example.com")

// Check via RunnerInterface, then act
nginxInstall := skills.NewAptInstall()
results := node.Check(nginxInstall)
result := results.Results["server.example.com"]

if result.Changed {
    results = node.Run(nginxInstall)
}

// Direct command execution
results = node.RunCommand("sudo systemctl restart nginx")
result = results.Results["server.example.com"]
```

## Model Comparison

### CFEngine (Promise-based)
```cfengine3
# Agent continuously ensures these promises are kept
bundle agent manage_user {
    vars:
        "user_name" string => "deploy";
        "user_shell" string => "/bin/bash";

    users:
        "$(user_name)"
            policy => "present",
            shell => "$(user_shell)",
            home_dir => "/home/$(user_name)",
            manage_home => "true";
}
```

- **Agent runs every 5 minutes** by default
- **Promises are continuously verified**
- **No explicit execution order**
- **Convergent by design**

### Ork (Procedural)
```go
// Your code decides when and what to run
node := ork.NewNodeForHost("server.example.com")

if shouldCreateUser {
    userPb := skills.NewUserCreate()
    userPb.SetArg("username", "deploy")
    results := node.Run(userPb)
    result := results.Results["server.example.com"]

    if result.Error != nil {
        log.Fatal(result.Error)
    }
}
```

- **Execute when you call it**
- **Explicit control flow**
- **Sequential execution**
- **Error handling in Go**

## Resource Efficiency

### CFEngine (Extremely Lightweight)
```
CFEngine Agent:
- Binary size: ~1 MB
- Memory: ~5-10 MB
- CPU: Minimal (sleeping most of time)
- Boot time: Milliseconds
- Dependencies: None (static binary)
```

**Ideal for:**
- Embedded systems
- IoT devices
- Containers (minimal overhead)
- Environments with strict resource constraints

### Ork (Standard)
```
Ork:
- Binary size: Depends on Go build
- Memory: Per-connection
- CPU: During SSH operations
- No background process (no agent)
```

**Characteristics:**
- SSH connection overhead
- No persistent agent
- Resource usage only during execution

## Speed Comparison

### CFEngine (Fast Convergence)
```
Agent Check Cycle (~5 seconds):
1. Wake up (every 5 min by default)
2. Read promises
3. Verify state (C code, very fast)
4. Apply if needed
5. Sleep
```

- C implementation = very fast execution
- Local operations only (no SSH)
- Continuous verification

### Ork (SSH Latency)
```go
// Each operation requires SSH connection setup
node := ork.NewNodeForHost("server.example.com")

// SSH handshake + authentication
result := node.Run(skills.NewPing())

// New command = reuse connection or new handshake
output, _ := node.RunCommand("uptime")
```

- SSH connection overhead
- Network latency dependent
- On-demand execution

## Configuration Distribution

### CFEngine (Policy Distribution)
```
┌──────────────┐
│ Policy Hub   │──► Distributes promises
│ (optional)   │    to all agents
└──────────────┘
         │
         ▼
   ┌─────────────┐
   │ Git/CM repo │──► Can pull policies from Git
   └─────────────┘
```

**Options:**
1. **Policy Hub** - Central distribution server
2. **Git pull** - Agents pull from Git repository
3. **Local files** - Policies on each node

```cfengine3
# Masterfiles (policy files) distributed to all agents
body common control {
    inputs => { "libraries/cfengine_stdlib.cf", "services/*.cf" };
}
```

### Ork (Embedded or External)
```go
// Policies embedded in Go code
func configureServer(node ork.NodeInterface) {
    node.Run(skills.NewAptUpdate())
    node.Run(skills.NewFail2banInstall())
}

// Or load from external source
cfg := loadConfigFromGit("git@github.com:org/config.git")
```

- Go code is the policy
- Can load from any source (Git, DB, API)
- Compiled binary contains logic

## When to Choose

### Use CFEngine when:
- Managing 10,000+ nodes (proven at massive scale)
- Need extreme resource efficiency
- Want autonomous agents (no central coordination)
- Environment has bandwidth constraints
- Running on embedded/IoT devices
- Value Promise Theory model
- Want 30+ years of battle-tested reliability

**CFEngine users:** LinkedIn, AT&T, Verizon (massive scale)

### Use Ork when:
- Building Go applications
- Want type safety and compile-time checking
- Need programmatic control flow
- Embedding automation in larger projects
- Prefer SSH-based simplicity
- Managing smaller to medium fleets
- Want on-demand execution

## Unique CFEngine Features

**Knowledge Management:**
```cfengine3
# Built-in CMDB-like functionality
bundle agent inventory {
    vars:
        "os" string => canonify("$(sys.flavor)");
        "ipv4[eth0]" string => "$(sys.ipv4[eth0])";

    reports:
        "System $(sys.fqdn) runs $(os)"
            inform => "true";
}
```

**Compliance Reporting:**
- Built-in compliance dashboard (Enterprise)
- Automated drift detection
- Historical state tracking

**Mission-Critical Reliability:**
- Designed for "never touch" systems
- Self-healing even without connectivity
- Minimal attack surface (small C binary)

## Feature Comparison Table

| Feature | CFEngine | Ork | Notes |
|---------|----------|-----|-------|
| **Architecture** | ✅ Agent-based (C binary) | ✅ Agentless (SSH) | CFEngine has local agent; Ork uses SSH |
| **Execution Model** | ✅ Pull (continuous) | ✅ Push (on-demand) | CFEngine runs continuously; Ork runs when invoked |
| **Parallel Execution** | ✅ Native (agent-local) | ✅ Configurable concurrency | CFEngine parallel per-agent; Ork via SetMaxConcurrency() |
| **Resource Usage** | ✅ Extremely lightweight (~1MB) | ⚠️ Connection-based | CFEngine minimal overhead; Ork uses SSH connections |
| **Speed** | ✅ Very fast (C code) | ⚠️ Standard SSH speed | CFEngine local execution; Ork network latency |
| **State Model** | ✅ Declarative (Promise Theory) | ✅ Procedural | CFEngine promises; Ork explicit execution |
| **Idempotency** | ✅ Built-in (promises) | ✅ Skill-level | Both support idempotent operations |
| **Secrets Management** | ❌ Manual | ✅ envenc vault | Ork has built-in vault support |
| **Configuration Distribution** | ✅ Policy Hub / Git | ⚠️ Go code / external | CFEngine has distribution system; Ork uses Go |
| **Scalability** | ✅ 10,000+ nodes | ⚠️ Smaller scale | CFEngine proven at massive scale |
| **Server Required** | ⚠️ Optional (Policy Hub) | ✅ No | CFEngine can run without central server |
| **Type Safety** | ❌ No | ✅ Yes (Go) | Ork has compile-time type checking |
| **Learning Curve** | ⚠️ Steep (unique concepts) | ✅ Low (Go knowledge) | CFEngine Promise Theory is unique |
| **Compliance Reporting** | ✅ Built-in (Enterprise) | ❌ Manual | CFEngine has enterprise reporting |
| **Drift Detection** | ✅ Built-in (continuous) | ⚠️ Manual (via Check) | CFEngine detects continuously; Ork on-demand |

## Summary

**CFEngine Philosophy:**
- "Make promises, let agents converge autonomously"
- Extreme efficiency and reliability
- Set up once, runs forever
- Unique Promise Theory approach

**Ork Philosophy:**
- "Execute commands when needed"
- Simple, explicit, type-safe
- Go-native, embeddable
- No infrastructure required

**Key Differences:**
- CFEngine: C-based, extremely lightweight, continuous convergence
- Ork: Go-based, SSH-based, on-demand execution

**Historical Note:**
CFEngine (1993) predates most config management tools:
- CFEngine: 1993
- Puppet: 2005
- Chef: 2009
- Ansible: 2012
- SaltStack: 2011

CFEngine invented the category and remains the most resource-efficient option.
