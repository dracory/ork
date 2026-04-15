# Ork vs Chef Comparison

## Quick Comparison

| Aspect | Chef | Ork |
|--------|------|-----|
| **Language** | Ruby | Go |
| **Architecture** | Agent-based | Agentless (SSH) |
| **Execution** | Pull | Push |
| **Inventory** | Chef Server registry | Go structs (programmatic) |
| **Automation Unit** | Cookbooks + Recipes | Playbooks |
| **State Model** | Declarative / Convergent | Procedural |
| **Idempotency** | Built into resources | Playbook-level |
| **Server Required** | Yes (Chef Server) | No |
| **Learning Curve** | Steep | Low (Go knowledge) |

## Architecture & Execution

### Chef (Agent-based, Pull)
```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ Chef Server │◄────│ Chef Client  │     │ Chef Client │
│  (API/DB)   │     │  (on node)   │     │  (on node)  │
└─────────────┘     └──────────────┘     └─────────────┘
        ▲                    │                  │
        │                    ▼                  ▼
        │            Runs every 30 min   Runs every 30 min
        │            Pulls cookbooks     Pulls cookbooks
        │            Converges state     Converges state
        │
┌───────┴───────┐
│ Workstation   │
│ (knife, edit  │
│  cookbooks)   │
└───────────────┘
```

- **Chef Server** - Central API storing cookbooks, node data, policies
- **Chef Client** - Agent installed on each node
- **Workstation** - Developer machine for authoring cookbooks

### Ork (Agentless, Push)
```
┌─────────────┐     SSH      ┌─────────────┐
│   Your Go   │─────────────►│ Target Node │
│   Program   │              │  (no agent) │
│             │─────────────►│             │
└─────────────┘              └─────────────┘
        │
        │ SSH to any node
        │
   ┌────┴────┐
   │ Node 2  │
   │ Node 3  │
   │ Node N  │
   └─────────┘
```

- **No agent** - Pure SSH connection
- **Push model** - Your program initiates operations
- **On-demand** - Runs when you invoke it

## Inventory Management

### Chef Nodes (Dynamic Registry)
- No static inventory files
- Nodes self-register with Chef Server
- Query nodes via search

```bash
# Search for nodes
knife search node 'role:webserver AND chef_environment:production'

# Node attributes stored centrally
knife node show web1.example.com -F json
```

```ruby
# In a recipe, target nodes dynamically
search(:node, 'role:webserver').each do |web_node|
  # Configure each web server
end
```

### Ork Inventory (Implemented)
```go
// Programmatic creation
inv := ork.NewInventory()

// Create and add group
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
inv.AddGroup(webGroup)

// Target specific group
results := inv.GetGroupByName("webservers").RunPlaybook(playbooks.NewPing())
summary := results.Summary()
```

## Automation Units

### Chef Cookbook Structure
```
cookbooks/
└── myapp/
    ├── recipes/
    │   ├── default.rb      # Main recipe
    │   └── configure.rb    # Sub-recipe
    ├── templates/
    │   └── config.erb      # ERB templates
    ├── files/
    │   └── script.sh       # Static files
    ├── attributes/
    │   └── default.rb      # Default values
    └── metadata.rb         # Cookbook metadata
```

### Chef Recipe (Declarative)
```ruby
# cookbooks/nginx/recipes/default.rb
# Declares desired state, not execution steps

package 'nginx' do
  action :install
end

service 'nginx' do
  action [:enable, :start]
end

template '/etc/nginx/sites-enabled/default' do
  source 'site.erb'
  variables(
    server_name: node['fqdn'],
    port: 80
  )
  notifies :restart, 'service[nginx]', :delayed
end

# Chef decides HOW to achieve this state
# Different on Ubuntu vs CentOS vs Windows
```

### Ork Playbook (Procedural)
```go
// Explicit steps executed in order
ping := playbooks.NewPing()
results := node.RunPlaybook(ping)

// Access result for specific node
result := results.Results["server.example.com"]
if result.Error != nil {
    return err
}

// Chain multiple playbooks
results = node.RunPlaybook(playbooks.NewAptUpdate())
results = node.RunPlaybook(playbooks.NewAptUpgrade())
```

## State Model

### Chef (Declarative / Convergent)
- **Declare desired state** - "Nginx should be installed and running"
- **Resources handle implementation** - Chef knows how to install nginx on different OSes
- **Convergent** - Repeated runs only fix drift
- **Periodic execution** - Chef client runs every 30 minutes by default

```ruby
# Same recipe works on Ubuntu, CentOS, Windows
package 'nginx'  # Chef picks apt, yum, or msi
```

### Ork (Procedural)
- **Execute explicit steps** - "Run apt update, then apt upgrade"
- **You define the implementation** - Write Go code for each step
- **Run once** - Executed when you call it
- **On-demand** - No background processes

```go
// You control exactly what runs and when
results := node.RunPlaybook(playbooks.NewAptUpgrade())
result := results.Results["server.example.com"]
```

## Idempotency

### Chef (Resource-level)
```ruby
# Chef resources are inherently idempotent
package 'nginx' do
  action :install
end
# First run: installs nginx
# Second run: checks, already installed, does nothing

# Idempotency built into resource providers
template '/etc/config' do
  source 'config.erb'
  # Only updates if template content changed
end
```

### Ork (Playbook-level)
```go
// Check pattern via RunnableInterface
aptUpgrade := playbooks.NewAptUpgrade()

// Check if upgrade needed - works on Node, Group, or Inventory
results := node.CheckPlaybook(aptUpgrade)
result := results.Results["server.example.com"]

if result.Changed {
    log.Printf("Packages would be updated")
}
```

## Key Concepts Comparison

| Chef Concept | Ork Equivalent |
|--------------|----------------|
| Cookbook | Package of related playbooks |
| Recipe | Individual playbook |
| Resource | Playbook implementation |
| Attribute | Node configuration / Args |
| Role | Group with shared playbooks |
| Data Bag | Configuration data |
| Environment | Group-level variables |
| Node | Node |
| Knife CLI | Go API / future CLI |

## Feature Comparison Table

| Feature | Chef | Ork | Notes |
|---------|------|-----|-------|
| **Architecture** | ✅ Agent-based (Chef Client) | ✅ Agentless (SSH) | Chef requires agent installation; Ork uses SSH |
| **Execution Model** | ✅ Pull (every 30 min) | ✅ Push (on-demand) | Chef runs continuously; Ork runs when invoked |
| **Server Required** | ✅ Yes (Chef Server) | ✅ No | Chef needs central server; Ork is standalone |
| **Parallel Execution** | ✅ Native (Chef Server) | ✅ Configurable concurrency | Chef parallel via server; Ork via SetMaxConcurrency() |
| **State Model** | ✅ Declarative / Convergent | ✅ Procedural | Chef declares desired state; Ork explicit execution |
| **Idempotency** | ✅ Built-in (resources) | ✅ Playbook-level | Both support idempotent operations |
| **Secrets Management** | ✅ Chef Vault / Data Bags | ✅ envenc vault | Chef has encrypted data bags; Ork uses envenc |
| **Inventory** | ✅ Dynamic (node registry) | ✅ Programmatic (structs) | Chef nodes self-register; Ork explicit creation |
| **Configuration Language** | ✅ Ruby (DSL) | ✅ Go | Chef uses Ruby DSL; Ork uses Go |
| **Package Management** | ✅ Cross-platform (resource providers) | ⚠️ Platform-specific playbooks | Chef handles OS differences; Ork needs playbooks per OS |
| **Templates** | ✅ ERB templates | ✅ Go templates | Both support templating |
| **Search/Query** | ✅ Built-in search API | ❌ Manual filtering | Chef can query node attributes; Ork manual |
| **Role Management** | ✅ Built-in roles | ⚠️ Manual (groups) | Chef has role system; Ork uses groups |
| **Environments** | ✅ Built-in environments | ⚠️ Group-level args | Chef has environment concept; Ork uses args |
| **Data Bags** | ✅ Built-in data storage | ⚠️ Go structs / config | Chef has data bags; Ork uses Go structures |
| **Cookbook Ecosystem** | ✅ Supermarket (community) | ⚠️ Built-in + custom | Chef has large cookbook repository |
| **Test Framework** | ✅ Test Kitchen, InSpec | ⚠️ Go testing + testcontainers | Chef has mature testing tools |
| **Compliance** | ✅ InSpec (built-in) | ❌ Manual | Chef has compliance scanning |
| **Type Safety** | ❌ No | ✅ Yes (Go) | Ork has compile-time type checking |
| **Learning Curve** | ⚠️ Steep (Ruby + Chef concepts) | ✅ Low (Go knowledge) | Chef requires Ruby and DSL knowledge |
| **Scalability** | ✅ Hundreds/thousands of nodes | ⚠️ Smaller scale | Chef designed for large fleets |
| **Continuous Enforcement** | ✅ Yes (client runs) | ❌ No (on-demand) | Chef enforces continuously; Ork on-demand |

## When to Choose Each

### Choose Chef when:
- Managing hundreds/thousands of nodes
- Need continuous enforcement (not just on-demand)
- Want declarative "desired state" model
- Team has Ruby expertise
- Need complete lifecycle management
- Can invest in Chef Server infrastructure

### Choose Ork when:
- Building Go applications
- Need programmatic/automated workflows
- Want type safety and compile-time checking
- Embedding automation in larger projects
- Prefer SSH-based, agentless approach
- Need simple, lightweight solution

## Summary

**Chef Philosophy:**
- Set up once, converge continuously
- Nodes self-heal to desired state
- Heavy infrastructure, powerful for scale

**Ork Philosophy:**
- Run when needed, explicit control
- Simple SSH-based execution
- Lightweight, embeddable, type-safe

**Fundamental Difference:**
- Chef: "Ensure the system is always in this state" (background agent)
- Ork: "Run this task now" (on-demand execution)
