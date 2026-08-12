# Ork — Project Guide for Agents

Ork is a Go library for SSH-based server automation ("Ansible in Go"). It also ships a `ork` CLI focused on vault management. Licensed under AGPL-3.0 (+ commercial license available).

## Module

- Module path: `github.com/dracory/ork`
- Go version: `1.26`
- Key deps: `golang.org/x/crypto/ssh` (SSH), `github.com/testcontainers/testcontainers-go` (integration tests), `github.com/dracory/omni` (Atom state container), `github.com/dracory/uid`, `github.com/samber/lo`.
- No external linter config; standard `gofmt`/`go vet` style.

## Architecture (layered)

1. **`types/`** — Core shared types and interfaces. No business logic, no SSH.
   - `NodeConfig` — SSH connection + execution config (host, port, user, key, args, dry-run, become user/password, chdir, KEX/host-key algorithms, logger). Fluent `WithX` setters.
   - `RunnableInterface` — anything runnable on a remote host (commands + skills). Implements `Check() (bool, error)` and `Run() Result`. Embeds `BecomeInterface`.
   - `BaseSkill` — embeddable base providing default `RunnableInterface` impls. State stored in `omni.Atom` (thread-safe). Skills embed `*types.BaseSkill` and only implement `Check()`/`Run()`. The framework clones skills via `ToMap()`/`FromMap()` before mutation so each goroutine gets an isolated copy.
   - `BasePlaybook` — base for playbooks (orchestration of runnables).
   - `Registry` — concurrent-safe registry of runnables with disable/enable and three merge strategies (`MergeReplaceOnOverlap`, `MergeKeepOnOverlap`, `MergeNoOverlap`).
   - `Result` / `Results` / `Summary` — execution outcomes; `Results.Summary()` aggregates changed/unchanged/failed.
   - `RunnerInterface` — common `Run`/`RunCommand`/`Check`/logger/dry-run API shared by Node, Group, Inventory.
   - `Command` — shell command struct (`Command`, `Description`, `Required`).

2. **`ssh/`** — SSH client built on `golang.org/x/crypto/ssh`.
   - `Client` with `Connect`/`Run`/`Close`, key resolved to `~/.ssh/<key>`, known_hosts verification, configurable KEX/host-key algorithms.
   - `ssh.Run(cfg, cmd)` — functional entry point used by skills.
   - `become.go` — privilege escalation via `sudo` (with prompt-detected password via `-S`, or `-n` for NOPASSWD).
   - `classifySSHError` — translates raw SSH errors into actionable messages (host key, auth, KEX mismatch, timeout, ...).
   - `exit_error.go` — `ExitError` for non-zero exit codes.

3. **`skills/`** — Built-in skill implementations, each in its own subpackage (`apt`, `caddy`, `dpkg`, `fail2ban`, `fs`, `mariadb`, `ncdu`, `php`, `ping`, `reboot`, `security`, `swap`, `systemctl`, `ufw`, `user`). Each skill:
   - Embeds `*types.BaseSkill`.
   - Exposes a `NewX` constructor that sets a stable ID (from `skills/constants.go`) and description.
   - Implements `Check()` (idempotency probe) and `Run()` (execute with `Changed` flag).
   - Uses `skills.ShellEscapeArg` to prevent shell injection when interpolating user input.
   - Honors `cfg.IsDryRunMode` (logs and returns without executing).
   - Each skill file has a matching `_test.go` (table-driven, often using testcontainers).

4. **Top-level package `ork`** — Public fluent API.
   - `NewNode()` / `NewNodeForHost(host)` / `NewNodeFromConfig(cfg)` → `NodeInterface` (embeds `RunnerInterface`).
   - `NewGroup(name)` → `GroupInterface`; `NewInventory()` → `InventoryInterface`. Inventory runs concurrently across nodes, controlled by `SetMaxConcurrency` (default 1 = sequential).
   - `NewCommand()` — fluent builder for one-off shell commands.
   - `registry.go` — `GetGlobalSkillRegistry()` (singleton, lazily initialized via `sync.Once`) and `NewDefaultRegistry()` (fresh registry pre-loaded with all built-in skills). Skills are registered by ID.

5. **`cmd/ork/`** — CLI binary. Currently only exposes `ork vault ...` subcommands (init/set/get/delete/list/changepassword/ui). `ui.go` runs a local HTTP UI for the vault. No CLI for running skills/nodes yet.

6. **`vault/`** — Encrypted secrets store. AES-256-GCM + Argon2id key derivation (`crypto.go`). File-backed, password-protected, with `KeySet/KeyGet/KeyDelete/KeyList/Save/ChangePassword`.

7. **`internal/`** — Test helpers (`skilltest`, `sshtest`).

8. **`test/integration/`** — End-to-end integration tests using testcontainers (linuxserver/openssh-server, geerlingguy/docker-ubuntu2404-ansible).

9. **`docs/`** — Static HTML documentation site (rendered via html-preview). Includes comparison pages vs Ansible/Chef/Puppet/SaltStack/CFEngine/Terraform/Pulumi/CloudFormation/SSH libs, plus proposals/tasks subdirs.

## Conventions

- **Fluent API**: every setter returns the receiver (`SetX`/`WithX` aliases). `WithX` is the "convenience" alias of `SetX`.
- **Interface + implementation split**: `*_interface.go` defines the public interface and constructors; `*_implementation.go` contains the private struct. Pattern repeated for node/group/inventory/command.
- **Idempotency**: every skill implements `Check()`; `Run()` calls `Check()` first and skips work if the system is already in the desired state. `Result.Changed` reports whether work was done.
- **Dry-run**: `NodeConfig.IsDryRunMode` (and `RunnerInterface.SetDryRunMode`) propagate to skills; skills log and return without mutating.
- **Concurrency safety**: skills are cloned per goroutine via `ToMap`/`FromMap`; the shared registry instance is never mutated during parallel execution.
- **Shell injection**: never interpolate raw user input into shell strings — use `skills.ShellEscapeArg` (see `shellEscapePackages` in `skills/apt/pkg_install.go`).
- **Error wrapping**: use `fmt.Errorf("...: %w", err)`. SSH errors flow through `ssh.classifySSHError` for actionable messages.
- **Comments**: do not add/remove comments unless asked. Existing files are heavily documented with package-level and per-method doc comments — match that style.
- **Language**: English only in all code and docs (per global rule).

## Build / Test / Verify

```powershell
go build -v ./...                                  # build everything
go build -o ork.exe ./cmd/ork                       # build CLI
go test ./...                                      # unit tests (no Docker needed for most)
go test -timeout 30m -v ./...                       # full suite incl. integration (needs Docker)
go test -race ./...                                 # race detector (see race_detector_test.go)
go vet ./...
```

Integration tests require Docker (testcontainers pulls `linuxserver/openssh-server` and `geerlingguy/docker-ubuntu2404-ansible`). CI: `.github/workflows/go.yml` runs `go build` then `go test -timeout 30m -v ./...` on Ubuntu with Go 1.26.

## Where to look for examples

- `examples/` — runnable examples for command, node, inventory, skill, playbook (each with `_test.go`).
- `skills/apt/pkg_install.go` — canonical skill implementation pattern (Check/Run, dry-run, shell escaping, fluent setters, `NewX` constructor).
- `ork.go` package doc — top-level usage cheatsheet.
- `docs/overview.html`, `docs/architecture.html`, `docs/api-reference.html` — long-form reference.
