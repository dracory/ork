# Ork vs SaltStack Comparison

## Quick Comparison

| Aspect | SaltStack | Ork |
|--------|-----------|-----|
| **Language** | Python | Go |
| **Architecture** | Master-Minion (agent) or Agentless (SSH) | Agentless (SSH) |
| **Execution** | Push or Pull | Push |
| **Speed** | Very fast (ZeroMQ) | Standard SSH speed |
| **Scalability** | 10,000+ nodes | Smaller scale |
| **Model** | Procedural + Declarative | Procedural |
| **Server Required** | Yes (Salt Master) | No |
| **Learning Curve** | Medium | Low |

## Architecture Options

### SaltStack Master-Minion (Default)
```
┌──────────────────┐
│   Salt Master    │
│   (pub/sub via   │
│    ZeroMQ)       │
└────────┬─────────┘
         │
    ZeroMQ (fast)
         │
    ┌────┴────┐
    │         │
┌───┴───┐ ┌───┴───┐
│Minion │ │Minion │
│Node 1 │ │Node 2 │
└───────┘ └───────┘
```

- **Salt Master** - Central server using ZeroMQ for fast communication
- **Salt Minion** - Agent running on each node
- **Extremely fast** - Can execute commands on thousands of nodes in seconds
- **Real-time** - Event-driven, immediate execution

### SaltStack Agentless (Salt SSH)
```
┌──────────────────┐
│   Salt Master    │
│   (uses SSH      │
│    instead of    │
│    ZeroMQ)       │
└────────┬─────────┘
         │ SSH
    ┌────┴────┐
    │         │
┌───┴───┐ ┌───┴───┐
│Node 1 │ │Node 2 │
│(no    │ │(no    │
│ agent)│ │ agent)│
└───────┘ └───────┘
```

- **Salt SSH** - Agentless mode using SSH
- Slower than minions but no agent installation
- Good for bootstrapping or environments where agents aren't allowed

### Ork (Pure SSH)
```
┌─────────────┐     SSH      ┌─────────────┐
│   Your Go   │─────────────►│ Target Node │
│   Program   │              │  (no agent) │
└─────────────┘              └─────────────┘
```

- No master server
- SSH to each node individually
- Simple, no infrastructure

## Speed Comparison

### SaltStack (Extremely Fast)
```bash
# Execute on 1000 nodes in under 5 seconds
salt '*' cmd.run 'uptime'

# Parallel execution via ZeroMQ
salt -G 'os:Ubuntu' state.apply
```

### Ork (Standard SSH)
```go
// Sequential execution (for now)
for _, host := range hosts {
    node := ork.NewNodeForHost(host)
    node.RunCommand("uptime")  // One at a time
}

// Inventory executes across all nodes (sequential for now, parallel planned)
```

## Configuration Model

### SaltStack (States - YAML + Jinja2)
```yaml
# /srv/salt/webserver.sls - State file
install_nginx:
  pkg.installed:
    - name: nginx

start_nginx:
  service.running:
    - name: nginx
    - enable: True
    - require:
      - pkg: install_nginx

deploy_config:
  file.managed:
    - name: /etc/nginx/nginx.conf
    - source: salt://nginx/nginx.conf
    - template: jinja
    - context:
        server_name: {{ grains['fqdn'] }}
    - watch_in:
      - service: start_nginx

create_user:
  user.present:
    - name: deploy
    - shell: /bin/bash
    - createhome: True
```

**Key concepts:**
- **States** - Declarative YAML files defining desired state
- **Grains** - Static node data (OS, CPU, etc.)
- **Pillars** - Secure data storage (variables, secrets)
- **Modules** - Functions for execution (cmd.run, pkg.install)
- **Formulas** - Pre-built state collections

### Ork (Procedural Go)
```go
// Explicit steps in Go
node := ork.NewNodeForHost("server.example.com")

// Run playbooks
node.RunPlaybook(playbooks.NewAptUpdate())
node.RunPlaybook(playbooks.NewAptUpgrade())

// Custom logic in Go
if node.GetConfig().SSHHost == "production" {
    node.RunPlaybook(playbooks.NewFail2banInstall())
}
```

## Execution Patterns

### SaltStack (Multiple Modes)

**Ad-hoc commands (imperative):**
```bash
salt 'web*' cmd.run 'systemctl restart nginx'
salt '*' test.ping
salt -G 'os:Ubuntu' pkg.upgrade
```

**State application (declarative):**
```bash
salt 'web*' state.apply webserver
salt '*' state.highstate  # Apply all states
```

**Orchestration (across multiple nodes):**
```yaml
# /srv/salt/orch/deploy.sls
deploy_app:
  salt.state:
    - tgt: 'web*'
    - sls: app.deploy
    - require:
      - salt: update_db

update_db:
  salt.state:
    - tgt: 'db*'
    - sls: db.migrate
```

### Ork (Procedural)
```go
// One node at a time
node := ork.NewNodeForHost("server.example.com")
result := node.RunPlaybook(playbooks.NewPing())

// Inventory executes across all nodes
inv := ork.NewInventory()
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
inv.AddGroup(webGroup)
results := inv.RunPlaybook(playbooks.NewPing())  // Runs on all nodes
```

## Targeting / Inventory

### SaltStack (Powerful Targeting)
```bash
# By glob
salt 'web*' test.ping

# By grain (OS, etc.)
salt -G 'os:Ubuntu' state.apply

# By pillar
salt -I 'role:webserver' cmd.run 'uptime'

# By IP range
salt -S '192.168.1.0/24' test.ping

# Compound (complex logic)
salt -C 'G@os:Ubuntu and web* or db*' state.apply
```

### Ork (Implemented)
```go
// By host (single node)
node := ork.NewNodeForHost("server.example.com")
results := node.RunPlaybook(playbooks.NewPing())

// By group
inv := ork.NewInventory()
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
inv.AddGroup(webGroup)
results := inv.GetGroupByName("webservers").RunPlaybook(playbooks.NewPing())

// Access per-node results
for host, result := range results.Results {
    fmt.Printf("%s: Changed=%v, Error=%v\n", host, result.Changed, result.Error)
}
```

## When to Choose

### Use SaltStack when:
- Managing 1000+ nodes
- Need extremely fast parallel execution
- Want both agent and agentless options
- Need powerful targeting capabilities
- Want real-time event-driven automation
- Have Python expertise
- Need reactive automation (beacons, reactors)

### Use Ork when:
- Building Go applications
- Managing smaller fleets (< 500 nodes)
- Want type safety and compile-time checking
- Embedding automation in larger projects
- Prefer simple, no-infrastructure approach
- Need programmatic control flow

## Unique SaltStack Features

**Event System:**
```bash
# React to events automatically
salt -G 'role:webserver' event.fire_master 'new_deployment'
```

**Beacons:**
```yaml
# Monitor system and react
beacons:
  service:
    - services:
        nginx: stopped
    - interval: 30
```

**Salt SSH (Agentless):**
```bash
# No minion installation needed
salt-ssh '*' test.ping
salt-ssh 'web*' state.apply
```

## Summary

**SaltStack Strengths:**
- Blazing fast (ZeroMQ)
- Massive scale (10K+ nodes)
- Flexible targeting
- Both agent and agentless modes
- Rich module ecosystem

**Ork Strengths:**
- Simple, no infrastructure
- Go-native, type-safe
- Embeddable in applications
- Explicit control flow
- Easy to understand

**Key Difference:**
- SaltStack: Enterprise-grade speed and scale, requires infrastructure
- Ork: Lightweight, embeddable, focused on simplicity
