# Source: Terraform Image Building — Does It Have an OCI Factory Equivalent?

**Source URLs:**
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs/resources/image
- https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs/resources/registry_image
- https://github.com/kreuzwerker/terraform-provider-docker
- https://github.com/cruxstack/terraform-provider-buildkit
- https://github.com/dskiff/tko
- https://ocibuild.hexdocs.pm/ocibuild.html
**Retrieved:** 2026-08-12 (via terraform registry + GitHub + web search)

## Summary

**Terraform itself has no "local OCI image factory" equivalent** — its primary Docker provider (`kreuzwerker/terraform-provider-docker`) requires a Docker daemon and Dockerfile for image building, same as Ansible and Pulumi (Pulumi wraps this provider).

However, the Terraform ecosystem has two interesting additions:
1. **`terraform-provider-buildkit`** — a third-party provider that builds images using BuildKit's gRPC API directly, without driving the Docker CLI or daemon as a build mechanism (but still needs a `buildkitd`).
2. **`tko`** — not a Terraform provider but a standalone daemonless OCI image builder (base image + build artifacts → registry push).

This research documents Terraform's image-building approach and surveys these additional tools.

## Part 1: Terraform's `docker_image` Resource — `build` Block

The `kreuzwerker/terraform-provider-docker` provider's `docker_image` resource can build from a Dockerfile:

```hcl
resource "docker_image" "myapp" {
  name = "myapp:latest"
  build {
    context = "."
    dockerfile = "Dockerfile"
    build_arg = {
      version = "1.2.3"
    }
    label = {
      author = "me"
    }
    cache_from = ["nginx:latest", "alpine:3.8"]
  }
  triggers = {
    "src_hash" = sha256(join("", [for f in fileset(".", "src/**/*") : filesha256(f)]))
  }
}
```

**Key limitation:** "Building images is done using Docker daemon's API. It is not possible to use BuildKit / buildx this way."

This resource:
- Requires a Dockerfile
- Requires a running Docker daemon (uses the Docker Engine API)
- Cannot build images from scratch without a Dockerfile
- Cannot construct OCI images programmatically

### The `triggers` Argument

> "You can use the `triggers` argument to specify when the image should be rebuild. This is for example helpful when you want to rebuild the docker image whenever the source code changes."

The `triggers` map is a general-purpose force-rebuild mechanism. When any value in the map changes, Terraform treats the image as needing a rebuild. This is more flexible than Pulumi's automatic build context digest (which is computed by the provider).

### The `pull_triggers` Argument

> "List of values which cause an image pull when changed. This is used to store the image digest from the registry when using the docker_registry_image data source."

Separate from `triggers` (which controls rebuilds), `pull_triggers` controls re-pulls. Used with the `docker_registry_image` data source:

```hcl
data "docker_registry_image" "ubuntu" {
  name = "ubuntu:precise"
}

resource "docker_image" "ubuntu" {
  name          = data.docker_registry_image.ubuntu.name
  pull_triggers = [data.docker_registry_image.ubuntu.sha256_digest]
}
```

### The `keep_locally` Argument

| Value | Behavior |
|-------|----------|
| `true` | Image won't be deleted when the resource is destroyed |
| `false` (default) | Image will be deleted from local Docker storage on destroy |

## Part 2: Terraform's `docker_registry_image` Resource — Push to Registry

The `docker_registry_image` resource builds and pushes an image to a registry:

```hcl
resource "docker_registry_image" "myapp" {
  name = "myregistry.com/myapp:latest"
  build {
    context = "."
    dockerfile = "Dockerfile"
  }
  keep_remotely = false
}
```

| Parameter | Description |
|-----------|-------------|
| `name` | Image name including registry and tag |
| `build` | Build configuration (context, dockerfile, build args, labels) |
| `keep_remotely` | If `true`, image won't be deleted from registry on destroy |
| `insecure_skip_verify` | Skip TLS verification |
| `auth_config` | Registry authentication |

**Key limitation:** Same as `docker_image` — requires a Dockerfile and Docker daemon.

## Part 3: `terraform-provider-buildkit` — BuildKit gRPC (No Docker CLI)

**Repository:** https://github.com/cruxstack/terraform-provider-buildkit

A third-party Terraform/OpenTofu provider that builds container images using BuildKit **directly over its gRPC API** — without driving the Docker CLI, the Docker daemon as a build mechanism, or `local-exec`.

### Key Features
- **No Docker CLI, no Docker daemon as build mechanism** — speaks to BuildKit gRPC directly
- **Can auto-discover BuildKit** embedded in OrbStack / Docker Desktop / Colima
- **Can supervise an embedded rootless `buildkitd`** on Linux
- **Multi-platform images** — build and push in a single build
- **Build secrets / SSH forwarding**
- **SBOM + provenance attestations**
- **Cache import/export** (registry/local/gha)
- **Extract artifacts to host** — `buildkit_artifact` resource extracts a file/dir from a built stage

### Resources
| Resource | Purpose |
|----------|---------|
| `buildkit_image` | Build and push multi-platform images |
| `buildkit_artifact` | Build and extract a file/dir to the host filesystem |

### Comparison (from the provider's README)

| Capability | `terraform-provider-buildkit` | `kreuzwerker/docker` |
|------------|-------------------------------|----------------------|
| Build + push images via BuildKit | yes (direct gRPC) | via daemon |
| Multi-platform images | yes | limited |
| Build secrets / SSH forwarding | yes | limited |
| SBOM / provenance attestations | yes | no |
| Cache import/export | yes | limited |
| Extract file/dir artifact to host | yes | no |
| Endpoint auto-discovery | yes | n/a |
| Embedded rootless buildkitd (Linux) | yes | no |

### Relevance to OCI Factory

`terraform-provider-buildkit` is closer to daemonless than `kreuzwerker/docker`, but it still:
- **Requires a Dockerfile** — no programmatic image construction
- **Requires a `buildkitd`** — either embedded, auto-discovered, or remote
- **Cannot build from scratch** without a Dockerfile

It's a **Dockerfile-based builder that bypasses the Docker daemon**, not a programmatic OCI image factory. The gocker pattern (constructing images from tar + JSON) is fundamentally different — it doesn't need BuildKit at all.

## Part 4: `tko` — Daemonless, Privilege-Free OCI Image Builder

**Repository:** https://github.com/dskiff/tko

tko builds OCI images without elevated privileges — no `privileged`, no DinD, no root/sudo, no capabilities.

### Key Features
- **No Docker daemon, no Dockerfile** — takes build artifacts + base image
- **Truly rootless** — no sudo/daemon/chroot/caps needed
- **Low footprint** — ~15MiB single static binary, no runtime deps
- **Reproducible** — same base + build artifacts → same image digest
- **Registry push** — pushes directly to registry

### How It Works

```bash
# Compile and store output in ./build-artifacts
your-build-tool build --output ./build-artifacts

# Take compiled output and add it to a base image, push to remote repo
tko build --target-repo="destination/repo" ./build-artifacts
```

> "tko takes `(base image) + (your build artifacts) + (metadata)` and pushes it to your repo."

### Configuration (`.tko.yml`)

```yaml
build:
  base-ref: ubuntu:jammy@sha256:6d7b5d3317a71adb5e175640150e44b8b9a9401a7dd394f44840626aff9fa94d
  author: my name
  default-annotations:
    org.opencontainers.image.source: github.com/my-org/my-project
```

### What tko Is NOT

> "tko is NOT a full replacement for docker build (or buildah, kaniko, etc). Rather than running a Dockerfile in an isolated container, you need to build your artifacts in your native build environment and tko combines your artifacts with a base image."

### Relevance to OCI Factory

tko is very close to the gocker pattern in spirit:
- Takes pre-built artifacts + base image → produces OCI image
- No Docker daemon, no Dockerfile
- Single binary, no runtime deps
- Pushes to registry

The difference:
- **tko** uses a base image (pulls it, adds a layer) — it's an "append" tool
- **gocker** builds from scratch (creates empty image, adds layers) — it's a "create" tool
- **tko** is production-grade (handles registry auth, multi-platform push, annotations)
- **gocker** is educational (~150 lines, stdlib only)

tko validates the "base image + artifacts" pattern as a production approach. The OCI factory could support both modes:
- **From scratch** (gocker pattern): empty image + layers
- **From base image** (tko pattern): pull base + append layers

## Part 5: `ocibuild` — Erlang/Elixir OCI Image Builder

**Repository:** https://github.com/ocibuild/ocibuild (Erlang/Elixir)

> "Build and publish OCI container images from the BEAM. This library provides the public API for building OCI-compliant container images without requiring Docker or any container runtime."

### Key Features
- **No Docker, no container runtime** — builds OCI images from Erlang/Elixir
- **Programmatic** — `ocibuild:from/1`, `ocibuild:copy/2`, `ocibuild:entrypoint/2`, `ocibuild:env/2`
- **From scratch** — `ocibuild:scratch()` creates an empty image
- **From base image** — `ocibuild:from("docker.io/library/alpine:3.19")`
- **Multi-platform push** — `ocibuild:push_multi/4`
- **Save as tarball** — `ocibuild:save/2`

### Example

```erlang
%% Build from a base image
{ok, Image0} = ocibuild:from(~"docker.io/library/alpine:3.19"),

%% Add your application
Image1 = ocibuild:copy(Image0, [{~"myapp", AppBinary}], ~"/app"),

%% Configure the container
Image2 = ocibuild:entrypoint(Image1, [~"/app/myapp", ~"start"]),
Image3 = ocibuild:env(Image2, #{~"MIX_ENV" => ~"prod"}),

%% Push to registry
ocibuild:push(Image3, ~"ghcr.io", ~"myorg/myapp:v1", Auth).
```

### Relevance to OCI Factory

`ocibuild` is the Erlang/Elixir equivalent of the gocker pattern — a programmatic, daemonless, runtime-free OCI image builder. It validates the approach:
- Build images as data structures (layers + config + manifest)
- No Docker daemon needed
- Both "from scratch" and "from base image" modes
- Push directly to registry

The OCI factory proposal's Go implementation would be the Go equivalent of what `ocibuild` is for Erlang/Elixir.

## Part 6: Updated Comparison Matrix — All Image Building Approaches

| Tool | Daemon Required | Dockerfile Required | From Scratch | From Base | Language | Go Library | Production-Grade |
|------|----------------|---------------------|--------------|-----------|----------|------------|-----------------|
| **Docker build** | Yes | Yes | No (FROM scratch) | Yes | Any | No | Yes |
| **Ansible docker_image** | Yes | Yes | No | Yes | Any | No | Yes |
| **Ansible docker_image_build** | Yes (buildx) | Yes | No | Yes | Any | No | Yes |
| **Terraform docker_image** | Yes | Yes | No | Yes | Any | No | Yes |
| **Terraform docker_registry_image** | Yes | Yes | No | Yes | Any | No | Yes |
| **terraform-provider-buildkit** | No (needs buildkitd) | Yes | No | Yes | Any | No | Yes |
| **Pulumi docker.Image** | Yes | Yes | No | Yes | Any | No | Yes |
| **Pulumi docker_build.Image** | Yes (buildx) | Yes | No | Yes | Any | No | Yes |
| **buildah** | No | Optional | Yes | Yes | Any | No (CLI) | Yes |
| **ko** | No | No | Yes (Go binaries) | Yes | Go only | No (CLI) | Yes |
| **jib** | No | No | Yes (Java apps) | Yes | Java only | No (Maven/Gradle) | Yes |
| **apko** | No | No (YAML manifest) | Yes (APK packages) | No | Any | No (CLI) | Yes |
| **go-containerregistry** | No | No | Yes (programmatic) | Yes | Go | **Yes** | Yes |
| **gocker (proposed)** | No | No | Yes (stdlib only) | No | Go | **Yes** | No (demo) |
| **tko** | No | No | No | Yes (base + artifacts) | Any | No (CLI) | Yes |
| **ocibuild** | No | No | Yes | Yes | Erlang/Elixir | No (Erlang lib) | Yes |
| **img** | No | Yes | No | Yes | Any | No (CLI) | No (unmaintained) |

## Part 7: What This Means for the OCI Factory Proposal

### Terraform Has No OCI Factory Equivalent

Terraform's primary Docker provider (`kreuzwerker/terraform-provider-docker`) requires a Docker daemon and Dockerfile — same as Ansible and Pulumi. The `terraform-provider-buildkit` provider is closer to daemonless but still requires a Dockerfile and a `buildkitd`.

### The True Peers (Updated)

The OCI factory's true peers are the daemonless, Dockerfile-free image builders:

| Tool | Approach | Closest to gocker? |
|------|----------|-------------------|
| `go-containerregistry` | Go library, programmatic | **Yes** (Go, from scratch, library) |
| `ocibuild` | Erlang library, programmatic | Yes (from scratch, library, different language) |
| `tko` | CLI, base image + artifacts | Partial (from base, not from scratch) |
| `buildah` | CLI, scriptable | Partial (from scratch, but CLI not library) |
| `ko` | CLI, Go binaries only | Partial (Go-specific, not general-purpose) |
| `jib` | Maven/Gradle, Java only | No (Java-specific) |
| `apko` | CLI, APK packages | No (APK-specific) |

### Lessons from Terraform (New, Beyond Ansible/Pulumi)

1. **The `triggers` argument is a general-purpose force-rebuild mechanism**
   - Terraform's `triggers` map on `docker_image` lets users specify arbitrary values that force a rebuild when changed
   - More flexible than Pulumi's automatic build context digest
   - The OCI factory should support a similar concept: a `trigger` or `input-hash` arg that, when changed, forces a rebuild

2. **`terraform-provider-buildkit` shows the "bypass Docker daemon" approach**
   - It speaks to BuildKit gRPC directly, bypassing the Docker CLI and daemon
   - But it still needs a Dockerfile and a `buildkitd`
   - The OCI factory goes further: no Dockerfile, no buildkitd, no daemon at all

3. **`tko` validates the "base image + artifacts" pattern**
   - Production-grade tool that takes pre-built artifacts + base image → OCI image
   - The OCI factory could support both "from scratch" (gocker) and "from base" (tko) modes
   - This is a Phase 2 enhancement: "from base" mode requires pulling and parsing a base image

4. **`ocibuild` validates the programmatic library approach**
   - Erlang/Elixir library that builds OCI images as data structures
   - Same conceptual model as the gocker pattern (layers + config + manifest)
   - Validates that the "programmatic image construction" approach works in production

### Updated Recommendation for the OCI Factory Proposal

Based on this research (combined with `research/11-ansible-pulumi-image-building.md`):

1. **Reference `terraform-provider-buildkit`, `tko`, and `ocibuild` as additional peers** — alongside `buildah`, `ko`, `jib`, `apko`, and `go-containerregistry`.

2. **Consider supporting both "from scratch" and "from base image" modes** — the gocker pattern is "from scratch" only; `tko` and `ocibuild` show that "from base" is also valuable. Phase 1: from scratch only. Phase 2: from base image (pull base, append layers).

3. **Adopt Terraform's `triggers` concept** — a general-purpose `trigger` arg that forces a rebuild when changed. More flexible than Pulumi's automatic digest.

4. **Use `google/go-containerregistry` as the primary dependency** — it's the production-grade Go library, the closest peer to the gocker pattern, and supports both "from scratch" and "from base" modes.

5. **Be non-idempotent** — building is inherently non-idempotent (like Ansible's `docker_image_build` and Terraform's `docker_image` with `triggers`). Use input hash comparison for "should I rebuild?" decisions.

## Summary Table: Terraform Image Building vs OCI Factory

| Aspect | Terraform (kreuzwerker/docker) | Terraform (buildkit provider) | OCI Factory (proposed) |
|--------|-------------------------------|-------------------------------|------------------------|
| Build from Dockerfile | Yes (`docker_image` build block) | Yes (BuildKit gRPC) | No (programmatic) |
| Build from scratch | No | No | **Yes** |
| Build from base image | Yes (FROM in Dockerfile) | Yes (FROM in Dockerfile) | Phase 2 (pull + append) |
| Docker daemon required | Yes | No (needs buildkitd) | **No** |
| Dockerfile required | Yes | Yes | **No** |
| Daemonless | No | Partial (needs buildkitd) | **Yes** |
| Go library | No (HCL + provider) | No (HCL + provider) | **Yes** |
| Programmatic image construction | No | No | **Yes** |
| `triggers` for rebuild | Yes | No | **Recommended** |
| `pull_triggers` for re-pull | Yes | N/A | N/A (imperative) |
| Multi-platform | Limited | Yes (buildx) | Possible (set config) |
| Registry push | Yes (`docker_registry_image`) | Yes | Possible (go-containerregistry) |
| SBOM / provenance | No | Yes | Possible (Phase 2) |
| Cache import/export | Limited | Yes | Possible (manual) |
| Idempotency | `triggers`-based | Non-idempotent | Non-idempotent (or input-hash) |

**Bottom line:** Terraform has no OCI factory equivalent. The `terraform-provider-buildkit` provider is closer to daemonless but still requires a Dockerfile and `buildkitd`. The true peers remain `buildah`, `ko`, `jib`, `apko`, `go-containerregistry`, and now also `tko` and `ocibuild`. The OCI factory fills a gap that Terraform, Ansible, and Pulumi all fail to address: programmatic, daemonless, Dockerfile-free OCI image construction.
