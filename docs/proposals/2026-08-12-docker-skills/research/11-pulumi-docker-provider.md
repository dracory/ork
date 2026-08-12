# Source: Pulumi Docker Provider (`pulumi/docker`)

**Source URLs:**
- https://www.pulumi.com/registry/packages/docker/api-docs/container/
- https://www.pulumi.com/registry/packages/docker/api-docs/remoteimage/
- https://www.pulumi.com/registry/packages/docker/api-docs/image/
- https://www.pulumi.com/registry/packages/docker/api-docs/registryimage/
- https://www.pulumi.com/registry/packages/docker/api-docs/provider/
- https://www.pulumi.com/blog/fast-docker-image-builds-with-pulumi/
- https://github.com/pulumi/pulumi-docker/issues/342 (RemoteImage pullTriggers bug)
- https://github.com/pulumi/pulumi-docker/issues/872 (build context digest)
**Retrieved:** 2026-08-12 (via pulumi.com + web search)

## Summary

Pulumi's Docker provider (`pulumi/docker`, v5.1.0 as of Jun 2026) is a Terraform-derived provider that manages Docker resources declaratively from real programming languages (TypeScript, Python, Go, C#, Java). It wraps the `kreuzwerker/terraform-provider-docker` Go provider under the hood.

The provider uses the **Docker Engine API** (HTTP over Unix socket or TCP), NOT the `docker` CLI — same as Ansible. However, Pulumi's paradigm is fundamentally different from both Ansible and Ork: it's **declarative IaC with a state file** (like Terraform), not imperative task execution.

This research focuses on the design patterns Pulumi uses for **idempotency**, **image pulling**, **container lifecycle**, and **state management** — and what Ork can learn from them.

## Resource Inventory

### Container Resources
| Resource | Purpose | Ork Equivalent (proposed) |
|----------|---------|---------------------------|
| `docker.Container` | Manage container lifecycle (create/start/stop/destroy) | `docker-run`, `docker-stop`, `docker-rm` |
| `docker.RemoteImage` | Pull or build an image on the Docker host | `docker-pull` (pull mode) |
| `docker.Image` | Build + push an image (build context on Pulumi machine) | (out of scope — build-time) |
| `docker.RegistryImage` | Push an image to a registry | (future — `docker-push`) |

### Data Sources (Read-Only)
| Data Source | Purpose | Ork Equivalent (proposed) |
|-------------|---------|---------------------------|
| `docker.getRegistryImage` | Read image metadata from a registry (returns sha256Digest) | (future — `docker-registry-info`) |
| `docker.getContainer` | Read container info | `docker-ps` (partial) |
| `docker.getImage` | Read local image info | `docker-images` (partial) |
| `docker.getNetwork` | Read network info | (future) |
| `docker.getVolume` | Read volume info | (future) |

### Provider Configuration
| Parameter | Description |
|-----------|-------------|
| `host` | Docker daemon address (env: `DOCKER_HOST`) |
| `context` | Docker context name (env: `DOCKER_CONTEXT`) |
| `certPath` | Path to TLS config directory |
| `caMaterial` / `certMaterial` / `keyMaterial` | PEM-encoded TLS materials |
| `registryAuth` | List of registry authentication configs |
| `sshOpts` | Additional SSH option flags for `ssh://` protocol |
| `disableDockerDaemonCheck` | Skip daemon check (for resources that don't need it) |

## Key Design Pattern 1: Declarative State with `must_run`

Pulumi's `docker.Container` resource uses a **`must_run` boolean** to control the container's desired running state. This is Pulumi's equivalent of Ansible's `state` parameter, but simplified to a single boolean.

### `must_run` Behavior
| Value | Behavior |
|-------|----------|
| `true` (default) | Container will be kept running. If stopped, Pulumi detects drift and restarts it. |
| `false` | Pulumi leaves the container alone (doesn't start it). Used to trigger a restart of a stopped container. |

> "If `true`, then the Docker container will be kept running. If `false`, Terraform leaves the container alone. This attribute is also used to trigger a restart of a stopped container. If your container is stopped, Terraform will set `mustRun` to `false` and this will trigger a change." — Pulumi docs

### `start` Parameter
| Value | Behavior |
|-------|----------|
| `true` (default) | Container will be started after creation |
| `false` | Container is only created (not started) |

### `restart` Parameter (Restart Policy)
| Value | Description |
|-------|-------------|
| `no` (default) | No restart policy |
| `on-failure` | Restart on failure |
| `always` | Always restart |
| `unless-stopped` | Restart unless explicitly stopped |

This maps to Docker's `--restart` flag. Ork's `docker-run` skill already has this as `ArgRestart` with `DefaultRestart = "unless-stopped"`.

### Relevance to Ork

Pulumi's `must_run` + `start` pattern is the declarative equivalent of Ork's imperative `docker-run` (start) vs `docker-stop` (stop) skills. Ork doesn't need a `must_run` parameter because the skill choice itself encodes the desired state:
- `docker-run` = `must_run: true, start: true`
- `docker-stop` = `must_run: false` (stop the container)
- `docker-rm` = destroy the resource

However, Pulumi's `must_run` drift detection is noteworthy: if a container is stopped externally (e.g., `docker stop` run manually), Pulumi detects the drift on the next `pulumi up` and restarts it. Ork's `docker-run` skill does the same: `Check()` returns `true` if the container is stopped, and `Run()` calls `docker start`.

## Key Design Pattern 2: `pullTriggers` — Digest-Based Image Updates

This is Pulumi's most innovative pattern for image idempotency. The `docker.RemoteImage` resource has a `pullTriggers` parameter:

> "List of values which cause an image pull when changed. This is used to store the image digest from the registry when using the docker.getRegistryImage data source."

### How It Works

1. `docker.getRegistryImage` data source reads the registry and returns the current `sha256Digest` of the image
2. The digest is passed as a `pullTrigger` to `docker.RemoteImage`
3. When the registry digest changes (new image pushed), the `pullTrigger` value changes
4. Pulumi detects the change and pulls the new image

### Example (TypeScript)
```typescript
const ubuntu = docker.getRegistryImage({ name: "ubuntu:precise" });
const ubuntuRemoteImage = new docker.RemoteImage("ubuntu", {
    name: ubuntu.then(ubuntu => ubuntu.name),
    pullTriggers: [ubuntu.then(ubuntu => ubuntu.sha256Digest)],
});
```

### The Problem with `:latest`

Without `pullTriggers`, `docker.RemoteImage` pulls the image once and never checks for updates — even for `:latest`. This is because Pulumi stores the image ID in its state file, and if the image name hasn't changed, Pulumi sees no diff.

The `pullTriggers` pattern solves this by introducing an external change signal (the registry digest) that forces a re-pull when the actual image content changes, even if the tag (`:latest`) stays the same.

### Known Bug (GitHub Issue #342)
There's a known issue where `pullTriggers` changes don't always trigger an actual pull on the Docker host — the state is updated but the image isn't re-pulled. This is a reminder that digest-based pulling is complex and bug-prone.

### Relevance to Ork

Ork's `docker-pull` skill has an `always-pull` arg that's simpler:
- `always-pull=false` (default): only pull if image doesn't exist locally
- `always-pull=true`: always pull (for `:latest` updates)

Pulumi's `pullTriggers` pattern is more sophisticated — it only pulls when the registry digest actually changes, not on every run. Ork could adopt this in the future:
- **Phase 1:** `always-pull` arg (simple, works, may pull unnecessarily)
- **Phase 2 (future):** A `docker-registry-info` read-only skill that returns the registry digest, which users can compare to decide whether to pull. This would be the Ork equivalent of `docker.getRegistryImage` + `pullTriggers`.

However, Ork's imperative model means the user controls when to pull — there's no state file tracking. The `always-pull` arg is sufficient for Phase 1.

## Key Design Pattern 3: `keep_locally` — Image Retention on Destroy

Pulumi's `docker.RemoteImage` has a `keep_locally` parameter:

| Value | Behavior |
|-------|----------|
| `true` | Image won't be deleted when the resource is destroyed |
| `false` (default) | Image will be deleted from local Docker storage on destroy |

### Relevance to Ork

Ork doesn't have a "destroy" concept (skills are imperative, not declarative). But the concept maps to `docker-rmi`:
- Ork's `docker-rmi` skill explicitly removes an image (user chooses when)
- Pulumi's `keep_locally: false` removes the image automatically when the resource is removed from the Pulumi program

No change needed to Ork's design — the imperative model gives users explicit control over when images are removed.

## Key Design Pattern 4: `destroy_grace_seconds` — Graceful Destruction

Pulumi's `docker.Container` has a `destroy_grace_seconds` parameter:

> "If defined will attempt to stop the container before destroying. Container will be destroyed after `n` seconds or on successful stop."

### Relevance to Ork

This maps to Ork's `docker-stop` skill with a `timeout` arg:
- `docker-stop --timeout 30` = stop the container, wait up to 30 seconds, then SIGKILL
- Pulumi's `destroy_grace_seconds` = same concept, but applied at destroy time

Ork's `docker-rm` skill with `force=true` should also support a timeout: stop the container gracefully first, then remove it. This matches Pulumi's destroy behavior.

## Key Design Pattern 5: `wait` and `wait_timeout` — Health Check Waiting

Pulumi's `docker.Container` has `wait` and `wait_timeout` parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `wait` | `false` | If `true`, wait for the container to reach healthy state after creation. Requires a healthcheck. |
| `wait_timeout` | `60` | Timeout in seconds to wait for healthy state |
| `container_read_refresh_timeout_milliseconds` | — | Milliseconds to wait for container to reach 'running' status |

### Relevance to Ork

Ork's `docker-run` skill could optionally support a `wait` arg:
- `wait=false` (default): return immediately after `docker run -d`
- `wait=true`: after `docker run -d`, poll `docker inspect --format '{{.State.Health.Status}}'` until healthy or timeout

This is a Phase 2 enhancement. For Phase 1, `docker-run` returns immediately after starting the container (matching `docker run -d` behavior). Users can check health separately via `docker-exec` or `docker-ps`.

## Key Design Pattern 6: `docker.Image` — Build + Push (Local Context)

Pulumi's `docker.Image` resource is unique: it builds a Docker image from a **local** build context (on the machine running Pulumi) and pushes it to a registry. This is a **build-time** operation, not a deployment operation.

### Key Parameters
| Parameter | Description |
|-----------|-------------|
| `imageName` | Image name in `repository:tag` format |
| `build` | Build context (local path, Dockerfile, build args, etc.) |
| `registry` | Registry to push to |
| `skipPush` | Skip the push (build only) |
| `buildOnPreview` | Build during `pulumi preview` (default: false) |

### `repoDigest` Output
The `repoDigest` output is the unique manifest SHA of a pushed image (`repository@sha256:<hash>`). This is important because `imageName` (`repository:tag`) is NOT unique per build — the same tag can point to different images over time. `repoDigest` is the immutable identifier.

### Relevance to Ork

This is **out of scope** for Ork's Docker skills. Building images is a local operation, not a remote SSH operation. The OCI Image Factory proposal (Part 3) covers local image building. Ork's Docker skills are for managing containers on remote servers, not building images.

However, the `repoDigest` concept is relevant: Ork's `docker-pull` skill could optionally return the pulled image's digest in `Result.Details`, which users could use for tracking or drift detection.

## Key Design Pattern 7: `docker.Image` Build Context Digest (Issue #872)

Pulumi's `docker.Image` resource only rebuilds when the build context changes. This is determined by a build context digest computed by the provider. Issue #872 requests exposing this digest so users can tag images with it, ensuring image-dependent resources only update when the image actually changes.

### Relevance to Ork

Not directly relevant (Ork doesn't build images), but the concept of content-addressed digests for change detection is the same pattern as `pullTriggers`. The lesson: **use digests, not tags, for change detection**.

## Key Design Pattern 8: Provider Configuration — Remote Docker Daemons

Pulumi's Docker provider can connect to remote Docker daemons via:
- `host: tcp://192.0.2.23:2376` (TCP + TLS)
- `host: ssh://user@host` (SSH transport)
- `sshOpts: ["-o", "StrictHostKeyChecking=no"]` (additional SSH options)

### Relevance to Ork

Pulumi's `ssh://` transport is interesting — it connects to a remote Docker daemon via SSH. This is architecturally similar to Ork's approach (SSH to remote server), but Pulumi tunnels the Docker API over SSH, while Ork runs `docker` CLI commands over SSH.

Ork's approach is simpler and more transparent:
- No need to configure Docker daemon for TCP/SSH remote access
- No TLS certificate management
- Uses the standard `docker` CLI that's already on the server
- Works with any SSH-accessible server, no Docker daemon configuration needed

## Architectural Comparison: Pulumi vs Ansible vs Ork

| Aspect | Pulumi | Ansible | Ork |
|--------|--------|---------|-----|
| Paradigm | Declarative IaC (state file) | Declarative tasks (no state file) | Imperative skills (no state file) |
| Language | Real languages (TS, Python, Go, C#, Java) | YAML playbooks | Go (skills are Go structs) |
| Docker access | Docker Engine API (HTTP) | Docker Engine API (HTTP) | `docker` CLI over SSH |
| Target dependency | Docker daemon accessible | Python + Docker SDK on target | `docker` CLI on target |
| Idempotency | State file diff (drift detection) | `Check()` + state parameter | `Check()` + skill choice |
| Image pulling | `pullTriggers` (digest-based) | `pull` parameter (missing/always/never) | `always-pull` arg (simple) |
| Container lifecycle | `must_run` + `start` + destroy | `state` parameter | Separate skills (run/stop/rm) |
| Config drift | State file comparison | `comparisons` dictionary | `force` arg (Phase 2) |
| Build images | `docker.Image` (local context) | `docker_image` (build) | Out of scope (OCI factory) |
| Remote daemon | TCP/SSH tunnel to Docker API | TCP/SSH to Docker API | SSH + `docker` CLI |
| Complexity | High (state file, provider config) | Medium (Python SDK, API versioning) | Low (shell commands) |

## Lessons for Ork's Docker Skills

1. **No state file — that's fine**
   - Pulumi's state file enables drift detection (container stopped externally → Pulumi restarts it)
   - Ork doesn't have a state file, but `Check()` achieves the same result: if container is stopped, `docker-run`'s `Check()` returns `true` and `Run()` starts it
   - The difference: Pulumi detects drift automatically on `pulumi up`; Ork requires the user to run the skill again
   - This is acceptable for Ork's imperative model

2. **`pullTriggers` is sophisticated but overkill for Phase 1**
   - Pulumi's digest-based pulling only re-pulls when the registry digest changes
   - Ork's `always-pull` arg is simpler: always pull or only pull if missing
   - Phase 2 (future): could add a `docker-registry-info` read-only skill that returns the registry digest, enabling digest-based pull decisions

3. **`must_run` = skill choice in Ork**
   - Pulumi's `must_run: true/false` maps to Ork's `docker-run` vs `docker-stop` skill choice
   - No need for a `must_run` parameter in Ork

4. **`wait` for health checks is a good Phase 2 enhancement**
   - Pulumi waits for the container to be healthy before returning
   - Ork's `docker-run` could optionally do the same with a `wait` arg + `wait-timeout` arg
   - Phase 1: return immediately after `docker run -d` (simpler, matches Docker CLI behavior)

5. **`destroy_grace_seconds` = `docker-rm --force` with timeout**
   - Pulumi stops the container gracefully before destroying
   - Ork's `docker-rm` with `force=true` should also stop gracefully first (with configurable timeout)
   - This is already in the proposal: `docker-rm` accepts `force` and `docker-stop` accepts `timeout`

6. **`repoDigest` concept — return image digest in `Result.Details`**
   - Pulumi exposes `repoDigest` as an output for tracking
   - Ork's `docker-pull` and `docker-run` skills could return the image digest in `Result.Details`
   - Useful for logging, auditing, and future drift detection

7. **Build-time vs deployment-time separation**
   - Pulumi's `docker.Image` builds locally and pushes — this is build-time
   - Ork's Docker skills are deployment-time (manage containers on remote servers)
   - The OCI factory proposal covers build-time (local image factory)
   - This separation is correct and should be maintained

8. **SSH transport for Docker API vs CLI over SSH**
   - Pulumi can tunnel the Docker API over SSH (`ssh://` host)
   - Ork runs `docker` CLI commands over SSH
   - Ork's approach is simpler (no Docker daemon TCP/SSH configuration needed)
   - Trade-off: Ork can't do API-level operations (like `comparisons` drift detection) without parsing `docker inspect` JSON

## What to Borrow, What to Skip

### Borrow from Pulumi
- The `wait` + `wait_timeout` concept for health-check-aware container start → Phase 2 enhancement for `docker-run`
- The `destroy_grace_seconds` concept → `docker-rm --force` should stop gracefully first (with timeout)
- The `repoDigest` concept → return image digest in `Result.Details` of `docker-pull` and `docker-run`
- The `pullTriggers` concept (simplified) → `always-pull` arg on `docker-pull` (already in proposal)

### Skip from Pulumi
- The state file / declarative model (Ork is imperative)
- The `must_run` parameter (Ork uses skill choice)
- The Docker Engine API approach (Ork uses CLI over SSH)
- `docker.Image` build + push (out of scope; covered by OCI factory proposal)
- `docker.RegistryImage` push to registry (future skill, not Phase 1)
- Provider configuration (TCP/TLS/SSH tunnel to Docker daemon) — Ork uses SSH directly
- `pullTriggers` full digest-based pulling (too complex for Phase 1; `always-pull` is sufficient)
- Build context digest tracking (out of scope; Ork doesn't build images)

## Summary

Pulumi's Docker provider is the most sophisticated of the three (Pulumi > Ansible > Ork in terms of idempotency features), but also the most complex. It requires:
- A state file (for drift detection)
- The Docker Engine API (for config comparison)
- Provider configuration (for remote daemon access)
- Understanding of `pullTriggers`, `must_run`, `repoDigest`, etc.

Ork's approach is deliberately simpler:
- No state file (imperative, run-once skills)
- `docker` CLI over SSH (no API access needed)
- No provider configuration (SSH connection is already configured)
- Simple args (`force`, `pull`, `always-pull`) instead of complex parameters

The trade-off is that Ork can't do sophisticated drift detection (Ansible's `comparisons` or Pulumi's state-file diff) without parsing `docker inspect` JSON. But for Phase 1, the simple approach is sufficient. Drift detection can be added in Phase 2+ if there's demand.
