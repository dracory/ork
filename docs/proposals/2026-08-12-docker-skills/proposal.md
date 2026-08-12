# Proposal: Docker Management Skills

**Date:** 2026-08-12
**Status:** Draft
**Author:** Devin (assisted)
**Research:** See `research/` subdirectory (10 source files, including Ansible `community.docker` analysis)
**Related:** [OCI Image Factory proposal](../2026-08-12-oci-image-factory/proposal.md) — Part 3 (local OCI factory) depends on this proposal's `docker-import` skill

## Problem Statement

Ork has **no Docker skills**. A grep across `skills/` for "docker", "container", or "oci" returns zero relevant matches. Users who want to deploy or manage containerized applications via Ork must manually shell out to `docker` commands — there are no idempotent, check-run skills for Docker operations.

This is a significant gap because Docker is the dominant container runtime. Ork already manages Caddy (web server), MariaDB (database), PHP-FPM (application server), systemd units, UFW firewall, and apt packages — all via the same idempotent check-run pattern. Docker management is a natural fit for this pattern, and its absence is conspicuous.

Every existing Ork skill package follows the same architecture:
1. A `constants.go` with arg keys and defaults
2. Skill structs embedding `*types.BaseSkill`, implementing `Check()` and `Run()`
3. Shell escaping via `skills.ShellEscapeArg()` for all user-supplied values
4. Dry-run mode support (`cfg.IsDryRunMode`)
5. Privilege escalation via `BecomeUser` (for root-only operations)
6. Registration in `registry.go` → `NewDefaultRegistry()`

Docker skills should follow this exact pattern. Each Docker CLI command maps to one skill, with `Check()` probing the current state and `Run()` executing the change.

## Proposed Solution

Add a `skills/docker/` package with 13 skills covering the core Docker lifecycle: installation, image management, container lifecycle, and inspection.

### Skill List

| Skill | ID | Type | Description |
|-------|----|------|-------------|
| `docker-install` | `docker-install` | Idempotent | Install Docker Engine via apt (Ubuntu/Debian) |
| `docker-pull` | `docker-pull` | Idempotent | Pull an image from a registry |
| `docker-tag` | `docker-tag` | Idempotent | Tag a local image |
| `docker-import` | `docker-import` | Idempotent | Import a flat tarball as an image |
| `docker-load` | `docker-load` | Idempotent | Load a Docker save archive |
| `docker-run` | `docker-run` | Idempotent | Run a container (create if absent, start if stopped) |
| `docker-stop` | `docker-stop` | Idempotent | Stop a running container |
| `docker-restart` | `docker-restart` | Non-idempotent | Restart a container (always runs) |
| `docker-rm` | `docker-rm` | Idempotent | Remove a container |
| `docker-rmi` | `docker-rmi` | Idempotent | Remove an image |
| `docker-exec` | `docker-exec` | Non-idempotent | Execute a command in a running container |
| `docker-ps` | `docker-ps` | Read-only | List containers |
| `docker-images` | `docker-images` | Read-only | List images |

### Skill Categories

**Installation (1 skill):**
- `docker-install` — prerequisite for all others. Adds Docker's apt repo, installs packages, verifies daemon is running.

**Image Management (4 skills):**
- `docker-pull` — download from registry
- `docker-tag` — create a tag alias
- `docker-import` — import flat tarball as image (no metadata)
- `docker-load` — load Docker save archive (with metadata)

**Container Lifecycle (4 skills):**
- `docker-run` — create and start a container (idempotent: checks if already running)
- `docker-stop` — stop a running container
- `docker-restart` — restart a container (non-idempotent, always runs)
- `docker-rm` — remove a container

**Image Removal (1 skill):**
- `docker-rmi` — remove an image

**Inspection/Execution (3 skills):**
- `docker-exec` — run a command inside a running container (non-idempotent)
- `docker-ps` — list containers (read-only)
- `docker-images` — list images (read-only)

## Implementation

### Package Structure

```
skills/docker/
├── constants.go           # Arg key constants + defaults
├── install.go             # docker-install skill
├── install_test.go
├── pull.go                # docker-pull skill
├── pull_test.go
├── tag.go                 # docker-tag skill
├── tag_test.go
├── import.go              # docker-import skill
├── import_test.go
├── load.go                # docker-load skill
├── load_test.go
├── run.go                 # docker-run skill
├── run_test.go
├── stop.go                # docker-stop skill
├── stop_test.go
├── restart.go             # docker-restart skill
├── restart_test.go
├── rm.go                  # docker-rm skill
├── rm_test.go
├── rmi.go                 # docker-rmi skill
├── rmi_test.go
├── exec.go                # docker-exec skill
├── exec_test.go
├── ps.go                  # docker-ps skill
├── ps_test.go
├── images.go              # docker-images skill
├── images_test.go
└── helpers.go             # Shared helpers (containerExists, imageExists, etc.)
```

### Constants (`skills/docker/constants.go`)

```go
package docker

// Argument key constants
const (
    ArgName        = "name"         // Container name
    ArgImage       = "image"        // Image name:tag
    ArgCommand     = "command"      // Command to run
    ArgArgs        = "args"         // Command arguments
    ArgTarball     = "tarball"      // Path to tarball (remote)
    ArgSource      = "source"       // Source image (for tag)
    ArgTarget      = "target"       // Target image:tag (for tag)
    ArgPorts       = "ports"        // Port mappings (e.g., "8080:80,443:443")
    ArgEnv         = "env"          // Environment variables (e.g., "DEBUG=true,FOO=bar")
    ArgVolumes     = "volumes"      // Volume mounts (e.g., "/host:/container")
    ArgNetwork     = "network"      // Network name
    ArgRestart     = "restart"      // Restart policy
    ArgUser        = "user"         // Container user
    ArgWorkdir     = "workdir"      // Working directory
    ArgForce       = "force"        // Force operation (recreate container, remove running)
    ArgDetach      = "detach"       // Run detached (default: true)
    ArgChanges     = "changes"      // Dockerfile instructions for import --change
    ArgTimeout     = "timeout"      // Stop/restart timeout in seconds
    ArgSignal      = "signal"       // Stop/restart signal
    ArgAll         = "all"          // Show all containers (ps -a)
    ArgFormat      = "format"       // Output format
    ArgQuiet       = "quiet"        // Quiet output
    ArgPlatform    = "platform"     // Platform (for pull/import/load)
    ArgAddDockerGroup = "add-docker-group" // Add user to docker group (install)
    ArgAlwaysPull  = "always-pull"  // Force pull even if image exists (pull)
    ArgPull        = "pull"         // Pull policy for docker-run: missing|always|never (borrowed from Ansible)
)

// Defaults
const (
    DefaultDetach  = "true"
    DefaultRestart = "unless-stopped"
    DefaultTimeout = "10"
    DefaultAll     = "false"
    DefaultForce   = "false"
    DefaultPull    = "missing"  // Only pull if image not present (matches Ansible's default)
)
```

### Skill ID Constants (added to `skills/constants.go`)

```go
// Docker skills (Docker Engine management)
// IDDockerInstall installs Docker Engine via apt (Ubuntu/Debian)
IDDockerInstall = "docker-install"

// IDDockerPull pulls an image from a registry
IDDockerPull = "docker-pull"

// IDDockerTag tags a local image
IDDockerTag = "docker-tag"

// IDDockerImport imports a flat tarball as a Docker image
IDDockerImport = "docker-import"

// IDDockerLoad loads a Docker save archive
IDDockerLoad = "docker-load"

// IDDockerRun runs a container (idempotent: creates if absent, starts if stopped)
IDDockerRun = "docker-run"

// IDDockerStop stops a running container
IDDockerStop = "docker-stop"

// IDDockerRestart restarts a container (non-idempotent)
IDDockerRestart = "docker-restart"

// IDDockerRm removes a container
IDDockerRm = "docker-rm"

// IDDockerRmi removes an image
IDDockerRmi = "docker-rmi"

// IDDockerExec executes a command in a running container (non-idempotent)
IDDockerExec = "docker-exec"

// IDDockerPs lists containers (read-only)
IDDockerPs = "docker-ps"

// IDDockerImages lists images (read-only)
IDDockerImages = "docker-images"
```

### Shared Helpers (`skills/docker/helpers.go`)

```go
package docker

import (
    "fmt"
    "strings"

    "github.com/dracory/ork/skills"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/types"
)

// containerExists checks if a container exists (running or stopped).
func containerExists(cfg types.NodeConfig, name string) bool {
    cmd := types.Command{
        Command:     fmt.Sprintf("docker ps -a -q -f name=^/%s$", skills.ShellEscapeArg(name)),
        Description: "Check if container exists: " + name,
        Required:    false,
        BecomeUser:  "root",
    }
    output, err := ssh.Run(cfg, cmd)
    return err == nil && strings.TrimSpace(output) != ""
}

// containerRunning checks if a container is currently running.
func containerRunning(cfg types.NodeConfig, name string) bool {
    cmd := types.Command{
        Command:     fmt.Sprintf("docker ps -q -f name=^/%s$", skills.ShellEscapeArg(name)),
        Description: "Check if container is running: " + name,
        Required:    false,
        BecomeUser:  "root",
    }
    output, err := ssh.Run(cfg, cmd)
    return err == nil && strings.TrimSpace(output) != ""
}

// imageExists checks if an image exists locally.
func imageExists(cfg types.NodeConfig, image string) bool {
    cmd := types.Command{
        Command:     fmt.Sprintf("docker image inspect %s >/dev/null 2>&1", skills.ShellEscapeArg(image)),
        Description: "Check if image exists: " + image,
        Required:    false,
        BecomeUser:  "root",
    }
    _, err := ssh.Run(cfg, cmd)
    return err == nil
}

// dockerInstalled checks if Docker Engine is installed and the daemon is running.
func dockerInstalled(cfg types.NodeConfig) bool {
    cmd := types.Command{
        Command:     "docker info >/dev/null 2>&1",
        Description: "Check if Docker is installed and daemon is running",
        Required:    false,
    }
    _, err := ssh.Run(cfg, cmd)
    return err == nil
}
```

### Skill: docker-run (Idempotent — the key skill)

```go
package docker

import (
    "fmt"
    "strings"
    "time"

    "github.com/dracory/ork/skills"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/types"
)

// DockerRun runs a Docker container idempotently.
//
// If the container doesn't exist, it creates and starts it (docker run -d).
// If the container exists but is stopped, it starts it (docker start).
// If the container is already running, Check() returns false (no change needed).
// If force=true, the container is stopped, removed, and recreated (matches
// Ansible's recreate=true parameter).
//
// The pull arg controls image pulling before run (matches Ansible's pull
// parameter and Docker CLI's --pull flag):
//   - missing (default): only pull if image not present locally
//   - always: always pull latest version before run
//   - never: never pull; fail if image not present
//
// Usage:
//
//	node.Run(docker.NewDockerRun().
//	    SetName("myapp").
//	    SetImage("myapp:v1").
//	    SetPorts("8080:80").
//	    SetEnv("DEBUG=true").
//	    SetRestart("unless-stopped"))
//	// Force recreate (like Ansible recreate=true):
//	node.Run(docker.NewDockerRun().SetName("myapp").SetImage("myapp:v2").SetForce("true"))
//	// Always pull latest before run:
//	node.Run(docker.NewDockerRun().SetName("myapp").SetImage("myapp:latest").SetPull("always"))
type DockerRun struct {
    *types.BaseSkill
}

var _ types.RunnableInterface = (*DockerRun)(nil)

func (d *DockerRun) Check() (bool, error) {
    cfg := d.GetNodeConfig()
    name := d.GetArg(ArgName)
    image := d.GetArg(ArgImage)

    if name == "" {
        return false, fmt.Errorf("no container name specified: set the %q argument", ArgName)
    }
    if image == "" {
        return false, fmt.Errorf("no image specified: set the %q argument", ArgImage)
    }

    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would check if container is running",
            "name", name)
        return true, nil
    }

    // If container is already running, no change needed
    if containerRunning(cfg, name) {
        return false, nil
    }

    // Container doesn't exist or is stopped — needs action
    return true, nil
}

func (d *DockerRun) Run() types.Result {
    name := d.GetArg(ArgName)
    image := d.GetArg(ArgImage)
    ports := d.GetArg(ArgPorts)
    env := d.GetArg(ArgEnv)
    volumes := d.GetArg(ArgVolumes)
    network := d.GetArg(ArgNetwork)
    restart := d.GetArg(ArgRestart)
    user := d.GetArg(ArgUser)
    workdir := d.GetArg(ArgWorkdir)
    command := d.GetArg(ArgCommand)

    if restart == "" {
        restart = DefaultRestart
    }

    // ... validation ...

    cfg := d.GetNodeConfig()

    needsChange, err := d.Check()
    if err != nil {
        return types.Result{Changed: false, Message: "Failed to check container state", Error: err}
    }
    if !needsChange {
        return types.Result{Changed: false, Message: "Container already running: " + name}
    }

    // If container exists but is stopped, start it
    if containerExists(cfg, name) && !containerRunning(cfg, name) {
        cmdStart := types.Command{
            Command:     fmt.Sprintf("docker start %s", skills.ShellEscapeArg(name)),
            Description: "Start existing container: " + name,
            BecomeUser:  "root",
        }
        if cfg.IsDryRunMode {
            cfg.GetLoggerOrDefault().Info("dry-run: would start container", "cmd", cmdStart.Command)
            return types.Result{Changed: true, Message: "Would start container: " + name}
        }
        _, err := ssh.Run(cfg, cmdStart)
        if err != nil {
            return types.Result{Changed: false, Message: "Failed to start container", Error: err}
        }
        return types.Result{Changed: true, Message: "Container started: " + name}
    }

    // Container doesn't exist — create and run
    var cmdParts []string
    cmdParts = append(cmdParts, "docker", "run", "--name", skills.ShellEscapeArg(name), "-d")

    if restart != "" {
        cmdParts = append(cmdParts, "--restart", skills.ShellEscapeArg(restart))
    }
    // ... add ports, env, volumes, network, user, workdir ...
    cmdParts = append(cmdParts, skills.ShellEscapeArg(image))
    if command != "" {
        cmdParts = append(cmdParts, command) // user's command
    }

    cmdRun := types.Command{
        Command:     strings.Join(cmdParts, " "),
        Description: "Run container: " + name + " from " + image,
        BecomeUser:  "root",
    }

    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would run container", "cmd", cmdRun.Command)
        return types.Result{Changed: true, Message: "Would run container: " + name}
    }

    output, err := ssh.Run(cfg, cmdRun)
    if err != nil {
        return types.Result{
            Changed: false,
            Message: "Failed to run container",
            Error:   fmt.Errorf("failed to run container: %w\nOutput: %s", err, output),
        }
    }

    return types.Result{
        Changed: true,
        Message: "Container running: " + name,
        Details: map[string]string{
            "name":  name,
            "image": image,
        },
    }
}

// Fluent setters, SetArgs, SetArg, SetID, etc. follow the standard pattern.
// ... (see research/09-ork-skill-conventions.md for the full list)

func NewDockerRun() *DockerRun {
    pb := types.NewBaseSkill()
    pb.SetID(skills.IDDockerRun)
    pb.SetDescription("Run a Docker container (idempotent: creates if absent, starts if stopped)")
    return &DockerRun{BaseSkill: pb}
}
```

### Skill: docker-install (Idempotent)

```go
func (d *DockerInstall) Check() (bool, error) {
    cfg := d.GetNodeConfig()

    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would check if Docker is installed")
        return true, nil
    }

    if dockerInstalled(cfg) {
        return false, nil // Already installed
    }
    return true, nil
}

func (d *DockerInstall) Run() types.Result {
    cfg := d.GetNodeConfig()

    needsChange, err := d.Check()
    if err != nil {
        return types.Result{Changed: false, Error: err}
    }
    if !needsChange {
        return types.Result{Changed: false, Message: "Docker already installed"}
    }

    // Multi-step installation:
    // 1. Remove conflicting packages
    // 2. Add Docker's GPG key
    // 3. Add Docker's apt repository
    // 4. apt update
    // 5. Install docker-ce, docker-ce-cli, containerd.io, plugins
    // 6. Verify daemon is running
    // 7. Optionally add user to docker group

    // Each step is a separate types.Command with BecomeUser: "root"
    // Using skills.DebianNonInteractive and skills.DpkgConfOptions
    // ... (full implementation follows caddy.Install pattern)
}
```

### Skill: docker-ps (Read-Only)

```go
func (d *DockerPs) Check() (bool, error) {
    return false, nil // Read-only skill — never changes state
}

func (d *DockerPs) Run() types.Result {
    cfg := d.GetNodeConfig()
    all := d.GetArg(ArgAll)
    format := d.GetArg(ArgFormat)

    var cmdStr string
    cmdStr = "docker ps"
    if isTrue(all) {
        cmdStr += " -a"
    }
    if format != "" {
        cmdStr += " --format " + skills.ShellEscapeArg(format)
    }

    cmd := types.Command{
        Command:     cmdStr,
        Description: "List containers",
        BecomeUser:  "root",
    }

    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would list containers", "cmd", cmd.Command)
        return types.Result{Changed: false, Message: "Would list containers"}
    }

    output, err := ssh.Run(cfg, cmd)
    if err != nil {
        return types.Result{Changed: false, Error: err}
    }

    return types.Result{
        Changed: false, // Read-only — never reports changes
        Message: output,
        Details: map[string]string{"output": output},
    }
}
```

### Registry Registration (in `registry.go`)

```go
import (
    // ... existing imports ...
    "github.com/dracory/ork/skills/docker"
)

func NewDefaultRegistry() (*types.Registry, error) {
    reg := NewSkillRegistry()
    skills := []types.RunnableInterface{
        // ... existing skills ...
        docker.NewDockerInstall(),
        docker.NewDockerPull(),
        docker.NewDockerTag(),
        docker.NewDockerImport(),
        docker.NewDockerLoad(),
        docker.NewDockerRun(),
        docker.NewDockerStop(),
        docker.NewDockerRestart(),
        docker.NewDockerRm(),
        docker.NewDockerRmi(),
        docker.NewDockerExec(),
        docker.NewDockerPs(),
        docker.NewDockerImages(),
    }
    // ...
}
```

## Benefits

1. **Fills a conspicuous gap** — Ork manages web servers, databases, PHP, firewalls, but not Docker, the dominant container runtime
2. **Idempotent Docker management** — safe to run multiple times (check-run pattern), matching all existing Ork skills
3. **Consistent with Ork's design** — pure SSH command execution, same architecture as `caddy`, `mariadb`, `php`, `systemctl` packages
4. **No new dependencies** — Docker skills shell out to the `docker` CLI on the remote server via SSH; no Go libraries needed
5. **Inventory support** — deploy containers across a fleet concurrently (Ork's parallel execution)
6. **Shell injection prevention** — all user-supplied values escaped via `skills.ShellEscapeArg()`
7. **Dry-run support** — all skills honor `cfg.IsDryRunMode`
8. **Privilege escalation** — all skills use `BecomeUser: "root"` (Docker requires root or docker group)
9. **Composable with existing skills** — `docker-install` uses apt patterns; `docker-run` can follow `fs.NewFileCreate()` for config file deployment; `systemctl` skills can manage the Docker daemon

## Challenges & Solutions

### Challenge 1: Docker requires root or docker group membership
**Problem:** Most `docker` commands require root or membership in the `docker` group.
**Solution:** All Docker skills default to `BecomeUser: "root"`. The `docker-install` skill optionally adds the SSH user to the `docker` group (via `ArgAddDockerGroup`), after which `BecomeUser` can be omitted.

### Challenge 2: Docker may not be installed
**Problem:** Docker skills assume Docker is present on the remote server.
**Solution:** `docker-install` skill handles installation. All other skills' `Check()` methods call `dockerInstalled()` first and return a clear error if Docker is missing.

### Challenge 3: `docker-run` idempotency is complex
**Problem:** `docker run` is not idempotent by default. Running it twice creates two containers.
**Solution:** The `docker-run` skill uses `--name` (required arg) and probes state in `Check()`:
- Container running → `false` (no change)
- Container exists but stopped → `true`, `Run()` calls `docker start`
- Container doesn't exist → `true`, `Run()` calls `docker run -d --name ...`
- Optionally: if container exists with different config (image, ports), `--force` arg removes and recreates

### Challenge 4: `docker-import` and `docker-load` need tarballs on the remote server
**Problem:** These skills require tarballs to already be on the remote server. Ork can write text files today (`file_create`, `Command.Stdin`) but lacks streaming binary transfer for large tarballs.
**Solution:** These skills accept a **remote path** (the tarball must already be on the server). File transfer is a separate concern — see the [SFTP proposal](../2026-08-12-oci-image-factory/proposal.md) in the OCI factory proposal. Users can also transfer tarballs via `scp` manually or use `docker pull` from a registry instead.

### Challenge 5: `docker-exec` and `docker-restart` are non-idempotent
**Problem:** These operations should run every time, not skip if "already done."
**Solution:** `Check()` always returns `true` for these skills, matching the pattern of `caddy.Restart` (line 58 of `skills/caddy/restart.go`: "Check always returns true since Restart is intentionally non-idempotent"). This is documented in each skill's doc comment.

### Challenge 6: Docker daemon not running after install
**Problem:** Docker might be installed but the daemon not started.
**Solution:** `docker-install` verifies the daemon is running via `docker info` after installation. If not running, it starts the service via `systemctl start docker` (using Ork's existing `systemctl` skills).

## Implementation Plan

### Phase 1: Core Skills (MVP)
1. Create `skills/docker/constants.go` with arg keys and defaults
2. Create `skills/docker/helpers.go` with `containerExists`, `containerRunning`, `imageExists`, `dockerInstalled`
3. Implement `docker-install` (prerequisite for all others)
4. Implement `docker-run` (the primary deployment skill)
5. Implement `docker-ps` (read-only, needed for debugging)
6. Implement `docker-stop` (companion to `docker-run`)
7. Add Docker skill ID constants to `skills/constants.go`
8. Register skills in `registry.go`
9. Write tests for Phase 1 skills

### Phase 2: Image Management
10. Implement `docker-pull`
11. Implement `docker-tag`
12. Implement `docker-rmi`
13. Implement `docker-images` (read-only)
14. Write tests for Phase 2 skills

### Phase 3: Advanced Operations
15. Implement `docker-import`
16. Implement `docker-load`
17. Implement `docker-rm`
18. Implement `docker-restart` (non-idempotent)
19. Implement `docker-exec` (non-idempotent)
20. Write tests for Phase 3 skills

### Phase 4: Integration & Documentation
21. Run full test suite (`go test ./skills/docker/...`)
22. Run `go vet` and `golangci-lint`
23. Update `docs/skills.html` with Docker skills documentation
24. Add `examples/example_docker_deploy.go` — end-to-end: install → pull → run → verify
25. Update `AGENTS.md` with Docker skills build/test commands

## Success Metrics

- [ ] All 13 Docker skills implemented with `Check()` and `Run()`
- [ ] All skills pass `go test ./skills/docker/...`
- [ ] `go vet ./...` passes with no new warnings
- [ ] Idempotent skills return `Changed: false` on second run
- [ ] Non-idempotent skills (`docker-restart`, `docker-exec`) always return `Changed: true`
- [ ] Read-only skills (`docker-ps`, `docker-images`) always return `Changed: false`
- [ ] All skills honor `cfg.IsDryRunMode` (log and return without executing)
- [ ] All user-supplied values escaped via `skills.ShellEscapeArg()`
- [ ] All skills use `BecomeUser: "root"` by default
- [ ] All skills registered in `NewDefaultRegistry()` and discoverable via `GetGlobalSkillRegistry()`
- [ ] All skill ID constants added to `skills/constants.go`
- [ ] No new external dependencies added to `go.mod`

## Open Questions

1. **Should `docker-run` support config drift detection?**
   - If a container is running with different ports/env/image than specified, should `Check()` return `true` and `Run()` recreate it?
   - Option A: No — if container is running, leave it alone (safe, simple)
   - Option B: Yes, with `--force` arg — detect drift and recreate if `force=true`
   - Recommendation: Option A for Phase 1 (safe default), Option B as a future enhancement

2. **Should `docker-install` support non-Debian distributions?**
   - The current scope is Ubuntu/Debian (apt-based), matching Ork's existing apt-focused skills
   - CentOS/RHEL/Fedora would use `dnf`/`yum` — different package manager
   - Recommendation: Ubuntu/Debian only for now. Add RHEL family as a separate proposal if needed.

3. **Should there be a `docker-login` skill for registry authentication?**
   - `docker pull` from private registries requires `docker login` first
   - `docker login` involves credentials (sensitive)
   - Recommendation: Track as a future skill. For now, users handle `docker login` manually or via `docker-exec`.

4. **Should there be a `docker-compose` skill?**
   - Docker Compose manages multi-container applications
   - Out of scope for this proposal (single-container focus)
   - Recommendation: Track as a separate future proposal

5. **Should `docker-ps` and `docker-images` support `--format json`?**
   - JSON output is structured and easier to parse programmatically
   - Could return structured data in `Result.Details`
   - Recommendation: Yes — support `--format` arg, default to table format, allow `json` for structured output

## Research Sources

See `research/` subdirectory:
- `01-docker-run.md` — `docker container run` command reference
- `02-docker-ps.md` — `docker container ls` / `docker ps` command reference
- `03-docker-stop-restart.md` — `docker stop` and `docker restart` command references
- `04-docker-rm-rmi.md` — `docker rm` and `docker rmi` command references
- `05-docker-pull-tag.md` — `docker pull` and `docker tag` command references
- `06-docker-exec.md` — `docker exec` command reference
- `07-docker-install-ubuntu.md` — Docker Engine installation on Ubuntu
- `08-docker-import-load.md` — `docker import` and `docker load` (cross-referenced from OCI factory proposal)
- `09-ork-skill-conventions.md` — Ork's skill patterns derived from codebase analysis

## Related Proposals

- [OCI Image Factory](../2026-08-12-oci-image-factory/proposal.md) — Part 3 (local OCI factory) produces tarballs that `docker-import` consumes. The SFTP proposal (Part 2 of the OCI factory proposal) enables transferring those tarballs to remote servers.
- [Playbooks](../implemented/2026-04-15-playbooks.md) — Could orchestrate Docker skills in sequences (e.g., install → pull → run → verify)
- [Privilege Escalation](../implemented/2026-04-15-privilege-escalation.md) — Used by all Docker skills for `sudo docker` operations
- [Parallel Execution](../implemented/2026-04-15-parallel-execution.md) — Enables fleet-wide Docker deployment

## Related Resources

- [Project README](../../README.md)
- [Skills Documentation](../../docs/skills.html)
- [Architecture Documentation](../../docs/architecture.html)
- [Existing proposals](../README.md)
