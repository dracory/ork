# Proposal: Configuration Management

**Date:** 2026-04-12  
**Status:** Rejected. Out of scope. Already implemented in Go.  
**Author:** System Review

> **Note:** Configuration is done programmatically via Go structs. No YAML/JSON file loading.

## Problem Statement

Currently, configuration is passed manually via `Config` structs in code. This creates several issues:

- No standard way to load configuration from files
- Secrets (SSH keys, passwords) are hardcoded or manually managed
- No environment variable support
- Difficult to share configurations across teams
- No validation or defaults

## Proposed Solution

Implement a comprehensive configuration management system supporting multiple sources with precedence rules.

## Configuration Sources (Priority Order)

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Config files** (workspace, user, system)
4. **Defaults** (lowest priority)

## Configuration File Formats

### YAML Format (Recommended)

```yaml
# .ork.yml or ork.yml
version: "1.0"

# SSH defaults
ssh:
  port: 22
  user: root
  key: id_rsa
  timeout: 30s

# Hosts
hosts:
  production-db:
    host: db3.sinevia.com
    port: 40022
    key: 2024_sinevia.prv
  
  staging-web:
    host: staging.example.com
    user: deploy

# Inventory file location
# Inventory is created programmatically
# See 2026-04-13-inventory.md for Inventory API

# Execution settings
execution:
  parallel: 5
  stop_on_error: false
  retry: 3
  retry_delay: 5s

# Logging
logging:
  level: info
  format: text
  file: /var/log/ork.log

# Secrets (can reference vault)
secrets:
  db_password: "{{ vault.db_password }}"
```

### JSON Format

```json
{
  "version": "1.0",
  "ssh": {
    "port": 22,
    "user": "root",
    "key": "id_rsa",
    "timeout": "30s"
  },
  "hosts": {
    "production-db": {
      "host": "db3.sinevia.com",
      "port": 40022,
      "key": "2024_sinevia.prv"
    }
  }
}
```

### TOML Format

```toml
version = "1.0"

[ssh]
port = 22
user = "root"
key = "id_rsa"
timeout = "30s"

[hosts.production-db]
host = "db3.sinevia.com"
port = 40022
key = "2024_sinevia.prv"
```

## Environment Variables

```bash
# SSH settings
export ORK_SSH_HOST=server.example.com
export ORK_SSH_PORT=22
export ORK_SSH_USER=root
export ORK_SSH_KEY=id_rsa

# Execution settings
export ORK_PARALLEL=10
export ORK_TIMEOUT=5m

# Logging
export ORK_LOG_LEVEL=debug
export ORK_LOG_FORMAT=json

# Config file location
export ORK_CONFIG=/etc/ork/config.yml
```

## Configuration Loading

### File Search Paths

1. `./ork.yml` (current directory)
2. `./.ork.yml` (hidden file in current directory)
3. `~/.ork/config.yml` (user config)
4. `/etc/ork/config.yml` (system config)

### Implementation

```go
package config

import (
    "github.com/spf13/viper"
)

type Manager struct {
    v *viper.Viper
}

func NewManager() *Manager {
    v := viper.New()
    
    // Set config name and paths
    v.SetConfigName("ork")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.AddConfigPath("$HOME/.ork")
    v.AddConfigPath("/etc/ork")
    
    // Environment variables
    v.SetEnvPrefix("ORK")
    v.AutomaticEnv()
    
    // Defaults
    v.SetDefault("ssh.port", 22)
    v.SetDefault("ssh.user", "root")
    v.SetDefault("ssh.key", "id_rsa")
    v.SetDefault("ssh.timeout", "30s")
    v.SetDefault("execution.parallel", 5)
    v.SetDefault("execution.retry", 3)
    v.SetDefault("logging.level", "info")
    
    return &Manager{v: v}
}

func (m *Manager) Load() error {
    return m.v.ReadInConfig()
}

func (m *Manager) LoadFile(path string) error {
    m.v.SetConfigFile(path)
    return m.v.ReadInConfig()
}

func (m *Manager) GetConfig(hostName string) (Config, error) {
    cfg := Config{
        SSHPort:  m.v.GetString("ssh.port"),
        RootUser: m.v.GetString("ssh.user"),
        SSHKey:   m.v.GetString("ssh.key"),
    }
    
    // Load host-specific config
    if hostName != "" {
        hostKey := "hosts." + hostName
        if m.v.IsSet(hostKey) {
            cfg.SSHHost = m.v.GetString(hostKey + ".host")
            if m.v.IsSet(hostKey + ".port") {
                cfg.SSHPort = m.v.GetString(hostKey + ".port")
            }
            if m.v.IsSet(hostKey + ".user") {
                cfg.RootUser = m.v.GetString(hostKey + ".user")
            }
            if m.v.IsSet(hostKey + ".key") {
                cfg.SSHKey = m.v.GetString(hostKey + ".key")
            }
        }
    }
    
    return cfg, nil
}

func (m *Manager) GetHost(name string) (HostConfig, error) {
    hostKey := "hosts." + name
    if !m.v.IsSet(hostKey) {
        return HostConfig{}, fmt.Errorf("host '%s' not found", name)
    }
    
    return HostConfig{
        Name: name,
        Host: m.v.GetString(hostKey + ".host"),
        Port: m.v.GetString(hostKey + ".port"),
        User: m.v.GetString(hostKey + ".user"),
        Key:  m.v.GetString(hostKey + ".key"),
    }, nil
}

func (m *Manager) ListHosts() []string {
    hosts := m.v.GetStringMap("hosts")
    names := make([]string, 0, len(hosts))
    for name := range hosts {
        names = append(names, name)
    }
    return names
}
```

## Usage Examples

### Programmatic Usage

```go
package main

import (
    "log"
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbooks"
)

func main() {
    // Load configuration
    mgr := config.NewManager()
    if err := mgr.Load(); err != nil {
        log.Fatal(err)
    }
    
    // Get config for specific host
    cfg, err := mgr.GetConfig("production-db")
    if err != nil {
        log.Fatal(err)
    }
    
    // Run playbook
    playbook := playbooks.NewPing()
    playbook.SetNodeConfig(cfg)
    result := playbook.Run()
    if result.Error != nil {
        log.Fatal(result.Error)
    }
}
```

### CLI Usage

```bash
# Configuration is done programmatically in Go
node := ork.NewNodeForHost("production-db").
    SetPort("2222").
    SetUser("deploy")
```

## Secrets Management

### Vault Integration

```yaml
# config.yml
secrets:
  provider: vault
  vault:
    address: https://vault.example.com
    token: "{{ env.VAULT_TOKEN }}"
    path: secret/ork

# Reference secrets
hosts:
  production-db:
    host: db.example.com
    password: "{{ vault.db_password }}"
```

### Environment Variable Secrets

```yaml
hosts:
  production-db:
    host: db.example.com
    password: "{{ env.DB_PASSWORD }}"
```

### Encrypted Files

```bash
# Encrypt config file
ork encrypt config.yml

# Decrypt and use
ork run ping --config config.yml.encrypted --decrypt
```

## Validation

```go
type ConfigValidator struct{}

func (v *ConfigValidator) Validate(cfg Config) error {
    if cfg.SSHHost == "" {
        return fmt.Errorf("ssh.host is required")
    }
    
    port, err := strconv.Atoi(cfg.SSHPort)
    if err != nil || port < 1 || port > 65535 {
        return fmt.Errorf("invalid ssh.port: %s", cfg.SSHPort)
    }
    
    if cfg.SSHKey != "" {
        keyPath := ssh.PrivateKeyPath(cfg.SSHKey)
        if _, err := os.Stat(keyPath); os.IsNotExist(err) {
            return fmt.Errorf("ssh key not found: %s", keyPath)
        }
    }
    
    return nil
}
```

## Configuration Profiles

```yaml
# config.yml
profiles:
  development:
    ssh:
      user: vagrant
      key: vagrant_key
    execution:
      parallel: 2
  
  production:
    ssh:
      user: root
      key: production_key
    execution:
      parallel: 10
      stop_on_error: true

# Use profile
current_profile: development
```

```bash
# CLI usage
ork run ping --profile production --host db1
```

## Implementation Plan

### Phase 1: Basic Config Loading
- Add `config.Manager` with viper
- Support YAML/JSON config files
- Environment variable support (`ORK_SSH_HOST`, etc.)

### Phase 2: Advanced Features
- Host definitions in config
- Profile support (dev/prod)
- Config validation

### Phase 3: Secrets Management
- Vault integration
- Encrypted config files

## Benefits

- **Convenience**: Load config from files instead of hardcoding
- **Security**: Proper secrets management
- **Flexibility**: Multiple config sources with clear precedence
- **Sharing**: Easy to share configurations across teams
- **Validation**: Catch configuration errors early

## Success Metrics

- Config files work for 100% of use cases
- Clear error messages for invalid configs
- Secrets never logged or exposed
- Documentation covers all config options

## Open Questions

1. Should we support remote config sources (HTTP, S3)?
2. How to handle config migrations between versions?
3. Should we support config inheritance/includes?
4. How to handle sensitive data in logs?
