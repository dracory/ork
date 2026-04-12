# Ork

Ork is a Go package for SSH-based server automation playbooks. It provides common utilities for connecting to remote servers via SSH and running automation tasks.

## Installation

```bash
go get github.com/dracory/ork
```

## Packages

- `ssh` - SSH connection utilities and command execution
- `config` - Configuration types for remote operations
- `playbook` - Base interfaces and registry for organizing playbooks
- `playbooks` - Reusable playbook implementations (ping, apt, reboot, swap, user)

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/ssh"
)

func main() {
    // Create config
    cfg := config.Config{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }
    
    // Run a command
    output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(output)
}
```

## Using Reusable Playbooks

```go
package main

import (
    "log"

    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbooks"
)

func main() {
    cfg := config.Config{
        SSHHost:  "db3.sinevia.com",
        SSHPort:  "40022",
        SSHKey:   "2024_sinevia.prv",
        RootUser: "root",
    }

    // Ping server to check connectivity
    ping := playbooks.NewPing()
    if err := ping.Run(cfg); err != nil {
        log.Fatal(err)
    }

    // Update packages
    aptUpdate := playbooks.NewAptUpdate()
    if err := aptUpdate.Run(cfg); err != nil {
        log.Fatal(err)
    }

    // Create a 2GB swap file
    cfg.Args = map[string]string{"size": "2"}
    swapCreate := playbooks.NewSwapCreate()
    if err := swapCreate.Run(cfg); err != nil {
        log.Fatal(err)
    }
}
```

### Available Playbooks

| Playbook | Description | Args |
|----------|-------------|------|
| `ping` | Check SSH connectivity | - |
| `apt-update` | Refresh package database | - |
| `apt-upgrade` | Install available updates | - |
| `apt-status` | Show available updates | - |
| `reboot` | Reboot server | - |
| `swap-create` | Create swap file | `size` (GB, default 1) |
| `swap-delete` | Remove swap file | - |
| `swap-status` | Show swap status | - |
| `user-create` | Create user with sudo | `username` |
| `user-delete` | Delete user | `username` |
| `user-status` | Show user info | `username` (optional) |

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
