# Ork

<p align="center">
  <img src="docs/images/ork.jpg" alt="Ork Logo" width="100%"/>
</p>

Ork is a Go package for SSH-based server automation. Think of it like Ansible, but in Go - you define **Nodes** (remote servers), organize them into **Groups**, and run commands, skills, or playbooks against them individually or at scale via **Inventory**.

## Installation

### Library

```bash
go get github.com/dracory/ork
```

### CLI Tool

```bash
go install github.com/dracory/ork/cmd/ork@latest
```

Or build from source:

```bash
git clone https://github.com/dracory/ork.git
cd ork
go build -o ork ./cmd/ork
```

## Documentation

> **📖 Read the full documentation online:**
> **[Ork Docs — Live Preview](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/index.html)**

The documentation is a static HTML site in the [docs](docs/) directory. The links below open the rendered pages in your browser via html-preview (raw `.html` files do not render on GitHub).

### Getting Started

- [Overview](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/overview.html) - High-level introduction to Ork and core concepts
- [Getting Started](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/getting-started.html) - Step-by-step installation and first node guide
- [Quick Start](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/quick-start.html) - Fastest path to running commands and skills

### Core Reference

- [Architecture](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/architecture.html) - Layered architecture, design patterns, and concurrency model
- [API Reference](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/api-reference.html) - Complete reference for all public interfaces and types
- [Configuration](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/configuration.html) - NodeConfig structure, SSH settings, and dry-run mode
- [Commands](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/commands.html) - Fluent API for one-off shell command execution
- [Skills](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/skills.html) - Reusable, idempotent automation tasks and the full skill catalog
- [Playbooks](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/playbooks.html) - Complex orchestration with full Go power

### Features

- [Idempotency](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/idempotency.html) - Check-Run pattern for safe, repeatable operations
- [Dry-Run Mode](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/dry-run.html) - Preview changes without execution
- [Privilege Escalation](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/privilege-escalation.html) - Run commands as different users via sudo
- [Vault](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/vault.html) - Secure secrets management (AES-256-GCM + Argon2id)
- [Advanced Usage](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/advanced-usage.html) - Custom skills, internal packages, and advanced patterns

### Operations & Guides

- [Troubleshooting](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/troubleshooting.html) - Common issues and solutions
- [Cheatsheet](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/cheatsheet.html) - Quick reference for common operations
- [Conventions](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/conventions.html) - Coding and documentation standards for contributors

### Comparison

- [Comparison Index](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/index.html) - Overview of all comparisons
- [vs Ansible](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/ansible.html) - Agentless SSH automation: YAML+Jinja2 vs Go
- [vs Chef](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/chef.html) - Agent-based pull: Ruby DSL vs Go
- [vs Puppet](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/puppet.html) - Agent-based with Puppet DSL vs Go
- [vs SaltStack](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/saltstack.html) - Event-driven with Salt master vs Go
- [vs CFEngine](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/cfengine.html) - C-based Promise Theory vs Go
- [vs Terraform](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/terraform.html) - Declarative cloud provisioning vs SSH automation
- [vs Pulumi](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/pulumi.html) - Multi-language IaC vs server configuration
- [vs CloudFormation](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/cloudformation.html) - AWS-native templates vs Go
- [Infrastructure as Code](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/infrastructure-as-code.html) - General IaC comparison
- [vs SSH Libraries](https://html-preview.github.io/?url=https://github.com/dracory/ork/blob/main/docs/comparison/ssh-libraries.html) - Raw Go SSH libraries vs Ork

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at https://www.gnu.org/licenses/agpl-3.0.en.html

For commercial use, please use my contact page to obtain a commercial license.
