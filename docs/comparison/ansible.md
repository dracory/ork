# Ork vs Ansible Comparison

## Quick Comparison

| Aspect | Ansible | Ork |
|--------|---------|-----|
| **Language** | YAML + Jinja2 | Go |
| **Architecture** | Agentless (SSH) | Agentless (SSH) |
| **Execution** | Push | Push |
| **Inventory** | Static files (INI/YAML) | Go structs / YAML (planned) |
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
- Programmatic inventory (structs) or YAML

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

### Ork Inventory (Planned)
```go
// Programmatic creation
inv := ork.NewInventory()
webGroup := inv.AddGroup("webservers")
webGroup.AddNode("web1.example.com").
    SetPort("2222").
    SetUser("deploy")
webGroup.SetVar("env", "production")

// Or load from YAML
inv, _ := ork.NewInventoryFromYAML("inventory.yaml")
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
result := node.RunPlaybook(playbooks.NewPing())

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
- Built into playbooks via `Check()` method
- `Result.Changed` field indicates if change occurred
- Explicit pattern: check → decide → execute

```go
// Check if changes needed before running
ping := playbooks.NewPing()
needsChange, _ := ping.Check(cfg)

if needsChange {
    result := node.RunPlaybook(ping)
    // result.Changed tells if modifications were made
}
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
3. Group variables (planned)
4. Inventory variables (planned)
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
