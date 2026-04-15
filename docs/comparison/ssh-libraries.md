# Ork vs SSH Libraries & Deployment Tools

Comparison with lightweight SSH-based tools for task execution and deployment.

---

## Quick Comparison

| Tool | Language | Level | Primary Use | Architecture |
|------|----------|-------|-------------|--------------|
| **Ork** | Go | High | Server automation, skills | Library + patterns |
| **Fabric** | Python | Medium | SSH task execution | Library |
| **Paramiko** | Python | Low | SSH protocol implementation | Library |
| **Capistrano** | Ruby | High | Application deployment | Framework |

---

## Fabric

### Overview
Python library for SSH task execution. Simpler than Ansible, focused on running commands over SSH.

### Fabric Example (v2+)
```python
# fabfile.py
from fabric import task, Connection

@task
def update(c):
    """Update packages on remote server"""
    c.run("sudo apt-get update")
    c.run("sudo apt-get upgrade -y")

@task
def deploy(c):
    """Deploy application"""
    c.run("git pull origin main")
    c.run("pip install -r requirements.txt")
    c.run("systemctl restart myapp")

@task
def diskspace(c):
    """Check disk space"""
    c.run("df -h")
```

### Usage
```bash
# Run tasks
fab -H server1.example.com update
fab -H server1.example.com,server2.example.com deploy

# With different user
fab -H user@host -i /path/to/key.pem diskspace
```

### Ork vs Fabric

| Aspect | Fabric | Ork |
|--------|--------|-----|
| **Language** | ✅ Python | ✅ Go |
| **Level** | ⚠️ Command execution | ✅ Higher-level skills |
| **SSH Handling** | ✅ Built-in (Invoke+Paramiko) | ✅ Built-in (custom) |
| **Idempotency** | ❌ Manual | ✅ Built into skills |
| **Type Safety** | ❌ No | ✅ Yes (Go) |
| **Parallel** | ⚠️ Limited | ✅ Inventory (configurable concurrency) |
| **Library/CLI** | ✅ Both CLI and library | ✅ Library |

### Fabric with Multiple Hosts
```python
from fabric import SerialGroup, task

@task
def uptime(c):
    # Runs on all hosts in parallel
    cxn = SerialGroup('host1', 'host2', 'host3')
    cxn.run('uptime')
```

### Ork Equivalent
```go
// Sequential (single nodes)
hosts := []string{"host1", "host2", "host3"}
for _, host := range hosts {
    node := ork.NewNodeForHost(host)
    results := node.RunCommand("uptime")
    result := results.Results[host]
    fmt.Printf("%s: %s\n", host, result.Message)
}

// Inventory (runs on all nodes, sequential)
inv := ork.NewInventory()
webGroup := ork.NewGroup("webservers")
for _, host := range hosts {
    webGroup.AddNode(ork.NewNodeForHost(host))
}
inv.AddGroup(webGroup)
results := inv.RunCommand("uptime")
for host, result := range results.Results {
    fmt.Printf("%s: %s\n", host, result.Message)
}
```

---

## Paramiko

### Overview
Low-level Python SSH library implementing the SSHv2 protocol. What Fabric uses under the hood.

### Paramiko Example
```python
import paramiko

# Create SSH client
client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

# Connect
client.connect('server.example.com', username='root', key_filename='id_rsa')

# Execute command
stdin, stdout, stderr = client.exec_command('uptime')
print(stdout.read().decode())

# SFTP file transfer
sftp = client.open_sftp()
sftp.put('local.txt', '/remote/path.txt')
sftp.get('/remote/file.txt', 'local.txt')

# Close
client.close()
```

### Paramiko vs Ork

| Aspect | Paramiko | Ork |
|--------|----------|-----|
| **Level** | ⚠️ Very low-level | ✅ High-level |
| **Protocol** | ✅ SSH protocol implementation | ⚠️ Uses SSH library |
| **Use Case** | ⚠️ Building SSH tools | ✅ End-user automation |
| **Connection** | ❌ Manual management | ✅ Managed (persistent option) |
| **Abstractions** | ❌ None | ✅ Nodes, Skills, Results |

### Paramiko SFTP Example
```python
# Direct file operations over SSH
transport = paramiko.Transport(('server.example.com', 22))
transport.connect(username='root', pkey=my_key)

sftp = paramiko.SFTPClient.from_transport(transport)
sftp.put('/local/file.txt', '/remote/file.txt')

sftp.close()
transport.close()
```

### Ork Equivalent
```go
// Ork doesn't expose low-level SFTP
// Focus on higher-level operations
node := ork.NewNodeForHost("server.example.com")
results := node.RunCommand("cat /remote/file.txt")
result := results.Results["server.example.com"]

// Or use skills for file operations
filePb := skills.NewFileDeploy()
filePb.SetArg("source", "/local/file.txt")
filePb.SetArg("destination", "/remote/file.txt")
results = node.Run(filePb)
```

---

## Capistrano

### Overview
Ruby-based deployment framework, primarily designed for Rails applications but works with any app.

### Capistrano Structure
```
capfile
├── config/
│   └── deploy/
│       ├── production.rb
│       └── staging.rb
├── lib/
│   └── capistrano/
│       └── tasks/
│           └── custom.rake
```

### Capistrano Example
```ruby
# config/deploy.rb
set :application, 'myapp'
set :repo_url, 'git@github.com:user/repo.git'
set :deploy_to, '/var/www/myapp'
set :user, 'deploy'

# Roles
role :web, %w{web1.example.com web2.example.com}
role :db, %w{db1.example.com}

# Tasks
namespace :deploy do
  desc 'Restart application'
  task :restart do
    on roles(:web) do
      execute :systemctl, :restart, :myapp
    end
  end

  desc 'Check disk space'
  task :check_disk do
    on roles(:all) do |host|
      disk_usage = capture(:df, '-h', '/')
      info "#{host}: #{disk_usage}"
    end
  end
end
```

### Usage
```bash
# Deploy to staging
cap staging deploy

# Deploy to production
cap production deploy

# Run custom task
cap production deploy:check_disk
```

### Capistrano vs Ork

| Aspect | Capistrano | Ork |
|--------|------------|-----|
| **Language** | ✅ Ruby | ✅ Go |
| **Focus** | ⚠️ Application deployment | ✅ Server automation |
| **Structure** | ⚠️ Opinionated framework | ✅ Flexible library |
| **Roles** | ✅ Built-in | ⚠️ Planned (Inventory groups) |
| **Workflow** | ⚠️ Deploy-specific (symlink, rollback) | ✅ General automation |
| **Asset Pipeline** | ⚠️ Rails-optimized | ⚠️ Generic |

### Capistrano Deployment Flow
```
deploy:cleanup
deploy:started
  └─> git:create_release
        ├─ git:clone
        ├─ git:update
        └─ git:set_current_revision
  └─ deploy:symlink:linked_files
  └─ deploy:symlink:linked_dirs
  └─ bundle:install
  └─ deploy:migrate
  └─ deploy:symlink:release
  └─ deploy:restart
```

### Ork Equivalent
```go
// Manual deployment flow
func deploy(node ork.NodeInterface, version string) {
    // Git operations
    results := node.RunCommand("git fetch origin")
    results = node.RunCommand("git checkout " + version)

    // Install dependencies
    results = node.RunCommand("pip install -r requirements.txt")

    // Run migrations
    results = node.RunCommand("python manage.py migrate")

    // Symlink (manual)
    results = node.RunCommand("ln -sfn releases/" + version + " current")

    // Restart
    results = node.RunCommand("sudo systemctl restart myapp")
}
```

---

## Ork-Specific Features

### Vault Support

Ork includes built-in vault integration for secure secrets management using envenc, which is not available in Fabric, Paramiko, or Capistrano.

```go
// Load secrets from encrypted vault file
keys, err := ork.VaultFileToKeys("vault.envenc", "password")
if err != nil {
    log.Fatal(err)
}

// Hydrate environment variables from vault
err = ork.VaultFileToEnv("vault.envenc", "password")

// Interactive password prompt
keys, err = ork.VaultFileToKeysWithPrompt("vault.envenc")
```

### Interactive Prompts

Ork provides comprehensive prompt functions for interactive user input, which other SSH libraries lack.

```go
// Prompt for various input types
username, _ := ork.PromptForString("Username")
password, _ := ork.PromptForPasswordWithConfirmation("Password")
port, _ := ork.PromptForIntWithDefault("Port", 8080)
enabled, _ := ork.PromptForBool("Enable feature")

// Multi-prompt with validation
prompts := []types.PromptConfig{
    {Name: "email", Prompt: "Email", Required: true,
     Validate: func(v string) error {
         if !strings.Contains(v, "@") {
             return fmt.Errorf("invalid email")
         }
         return nil
     }},
}
results, _ := ork.PromptMultiple(prompts)
```

---

## Key Differences Summary

### Level of Abstraction

```
┌─────────────────────────────────────────────────────────┐
│  High-Level                                             │
│  ┌─────────┐  ┌────────────┐  ┌──────────┐            │
│  │ Ork     │  │ Capistrano │  │ Fabric   │            │
│  │ (Go)    │  │ (Ruby)     │  │ (Python) │            │
│  └─────────┘  └────────────┘  └──────────┘            │
├─────────────────────────────────────────────────────────┤
│  Low-Level                                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │ Paramiko (Python SSH implementation)             │  │
│  │ Ork/SSH internal library                         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Use Case Matrix

| Use Case | Best Tool |
|----------|-----------|
| **SSH command execution** | Fabric, Ork |
| **Build SSH tool** | Paramiko |
| **Rails deployment** | Capistrano |
| **Go application automation** | Ork |
| **Server hardening** | Ork (skills) |
| **Custom deployment workflow** | Fabric, Ork |
| **Multi-server orchestration** | Ork (Inventory), Fabric |

---

## When to Choose Each

### Use Paramiko when:
- Building custom SSH tools
- Need low-level SSH protocol control
- Implementing custom authentication
- Direct SFTP operations
- Fine-grained connection management

### Use Fabric when:
- Python shop
- Simple SSH task execution
- Ad-hoc remote commands
- Quick deployment scripts
- Don't need complex orchestration

### Use Capistrano when:
- Ruby/Rails shop
- Standardized deployment workflow
- Need rollback capability
- Rails asset pipeline
- Multi-stage deployments (staging/production)

### Use Ork when:
- Go shop
- Need type safety and compile-time checking
- Building automation into applications
- Want skill-based reusable automation
- Cross-platform server configuration
- Embedding in larger Go projects

---

## Code Comparison: Simple Task

**Task:** Check uptime on multiple servers

### Paramiko (Verbose)
```python
import paramiko

hosts = ['server1', 'server2', 'server3']
key = paramiko.RSAKey.from_private_key_file('id_rsa')

for host in hosts:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, username='root', pkey=key)

    stdin, stdout, stderr = client.exec_command('uptime')
    print(f"{host}: {stdout.read().decode()}")

    client.close()
```

### Fabric (Simpler)
```python
from fabric import SerialGroup

c = SerialGroup('server1', 'server2', 'server3', user='root', connect_kwargs={"key_filename": "id_rsa"})
results = c.run('uptime', hide=True)
for connection, result in results.items():
    print(f"{connection.host}: {result.stdout}")
```

### Capistrano (Framework)
```ruby
# lib/capistrano/tasks/uptime.rake
namespace :check do
  desc "Check uptime on all servers"
  task :uptime do
    on roles(:all) do |host|
      uptime = capture(:uptime)
      info "#{host}: #{uptime}"
    end
  end
end
```
```bash
cap production check:uptime
```

### Ork (Type-safe)
```go
hosts := []string{"server1", "server2", "server3"}

for _, host := range hosts {
    node := ork.NewNodeForHost(host).
        SetUser("root").
        SetKey("id_rsa")

    results := node.RunCommand("uptime")
    result := results.Results[host]
    if result.Error != nil {
        log.Printf("%s: error: %v", host, result.Error)
        continue
    }
    fmt.Printf("%s: %s\n", host, result.Message)
}
```

---

## Summary

| Need | Choose |
|------|--------|
| Low-level SSH control | Paramiko |
| Python task execution | Fabric |
| Rails deployment | Capistrano |
| Go server automation | Ork |
| Type safety + compile-time checks | Ork |
| Reusable skills | Ork |
| Secure secrets management (vault) | Ork |
| Interactive user input (prompts) | Ork |
| Simple, no dependencies | Fabric (Python) or Ork (Go binary) |
