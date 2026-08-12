# Source: Ork Skill Conventions (from codebase analysis)

**Source:** Codebase analysis of `skills/` package
**Retrieved:** 2026-08-12 (via read/grep tools)

## Summary

This document captures the conventions and patterns used by existing Ork skills, so the Docker skills proposal can match them exactly. All patterns are derived from reading the actual source code of `skills/caddy/`, `skills/fs/`, `skills/constants.go`, and `registry.go`.

## Skill ID Constants

Defined in `skills/constants.go`. Format: `ID<Package><Action>` = `"<package>-<action>"`.

Examples:
```go
IDCaddyInstall    = "caddy-install"
IDCaddyRestart    = "caddy-restart"
IDFSFileCreate    = "fs-file-create"
IDFSFileCopy      = "fs-file-copy"
IDMariadbInstall  = "mariadb-install"
IDSystemctlRestart = "systemctl-restart"
```

For Docker skills, the IDs would be:
```go
IDDockerInstall  = "docker-install"
IDDockerImport   = "docker-import"
IDDockerLoad     = "docker-load"
IDDockerRun      = "docker-run"
IDDockerStop     = "docker-stop"
IDDockerRm       = "docker-rm"
IDDockerRmi      = "docker-rmi"
IDDockerPull     = "docker-pull"
IDDockerTag      = "docker-tag"
IDDockerPs       = "docker-ps"
IDDockerImages   = "docker-images"
IDDockerRestart  = "docker-restart"
IDDockerExec     = "docker-exec"
```

## Package Structure

Each skill package follows this structure:
```
skills/<package>/
├── constants.go      # Arg key constants (Arg<Path>, Arg<Name>, etc.) + defaults
├── <skill1>.go       # Skill struct + Check() + Run() + setters + constructor
├── <skill1>_test.go  # Tests
├── <skill2>.go       # Another skill in the same package
├── <skill2>_test.go
└── helpers.go        # Shared helper functions (optional)
```

## Skill Struct Pattern

Every skill embeds `*types.BaseSkill` and implements `types.RunnableInterface`:

```go
type DockerRun struct {
    *types.BaseSkill
}

// Compile-time assertion
var _ types.RunnableInterface = (*DockerRun)(nil)

func (d *DockerRun) Check() (bool, error) { ... }
func (d *DockerRun) Run() types.Result { ... }

// Fluent setters
func (d *DockerRun) SetName(name string) *DockerRun { ... }
func (d *DockerRun) SetImage(image string) *DockerRun { ... }

// Standard interface methods (required by RunnableInterface)
func (d *DockerRun) SetArgs(args map[string]string) types.RunnableInterface { ... }
func (d *DockerRun) SetArg(key, value string) types.RunnableInterface { ... }
func (d *DockerRun) SetID(id string) types.RunnableInterface { ... }
func (d *DockerRun) SetDescription(desc string) types.RunnableInterface { ... }
func (d *DockerRun) SetTimeout(timeout time.Duration) types.RunnableInterface { ... }
func (d *DockerRun) WithNodeConfig(cfg types.NodeConfig) *DockerRun { ... }

// Constructor
func NewDockerRun() *DockerRun {
    pb := types.NewBaseSkill()
    pb.SetID(skills.IDDockerRun)
    pb.SetDescription("Run a Docker container (idempotent)")
    return &DockerRun{BaseSkill: pb}
}
```

## Check() Pattern

`Check()` returns `(bool, error)`:
- `true` = change needed (Run will execute)
- `false` = no change needed (already in desired state)
- Always check `cfg.IsDryRunMode` first and return `true` without running SSH commands
- Validate required args before running any SSH commands
- Use `ssh.Run(cfg, cmdCheck)` to probe remote state

Example from `skills/fs/file_create.go`:
```go
func (f *FileCreate) Check() (bool, error) {
    cfg := f.GetNodeConfig()
    path := f.GetArg(ArgPath)
    if err := validatePath(path); err != nil {
        return false, err
    }
    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would check if file exists...")
        return true, nil
    }
    if !fileExists(cfg, path) {
        return true, nil  // File doesn't exist — needs creation
    }
    // ... further checks (content, mode, owner)
    return false, nil  // Everything matches — no change needed
}
```

## Run() Pattern

`Run()` returns `types.Result`:
```go
type Result struct {
    Changed  bool
    Message  string
    Error    error
    Details  map[string]string
}
```

- Always validate args first
- Call `Check()` to see if changes are needed
- If `!needsChange`, return `Changed: false` with a message
- Check `cfg.IsDryRunMode` and log intent without executing
- Use `skills.ShellEscapeArg()` for ALL user-supplied values in commands
- Use `ssh.Run(cfg, cmd)` to execute
- Return `Changed: true` on success, `Changed: false` + `Error` on failure

## Shell Escaping

**Critical:** All user-supplied values MUST be shell-escaped via `skills.ShellEscapeArg()`:
```go
escName := skills.ShellEscapeArg(containerName)
escImage := skills.ShellEscapeArg(image)
cmd := types.Command{
    Command: fmt.Sprintf("docker run --name %s -d %s", escName, escImage),
}
```

For content (file writes), use `skills.ShellEscapeContent()`.

## Dry-Run Mode

Every skill must honor `cfg.IsDryRunMode`:
```go
if cfg.IsDryRunMode {
    cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command)
    return types.Result{Changed: true, Message: "Would ..."}
}
```

Dry-run mode should:
- Log what would be done (including the command)
- NOT execute any SSH commands
- NOT read local files (for skills that upload)
- Return `Changed: true` (assuming the change would happen)
- Never log `Stdin` (sensitive data) — see `TestRun_DryRun_DoesNotLogStdin` in `become_test.go`

## Privilege Escalation

Docker commands require root or `docker` group membership. Skills should use `BecomeUser`:
```go
cmd := types.Command{
    Command:    fmt.Sprintf("docker run --name %s -d %s", escName, escImage),
    BecomeUser: "root",  // or leave empty if user is in docker group
}
```

The SSH layer handles `BecomeUser` + `BecomePassword` via the `becomeWriter` state machine.

## Registry Registration

Skills are registered in `registry.go` → `NewDefaultRegistry()`:
```go
func NewDefaultRegistry() (*types.Registry, error) {
    reg := NewSkillRegistry()
    skills := []types.RunnableInterface{
        // ... existing skills ...
        docker.NewDockerInstall(),
        docker.NewDockerImport(),
        docker.NewDockerLoad(),
        docker.NewDockerRun(),
        docker.NewDockerStop(),
        docker.NewDockerRm(),
        docker.NewDockerRmi(),
        docker.NewDockerPull(),
        docker.NewDockerTag(),
        docker.NewDockerPs(),
        docker.NewDockerImages(),
        docker.NewDockerRestart(),
        docker.NewDockerExec(),
    }
    if err := reg.SetAll(skills); err != nil {
        return nil, err
    }
    return reg, nil
}
```

Import added at top:
```go
import (
    // ... existing imports ...
    "github.com/dracory/ork/skills/docker"
)
```

## Read-Only Skills (Check-Only)

Some skills never change state (e.g., `docker-ps`, `docker-images`). These should:
- `Check()` always returns `false` (no changes to make)
- `Run()` executes the read-only command and returns the output in `Result.Details` or `Result.Message`
- Marked as read-only in the description

## Non-Idempotent Skills

Some skills are intentionally non-idempotent (e.g., `docker-restart`, `docker-exec`). These should:
- `Check()` always returns `true` (always run)
- Document this clearly in the skill's doc comment
- Matches the pattern of `caddy.Restart` (line 58: "Check always returns true since Restart is intentionally non-idempotent")

## Apt-Based Installation Skills

For `docker-install`, follow the pattern of `caddy.Install` and `mariadb.Install`:
- Use `skills.DebianNonInteractive` constant (`DEBIAN_FRONTEND=noninteractive`)
- Use `skills.DpkgConfOptions` for non-interactive apt operations
- Use `BecomeUser: "root"` for all apt commands
- Check if already installed first (idempotency)
- Multi-step: add repo → apt update → install → verify
