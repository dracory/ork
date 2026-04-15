---
path: modules/playbooks.md
page-type: module
summary: Built-in playbook implementations for system management, users, swap, MariaDB, security, and more.
tags: [module, playbooks, automation, tasks]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# playbooks Package

Reusable playbook implementations for common server automation tasks.

## Purpose

The `playbooks` package provides built-in implementations of `PlaybookInterface` for common server management tasks. Each subpackage focuses on a specific domain (system, users, security, databases, etc.).

## Structure

```
playbooks/
├── doc.go              # Package documentation
├── apt/                # Package management
├── ping/               # Connectivity checks
├── reboot/             # Server reboot
├── swap/               # Swap file management
├── user/               # User management
├── mariadb/            # MariaDB database
├── security/           # Security hardening
├── ufw/                # UFW firewall
└── fail2ban/           # Fail2ban intrusion prevention
```

## Quick Reference

| Category | Playbook | Constructor | Key Arguments |
|----------|----------|-------------|---------------|
| **System** | ping | `ping.NewPing()` | - |
| | apt-update | `apt.NewAptUpdate()` | - |
| | apt-upgrade | `apt.NewAptUpgrade()` | - |
| | apt-status | `apt.NewAptStatus()` | - |
| | reboot | `reboot.NewReboot()` | - |
| **Users** | user-create | `user.NewUserCreate()` | `username` (required) |
| | user-delete | `user.NewUserDelete()` | `username` (required) |
| | user-status | `user.NewUserStatus()` | `username` (optional) |
| **Swap** | swap-create | `swap.NewSwapCreate()` | `size`, `unit`, `swappiness` |
| | swap-delete | `swap.NewSwapDelete()` | - |
| | swap-status | `swap.NewSwapStatus()` | - |
| **Security** | ssh-harden | `security.NewSshHarden()` | `non-root-user` |
| | kernel-harden | `security.NewKernelHarden()` | - |
| | aide-install | `security.NewAideInstall()` | - |
| | auditd-install | `security.NewAuditdInstall()` | - |
| | ssh-change-port | `security.NewSshChangePort()` | `port` (required) |
| **UFW** | ufw-install | `ufw.NewUfwInstall()` | - |
| | ufw-status | `ufw.NewUfwStatus()` | - |
| | ufw-allow-mariadb | `ufw.NewAllowMariaDB()` | - |
| **Fail2ban** | fail2ban-install | `fail2ban.NewFail2banInstall()` | - |
| | fail2ban-status | `fail2ban.NewFail2banStatus()` | - |
| **MariaDB** | mariadb-install | `mariadb.NewInstall()` | `root-password` |
| | mariadb-secure | `mariadb.NewSecure()` | `root-password` |
| | mariadb-create-db | `mariadb.NewCreateDB()` | `database` (required) |
| | mariadb-create-user | `mariadb.NewCreateUser()` | `username`, `password` |
| | mariadb-status | `mariadb.NewStatus()` | - |
| | mariadb-list-dbs | `mariadb.NewListDBs()` | - |
| | mariadb-list-users | `mariadb.NewListUsers()` | - |
| | mariadb-backup | `mariadb.NewBackup()` | `database` (optional) |
| | mariadb-security-audit | `mariadb.NewSecurityAudit()` | - |
| | mariadb-change-port | `mariadb.NewChangePort()` | `port` (required) |
| | mariadb-enable-ssl | `mariadb.NewEnableSSL()` | - |
| | mariadb-enable-encryption | `mariadb.NewEnableEncryption()` | - |
| | mariadb-backup-encrypt | `mariadb.NewBackupEncrypt()` | - |

## System Playbooks

### ping

Tests SSH connectivity and returns server uptime.

```go
results := node.RunPlaybook(playbooks.NewPing())
// Result.Details["uptime"] contains uptime string
```

**Idempotent**: Yes (read-only)

### apt

Package management playbooks for Debian/Ubuntu systems.

```go
// Update package lists
results := node.RunPlaybook(playbooks.NewAptUpdate())

// Upgrade installed packages
results = node.RunPlaybook(playbooks.NewAptUpgrade())

// Check for available updates
results = node.RunPlaybook(playbooks.NewAptStatus())
// Result.Details contain update information
```

**Idempotent**: apt-update always reports Changed (cache timestamp updated)

### reboot

Reboots the server with optional wait for reconnection.

```go
results := node.RunPlaybook(playbooks.NewReboot())
```

**Warning**: Use with caution. Will disconnect active sessions.

## User Playbooks

All user playbooks require the `username` argument.

### user-create

Creates a non-root user with sudo access.

```go
node.SetArg("username", "alice")
node.SetArg("shell", "/bin/bash")           // optional, default: /bin/bash
node.SetArg("password", "initialpass")      // optional
node.SetArg("ssh-key", "ssh-rsa AAAAB3...") // optional

results := node.RunPlaybook(playbooks.NewUserCreate())
```

**Idempotent**: Returns unchanged if user already exists.

### user-delete

Removes a user from the system.

```go
node.SetArg("username", "bob")
results := node.RunPlaybook(playbooks.NewUserDelete())
```

**Idempotent**: Returns unchanged if user doesn't exist.

### user-status

Shows user information.

```go
node.SetArg("username", "alice")  // optional - shows all users if empty
results := node.RunPlaybook(playbooks.NewUserStatus())
// Result.Details contain user info
```

**Idempotent**: Yes (read-only)

## Swap Playbooks

### swap-create

Creates a swap file with configurable size.

```go
node.SetArg("size", "2")        // Size in GB (default: "1")
node.SetArg("unit", "gb")       // "gb" or "mb" (default: "gb")
node.SetArg("swappiness", "10") // 0-100 (default: "10")

results := node.RunPlaybook(playbooks.NewSwapCreate())
// Result.Details: size, file, swappiness, status
```

**Idempotent**: Returns unchanged if swap already exists.

### swap-delete

Removes the swap file.

```go
results := node.RunPlaybook(playbooks.NewSwapDelete())
```

**Idempotent**: Returns unchanged if no swap exists.

### swap-status

Shows current swap status.

```go
results := node.RunPlaybook(playbooks.NewSwapStatus())
// Result.Details contain swap info
```

**Idempotent**: Yes (read-only)

## Security Playbooks

### ssh-harden

Applies security hardening to SSH configuration.

**Changes applied**:
- Disable root login
- Disable password authentication
- Enable public key authentication
- Disable empty passwords
- Set MaxAuthTries to 3
- Disable X11 forwarding
- Set client alive settings

```go
node.SetArg("non-root-user", "deploy")  // Verify this user exists first
results := node.RunPlaybook(playbooks.NewSshHarden())
```

**Warning**: After running, you must use SSH key authentication. Ensure the non-root user exists and has sudo access.

### kernel-harden

Applies security-focused kernel parameters.

```go
results := node.RunPlaybook(playbooks.NewKernelHarden())
```

### aide-install

Installs AIDE (Advanced Intrusion Detection Environment).

```go
results := node.RunPlaybook(playbooks.NewAideInstall())
```

### auditd-install

Installs and configures the Linux Audit Framework (auditd).

```go
results := node.RunPlaybook(playbooks.NewAuditdInstall())
```

### ssh-change-port

Changes the SSH server port.

```go
node.SetArg("port", "2222")
results := node.RunPlaybook(playbooks.NewSshChangePort())
```

**Warning**: Ensure firewall rules are updated before changing ports.

## UFW (Firewall) Playbooks

### ufw-install

Installs and enables UFW (Uncomplicated Firewall).

```go
results := node.RunPlaybook(playbooks.NewUfwInstall())
```

### ufw-status

Shows UFW status and rules.

```go
results := node.RunPlaybook(playbooks.NewUfwStatus())
```

### ufw-allow-mariadb

Configures UFW to allow MariaDB connections.

```go
results := node.RunPlaybook(playbooks.NewAllowMariaDB())
// Allows port 3306/tcp
```

## Fail2ban Playbooks

### fail2ban-install

Installs and configures fail2ban intrusion prevention.

```go
results := node.RunPlaybook(playbooks.NewFail2banInstall())
```

### fail2ban-status

Shows fail2ban service and jail status.

```go
results := node.RunPlaybook(playbooks.NewFail2banStatus())
```

## MariaDB Playbooks

### mariadb-install

Installs and configures MariaDB server.

```go
node.SetArg("root-password", "secure_password")  // optional
results := node.RunPlaybook(playbooks.NewMariadbInstall())
```

Configures bind-address to 0.0.0.0 for remote access.

### mariadb-secure

Secures MariaDB installation (removes test data, disables remote root).

```go
node.SetArg("root-password", "secure_password")
results := node.RunPlaybook(playbooks.NewMariadbSecure())
```

### mariadb-create-db

Creates a new database.

```go
node.SetArg("database", "myapp")
node.SetArg("root-password", "secure_password")  // If not in config
results := node.RunPlaybook(playbooks.NewMariadbCreateDB())
```

**Idempotent**: Returns unchanged if database exists.

### mariadb-create-user

Creates a new database user.

```go
node.SetArg("username", "appuser")
node.SetArg("password", "apppassword")
node.SetArg("root-password", "secure_password")
results := node.RunPlaybook(playbooks.NewMariadbCreateUser())
```

**Idempotent**: Updates password if user exists.

### mariadb-status

Shows MariaDB server status.

```go
results := node.RunPlaybook(playbooks.NewMariadbStatus())
```

### mariadb-list-dbs

Lists all databases.

```go
results := node.RunPlaybook(playbooks.NewMariadbListDBs())
// Result.Details contain database list
```

### mariadb-list-users

Lists all database users.

```go
results := node.RunPlaybook(playbooks.NewMariadbListUsers())
```

### mariadb-backup

Creates a compressed SQL dump.

```go
node.SetArg("database", "myapp")  // optional - backs up all if empty
node.SetArg("root-password", "secure_password")
results := node.RunPlaybook(playbooks.NewMariadbBackup())
```

### mariadb-security-audit

Performs security audit of MariaDB configuration.

```go
results := node.RunPlaybook(playbooks.NewMariadbSecurityAudit())
// Result.Details contain audit findings
```

### mariadb-change-port

Changes MariaDB port.

```go
node.SetArg("port", "3307")
results := node.RunPlaybook(playbooks.NewMariadbChangePort())
```

### mariadb-enable-ssl

Enables SSL/TLS encryption for connections.

```go
results := node.RunPlaybook(playbooks.NewMariadbEnableSSL())
```

### mariadb-enable-encryption

Enables data-at-rest encryption.

```go
results := node.RunPlaybook(playbooks.NewMariadbEnableEncryption())
```

### mariadb-backup-encrypt

Creates an encrypted backup.

```go
node.SetArg("database", "myapp")  // optional
results := node.RunPlaybook(playbooks.NewMariadbBackupEncrypt())
```

## Usage Patterns

### Running Playbooks

```go
// Direct instance (recommended)
results := node.RunPlaybook(playbooks.NewAptUpdate())

// By ID via registry
results := node.RunPlaybookByID(playbooks.IDAptUpdate)

// Check mode (preview changes)
results := node.CheckPlaybook(playbooks.NewAptUpgrade())
```

### Setting Arguments

```go
// Via node
node.SetArg("username", "alice")
results := node.RunPlaybook(playbooks.NewUserCreate())

// Via config
cfg := config.NodeConfig{
    SSHHost: "server.example.com",
    Args: map[string]string{
        "username": "alice",
    },
}
pb := playbooks.NewUserCreate()
pb.SetConfig(cfg)
result := pb.Run()
```

### Handling Results

```go
results := node.RunPlaybook(playbooks.NewSwapCreate())
result := results.Results["server.example.com"]

if result.Error != nil {
    log.Fatalf("Failed: %v", result.Error)
}

if result.Changed {
    log.Printf("Created: %s", result.Message)
    for key, value := range result.Details {
        log.Printf("  %s: %s", key, value)
    }
} else {
    log.Println("No changes needed")
}
```

## Creating Custom Playbooks

See the [Development Guide](../development.md) for creating custom playbooks.

## See Also

- [playbook](playbook.md) - Playbook interface definition
- [API Reference](../api_reference.md) - Complete API
- [Cheatsheet](../cheatsheet.md) - Quick reference
