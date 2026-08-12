# Proposal: Container Image Skills and OCI Image Factory

**Date:** 2026-08-12
**Status:** Draft
**Author:** Devin (assisted)
**Research:** See `research/` subdirectory (10 source files)
**Related:** [Docker Management Skills proposal](../2026-08-12-docker-skills/proposal.md) — Part 1 (Docker skills) has been extracted into its own standalone proposal with dedicated research

> **Note:** Part 1 (Docker skills) has been extracted into a standalone proposal at [`../2026-08-12-docker-skills/proposal.md`](../2026-08-12-docker-skills/proposal.md) with its own 9-file research directory. This proposal focuses on Parts 2 (SFTP) and 3 (OCI factory), and references the standalone Docker skills proposal where relevant.

## Problem Statement

Ork has **no Docker or container image skills**. A grep across `skills/` for "docker", "container", "oci", or "image" returns zero relevant matches (the few hits are false positives in test fixtures and mariadb purge). Users who want to deploy containerized applications via Ork must manually shell out to `docker` commands — there are no idempotent, check-run skills for Docker operations.

Meanwhile, a blog post by Cheikh Seck ("Building a Docker Image Factory From Scratch in Go") demonstrated that OCI container images can be built using only Go's standard library (`archive/tar`, `crypto/sha256`, `encoding/json`) in ~150 lines of code, with zero external dependencies. This raises the question: should Ork integrate this factory pattern, and if so, how?

The core tension is that the factory is a **build-time tool** (runs locally, produces image artifacts), while Ork skills run **on remote servers via SSH**. These execution contexts don't naturally overlap, so the proposal must carefully scope what belongs where.

Additionally, Ork's `ssh` package has **no efficient binary file transfer**. It can write text files today via `fs.NewFileCreate()` (which uses `printf '%s' > path` over SSH) and via `types.Command.Stdin` (a `string` field piped to the remote command's stdin). Both work for config files and scripts. Neither is suitable for transferring large binary tarballs (Docker images can be hundreds of MB) — `Stdin` is a string loaded entirely into memory, not a streaming `io.Reader`, and `printf` breaks on null bytes. Any Docker deployment workflow that needs to get a tarball to a remote server requires a proper streaming file-transfer mechanism.

## Proposed Solution

A **three-part proposal** with clear separation of concerns:

### Part 1: Docker Skills Package (Primary — fills the real gap)

> **Extracted to standalone proposal:** [`../2026-08-12-docker-skills/proposal.md`](../2026-08-12-docker-skills/proposal.md)
>
> The Docker skills package has been extracted into its own proposal with 9 dedicated research files covering each Docker CLI command, idempotency patterns, and Ork's skill conventions. See the standalone proposal for the full skill list (13 skills), implementation details, and phased rollout plan.

Add `skills/docker/` with idempotent Docker management skills that run on remote servers via SSH. These are pure command-execution skills — exactly what Ork is built for. No factory dependency.

**Skills (13 total — see standalone proposal for details):**
- `docker-install` — Install Docker Engine (apt-based)
- `docker-import` — Import a flat tarball as an image (`docker import`)
- `docker-load` — Load a Docker save archive (`docker load`)
- `docker-run` — Run a container (idempotent: check if already running)
- `docker-stop` — Stop a container
- `docker-rm` — Remove a container
- `docker-rmi` — Remove an image
- `docker-pull` — Pull an image from a registry
- `docker-tag` — Tag an image
- `docker-ps` — List containers (read-only)
- `docker-images` — List images (read-only)
- `docker-restart` — Restart a container (non-idempotent)
- `docker-exec` — Execute a command in a running container (non-idempotent)

Each skill follows the existing pattern: embed `*types.BaseSkill`, implement `Check()` (idempotency probe) and `Run()`, honor `cfg.IsDryRunMode`, use `skills.ShellEscapeArg` for shell injection prevention.

### Part 2: SFTP File Transfer (Enabler — needed for Part 1)

Add SFTP file transfer to the `ssh` package using `github.com/pkg/sftp`. This is required for `docker-import` and `docker-load` skills to transfer large binary tarballs to remote servers efficiently.

Ork can already write **text files** today via `fs.NewFileCreate()` (uses `printf '%s' > path` over SSH) and via `types.Command.Stdin` (a `string` piped to the remote command's stdin). Both are fine for configs and scripts. Neither works for the Docker use case:
- `Stdin` is a `string` loaded entirely into memory — a 500MB tarball would consume 500MB of RAM
- `printf '%s'` with shell-escaped content breaks on null bytes (binary data)
- Neither provides streaming, progress tracking, or integrity verification

SFTP solves this by streaming via `io.Copy` without loading the full file into memory.

**New files:**
- `ssh/sftp.go` — `SFTPClient` type wrapping `pkg/sftp.Client`
- Method on `ssh.Client`: `func (c *Client) SFTP() (*SFTPClient, error)`
- `SFTPClient.UploadFile(localPath, remotePath) error`
- `SFTPClient.DownloadFile(remotePath, localPath) error`
- `SFTPClient.Close() error`

**New dependencies:**
- `github.com/pkg/sftp` (MIT, ~2K stars, mature)
- `github.com/kr/fs` (transitive, ~100 lines)

Ork already has `golang.org/x/crypto` and `golang.org/x/sys` — the only truly new dependency is `pkg/sftp` + its tiny `kr/fs` transitive.

### Part 3: Local OCI Image Factory (Optional — future phase)

Add a local OCI image factory as `pkg/oci` (or `internal/oci`) + `ork oci build` CLI subcommand. This is the gocker pattern, adapted and fixed for spec compliance. This is a **local build tool**, not an SSH skill — it runs on the control machine, not on remote servers.

**Scope:**
- `pkg/oci/builder.go` — OCI image layout builder (stdlib only)
- `pkg/oci/tarball.go` — flat tarball creator for `docker import`
- `cmd/ork/oci.go` — `ork oci build` CLI subcommand
- Optional: `ork oci build --cross-compile` to cross-compile a Go binary before packaging

**Spec compliance fixes over gocker:**
- Config referenced by descriptor (not embedded inline in manifest)
- Proper `mediaType` on config and layer descriptors
- Proper `index.json` as image index object (not bare array)
- Preserve original file permissions (`info.Mode().Perm()`, not hardcoded 0755)
- Streaming SHA-256 (avoid loading entire layers into memory)
- Optional gzip compression for layers

**This part is explicitly marked as future/optional** because:
1. It doesn't fill an urgent gap (Ork's core mission is SSH automation, not image building)
2. Users can build images with `docker build`, `buildah`, `ko`, or any existing tool
3. The value is ownership/control (tiny images, no daemon) — nice but not essential
4. It's a separate concern from Ork's skill system

## Implementation

### Part 1: Docker Skills

#### Skill: docker-import

```go
package docker

import (
    "fmt"
    "github.com/dracory/ork/skills"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/types"
)

// DockerImport imports a flat filesystem tarball as a Docker image.
// The tarball must already exist on the remote server (use SFTP to transfer it first).
//
// Args:
//   - tarball: Path to the tarball on the remote server (required)
//   - image: Image name with optional tag, e.g. "myapp:v1" (required)
//   - changes: Dockerfile instructions to apply, e.g. "CMD [\"/app/server\"]" (optional)
type DockerImport struct {
    *types.BaseSkill
}

func NewDockerImport() *DockerImport {
    pb := types.NewBaseSkill()
    pb.SetID(skills.IDDockerImport)
    pb.SetDescription("Import a tarball as a Docker image (docker import)")
    return &DockerImport{BaseSkill: pb}
}

func (d *DockerImport) Check() (bool, error) {
    tarball := d.GetArg(ArgTarball)
    if tarball == "" {
        return false, fmt.Errorf("no tarball specified: set the %q argument", ArgTarball)
    }

    image := d.GetArg(ArgImage)
    if image == "" {
        return false, fmt.Errorf("no image specified: set the %q argument", ArgImage)
    }

    cfg := d.GetNodeConfig()

    if cfg.IsDryRunMode {
        cfg.GetLoggerOrDefault().Info("dry-run: would check if image exists")
        return true, nil
    }

    // Check if the image already exists
    cmdCheck := types.Command{
        Command:     fmt.Sprintf("docker image inspect %s >/dev/null 2>&1", skills.ShellEscapeArg(image)),
        Description: "Check if image already exists: " + image,
        Required:    false, // non-zero exit means image doesn't exist, which is fine
    }
    _, err := ssh.Run(cfg, cmdCheck)
    if err == nil {
        // Image already exists
        return false, nil
    }
    return true, nil
}

func (d *DockerImport) Run() types.Result {
    tarball := d.GetArg(ArgTarball)
    image := d.GetArg(ArgImage)
    changes := d.GetArg(ArgChanges)

    // ... validation, dry-run check, build docker import command with --change flags,
    // execute via ssh.Run, return Result with Changed=true
    // ...
}
```

#### Skill: docker-run (idempotent)

```go
func (d *DockerRun) Check() (bool, error) {
    name := d.GetArg(ArgName)
    // Check if container is already running
    cmdCheck := types.Command{
        Command:     fmt.Sprintf("docker ps -q -f name=^/%s$", skills.ShellEscapeArg(name)),
        Description: "Check if container is running: " + name,
        Required:    false,
    }
    output, err := ssh.Run(cfg, cmdCheck)
    if err == nil && strings.TrimSpace(output) != "" {
        return false, nil // already running
    }
    return true, nil
}
```

#### Registry Registration

Add to `registry.go`:
```go
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
```

### Part 2: SFTP Integration

```go
// ssh/sftp.go
package ssh

import (
    "io"
    "os"

    sftp "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

type SFTPClient struct {
    client *sftp.Client
}

func (c *Client) SFTP() (*SFTPClient, error) {
    if c.client == nil {
        return nil, fmt.Errorf("not connected, call Connect() first")
    }
    sftpClient, err := sftp.NewClient(c.client)
    if err != nil {
        return nil, fmt.Errorf("failed to create SFTP session: %w", err)
    }
    return &SFTPClient{client: sftpClient}, nil
}

func (s *SFTPClient) UploadFile(localPath, remotePath string) error {
    localFile, err := os.Open(localPath)
    if err != nil {
        return fmt.Errorf("open local file: %w", err)
    }
    defer localFile.Close()

    remoteFile, err := s.client.Create(remotePath)
    if err != nil {
        return fmt.Errorf("create remote file: %w", err)
    }
    defer remoteFile.Close()

    _, err = io.Copy(remoteFile, localFile)
    return err
}

func (s *SFTPClient) DownloadFile(remotePath, localPath string) error { /* ... */ }
func (s *SFTPClient) Close() error { return s.client.Close() }
```

### Part 3: OCI Factory (Future)

```go
// pkg/oci/builder.go (future phase)
package oci

import (
    "archive/tar"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

type Builder struct {
    outputDir string
    layers    []Layer
    config    Config
}

// Spec-compliant manifest with config as descriptor (not inline)
type Manifest struct {
    SchemaVersion int          `json:"schemaVersion"`
    MediaType     string       `json:"mediaType"`
    Config        Descriptor   `json:"config"`
    Layers        []Descriptor `json:"layers"`
}

type Descriptor struct {
    MediaType string `json:"mediaType"`
    Digest    string `json:"digest"`
    Size      int64  `json:"size"`
}

// ... fixed implementation with proper descriptors, media types, index.json
```

## Benefits

### Part 1 (Docker Skills)
- **Fills a real gap** — Ork has zero Docker skills today
- **Idempotent Docker management** — safe to run multiple times (check-run pattern)
- **Consistent with Ork's design** — pure SSH command execution, same as existing skills
- **Works with any image source** — `docker import` (flat tarball), `docker load` (Docker save), `docker pull` (registry)
- **Inventory support** — deploy containers across a fleet concurrently

### Part 2 (SFTP)
- **Unlocks file transfer** — essential for Docker deployment and many other use cases
- **Minimal dependencies** — only `github.com/pkg/sftp` + tiny `kr/fs`
- **General-purpose** — useful for config deployment, log retrieval, backups, not just Docker
- **Aligns with existing SSH architecture** — works on top of Ork's existing `ssh.Client`

### Part 3 (OCI Factory)
- **Zero-dependency image building** — stdlib only, aligns with Ork's minimal-dep philosophy
- **Tiny images** — no base OS, just your binary (8MB for a Go server)
- **Ownership** — understand and control every byte of the image format
- **Cross-platform** — build Linux images on any host (Windows, macOS, Linux)

## Challenges & Solutions

### Challenge 1: Binary file transfer gap
**Problem:** `docker-import` and `docker-load` need large binary tarballs on the remote server. Ork can write text files today (`file_create`, `Command.Stdin`), but neither works for large binaries — `Stdin` is a string loaded into memory, and `printf` breaks on null bytes.
**Solution:** Part 2 (SFTP). Low risk — `pkg/sftp` is mature, works on top of existing SSH connections, and streams via `io.Copy` without loading the full file into memory.

### Challenge 2: SFTP may be disabled on hardened servers
**Problem:** Some SSH servers disable the SFTP subsystem.
**Solution:** Document the requirement. Provide a fallback skill that uses `cat > file` via SSH pipe (slower, no integrity check, but works everywhere). The SFTP path is preferred but not mandatory.

### Challenge 3: gocker factory is not spec-compliant
**Problem:** The blog's factory has bugs (inline config, missing media types, bare array index.json, hardcoded permissions).
**Solution:** Part 3 is marked as future. If implemented, fix all spec compliance issues (see research file 03 for the full list). Use `google/go-containerregistry` instead if full spec compliance and registry push are needed (see research file 08).

### Challenge 4: Docker may not be installed on remote servers
**Problem:** Docker skills assume Docker is present.
**Solution:** `docker-install` skill as part of Part 1. Check for Docker presence in `Check()` of all Docker skills and return a clear error if missing.

### Challenge 5: Docker operations require root or docker group membership
**Problem:** Most `docker` commands require root or membership in the `docker` group.
**Solution:** Use Ork's existing privilege escalation (`BecomeUser`, `BecomePassword`). Docker skills should default to `BecomeUser: "root"` unless the user is already in the `docker` group.

### Challenge 6: Scope creep
**Problem:** The factory pattern is interesting but doesn't fit Ork's core SSH-automation mission.
**Solution:** Part 3 is explicitly optional/future. Parts 1 and 2 are the priority. The factory can be revisited after Docker skills prove useful.

## Implementation Plan

### Phase 1: SFTP File Transfer (Part 2)
1. Add `github.com/pkg/sftp` to `go.mod`
2. Create `ssh/sftp.go` with `SFTPClient` type
3. Add `SFTP()` method to `ssh.Client`
4. Add `UploadFile`, `DownloadFile`, `Close` methods
5. Write tests (`ssh/sftp_test.go`) using testcontainers SSH server
6. Document in `docs/` that SFTP is now supported

### Phase 2: Docker Skills (Part 1 — see standalone proposal)
> **Full details:** [`../2026-08-12-docker-skills/proposal.md`](../2026-08-12-docker-skills/proposal.md)
>
> The Docker skills implementation has been extracted into a standalone proposal with a 4-phase rollout plan (Core MVP → Image Management → Advanced Operations → Integration & Docs). See the standalone proposal for the complete implementation plan, skill specifications, and test strategy.

1. Create `skills/docker/` package
2. Add `skills/docker/constants.go` with skill IDs and arg names
3. Implement `docker-install` skill (prerequisite for all others)
4. Implement `docker-import` skill (uses SFTP to transfer tarball, then `docker import`)
5. Implement `docker-load` skill (uses SFTP to transfer archive, then `docker load`)
6. Implement `docker-run` skill (idempotent: checks if container is running)
7. Implement `docker-stop`, `docker-rm`, `docker-rmi`, `docker-pull`, `docker-tag`
8. Implement `docker-ps`, `docker-images` (read-only skills)
9. Implement `docker-restart`, `docker-exec` (non-idempotent skills)
10. Write tests for each skill (`*_test.go`) using testcontainers
11. Register all skills in `registry.go`
12. Add `skills/docker/` to documentation

### Phase 3: OCI Factory (Part 3 — Future, Optional)
1. Evaluate whether to use stdlib-only approach or `google/go-containerregistry`
2. If stdlib: create `pkg/oci/builder.go` with spec-compliant implementation
3. Add `ork oci build` CLI subcommand in `cmd/ork/`
4. Add optional `--cross-compile` flag for Go binary cross-compilation
5. Write tests for OCI layout compliance
6. Document the factory and its relationship to Docker skills

### Phase 4: Integration Examples (Future)
1. Add `examples/example_docker_deploy.go` — end-to-end: build → transfer → import → run
2. Add `examples/example_oci_factory.go` — local OCI image building
3. Update `docs/` with Docker skills and OCI factory documentation

## Success Metrics

### Part 1 (Docker Skills)
- [ ] All 13 Docker skills implemented with `Check()` and `Run()`
- [ ] All skills pass tests with testcontainers (linuxserver/openssh-server + Docker-in-Docker)
- [ ] Skills are idempotent (running twice produces `Changed: false` the second time)
- [ ] Skills honor `IsDryRunMode` (log and return without executing)
- [ ] Skills use `ShellEscapeArg` for all user-supplied values
- [ ] Skills work with privilege escalation (`BecomeUser`)
- [ ] Skills registered in `NewDefaultRegistry()` and discoverable via `GetGlobalSkillRegistry()`

### Part 2 (SFTP)
- [ ] `SFTPClient` can upload and download files to/from testcontainers SSH server
- [ ] `SFTP()` method works on both persistent (`Connect()`) and one-time connections
- [ ] Error messages are actionable (connection refused, auth failed, permission denied)
- [ ] No regression in existing SSH command execution tests

### Part 3 (OCI Factory — Future)
- [ ] OCI layout passes `oci-image-tool validate` (if implemented)
- [ ] Produced images load successfully via `docker import` and/or `docker load`
- [ ] Round-trip test: `docker save` → `docker rmi` → `docker load` → `docker run` passes
- [ ] Zero external dependencies (stdlib only) if stdlib approach is chosen

## Open Questions

1. **Should `docker-import` skill include SFTP transfer, or should transfer be a separate skill?**
   - Option A: `docker-import` takes a local path, transfers via SFTP, then imports (one-step UX)
   - Option B: Separate `file-upload` skill, then `docker-import` takes a remote path (composable)
   - Recommendation: Option B (composable, follows Unix philosophy, transfer skill is reusable)

2. **Should the OCI factory (Part 3) use stdlib-only or `google/go-containerregistry`?**
   - Stdlib: zero deps, but not spec-compliant without significant work, no registry push
   - go-containerregistry: full compliance + registry push, but adds a large dependency tree
   - Recommendation: Start with stdlib (aligns with Ork's philosophy), add go-containerregistry only if registry push is needed

3. **Should Docker skills support Docker-in-Docker (DinD) or assume Docker is installed natively?**
   - DinD: works in more environments (CI, testcontainers) but has security implications
   - Native: simpler, but requires Docker installed on the target server
   - Recommendation: Native (Docker skills check for Docker presence; `docker-install` skill handles installation)

4. **Should the OCI factory support multi-platform images (image index)?**
   - The gocker factory only supports single-platform
   - Multi-platform requires building multiple images and an image index manifest
   - Recommendation: No, not in the initial implementation. Single-platform is sufficient for the SSH-automation use case.

5. **Should Ork add a `docker-compose` skill for multi-container orchestration?**
   - Out of scope for this proposal but worth considering as a follow-up
   - Recommendation: Track as a separate future proposal

## Research Sources

See `research/` subdirectory:
- `01-blog-post.md` — The original blog post by Cheikh Seck
- `02-gocker-source-code.md` — Full source code analysis of the gocker factory
- `03-oci-image-spec.md` — OCI Image Format Specification
- `04-oci-distribution-spec.md` — OCI Distribution Specification (registry push/pull)
- `05-go-stdlib-archive-tar.md` — Go `archive/tar` package
- `06-go-stdlib-crypto-sha256.md` — Go `crypto/sha256` package
- `07-docker-import-load.md` — Docker `import` vs `load` commands
- `08-existing-go-oci-libraries.md` — `google/go-containerregistry` and other Go OCI libraries
- `09-sftp-file-transfer.md` — `github.com/pkg/sftp` for SFTP file transfer
- `10-go-cross-compilation.md` — Go cross-compilation (`GOOS`/`GOARCH`)

## Comparison with Existing Proposals

This proposal fills a gap not covered by any existing proposal:
- **Vault** (implemented) — secrets management, unrelated
- **Privilege Escalation** (implemented) — used by Docker skills for `sudo docker`
- **Playbooks** (implemented) — could orchestrate Docker skills in sequences
- **Parallel Execution** (implemented) — enables fleet-wide Docker deployment
- **Roles** (rejected) — would have packaged Docker skills into reusable units
- **CLI Tool** (rejected) — the `ork` CLI currently only handles vault; `ork oci build` would extend it

## Related Resources

- [Project README](../../README.md)
- [Architecture Documentation](../../docs/architecture.html)
- [Skills Documentation](../../docs/skills.html)
- [Existing proposals](../README.md)
