# Source: Terraform Docker Provider (`kreuzwerker/terraform-provider-docker`)

**Source URLs:**
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs/resources/container
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs/resources/image
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs/data-sources/image
- https://github.com/kreuzwerker/terraform-provider-docker
- https://github.com/hashicorp/terraform-provider-docker/issues/174 (destroy_grace_seconds bug)
- https://github.com/kreuzwerker/terraform-provider-docker/issues/542 (must_run=false bug)
**Retrieved:** 2026-08-12 (via terraform registry + GitHub + web search)

## Summary

Terraform's Docker provider (`kreuzwerker/terraform-provider-docker`, v4.5.0+) is the **upstream** of Pulumi's Docker provider. Pulumi wraps this Terraform provider under the hood, so the resource schemas and parameters are nearly identical. This research confirms that lineage and documents Terraform-specific details (known bugs, the actions/ephemeral resources, the `triggers` parameter for rebuilds).

The provider uses the **Docker Engine API** (HTTP over Unix socket or TCP), same as Ansible and Pulumi. It's declarative IaC with a **state file** — the same paradigm as Pulumi.

## Resource Inventory

### Resources (stateful, managed lifecycle)
| Resource | Purpose | Ork Equivalent (proposed) |
|----------|---------|---------------------------|
| `docker_container` | Manage container lifecycle | `docker-run`, `docker-stop`, `docker-rm` |
| `docker_image` | Pull or build an image | `docker-pull` (pull mode) |
| `docker_registry_image` | Push an image to a registry | (future — `docker-push`) |
| `docker_tag` | Tag an image | `docker-tag` |
| `docker_compose` | Compose application (v2) | (future — `docker-compose`) |
| `docker_buildx_builder` | Manage buildx builders | (out of scope) |
| `docker_service` | Swarm service | (out of scope) |
| `docker_network` | Manage networks | (future) |
| `docker_volume` | Manage volumes | (future) |
| `docker_config` | Swarm config | (out of scope) |
| `docker_secret` | Swarm secret | (out of scope) |
| `docker_plugin` | Manage plugins | (out of scope) |

### Actions (ephemeral, one-off operations — Terraform v4+)
| Action | Purpose | Ork Equivalent (proposed) |
|--------|---------|---------------------------|
| `docker_exec` | Run a command in a container | `docker-exec` |
| `docker_image_import` | Import from a tarball | `docker-import` |
| `docker_image_load` | Load from a tar archive | `docker-load` |
| `docker_image_save` | Save to a tar archive | (future — `docker-save`) |
| `docker_container_export` | Export container filesystem | (future — `docker-export`) |
| `docker_system_prune` | Prune unused Docker objects | (future — `docker-prune`) |

### Data Sources (read-only)
| Data Source | Purpose | Ork Equivalent (proposed) |
|-------------|---------|---------------------------|
| `docker_image` | Read local image info | `docker-images` (partial) |
| `docker_container` | Read container info | `docker-ps` (partial) |
| `docker_registry_image` | Read registry image metadata (returns `sha256_digest`) | (future — `docker-registry-info`) |
| `docker_network` | Read network info | (future) |
| `docker_plugin` | Read plugin info | (out of scope) |
| `docker_tags` | List image tags in a registry | (future) |
| `docker_logs` | Read container logs | (future — `docker-logs`) |

## Key Design Pattern 1: `must_run` + `start` — Declarative Container Lifecycle

Same as Pulumi (inherited from this provider):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `must_run` | `true` | If `true`, container is kept running. If `false`, Terraform leaves it alone. Used to trigger restart of stopped containers. |
| `start` | `true` | If `true`, container is started after creation. If `false`, only created. |
| `restart` | `no` | Restart policy: `no`, `on-failure`, `always`, `unless-stopped` |
| `rm` | `false` | If `true`, container is automatically removed when it exits |

### Known Bug (Issue #542): `must_run = false` replacement fails every other run

> "The `docker_container` attempts to replace it on the second `terraform apply`... When the Terraform provider tries to destroy the `docker_container`, for some reason it waits for the container to be in the `not-running` state instead of the `removed` state... this will fail since it just removed the container."

**Lesson for Ork:** The `must_run = false` edge case is a known source of bugs in the Terraform provider. Ork's imperative approach (separate `docker-stop` and `docker-rm` skills) avoids this entirely — there's no declarative state to get out of sync.

## Key Design Pattern 2: `pull_triggers` — Digest-Based Image Updates

Same as Pulumi (inherited from this provider):

```hcl
data "docker_registry_image" "ubuntu" {
  name = "ubuntu:precise"
}

resource "docker_image" "ubuntu" {
  name          = data.docker_registry_image.ubuntu.name
  pull_triggers = [data.docker_registry_image.ubuntu.sha256_digest]
}
```

> "List of values which cause an image pull when changed. This is used to store the image digest from the registry when using the docker_registry_image data source."

### Relevance to Ork

Same conclusion as Pulumi: `pull_triggers` is sophisticated but overkill for Phase 1. Ork's `always-pull` arg is sufficient. The pattern could be adopted in Phase 2 with a `docker-registry-info` read-only skill.

## Key Design Pattern 3: `triggers` — Force Rebuild on Source Changes

The `docker_image` resource has a `triggers` argument (separate from `pull_triggers`):

> "You can use the `triggers` argument to specify when the image should be rebuild. This is for example helpful when you want to rebuild the docker image whenever the source code changes."

```hcl
resource "docker_image" "myapp" {
  name = "myapp:latest"
  build {
    context = "."
  }
  triggers = {
    "src_hash" = sha256(join("", [for f in fileset(".", "src/**/*") : filesha256(f)]))
  }
}
```

### Relevance to Ork

This is the Terraform equivalent of Pulumi's build context digest tracking. The `triggers` map is a general-purpose "rebuild when these values change" mechanism. Ork could adopt a similar concept for `docker-pull` or a future `docker-build` skill: accept a `trigger` arg (a hash or list of hashes) that, when changed, forces a re-pull or rebuild.

## Key Design Pattern 4: `keep_locally` — Image Retention on Destroy

Same as Pulumi (inherited from this provider):

| Value | Behavior |
|-------|----------|
| `true` | Image won't be deleted when the resource is destroyed |
| `false` (default) | Image will be deleted from local Docker storage on destroy |

### Relevance to Ork

Same as Pulumi — Ork is imperative, so `docker-rmi` is explicit. No `keep_locally` parameter needed.

## Key Design Pattern 5: `destroy_grace_seconds` — Graceful Destruction

Same as Pulumi (inherited from this provider):

> "If defined will attempt to stop the container before destroying. Container will be destroyed after `n` seconds or on successful stop."

### Known Bug (Issue #174): `destroy_grace_seconds` is not adhered to

> "terraform-provider-docker does not wait the full destroy_grace_seconds and kills the container before the destroy_grace_seconds period is over... Looking at the code, it looks like terraform-provider-docker waits a random period of time between 0 and destroy_grace_seconds instead of the full period."

The bug: the provider uses `rand.Int31n()` to generate a random timeout between 0 and `destroy_grace_seconds`, instead of waiting the full period. This was reported in 2019 and persisted through multiple versions.

**Lesson for Ork:** Ork's `docker-stop` skill should use the timeout directly (matching `docker stop -t <seconds>`), not a random value. This is a cautionary tale about how even mature providers get lifecycle details wrong.

## Key Design Pattern 6: `wait` + `wait_timeout` — Health Check Waiting

Same as Pulumi (inherited from this provider):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `wait` | `false` | If `true`, wait for container to reach healthy state. Requires a healthcheck. |
| `wait_timeout` | `60` | Timeout in seconds to wait for healthy state |
| `container_read_refresh_timeout_milliseconds` | — | Milliseconds to wait for container to reach 'running' status |

### Relevance to Ork

Same as Pulumi — Phase 2 enhancement for `docker-run`. Already added `ArgWait` and `ArgWaitTimeout` constants to the proposal.

## Key Design Pattern 7: Actions (Ephemeral Resources) — One-Off Operations

Terraform v4+ introduced "actions" (ephemeral resources) for one-off Docker operations that don't have a managed lifecycle:

- `docker_exec` — run a command in a container (non-idempotent)
- `docker_image_import` — import from a tarball
- `docker_image_load` — load from a tar archive
- `docker_image_save` — save to a tar archive
- `docker_container_export` — export container filesystem
- `docker_system_prune` — prune unused Docker objects

### Relevance to Ork

This is exactly how Ork models these operations — as separate skills, not managed resources:
- `docker-exec` = `docker_exec` action (non-idempotent, one-off)
- `docker-import` = `docker_image_import` action
- `docker-load` = `docker_image_load` action

Terraform's "actions" concept validates Ork's design: one-off Docker operations should be separate from managed lifecycle resources. Ork just makes them all separate skills (no managed resources at all).

## Key Design Pattern 8: Provider Configuration — Remote Docker Daemons

Same as Pulumi (inherited from this provider):

| Parameter | Description |
|-----------|-------------|
| `host` | Docker daemon address (env: `DOCKER_HOST`) |
| `context` | Docker context name (env: `DOCKER_CONTEXT`) |
| `cert_path` | Path to TLS config directory |
| `ca_material` / `cert_material` / `key_material` | PEM-encoded TLS materials |
| `registry_auth` | List of registry authentication configs |
| `ssh_opts` | Additional SSH option flags for `ssh://` protocol |
| `disable_docker_daemon_check` | Skip daemon check (for resources that don't need it) |

### Relevance to Ork

Same as Pulumi — Ork uses SSH directly + `docker` CLI, no Docker daemon TCP/SSH tunnel needed.

## Architectural Comparison: Terraform vs Pulumi vs Ansible vs Ork

| Aspect | Terraform | Pulumi | Ansible | Ork |
|--------|-----------|--------|---------|-----|
| Paradigm | Declarative IaC (state file) | Declarative IaC (state file) | Declarative tasks | Imperative skills |
| Language | HCL | Real languages (TS, Python, Go, etc.) | YAML playbooks | Go (skills are Go structs) |
| Docker access | Docker Engine API | Docker Engine API (wraps Terraform provider) | Docker Engine API | `docker` CLI over SSH |
| Origin | **Original** | Wraps Terraform provider | Independent | Independent |
| Idempotency | State file diff (drift detection) | State file diff (inherited) | `Check()` + state parameter | `Check()` + skill choice |
| Image pulling | `pull_triggers` (digest-based) | `pull_triggers` (inherited) | `pull` parameter | `always-pull` arg |
| Container lifecycle | `must_run` + `start` + destroy | `must_run` + `start` (inherited) | `state` parameter | Separate skills |
| One-off ops | Actions (ephemeral resources) | Not available | Modules (tasks) | Separate skills |
| Config drift | State file comparison | State file comparison (inherited) | `comparisons` dictionary | `force` arg (Phase 2) |
| Build images | `docker_image` with `build` block | `docker.Image` (wraps Terraform) | `docker_image` module | Out of scope (OCI factory) |
| Complexity | High (state file, HCL, provider) | High (state file, provider) | Medium | Low |
| Fit with Ork | N/A | N/A | N/A | Perfect |

## Lessons for Ork's Docker Skills

### Confirmed from Terraform (same as Pulumi, since Pulumi inherits from Terraform)

1. **`must_run` = skill choice in Ork** — confirmed (same as Pulumi)
2. **`pull_triggers` is overkill for Phase 1** — confirmed (same as Pulumi)
3. **`wait`/`wait_timeout` for health checks is a Phase 2 enhancement** — confirmed (same as Pulumi)
4. **`destroy_grace_seconds` = `docker-rm --force` with timeout** — confirmed (same as Pulumi)
5. **`keep_locally` is N/A (Ork is imperative)** — confirmed (same as Pulumi)

### New Lessons Specific to Terraform

6. **The `triggers` argument is a general-purpose force-rebuild mechanism**
   - Terraform's `triggers` map on `docker_image` lets users specify arbitrary values that, when changed, force a rebuild
   - This is more general than Pulumi's build context digest (which is automatic)
   - Ork could adopt this: a `trigger` arg on `docker-pull` that, when changed, forces a re-pull (even without `always-pull`)
   - Phase 2 enhancement

7. **Actions (ephemeral resources) validate Ork's separate-skills design**
   - Terraform v4+ introduced "actions" for one-off operations (`docker_exec`, `docker_image_import`, etc.)
   - This is exactly how Ork models them: as separate skills, not managed resources
   - Ork's design is validated by Terraform's evolution toward the same pattern

8. **Known bugs are cautionary tales**
   - `destroy_grace_seconds` bug (Issue #174): random timeout instead of full wait — Ork must use the timeout directly
   - `must_run = false` bug (Issue #542): state desync on replacement — Ork's imperative approach avoids this entirely
   - These bugs show that declarative state management for containers is error-prone; Ork's imperative approach is simpler and less bug-prone

9. **`docker_compose` and `docker_service` (Swarm) are out of scope for Phase 1**
   - Terraform has resources for both Compose and Swarm
   - Ork's Phase 1 focuses on basic container management
   - Compose and Swarm are future proposals (already noted in the proposal's open questions)

10. **Data sources for read-only info validate Ork's read-only skills**
    - Terraform has `docker_image`, `docker_container`, `docker_registry_image`, `docker_logs` data sources
    - Ork's `docker-ps` and `docker-images` skills serve the same purpose
    - A future `docker-registry-info` skill would match `docker_registry_image` data source
    - A future `docker-logs` skill would match `docker_logs` data source

## What to Borrow, What to Skip

### Borrow from Terraform (in addition to Pulumi)
- The `triggers` concept (general-purpose force-rebuild/repull) → Phase 2 enhancement for `docker-pull`
- The "actions" pattern validation → confirms Ork's separate-skills design for one-off operations
- Data source patterns → validates read-only skills (`docker-ps`, `docker-images`, future `docker-logs`, `docker-registry-info`)

### Skip from Terraform (same as Pulumi)
- The state file / declarative model (Ork is imperative)
- The `must_run` parameter (Ork uses skill choice)
- The Docker Engine API approach (Ork uses CLI over SSH)
- `docker_compose` and `docker_service` (out of scope for Phase 1)
- Provider configuration (TCP/TLS/SSH tunnel to Docker daemon)
- `pull_triggers` full digest-based pulling (too complex for Phase 1)

## Summary

Terraform's Docker provider is the **origin** of Pulumi's Docker provider — Pulumi wraps `kreuzwerker/terraform-provider-docker`. The resource schemas, parameters (`must_run`, `start`, `restart`, `wait`, `wait_timeout`, `destroy_grace_seconds`, `pull_triggers`, `keep_locally`), and design patterns are identical.

The key new insights from researching Terraform (beyond what Pulumi already taught us):

1. **The `triggers` argument** — a general-purpose force-rebuild mechanism that's more flexible than Pulumi's automatic build context digest. Ork could adopt this as a Phase 2 enhancement.

2. **Actions (ephemeral resources)** — Terraform v4+ introduced one-off operations (`docker_exec`, `docker_image_import`, etc.) as separate from managed resources. This validates Ork's design of having separate skills for one-off operations.

3. **Known bugs** — `destroy_grace_seconds` (random timeout bug) and `must_run = false` (state desync bug) are cautionary tales. Ork's imperative approach avoids the state desync bug entirely, and Ork's `docker-stop` should use the timeout directly (not random).

4. **Data sources** — Terraform's read-only data sources (`docker_image`, `docker_container`, `docker_registry_image`, `docker_logs`) validate Ork's read-only skills and suggest future skills (`docker-logs`, `docker-registry-info`).

The four-way comparison (Terraform → Pulumi → Ansible → Ork) shows a clear spectrum:
- **Terraform**: Most mature, most complex (state file + HCL + Docker API + known bugs)
- **Pulumi**: Same as Terraform but with real languages (wraps the same provider)
- **Ansible**: Simpler (no state file, but still uses Docker API + Python SDK)
- **Ork**: Simplest (no state file, no Docker API, just CLI over SSH + `Check()`/`Run()`)

Ork's imperative CLI-over-SSH approach is the simplest and most robust — it avoids the state desync bugs that plague both Terraform and Pulumi, and it doesn't require Python (like Ansible) or a Docker daemon TCP connection (like all three).
