# Quick Start

The core concept is the **Node** - a representation of a remote server:

```go
package main

import (
    "log"
    "github.com/dracory/ork"
    "github.com/dracory/ork/config"
)

func main() {
    // Create a node (remote server) - multiple ways:
    
    // Option 1: From host (most common)
    node := ork.NewNodeForHost("server.example.com")
    
    // Option 2: From config (useful for complex setups)
    cfg := config.NodeConfig{
        SSHHost: "server.example.com",
        SSHPort: "22",
    }
    node := ork.NewNodeFromConfig(cfg)
    
    // Run a command
    results := node.RunCommand("uptime")
    result := results.Results["server.example.com"]
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    log.Println(result.Message)
}
```

## Configuration

Use the fluent API to configure connection settings:

```go
// From host
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy").
    SetKey("production.prv")

results := node.RunCommand("uptime")
result := results.Results["server.example.com"]
if result.Error != nil {
    log.Fatal(result.Error)
}
output := result.Message
```

## Persistent Connections

For multiple operations, establish a persistent connection:

```go
node := ork.NewNodeForHost("server.example.com").
    SetPort("2222").
    SetUser("deploy")

if err := node.Connect(); err != nil {
    log.Fatal(err)
}
defer node.Close()

// These commands reuse the same SSH connection
results1 := node.RunCommand("uptime")
results2 := node.RunCommand("df -h")
output1 := results1.Results["server.example.com"].Message
output2 := results2.Results["server.example.com"].Message
```

## Skills

Run pre-built automation tasks (skills) against a node:

```go
node := ork.NewNodeForHost("server.example.com").
    SetArg("username", "alice").
    SetArg("shell", "/bin/bash")

results := node.RunSkill(skills.NewUserCreate())
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

### Available Skills

| `ork` Package | `skill` Package | String | Args | Description |
|---------------|----------------|--------|------|-------------|
| `SkillPing` | `IDPing` | `ping` | - | Check SSH connectivity |
| `SkillAptUpdate` | `IDAptUpdate` | `apt-update` | - | Refresh package database |
| `SkillAptUpgrade` | `IDAptUpgrade` | `apt-upgrade` | - | Install available updates |
| `SkillAptStatus` | `IDAptStatus` | `apt-status` | - | Show available updates |
| `SkillReboot` | `IDReboot` | `reboot` | - | Reboot server |
| `SkillSwapCreate` | `IDSwapCreate` | `swap-create` | `size` (GB) | Create swap file |
| `SkillSwapDelete` | `IDSwapDelete` | `swap-delete` | - | Remove swap file |
| `SkillSwapStatus` | `IDSwapStatus` | `swap-status` | - | Show swap status |
| `SkillUserCreate` | `IDUserCreate` | `user-create` | `username` | Create user with sudo |
| `SkillUserDelete` | `IDUserDelete` | `user-delete` | `username` | Delete user |
| `SkillUserStatus` | `IDUserStatus` | `user-status` | `username` (opt) | Show user info |
| `SkillFail2banInstall` | `IDFail2banInstall` | `fail2ban-install` | - | Install and configure fail2ban |
| `SkillFail2banStatus` | `IDFail2banStatus` | `fail2ban-status` | - | Show fail2ban status |
| `SkillUfwInstall` | `IDUfwInstall` | `ufw-install` | - | Install UFW firewall |
| `SkillUfwStatus` | `IDUfwStatus` | `ufw-status` | - | Show UFW status |
| `SkillUfwAllowMariadb` | `IDUfwAllowMariadb` | `ufw-allow-mariadb` | - | Allow MariaDB through UFW |
| `SkillMariadbInstall` | `IDMariadbInstall` | `mariadb-install` | - | Install MariaDB server |
| `SkillMariadbStatus` | `IDMariadbStatus` | `mariadb-status` | - | Show MariaDB status |
| `SkillMariadbSecure` | `IDMariadbSecure` | `mariadb-secure` | - | Secure MariaDB installation |
| `SkillMariadbBackup` | `IDMariadbBackup` | `mariadb-backup` | `database` (opt) | Backup database |
| `SkillMariadbBackupEncrypt` | `IDMariadbBackupEncrypt` | `mariadb-backup-encrypt` | - | Encrypted backup |
| `SkillMariadbChangePort` | `IDMariadbChangePort` | `mariadb-change-port` | `port` | Change MariaDB port |
| `SkillMariadbCreateDB` | `IDMariadbCreateDB` | `mariadb-create-db` | `database` | Create database |
| `SkillMariadbCreateUser` | `IDMariadbCreateUser` | `mariadb-create-user` | `username`, `password` | Create DB user |
| `SkillMariadbEnableEncryption` | `IDMariadbEnableEncryption` | `mariadb-enable-encryption` | - | Enable encryption at rest |
| `SkillMariadbEnableSSL` | `IDMariadbEnableSSL` | `mariadb-enable-ssl` | - | Enable SSL connections |
| `SkillMariadbListDBs` | `IDMariadbListDBs` | `mariadb-list-dbs` | - | List databases |
| `SkillMariadbListUsers` | `IDMariadbListUsers` | `mariadb-list-users` | - | List DB users |
| `SkillMariadbSecurityAudit` | `IDMariadbSecurityAudit` | `mariadb-security-audit` | - | Run security audit |
| `SkillSecurityAideInstall` | `IDSecurityAideInstall` | `security-aide-install` | - | Install AIDE IDS |
| `SkillSecurityAuditdInstall` | `IDSecurityAuditdInstall` | `security-auditd-install` | - | Install audit daemon |
| `SkillSecurityKernelHarden` | `IDSecurityKernelHarden` | `security-kernel-harden` | - | Apply kernel hardening |
| `SkillSecuritySSHChangePort` | `IDSecuritySSHChangePort` | `security-ssh-change-port` | `port` | Change SSH port |
| `SkillSecuritySSHHarden` | `IDSecuritySSHHarden` | `security-ssh-harden` | - | Harden SSH config |

## Inventory (Multi-Node Operations)

Manage multiple servers with the same API:

```go
// Create inventory
inv := ork.NewInventory()

// Add nodes to groups
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))
webGroup.SetArg("env", "production")
inv.AddGroup(webGroup)

// Run skill on entire inventory
results := inv.RunSkill(skills.NewPing())

// Check summary
summary := results.Summary()
fmt.Printf("Total: %d, Changed: %d, Failed: %d\n",
    summary.Total, summary.Changed, summary.Failed)

// Check individual results
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s failed: %v", host, result.Error)
    }
}
```
