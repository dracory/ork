# Ork

Ork is a Go package for SSH-based server automation. Think of it like Ansible, but in Go - you define **Nodes** (remote servers), organize them into **Groups**, and run commands or skills against them individually or at scale via **Inventory**.

## Installation

```bash
go get github.com/dracory/ork
```

## Documentation

Full documentation is available in the [docs](docs/) directory.

### Getting Started

- [Quick Start](docs/quick_start.md) - Quick start guide with examples
- [Configuration](docs/quick_start.md) - Configuration options and settings

### Skills

- [Built-in Skills](docs/skills.md) - All available automation tasks
- [Playbooks](docs/playbooks.md) - Complex orchestration with full Go power

### Features

- [Vault](docs/vault.md) - Secure secrets management
- [Idempotency](docs/idempotency.md) - Understanding idempotent operations
- [Dry-Run Mode](docs/dry_run.md) - Preview changes without execution
- [Advanced Usage](docs/advanced_usage.md) - Custom skills and internal packages

### Reference

- [Comparison with Ansible](docs/comparison/ansible.md) - How Ork compares to Ansible

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
