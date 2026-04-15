# Privilege Escalation (Become)

Ork supports running commands as a different user using `sudo`. This feature is useful when you need to perform operations that require elevated privileges, such as installing packages, managing services, or accessing system resources.

## Overview

The privilege escalation feature (called "become") allows you to:
- Connect as a non-root user but run commands as root
- Run commands as specific service users (postgres, www-data, etc.)
- Configure become user at different hierarchy levels (inventory, group, node, skill)
- Automatically propagate become user settings through the hierarchy

## Basic Usage

### Connect as Deploy, Run as Root

```go
package main

import (
    "log"
    "github.com/dracory/ork"
)

func main() {
    // Connect as a non-root user, but run commands as root
    node := ork.NewNodeForHost("server.example.com").
        SetUser("deploy").
        SetBecomeUser("root")

    // This will run: sudo -u root apt-get update
    results := node.RunCommand("apt-get update")
    result := results.Results["server.example.com"]

    if result.Error != nil {
        log.Fatal(result.Error)
    }

    log.Println(result.Message)
}
```

### Running as Specific Users

```go
// Run database commands as the postgres user
node.SetBecomeUser("postgres")
results := node.RunCommand("psql -c 'SELECT version()'")

// Run web server commands as www-data
node.SetBecomeUser("www-data")
results := node.RunCommand("systemctl restart nginx")
```

## Hierarchy and Precedence

The become user setting follows this precedence (highest to lowest):

1. **Skill/Playbook level**: Skill-specific setting
2. **Node level**: Node-specific setting
3. **Group level**: Group setting (propagated to all nodes)
4. **Inventory level**: Inventory setting (propagated to all groups and nodes)

### Example: Precedence in Action

```go
// Set at inventory level - applies to all
inv := ork.NewInventory()
inv.SetBecomeUser("root")

// Create a group that overrides the inventory setting
dbGroup := ork.NewGroup("databases")
dbGroup.SetBecomeUser("postgres")  // Uses postgres, not root

// Add a node that overrides the group setting
node := ork.NewNodeForHost("special.example.com")
node.SetBecomeUser("admin")  // Uses admin, not postgres or root

dbGroup.AddNode(node)
inv.AddGroup(dbGroup)
```

## Configuration at Different Levels

### Inventory Level

```go
inv := ork.NewInventory()
inv.SetBecomeUser("root")

// All nodes in all groups will run as root
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

inv.AddGroup(webGroup)
```

### Group Level

```go
dbGroup := ork.NewGroup("databases")
dbGroup.SetBecomeUser("postgres")

// All nodes in this group run as postgres
dbGroup.AddNode(ork.NewNodeForHost("db1.example.com"))
dbGroup.AddNode(ork.NewNodeForHost("db2.example.com"))
```

### Node Level

```go
node := ork.NewNodeForHost("server.example.com").
    SetUser("deploy").
    SetBecomeUser("root")

// This specific node runs as root
results := node.RunCommand("apt-get update")
```

### Skill Level

```go
skill := skills.NewAptUpdate()
skill.SetBecomeUser("root")

// The skill will run as root regardless of node/group/inventory settings
results := node.Run(skill)
```

## With Skills and Playbooks

### Setting Become User on Skills

```go
skill := skills.NewAptUpdate()
skill.SetBecomeUser("root")

results := node.Run(skill)
```

### Setting Become User on Playbooks

```go
playbook := playbooks.NewUserCreate()
playbook.SetBecomeUser("root")

results := node.Run(playbook)
```

### Dynamic User Selection in Custom Skills

```go
type MyCustomSkill struct {
    types.BaseSkill
}

func (s *MyCustomSkill) Run() types.Result {
    cfg := s.GetNodeConfig()

    // Choose become user based on node configuration
    becomeUser := "root"
    if cfg.GetArg("environment") == "production" {
        becomeUser = "admin"
    }

    s.SetBecomeUser(becomeUser)

    // Execute commands as the selected user
    output, err := ssh.Run(cfg, types.Command{
        Command: "systemctl restart nginx",
    })
    // ...
}
```

## Security Considerations

### Requirements

- The connecting user must have `sudo` privileges to the target user
- Passwordless sudo must be configured for the connecting user
- The target user must exist on the remote server

### Configuration Example

To enable passwordless sudo for a user:

```bash
# On the remote server, as root
visudo

# Add this line (replace deploy with your user)
deploy ALL=(ALL) NOPASSWD: ALL

# Or restrict to specific users
deploy ALL=(postgres) NOPASSWD: ALL
```

### Best Practices

- **Principle of Least Privilege**: Only escalate when necessary
- **Use Specific Users**: Run as the specific service user (postgres, www-data) rather than root
- **Audit Sudo Access**: Regularly review sudo access configurations
- **Limit Sudo Scope**: Configure sudo to only allow specific commands when possible

### Current Limitations

- No password support is currently implemented
- No custom sudo flags or options
- Only supports `sudo -u <user>` syntax
- Does not support `su`, `doas`, or other privilege escalation methods

## API Reference

### BecomeInterface

```go
type BecomeInterface interface {
    SetBecomeUser(user string) BecomeInterface
    GetBecomeUser() string
}
```

**Methods:**

- `SetBecomeUser(user string) BecomeInterface` - Sets the user to become when executing commands via sudo. Returns BecomeInterface for fluent method chaining.
- `GetBecomeUser() string` - Returns the configured become user. Returns empty string if not set.

### BaseBecome

```go
type BaseBecome struct {
    becomeUser string
}
```

Provides a default implementation of BecomeInterface. Embed this in custom skills or playbooks to automatically get privilege escalation support.

### NodeConfig

```go
type NodeConfig struct {
    // ... other fields ...
    BecomeUser string
}
```

The `BecomeUser` field in NodeConfig stores the user to become when executing commands via sudo.

## Examples

### Database Migration as PostgreSQL User

```go
node := ork.NewNodeForHost("db.example.com").
    SetUser("deploy").
    SetBecomeUser("postgres")

// Run database migration
results := node.RunCommand("psql -f /path/to/migration.sql")
```

### Web Server Management as www-data

```go
webGroup := ork.NewGroup("webservers")
webGroup.SetBecomeUser("www-data")

webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))
webGroup.AddNode(ork.NewNodeForHost("web2.example.com"))

// Restart nginx as www-data on all servers
results := webGroup.RunCommand("systemctl restart nginx")
```

### Mixed Privilege Levels in Inventory

```go
inv := ork.NewInventory()

// Web servers run as www-data
webGroup := ork.NewGroup("webservers")
webGroup.SetBecomeUser("www-data")
webGroup.AddNode(ork.NewNodeForHost("web1.example.com"))

// Database servers run as postgres
dbGroup := ork.NewGroup("databases")
dbGroup.SetBecomeUser("postgres")
dbGroup.AddNode(ork.NewNodeForHost("db1.example.com"))

inv.AddGroup(webGroup)
inv.AddGroup(dbGroup)
```

## Troubleshooting

### "sudo: no tty present" Error

If you encounter this error, the remote server's sudo configuration requires a TTY. You can disable this by editing the sudoers file:

```bash
# On remote server
visudo

# Comment out or remove this line:
# Defaults requiretty
```

### Permission Denied

If you get permission denied errors:

1. Verify the connecting user has sudo privileges
2. Check that passwordless sudo is configured
3. Ensure the target user exists on the remote server
4. Verify the sudoers configuration allows the specific user transition

### Command Not Found

If the command is not found when running as a different user:

1. The target user's PATH may be different
2. Use absolute paths for commands
3. Check the target user's shell configuration

## Future Enhancements

The proposal document (`docs/proposals/2026-04-15-privilege-escalation.md`) outlines potential future enhancements:

- Password support for sudo
- Custom sudo flags and options
- Support for `su`, `doas`, and other privilege escalation methods
- Sudo password prompt handling
- Sudoers configuration validation

## See Also

- [Getting Started](livewiki/getting_started.md) - Basic Ork usage
- [Advanced Usage](advanced_usage.md) - Custom skills and internal packages
- [API Reference](livewiki/api_reference.md) - Complete API documentation
- [Proposal](proposals/2026-04-15-privilege-escalation.md) - Original design proposal
