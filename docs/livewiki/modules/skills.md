---
path: modules/skills.md
page-type: module
summary: Built-in skill implementations for system management, users, swap, MariaDB, security, and more.
tags: [module, skills, automation, tasks]
created: 2025-04-14
updated: 2026-04-15
version: 2.0.0
---

## Changelog
- **v2.0.0** (2026-04-15): Major terminology refactoring - playbooks renamed to skills, PlaybookInterface renamed to RunnableInterface
- **v1.0.0** (2025-04-14): Initial creation

# skills Package

Reusable skill implementations for common server automation tasks.

## Purpose

The `skills` package provides built-in implementations of `RunnableInterface` for common server management tasks. Each subpackage focuses on a specific domain (system, users, security, databases, etc.).

## Structure

```
skills/
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

| Category | Skill | Constructor | Key Arguments |
|----------|-------|-------------|---------------|
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

## System Skills

### ping

Tests SSH connectivity and returns server uptime.

```go
results := node.Run(skills.NewPing())
// Result.Details["uptime"] contains uptime string
```

**Idempotent**: Yes (read-only)

### apt

Package management skills for Debian/Ubuntu systems.

```go
// Update package lists
results := node.Run(skills.NewAptUpdate())

// Upgrade installed packages
results = node.Run(skills.NewAptUpgrade())

// Check for available updates
results = node.Run(skills.NewAptStatus())
// Result.Details contain update information
```

**Idempotent**: apt-update always reports Changed (cache timestamp updated)

### reboot

Reboots the server with optional wait for reconnection.

```go
results := node.Run(skills.NewReboot())
```

**Warning**: Use with caution. Will disconnect active sessions.

## User Skills

All user skills require the `username` argument.

### user-create

Creates a non-root user with sudo access.

```go
node.SetArg("username", "alice")
node.SetArg("shell", "/bin/bash")           // optional, default: /bin/bash
node.SetArg("password", "initialpass")      // optional
node.SetArg("ssh-key", "ssh-rsa AAAAB3...") // optional

results := node.Run(skills.NewUserCreate())
```

**Idempotent**: Returns unchanged if user already exists.

### user-delete

Removes a user from the system.

```go
node.SetArg("username", "bob")
results := node.Run(skills.NewUserDelete())
```

**Idempotent**: Returns unchanged if user doesn't exist.

### user-status

Shows user information.

```go
node.SetArg("username", "alice")  // optional - shows all users if empty
results := node.Run(skills.NewUserStatus())
// Result.Details contain user info
```

**Idempotent**: Yes (read-only)

## Swap Skills

### swap-create

Creates a swap file with configurable size.

```go
node.SetArg("size", "2")        // Size in GB (default: "1")
node.SetArg("unit", "gb")       // "gb" or "mb" (default: "gb")
node.SetArg("swappiness", "10") // 0-100 (default: "10")

results := node.Run(skills.NewSwapCreate())
// Result.Details: size, file, swappiness, status
```

**Idempotent**: Returns unchanged if swap already exists.

### swap-delete

Removes the swap file.

```go
results := node.Run(skills.NewSwapDelete())
```

**Idempotent**: Returns unchanged if no swap exists.

### swap-status

Shows current swap status.

```go
results := node.Run(skills.NewSwapStatus())
// Result.Details contain swap info
```

**Idempotent**: Yes (read-only)

## Security Skills

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
results := node.Run(skills.NewSshHarden())
```

**Warning**: After running, you must use SSH key authentication. Ensure the non-root user exists and has sudo access.

### kernel-harden

Applies security-focused kernel parameters.

```go
results := node.Run(skills.NewKernelHarden())
```

### aide-install

Installs AIDE (Advanced Intrusion Detection Environment).

```go
results := node.Run(skills.NewAideInstall())
```

### auditd-install

Installs and configures the Linux Audit Framework (auditd).

```go
results := node.Run(skills.NewAuditdInstall())
```

### ssh-change-port

Changes the SSH server port.

```go
node.SetArg("port", "2222")
results := node.Run(skills.NewSshChangePort())
```

**Warning**: Ensure firewall rules are updated before changing ports.

## UFW (Firewall) Skills

### ufw-install

Installs and enables UFW (Uncomplicated Firewall).

```go
results := node.Run(skills.NewUfwInstall())
```

### ufw-status

Shows UFW status and rules.

```go
results := node.Run(skills.NewUfwStatus())
```

### ufw-allow-mariadb

Configures UFW to allow MariaDB connections.

```go
results := node.Run(skills.NewAllowMariaDB())
// Allows port 3306/tcp
```

## Fail2ban Skills

### fail2ban-install

Installs and configures fail2ban intrusion prevention.

```go
results := node.Run(skills.NewFail2banInstall())
```

### fail2ban-status

Shows fail2ban service and jail status.

```go
results := node.Run(skills.NewFail2banStatus())
```

## MariaDB Skills

### mariadb-install

Installs and configures MariaDB server.

```go
node.SetArg("root-password", "secure_password")  // optional
results := node.Run(skills.NewMariadbInstall())
```

Configures bind-address to 0.0.0.0 for remote access.

### mariadb-secure

Secures MariaDB installation (removes test data, disables remote root).

```go
node.SetArg("root-password", "secure_password")
results := node.Run(skills.NewMariadbSecure())
```

### mariadb-create-db

Creates a new database.

```go
node.SetArg("database", "myapp")
node.SetArg("root-password", "secure_password")  // If not in config
results := node.Run(skills.NewMariadbCreateDB())
```

**Idempotent**: Returns unchanged if database exists.

### mariadb-create-user

Creates a new database user.

```go
node.SetArg("username", "appuser")
node.SetArg("password", "apppassword")
node.SetArg("root-password", "secure_password")
results := node.Run(skills.NewMariadbCreateUser())
```

**Idempotent**: Updates password if user exists.

### mariadb-status

Shows MariaDB server status.

```go
results := node.Run(skills.NewMariadbStatus())
```

### mariadb-list-dbs

Lists all databases.

```go
results := node.Run(skills.NewMariadbListDBs())
// Result.Details contain database list
```

### mariadb-list-users

Lists all database users.

```go
results := node.Run(skills.NewMariadbListUsers())
```

### mariadb-backup

Creates a compressed SQL dump.

```go
node.SetArg("database", "myapp")  // optional - backs up all if empty
node.SetArg("root-password", "secure_password")
results := node.Run(skills.NewMariadbBackup())
```

### mariadb-security-audit

Performs security audit of MariaDB configuration.

```go
results := node.Run(skills.NewMariadbSecurityAudit())
// Result.Details contain audit findings
```

### mariadb-change-port

Changes MariaDB port.

```go
node.SetArg("port", "3307")
results := node.Run(skills.NewMariadbChangePort())
```

### mariadb-enable-ssl

Enables SSL/TLS encryption for connections.

```go
results := node.Run(skills.NewMariadbEnableSSL())
```

### mariadb-enable-encryption

Enables data-at-rest encryption.

```go
results := node.Run(skills.NewMariadbEnableEncryption())
```

### mariadb-backup-encrypt

Creates an encrypted backup.

```go
node.SetArg("database", "myapp")  // optional
results := node.Run(skills.NewMariadbBackupEncrypt())
```

## Usage Patterns

### Running Skills

```go
// Direct instance (recommended)
results := node.Run(skills.NewAptUpdate())

// By ID via registry
results := node.RunByID(skills.IDAptUpdate)

// Check mode (preview changes)
results := node.Check(skills.NewAptUpgrade())
```

### Setting Arguments

```go
// Via node
node.SetArg("username", "alice")
results := node.Run(skills.NewUserCreate())

// Via config
cfg := types.NodeConfig{
    SSHHost: "server.example.com",
    Args: map[string]string{
        "username": "alice",
    },
}
pb := skills.NewUserCreate()
pb.SetNodeConfig(cfg)
result := pb.Run()
```

### Handling Results

```go
results := node.Run(skills.NewSwapCreate())
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

## Creating Custom Skills

See the [Development Guide](../development.md) for creating custom skills.

## See Also

- [types](types.md) - BasePlaybook, BaseSkill, and RunnableInterface in types package
- [API Reference](../api_reference.md) - Complete API
- [Cheatsheet](../cheatsheet.md) - Quick reference
