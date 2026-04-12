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

## License

MIT
