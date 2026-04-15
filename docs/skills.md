# Skills

Skills are reusable automation tasks that can be run against nodes, groups, or inventory. They implement the `RunnableInterface` and provide idempotent operations for common server management tasks.

## Running Skills

### On a Single Node

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

results := node.Run(skills.NewUserCreate())
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatalf("Skill failed: %v", result.Error)
}
if result.Changed {
    log.Printf("User created: %s", result.Message)
} else {
    log.Println("User already exists - no changes made")
}
```

### On a Group

```go
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

results := webGroup.Run(skills.NewAptUpdate())

// Check results for all nodes
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    } else {
        log.Printf("%s: %s (changed: %v)", host, result.Message, result.Changed)
    }
}
```

### On Inventory

```go
inv := ork.NewInventory()
inv.AddGroup(webGroup)

results := inv.Run(skills.NewPing())

// Get summary
summary := results.Summary()
log.Printf("Total: %d, Changed: %d, Failed: %d", 
    summary.Total, summary.Changed, summary.Failed)
```

## Available Skills

### System Skills

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| Ping | `ping` | - | Check SSH connectivity |
| Reboot | `reboot` | - | Reboot server |
| Apt Update | `apt-update` | - | Refresh package database |
| Apt Upgrade | `apt-upgrade` | - | Install available updates |
| Apt Status | `apt-status` | - | Show available updates |

### User Management

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| User Create | `user-create` | `username` | Create user with sudo |
| User Delete | `user-delete` | `username` | Delete user |
| User Status | `user-status` | `username` (optional) | Show user info |

### Swap Management

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| Swap Create | `swap-create` | `size` (GB) | Create swap file |
| Swap Delete | `swap-delete` | - | Remove swap file |
| Swap Status | `swap-status` | - | Show swap status |

### Security

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| AIDE Install | `security-aide-install` | - | Install AIDE IDS |
| Auditd Install | `security-auditd-install` | - | Install audit daemon |
| Kernel Harden | `security-kernel-harden` | - | Apply kernel hardening |
| SSH Change Port | `security-ssh-change-port` | `port` | Change SSH port |
| SSH Harden | `security-ssh-harden` | - | Harden SSH config |

### Firewall (UFW)

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| UFW Install | `ufw-install` | - | Install UFW firewall |
| UFW Status | `ufw-status` | - | Show UFW status |
| UFW Allow MariaDB | `ufw-allow-mariadb` | - | Allow MariaDB through UFW |

### Fail2Ban

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| Fail2ban Install | `fail2ban-install` | - | Install and configure fail2ban |
| Fail2ban Status | `fail2ban-status` | - | Show fail2ban status |

### MariaDB

| Skill | ID | Args | Description |
|-------|-----|------|-------------|
| MariaDB Install | `mariadb-install` | - | Install MariaDB server |
| MariaDB Status | `mariadb-status` | - | Show MariaDB status |
| MariaDB Secure | `mariadb-secure` | - | Secure MariaDB installation |
| MariaDB Backup | `mariadb-backup` | `database` (optional) | Backup database |
| MariaDB Backup Encrypt | `mariadb-backup-encrypt` | - | Encrypted backup |
| MariaDB Change Port | `mariadb-change-port` | `port` | Change MariaDB port |
| MariaDB Create DB | `mariadb-create-db` | `database` | Create database |
| MariaDB Create User | `mariadb-create-user` | `username`, `password` | Create DB user |
| MariaDB Enable Encryption | `mariadb-enable-encryption` | - | Enable encryption at rest |
| MariaDB Enable SSL | `mariadb-enable-ssl` | - | Enable SSL connections |
| MariaDB List DBs | `mariadb-list-dbs` | - | List databases |
| MariaDB List Users | `mariadb-list-users` | - | List DB users |
| MariaDB Security Audit | `mariadb-security-audit` | - | Run security audit |

## Setting Arguments

Many skills accept arguments to customize their behavior:

```go
node := ork.NewNodeForHost("server.example.com")

// Set single argument
node.SetArg("username", "alice")

// Set multiple arguments
node.SetArg("username", "alice").
    SetArg("shell", "/bin/bash").
    SetArg("environment", "production")

// Run skill with arguments
results := node.Run(skills.NewUserCreate())
```

## Running by ID

You can also run skills by their string ID:

```go
// Run by ID (registry lookup)
results := node.RunByID("ping")
```

## Checking Before Running

Use `Check()` to preview if changes would be made (idempotency):

```go
// Check if changes would be made
results := node.Check(skills.NewAptUpgrade())
result := results.Results["server.example.com"]

if result.Changed {
    log.Println("Would upgrade packages")
    // Now actually run it
    results = node.Run(skills.NewAptUpgrade())
} else {
    log.Println("No upgrades needed")
}
```

## Result Structure

Results are returned as `types.Results` with per-node access:

```go
type Results struct {
    Results map[string]Result  // Key is node hostname
}

func (r Results) Summary() Summary

type Result struct {
    Changed bool              // Whether changes were made
    Message string            // Human-readable description
    Details map[string]string // Additional information
    Error   error             // Non-nil if execution failed
}

type Summary struct {
    Total     int
    Changed   int
    Unchanged int
    Failed    int
}
```

## Custom Skills

For complex orchestration logic, see [Playbooks](playbooks.md). For creating custom skills, see [Advanced Usage](advanced_usage.md#custom-skills).
