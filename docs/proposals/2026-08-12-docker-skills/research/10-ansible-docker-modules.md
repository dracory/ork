# Source: Ansible `community.docker` Collection

**Source URLs:**
- https://docs.ansible.com/ansible/latest/collections/community/docker/index.html
- https://docs.ansible.com/ansible/latest/collections/community/docker/docker_container_module.html
- https://docs.ansible.com/ansible/latest/collections/community/docker/docker_container_exec_module.html
- https://docs.ansible.com/ansible/latest/collections/community/docker/docker_image_pull_module.html
- https://ansible-collections.github.io/community.docker/branch/main/docker_image_module.html
- https://ansible-collections.github.io/community.docker/branch/main/docker_image_load_module.html
- https://github.com/ansible-collections/community.docker (README module list)
**Retrieved:** 2026-08-12 (via docs.ansible.com + web search)

## Summary

Ansible's `community.docker` collection is the most mature, widely-used idempotent Docker management interface. It contains **30+ modules** covering containers, images, networks, volumes, plugins, Compose, and Swarm. The collection uses the **Docker Engine API** directly (via Python `requests` or the Docker SDK), NOT the `docker` CLI. This is a fundamental architectural difference from Ork's SSH-based approach.

This research focuses on the design patterns Ansible uses for **idempotency**, **state management**, and **argument structure** — patterns that Ork's Docker skills can learn from even though the execution mechanism differs.

## Module Inventory (community.docker)

### Container Modules
| Module | Purpose | Ork Equivalent (proposed) |
|--------|---------|---------------------------|
| `docker_container` | Manage container lifecycle (create/start/stop/recreate/remove) | `docker-run`, `docker-stop`, `docker-rm` |
| `docker_container_exec` | Execute command in a running container | `docker-exec` |
| `docker_container_copy_into` | Copy a file into a container | (future) |
| `docker_container_info` | Retrieve container info (read-only) | `docker-ps` (partial) |
| `docker_host_info` | Retrieve Docker daemon info | (future) |

### Image Modules
| Module | Purpose | Ork Equivalent (proposed) |
|--------|---------|---------------------------|
| `docker_image` | Manage images (build/load/pull/push/tag/archive) | `docker-pull`, `docker-tag`, `docker-load` |
| `docker_image_build` | Build images with buildx | (out of scope) |
| `docker_image_export` | Archive images to tar | (future) |
| `docker_image_info` | Retrieve image info (read-only) | `docker-images` (partial) |
| `docker_image_load` | Load images from tar archive | `docker-load` |
| `docker_image_pull` | Pull images from registries | `docker-pull` |
| `docker_image_push` | Push images to registries | (future) |
| `docker_image_remove` | Remove images | `docker-rmi` |
| `docker_image_tag` | Tag images | `docker-tag` |
| `docker_login` | Log in/out of registries | (future — see Open Questions) |
| `docker_prune` | Prune unused Docker resources | (future) |

### Network/Volume/Plugin Modules
| Module | Purpose |
|--------|---------|
| `docker_network` / `docker_network_info` | Manage/inspect networks |
| `docker_volume` / `docker_volume_info` | Manage/inspect volumes |
| `docker_plugin` | Manage plugins |

### Compose Modules
| Module | Purpose |
|--------|---------|
| `docker_compose_v2` | Manage multi-container apps via Compose CLI |
| `docker_compose_v2_exec` | Exec in a Compose service container |
| `docker_compose_v2_pull` | Pull a Compose project |
| `docker_compose_v2_run` | Run command in a new Compose service container |

### Swarm Modules
| Module | Purpose |
|--------|---------|
| `docker_config` / `docker_secret` | Manage Swarm configs/secrets |
| `docker_node` | Manage Swarm nodes |
| `docker_swarm` / `docker_swarm_info` | Manage/inspect Swarm |
| `docker_swarm_service` | Manage Swarm services |

## Key Design Pattern 1: The `state` Parameter

Ansible's `docker_container` module uses a **`state` parameter** to define the desired state, rather than having separate modules for each action. This is the core of Ansible's declarative model.

### `docker_container` States

| State | Behavior |
|-------|----------|
| `absent` | Stop and remove the container (like `docker-rm`) |
| `present` | Assert container exists with matching config; create if absent, update/recreate if config differs |
| `started` | Assert container is present AND running; start if stopped |
| `healthy` | Assert container is present, started, AND healthy (waits for healthcheck) |
| `stopped` | Assert container is present but stopped; stop if running |

### Comparison to Ork's Approach

**Ansible (declarative, single module):**
```yaml
- name: Ensure myapp is running
  community.docker.docker_container:
    name: myapp
    image: myapp:v1
    state: started
    ports: ["8080:80"]
```

**Ork (imperative, separate skills):**
```go
node.Run(docker.NewDockerRun().SetName("myapp").SetImage("myapp:v1").SetPorts("8080:80"))
node.Run(docker.NewDockerStop().SetName("myapp"))
node.Run(docker.NewDockerRm().SetName("myapp"))
```

**Key insight:** Ork's existing skill architecture is **imperative** (one skill per action), not declarative (one module with state parameter). This matches all existing Ork skills (`caddy.Restart`, `mariadb.Install`, `fs.FileCreate`, etc.). The Docker skills should follow Ork's imperative pattern, NOT Ansible's declarative pattern.

However, Ork's `Check()` + `Run()` pattern achieves the same idempotency as Ansible's state model — `Check()` probes the current state, and `Run()` transitions to the desired state. The difference is that Ork encodes the desired state in the **skill choice** (DockerRun vs DockerStop vs DockerRm), while Ansible encodes it in the **state parameter**.

## Key Design Pattern 2: The `comparisons` Parameter (Config Drift Detection)

This is Ansible's most sophisticated idempotency feature — and the most complex. The `comparisons` dictionary controls how the module decides whether an existing container's config matches the requested config.

### Comparison Modes

| Mode | Behavior |
|------|----------|
| `strict` | Values must be exactly equal; any difference triggers recreate/update |
| `ignore` | Changes to this property are ignored (no recreate) |
| `allow_more_present` | For lists/sets/dicts: only trigger update if module option has a value NOT present in the container |

### Example
```yaml
- name: Start container, but don't recreate if env vars differ
  community.docker.docker_container:
    name: myapp
    image: myapp:v1
    env:
      DEBUG: "true"
    comparisons:
      env: ignore           # Don't recreate if env differs
      image: strict         # Always recreate if image differs
      "*": strict           # All other options: strict comparison
```

### The `*` Wildcard
The `*` wildcard sets a default comparison mode for all properties not explicitly specified. This is powerful but complex.

### Relevance to Ork

Ork's Docker skills should adopt a **simplified version** of this concept:
- **Phase 1 (MVP):** No config drift detection. If container is running, `Check()` returns `false` (no change). This is safe and simple.
- **Phase 2 (future):** Add an optional `force` arg to `docker-run` that, when `true`, compares the running container's config (image, ports, env) with the requested config and recreates if they differ. This would be a simplified `comparisons: strict` mode.
- **Phase 3 (future):** Full `comparisons`-style per-property drift detection. This is complex and may not be worth the effort for Ork's SSH-based approach (Ansible can do it because it reads the full container config via the Docker API; Ork would need to parse `docker inspect` JSON output).

## Key Design Pattern 3: The `pull` Parameter

Ansible's `docker_container` module has a `pull` parameter that controls when to pull the image before running:

| Value | Behavior |
|-------|----------|
| `missing` (default) | Only pull if image not present on daemon |
| `always` | Always pull latest version |
| `never` | Never pull; fail if image not present |

### Relevance to Ork

Ork's `docker-run` skill should have a similar `pull` arg:
- `missing` (default): don't pull if image exists locally
- `always`: always `docker pull` before `docker run`
- `never`: never pull; fail if image doesn't exist

This maps to Docker CLI's `--pull` flag: `docker run --pull missing|always|never ...`

## Key Design Pattern 4: `recreate` vs `restart` Parameters

Ansible's `docker_container` module has two separate boolean parameters:

| Parameter | Default | Behavior |
|-----------|---------|----------|
| `recreate` | `false` | Force re-creation of a matching container (stop + rm + run) |
| `restart` | `false` | Force a matching container to be stopped and restarted |

### Relevance to Ork

Ork's `docker-run` skill should have a `force` arg (simplified `recreate`):
- `force=false` (default): if container is running, leave it alone
- `force=true`: stop + remove + recreate the container

Ork's `docker-restart` skill is the equivalent of `restart: true` — it always restarts.

## Key Design Pattern 5: `docker_image` Module's `source` Parameter

Ansible's `docker_image` module uses a `source` parameter to specify how to obtain the image:

| Source | Behavior |
|--------|----------|
| `build` | Build from Dockerfile |
| `load` | Load from a .tar archive |
| `pull` | Pull from a registry |
| `local` | Use image already present locally (for tagging) |

### Relevance to Ork

Ork uses **separate skills** instead of a `source` parameter:
- `docker-pull` = `source: pull`
- `docker-load` = `source: load`
- `docker-tag` = `source: local` (with repository/tag)
- `docker-import` = (no direct Ansible equivalent — Ansible doesn't have a `docker import` module)

This matches Ork's imperative design. Each skill does one thing.

## Key Design Pattern 6: `docker_image_pull` Idempotency

The dedicated `docker_image_pull` module has a `pull` parameter:

| Value | Behavior |
|-------|----------|
| `always` (default) | Always pull the image |
| `not_present` | Only pull if image doesn't exist or platform doesn't match |

**Idempotency attribute:** `full` — when run twice with `pull: not_present`, the second run reports no change.

### Relevance to Ork

Ork's `docker-pull` skill should have an `always-pull` arg:
- `always-pull=false` (default): only pull if image doesn't exist locally (idempotent)
- `always-pull=true`: always pull, even if image exists (for `:latest` updates)

This matches the pattern in research file `05-docker-pull-tag.md`.

## Key Design Pattern 7: `docker_container_exec` — Non-Idempotent by Design

Ansible's `docker_container_exec` module:
- **Idempotency attribute:** NOT listed (implicitly non-idempotent)
- **Check mode:** NOT supported (no `check_mode` attribute)
- Always runs the command, every time

### Key Parameters
| Parameter | Description |
|-----------|-------------|
| `container` (required) | Container name or ID |
| `argv` | Command as a list of arguments (no quoting needed) |
| `command` | Command as a string (alternative to `argv`) |
| `chdir` | Working directory inside container |
| `env` | Environment variables |
| `user` | User to run as |
| `detach` | Run in background |
| `tty` | Allocate a TTY |
| `stdin` | Content to pass to the command's stdin |

### Relevance to Ork

Ork's `docker-exec` skill matches this pattern:
- Non-idempotent (always runs)
- `Check()` always returns `true`
- Takes container name, command, optional user/workdir/env
- Should support `Stdin` via `types.Command.Stdin` (matching Ansible's `stdin`)

## Key Design Pattern 8: Read-Only Info Modules

Ansible separates **management** modules from **info** modules:
- `docker_container` (manage) vs `docker_container_info` (read-only)
- `docker_image` (manage) vs `docker_image_info` (read-only)
- `docker_network` (manage) vs `docker_network_info` (read-only)
- `docker_volume` (manage) vs `docker_volume_info` (read-only)
- `docker_host_info` (read-only)

Info modules:
- Always return `changed: false`
- Return data in structured return values (`exists`, `container`, `network`, etc.)
- Support check mode trivially (just return data)

### Relevance to Ork

Ork's `docker-ps` and `docker-images` skills are the equivalent of Ansible's info modules:
- `Check()` always returns `false` (read-only)
- `Run()` executes the read command and returns output in `Result.Details`
- Future: could add `docker-container-info` and `docker-image-info` skills that return `docker inspect` JSON

## Architectural Difference: API vs CLI

### Ansible's Approach
Ansible's `community.docker` modules use the **Docker Engine API** directly (via HTTP over Unix socket or TCP). This means:
- Modules run **on the Docker host** (or connect to a remote Docker daemon via TCP)
- No `docker` CLI needed on the target
- Full access to container config via API (enables `comparisons` drift detection)
- Can connect to remote Docker daemons via `docker_host: tcp://192.0.2.23:2376`

### Ork's Approach
Ork skills execute `docker` CLI commands **over SSH** on remote servers. This means:
- The `docker` CLI must be installed on the remote server
- Ork connects via SSH (its core competency), not via Docker API
- Container config must be obtained via `docker inspect` (JSON output) and parsed
- Simpler, but less powerful than API access

### Trade-off
| Aspect | Ansible (API) | Ork (CLI over SSH) |
|--------|--------------|---------------------|
| Dependency on target | Docker SDK (Python) | `docker` CLI |
| Config access | Full API (structured) | `docker inspect` JSON (must parse) |
| Drift detection | Sophisticated (`comparisons`) | Limited (parse `docker inspect`) |
| Remote daemon | TCP + TLS | SSH (Ork's core) |
| Complexity | High (Python SDK, API versioning) | Low (shell commands) |
| Fit with Ork | N/A | Perfect (matches all existing skills) |

**Conclusion:** Ork should use the CLI-over-SSH approach. It's simpler, matches all existing Ork skills, and doesn't require Python or the Docker SDK on the target. The trade-off is that sophisticated drift detection (Ansible's `comparisons`) is harder — but that's a Phase 2+ enhancement, not a Phase 1 requirement.

## Lessons for Ork's Docker Skills

1. **Use `state`-like behavior via skill choice, not a state parameter**
   - Ork uses separate skills (DockerRun, DockerStop, DockerRm) instead of a state parameter
   - This matches Ork's existing imperative architecture
   - `Check()` + `Run()` achieves the same idempotency as Ansible's state model

2. **Adopt the `pull` parameter concept**
   - `docker-run` should have a `pull` arg: `missing` (default), `always`, `never`
   - Maps directly to Docker CLI's `--pull` flag
   - Matches Ansible's `pull` parameter semantics

3. **Adopt the `recreate` concept as `force`**
   - `docker-run` should have a `force` arg (default: `false`)
   - When `true`: stop + remove + recreate the container, even if it's running
   - Simplified version of Ansible's `recreate: true`

4. **Defer `comparisons`-style drift detection to Phase 2+**
   - Ansible's per-property comparison is powerful but complex
   - Requires parsing `docker inspect` JSON and comparing each property
   - Phase 1: if container is running, leave it alone (safe default)
   - Phase 2: add `force` arg for full recreate
   - Phase 3: consider per-property drift detection if there's demand

5. **Separate read-only info skills from management skills**
   - `docker-ps` and `docker-images` are read-only (always `Changed: false`)
   - Matches Ansible's `docker_container_info` / `docker_image_info` pattern
   - Future: `docker-container-info` (returns `docker inspect` JSON)

6. **`docker-exec` is non-idempotent — match Ansible**
   - `Check()` always returns `true`
   - No check mode (can't predict what a command will do)
   - Support `Stdin` for piping data into the container

7. **`docker-pull` idempotency: `not_present` vs `always`**
   - Default: only pull if image doesn't exist (idempotent)
   - Optional `always-pull` arg: always pull (for `:latest` updates)
   - Matches Ansible's `pull: not_present` vs `pull: always`

8. **Consider a `docker-login` skill (future)**
   - Ansible has `docker_login` for registry authentication
   - Involves credentials (sensitive data)
   - Track as a future skill, handle carefully (don't log credentials)

9. **Consider `docker-prune` skill (future)**
   - Ansible has `docker_prune` for cleaning up unused resources
   - Useful for disk space management on Docker hosts
   - Track as a future skill

10. **Don't implement `docker_image_build` (out of scope)**
    - Ansible has `docker_image_build` for building images with buildx
    - This is a build-time operation, not a deployment operation
    - The OCI Image Factory proposal (Part 3) covers local image building
    - Ork skills are for remote server management, not local builds

## Summary: What to Borrow, What to Skip

### Borrow from Ansible
- The `pull` parameter concept (`missing`/`always`/`never`) → `docker-run` `pull` arg
- The `recreate` parameter concept → `docker-run` `force` arg
- The `docker_image_pull` `pull` parameter (`always`/`not_present`) → `docker-pull` `always-pull` arg
- Separation of info modules from management modules → `docker-ps`/`docker-images` as read-only
- `docker_container_exec` non-idempotency → `docker-exec` `Check()` always returns `true`
- Support for `stdin` in exec → `docker-exec` supports `types.Command.Stdin`

### Skip from Ansible
- The `state` parameter (Ork uses separate skills, not declarative state)
- The `comparisons` dictionary (too complex for Phase 1; defer to Phase 2+)
- The Docker Engine API approach (Ork uses CLI over SSH)
- `docker_image_build` (out of scope; covered by OCI factory proposal)
- `docker_compose_v2` (out of scope; separate future proposal)
- Swarm modules (out of scope; different orchestration model)
- `docker_network`, `docker_volume`, `docker_plugin` (out of scope for Phase 1; track as future skills)
