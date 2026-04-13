# Ork vs Ansible Comparison

## Quick Comparison

| Aspect | Ansible | Ork |
|--------|---------|-----|
| **Language** | YAML + Jinja2 | Go |
| **Architecture** | Agentless (SSH) | Agentless (SSH) |
| **Execution** | Push | Push |
| **Inventory** | Static files (INI/YAML) | Go structs (programmatic) |
| **Automation Unit** | Playbooks | Playbooks |
| **State Model** | Procedural | Procedural |
| **Idempotency** | Task-level | Playbook-level |
| **Server Required** | No | No |
| **Learning Curve** | Low | Low (Go knowledge) |

## Architecture & Execution

Both Ansible and Ork use the same fundamental architecture:

- **Agentless** - No software installed on target nodes
- **SSH-based** - Commands executed over SSH
- **Push model** - Control node initiates all operations
- **On-demand** - Runs when invoked, not continuously

### Ansible
- Control node pushes commands via SSH
- Inventory files define target hosts
- Runs from your machine or CI/CD

### Ork
- Similar SSH-based approach
- Compiled Go binary (no Python dependency)
- Can be embedded in Go applications
- Programmatic inventory (structs)

## Inventory Management

### Ansible Inventory (INI)
```ini
[webservers]
web1.example.com ansible_port=2222 ansible_user=deploy
web2.example.com

[dbservers]
db1.example.com

[production:children]
webservers
dbservers
```

### Ansible Inventory (YAML)
```yaml
all:
  children:
    webservers:
      hosts:
        web1.example.com:
          ansible_port: 2222
          ansible_user: deploy
        web2.example.com:
      vars:
        env: production
```

### Ork Inventory (Implemented)
```go
// Programmatic creation
inv := ork.NewInventory()

// Create group and add nodes
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com").
    SetPort("2222").
    SetUser("deploy"))
webGroup.SetArg("env", "production")
inv.AddGroup(webGroup)

// Run playbook on entire inventory
results := inv.RunPlaybook(playbooks.NewPing())
summary := results.Summary()
fmt.Printf("Changed: %d, Failed: %d\n", summary.Changed, summary.Failed)
```

## Automation Units

### Ansible Playbook
```yaml
- name: Configure web servers
  hosts: webservers
  become: yes
  vars:
    app_version: "1.2.3"
  
  tasks:
    - name: Install nginx
      apt:
        name: nginx
        state: present
    
    - name: Start nginx service
      service:
        name: nginx
        state: started
        enabled: yes
    
    - name: Deploy application
      template:
        src: app.conf.j2
        dest: /etc/app/config.conf
      notify: restart app
  
  handlers:
    - name: restart app
      service:
        name: myapp
        state: restarted
```

### Ork Playbook
```go
// Run a single playbook
node := ork.NewNodeForHost("server.example.com")
results := node.RunPlaybook(playbooks.NewPing())

// Get result for this specific node
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatal(result.Error)
}

if result.Changed {
    log.Printf("Changes made: %s", result.Message)
}

// Chain configuration
node.SetPort("2222").
    SetUser("deploy").
    SetArg("version", "1.2.3")

result = node.RunPlaybook(playbooks.NewAptUpgrade())
```

## Idempotency

### Ansible
- Modules handle idempotency internally
- `changed_when` for custom tasks
- Handlers trigger only on change

```yaml
- name: Create user
  user:
    name: deploy
    state: present
  # Only creates if doesn't exist

- name: Run custom script
  script: setup.sh
  changed_when: "'already configured' not in output"
```

### Ork
- Built into playbooks via `CheckPlaybook()` method (via RunnableInterface)
- `Result.Changed` field indicates if change occurred
- Works on Node, Group, and Inventory uniformly

```go
// Check if changes needed before running
ping := playbooks.NewPing()
results := node.CheckPlaybook(ping)

result := results.Results["server.example.com"]
if result.Changed {
    log.Printf("Changes would be made: %s", result.Message)
}

// Also works on groups and inventory
webServers := inv.GetGroupByName("webservers")
results := webServers.CheckPlaybook(ping)
```

## Configuration Patterns

### Variable Precedence (Ansible)
1. Command line
2. Task variables
3. Host variables
4. Group variables
5. Inventory variables
6. Role defaults

### Variable Precedence (Ork)
1. Playbook-level args (`SetArg()`)
2. Node-level args
3. Group args (via `SetArg()`)
4. Inventory-level args
5. Node defaults

## Extensibility

| Feature | Ansible | Ork |
|---------|---------|-----|
| **Custom Modules** | Python | Go (compile in) |
| **Module Repository** | Ansible Galaxy | Built-in + custom |
| **Templating** | Jinja2 | Go templates/text/template |
| **Secrets** | Ansible Vault | User implements |
| **Callbacks** | Custom callback plugins | Go interfaces |

## When to Choose Each

### Choose Ansible when:
- You prefer YAML configuration
- Team has ops background, no coding experience
- Quick ad-hoc commands needed frequently
- Existing Ansible ecosystem (roles, Galaxy)
- Need mature, battle-tested solution

### Choose Ork when:
- Building Go applications
- Need programmatic/automated workflows
- Want type safety and compile-time checking
- Embedding automation in larger Go projects
- Prefer explicit code over YAML

## Summary

**Similarities:**
- Both SSH-based and agentless
- Both push-based execution
- Both procedural (run tasks in order)
- Both support inventory concepts

**Differences:**
- Ansible: YAML-centric, mature ecosystem
- Ork: Go-native, compile-time safety, embeddable
