# Ork vs Puppet Comparison

## Quick Comparison

| Aspect | Puppet | Ork |
|--------|--------|-----|
| **Language** | Ruby (DSL) | Go |
| **Architecture** | Agent-based (master/agent) | Agentless (SSH) |
| **Execution** | Pull | Push |
| **Model** | Declarative | Procedural |
| **Server Required** | Yes (Puppet Master) | No |
| **State Management** | Catalog + Reports | No state |
| **Idempotency** | Built into resources | Playbook-level |
| **Learning Curve** | Steep | Low |

## Architecture

### Puppet (Master-Agent Model)
```
┌──────────────────┐
│  Puppet Master   │◄──────┐
│  (Compiles       │       │
│   catalogs)      │       │
└──────────────────┘       │
       ▲                   │ HTTPS
       │                   │ Pull every 30 min
       │            ┌──────┴──────┐
       │            │             │
┌──────┴─────┐  ┌───┴───┐  ┌─────┴──┐
│ Puppet DB  │  │ Agent │  │ Agent  │
│ (Reports)  │  │Node 1 │  │ Node 2 │
└────────────┘  └───────┘  └────────┘
```

- **Puppet Master** - Central server that compiles and serves configuration "catalogs"
- **Puppet Agent** - Runs on each node, pulls catalog every 30 minutes
- **PuppetDB** - Stores reports, facts, and exported resources

### Ork (SSH-based, Agentless)
```
┌─────────────┐     SSH      ┌─────────────┐
│   Your Go   │─────────────►│ Target Node │
│   Program   │              │  (no agent) │
└─────────────┘              └─────────────┘
```

- No master server
- No agents on nodes
- SSH connections initiated on demand

## Configuration Model

### Puppet (Declarative DSL)
```puppet
# site.pp - Define desired state
node 'webserver01.example.com' {
  package { 'nginx':
    ensure => 'installed',
  }

  service { 'nginx':
    ensure => 'running',
    enable => true,
    require => Package['nginx'],
  }

  file { '/etc/nginx/nginx.conf':
    ensure  => 'file',
    content => template('nginx/nginx.conf.erb'),
    notify  => Service['nginx'],
  }

  user { 'deploy':
    ensure     => 'present',
    managehome => true,
    shell      => '/bin/bash',
  }
}
```

**Key concepts:**
- **Resources** - Declarative units (package, service, file, user)
- **Catalog** - Compiled desired state for a node
- **Manifests** - Files containing Puppet code (.pp)
- **Templates** - ERB templates for dynamic content
- **Facter** - System for gathering node facts

### Ork (Procedural Go)
```go
// Explicit execution
node := ork.NewNodeForHost("webserver01.example.com")

// Install nginx
results := node.RunPlaybook(playbooks.NewAptInstall())

// Configure user
userPb := playbooks.NewUserCreate()
userPb.SetArg("username", "deploy")
userPb.SetArg("shell", "/bin/bash")
results = node.RunPlaybook(userPb)

// Direct command execution
results = node.RunCommand("sudo systemctl enable nginx")
```

## Execution Flow

### Puppet
1. Agent collects facts about the node
2. Agent sends facts to Master
3. Master compiles catalog (what should be)
4. Agent receives catalog
5. Agent applies catalog (converges to desired state)
6. Agent sends report back to Master
7. Repeats every 30 minutes

### Ork
1. Your Go program starts
2. SSH connection established
3. Commands/playbooks executed
4. Results returned
5. Connection closed (unless persistent)

## Idempotency

### Puppet (Resource-level)
```puppet
# First run: creates user, installs nginx
# Second run: verifies state, makes no changes
# Third run: verifies state, makes no changes

package { 'nginx':
  ensure => installed,  # Idempotent: checks if installed first
}

file { '/etc/config':
  ensure  => file,
  content => "...",
  # Only updates if content differs
}
```

### Ork (Playbook-level)
```go
// Check pattern via RunnableInterface
ping := playbooks.NewPing()
results := node.CheckPlaybook(ping)
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Would make changes: %s", result.Message)
}
```

## Key Concepts Comparison

| Puppet | Ork |
|--------|-----|
| Manifest (.pp) | Go program |
| Resource | Playbook |
| Facter | Direct configuration |
| Catalog | No equivalent (no state) |
| Class/Module | Package of playbooks |
| Node definition | Node instantiation |
| Hiera (data) | Go structs / config |

## Feature Comparison Table

| Feature | Puppet | Ork | Notes |
|---------|--------|-----|-------|
| **Architecture** | ✅ Master-Agent (centralized) | ✅ Agentless (SSH) | Puppet requires master/agent; Ork uses SSH |
| **Execution Model** | ✅ Pull (every 30 min) | ✅ Push (on-demand) | Puppet runs continuously; Ork runs when invoked |
| **Server Required** | ✅ Yes (Puppet Master) | ✅ No | Puppet needs master server; Ork is standalone |
| **Parallel Execution** | ✅ Native (master compiles catalogs) | ✅ Configurable concurrency | Puppet parallel at master level; Ork via SetMaxConcurrency() |
| **State Model** | ✅ Declarative (catalogs) | ✅ Procedural | Puppet declares desired state; Ork explicit execution |
| **Idempotency** | ✅ Built-in (resources) | ✅ Playbook-level | Both support idempotent operations |
| **Secrets Management** | ✅ Hiera (encrypted) | ✅ envenc vault | Puppet has Hiera for data; Ork uses envenc |
| **Inventory** | ✅ Node registry (PuppetDB) | ✅ Programmatic (structs) | Puppet nodes registered; Ork explicit creation |
| **Configuration Language** | ✅ Ruby (DSL) | ✅ Go | Puppet uses Ruby DSL; Ork uses Go |
| **Package Management** | ✅ Cross-platform (resource providers) | ⚠️ Platform-specific playbooks | Puppet handles OS differences; Ork needs playbooks per OS |
| **Templates** | ✅ ERB templates | ✅ Go templates | Both support templating |
| **Facts Gathering** | ✅ Built-in (Facter) | ❌ Manual commands | Puppet auto-collects facts; Ork manual |
| **Catalog Compilation** | ✅ Master compiles for each node | ❌ No catalog | Puppet compiles desired state; Ork executes directly |
| **Reporting** | ✅ PuppetDB (built-in) | ⚠️ Manual (results map) | Puppet has reporting database; Ork returns results |
| **Compliance** | ✅ Built-in (PuppetDB reports) | ⚠️ Manual (via Check) | Puppet tracks compliance; Ork on-demand checks |
| **Module Ecosystem** | ✅ Puppet Forge (community) | ⚠️ Built-in + custom | Puppet has large module repository |
| **Role/Profile Pattern** | ✅ Built-in pattern | ⚠️ Manual (groups) | Puppet has role/profile concept; Ork uses groups |
| **Type Safety** | ❌ No | ✅ Yes (Go) | Ork has compile-time type checking |
| **Learning Curve** | ⚠️ Steep (DSL + Puppet concepts) | ✅ Low (Go knowledge) | Puppet requires DSL knowledge |
| **Scalability** | ✅ 100+ nodes | ⚠️ Smaller scale | Puppet designed for medium-large fleets |
| **Continuous Enforcement** | ✅ Yes (agent runs) | ❌ No (on-demand) | Puppet enforces continuously; Ork on-demand |
| **Drift Detection** | ✅ Built-in (PuppetDB) | ⚠️ Manual (via Check) | Puppet detects drift via reports; Ork on-demand |

## When to Choose

### Use Puppet when:
- Managing 100+ nodes
- Need continuous enforcement
- Want declarative resource management
- Have Ruby expertise
- Can invest in Puppet Master infrastructure
- Need compliance reporting
- Want role/profile pattern

### Use Ork when:
- Managing smaller fleets (< 100 nodes)
- Need on-demand execution
- Want type safety and compile-time checking
- Embedding automation in Go applications
- Prefer explicit control over declarative
- No infrastructure for master server

## Summary

**Puppet Philosophy:**
- "Define desired state, let system converge"
- Continuous enforcement via agents
- Heavyweight but powerful for scale
- Full visibility via PuppetDB reports

**Ork Philosophy:**
- "Execute commands when I say so"
- On-demand SSH connections
- Lightweight, no infrastructure needed
- Go-native with type safety

**Key Difference:**
- Puppet: Background agents continuously ensure state
- Ork: Foreground execution, explicit control
