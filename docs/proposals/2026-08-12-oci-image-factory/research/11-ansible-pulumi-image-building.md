# Source: Ansible & Pulumi Image Building — Do They Have an OCI Factory Equivalent?

**Source URLs:**
- https://ansible-collections.github.io/community.docker/branch/main/docker_image_module.html
- https://ansible-collections.github.io/community.docker/branch/main/docker_image_build_module.html
- https://www.pulumi.com/registry/packages/docker-build/api-docs/image/
- https://www.pulumi.com/blog/docker-build/
- https://www.pulumi.com/blog/fast-docker-image-builds-with-pulumi/
- https://github.com/containers/buildah
- https://github.com/google/go-containerregistry/blob/main/pkg/v1/mutate/README.md
- https://ahmet.im/blog/building-container-images-in-go/
- https://github.com/GoogleContainerTools/jib
- https://latchkey.dev/learn/tool-comparisons/ko-vs-jib
- https://debugg.ai/resources/no-more-dockerfiles-reproducible-secure-container-builds-nix-buildpacks-apko-2025
- https://www.kenmuse.com/blog/building-oci-images-without-using-docker/
- https://github.com/genuinetools/img
**Retrieved:** 2026-08-12 (via web search + pulumi.com + ansible docs)

## Summary

**Neither Ansible nor Pulumi have a "local OCI image factory" equivalent.** Both rely on the Docker daemon (or buildx/BuildKit) to build images, and both require a Dockerfile. Neither can construct OCI images programmatically from scratch without a Docker daemon.

However, the broader ecosystem **does** have daemonless image-building tools that are the true peers to the proposed OCI factory: `buildah`, `ko`, `jib`, `apko`, and `google/go-containerregistry` (`crane`). This research documents what Ansible/Pulumi do for image building, and then surveys the daemonless alternatives that are more relevant comparisons.

## Part 1: Ansible's Image Building Approach

### `community.docker.docker_image` — `source: build`

Ansible's `docker_image` module can build images from a Dockerfile:

```yaml
- name: Build an image
  community.docker.docker_image:
    name: myapp:v1
    build:
      path: /path/to/build/dir
      args:
        log_volume: /var/log/myapp
    source: build
```

**Key limitation:** "Building images is done using Docker daemon's API. It is not possible to use BuildKit / buildx this way."

This module:
- Requires a Dockerfile
- Requires a running Docker daemon (uses the Docker Engine API)
- Cannot build images from scratch without a Dockerfile
- Cannot construct OCI images programmatically

### `community.docker.docker_image_build` — buildx/BuildKit

A newer module that uses buildx:

```yaml
- name: Build multi-platform image
  community.docker.docker_image_build:
    name: multi-platform-image
    tag: "1.5.2"
    path: /home/user/images/multi-platform
    platform:
      - linux/amd64
      - linux/arm64/v8
```

**Key limitation:** "Note that the module is not idempotent in the sense of classical Ansible modules. The only idempotence check is whether the built image already exists."

This module:
- Requires a Dockerfile
- Requires buildx (Docker daemon with BuildKit)
- Cannot build images from scratch without a Dockerfile
- Cannot construct OCI images programmatically

### Ansible's `docker commit` Pattern (Legacy)

An older Ansible pattern (pre-`docker_image`) builds images by:
1. Starting a container from a base image
2. Running commands inside it (via `docker exec` or Ansible's docker connection plugin)
3. Committing the container as a new image via `docker commit`

```yaml
- name: Make a new image from the Docker container
  shell: docker commit --author "{{ docker_author }}" {{ inventory_hostname }} {{ docker_repo }}/{{ inventory_hostname }}:{{ docker_tag }}
```

This is the "container commit" pattern — not programmatic image construction. It still requires a Docker daemon to run the container.

### Conclusion for Ansible

**Ansible has no equivalent to the OCI factory.** All Ansible image-building approaches require:
- A Docker daemon (running and accessible)
- A Dockerfile (or a running container to commit)

Ansible cannot build OCI images from scratch using only its standard library. It delegates all image building to the Docker daemon.

## Part 2: Pulumi's Image Building Approach

### `docker.Image` (pulumi/docker v4+)

Pulumi's `docker.Image` resource builds from a local Dockerfile and pushes to a registry:

```typescript
const image = new docker.Image("my-image", {
    imageName: "docker.io/username/demo-image:v1",
    build: {
        context: "./app",
        dockerfile: "./app/Dockerfile",
    },
    registry: {
        server: "docker.io",
        username: "username",
        password: pulumi.secret("password"),
    },
});
```

Since v4, it:
- Rebuilds only when the build context changes (content-addressed digest)
- Uses BuildKit by default
- Supports registry caching (`cacheFrom` / `cacheTo`)

**Key limitation:** Still requires a Dockerfile and a Docker daemon (or buildx).

### `docker_build.Image` (pulumi/docker-build — newer, dedicated provider)

The newer `docker-build` provider exposes the full buildx surface:

```typescript
const myImage = new docker_build.Image("my-image", {
    tags: ["docker.io/pulumibot/demo-image:latest"],
    context: { location: "./app" },
    dockerfile: { location: "./app/Dockerfile" },
    platforms: ["linux/amd64", "linux/arm64"],
    cacheFrom: [{ registry: { ref: "docker.io/pulumibot/demo-image:cache" } }],
    cacheTo: [{ registry: { ref: "docker.io/pulumibot/demo-image:cache", mode: "max" } }],
    push: true,
});
```

Features:
- Multi-platform builds
- Multiple cache backends (registry, local, inline)
- Build secrets
- Docker Build Cloud support
- SSH agent forwarding

**Key limitation:** Still requires a Dockerfile and buildx (Docker daemon with BuildKit).

### `docker.RemoteImage` with `build` (pulumi/docker)

`RemoteImage` can also build:

```typescript
const zoo = new docker.RemoteImage("zoo", {
    name: "zoo",
    build: {
        context: ".",
        tags: ["zoo:develop"],
        buildArgs: { foo: "zoo" },
        label: { author: "zoo" },
    },
});
```

**Key limitation:** Same — requires a Dockerfile and Docker daemon.

### Conclusion for Pulumi

**Pulumi has no equivalent to the OCI factory.** All Pulumi image-building approaches require:
- A Dockerfile
- A Docker daemon or buildx (with BuildKit)

Pulumi's innovation is in **caching** (registry cache, build context digest tracking) and **multi-platform** support, not in daemonless image construction. Pulumi delegates all image building to Docker buildx/BuildKit.

## Part 3: The Real Peers — Daemonless Image Builders

Since neither Ansible nor Pulumi can build OCI images without Docker, the true peers to the proposed OCI factory are these daemonless tools:

### 1. `buildah` (Red Hat / containers)

**Repository:** https://github.com/containers/buildah

Buildah is a daemonless OCI image builder that can build images with or without a Dockerfile.

**Key features:**
- **No Docker daemon required** — uses fork-exec model, no daemon
- **Can build from scratch** — `buildah from scratch` creates an empty working container
- **Scriptable** — `buildah mount` exposes the rootfs for direct manipulation
- **Supports Dockerfiles** — `buildah build` uses a Dockerfile (like `docker build`)
- **OCI and Docker formats** — can output either format
- **No root privileges** — can build unprivileged (with user namespaces)

**Example (no Dockerfile):**
```bash
ctr=$(buildah from alpine:3)
mnt=$(buildah mount "$ctr")
mkdir "$mnt/rootleveldir"
cp ./localfile "$mnt/rootleveldir"
buildah run -v "$PWD:$PWD" -- sh $PWD/init.sh
buildah config --entrypoint "/rootlevel/entrypoint.sh" "$ctr"
buildah commit "$ctr" myimage:v1
```

**Relevance to OCI factory:** Buildah is the most mature daemonless builder. It's the "scriptable Docker" — you can build images step-by-step without a Dockerfile. However, it's a CLI tool (not a Go library), requires Linux namespaces, and is heavier than the gocker pattern.

### 2. `ko` (Google / go-containerregistry)

**Repository:** https://github.com/ko-build/ko

ko builds Go applications directly into OCI images — no Dockerfile, no Docker daemon.

**Key features:**
- **No Docker daemon, no Dockerfile** — compiles Go binary and packages it into an image
- **Go-native** — integrates with Go modules, `go build`
- **Minimal images** — produces distroless-like images with just the Go binary
- **Fast** — skips Docker build entirely
- **Registry push** — pushes directly to registry
- **k8s integration** — can rewrite Kubernetes manifests to use the built image

**Example:**
```bash
ko build ./cmd/myapp
# Output: ko.local/myapp-<hash>:latest
```

**Relevance to OCI factory:** ko is the closest production-grade equivalent to the gocker pattern for Go applications. It uses `google/go-containerregistry` under the hood. However, it's Go-specific (builds Go binaries), not a general-purpose image factory.

### 3. `jib` (Google)

**Repository:** https://github.com/GoogleContainerTools/jib

Jib builds Java/JVM applications into OCI images — no Dockerfile, no Docker daemon.

**Key features:**
- **No Docker daemon, no Dockerfile** — builds from Maven/Gradle
- **Java-optimized layering** — separates dependencies from classes for incremental builds
- **Daemonless** — constructs images programmatically
- **Reproducible** — same input always produces same image
- **Maven and Gradle plugins** — integrates into Java build tools

**Relevance to OCI factory:** Jib is the Java equivalent of ko. It's language-specific (Java/JVM), not a general-purpose image factory.

### 4. `apko` (Chainguard)

**Repository:** https://github.com/chainguard-dev/apko

apko builds OCI images from APK packages (Wolfi/Alpine) — no Dockerfile, no Docker daemon.

**Key features:**
- **No Docker daemon, no Dockerfile** — builds from a YAML manifest
- **APK-based** — assembles images from Wolfi/Alpine packages
- **Reproducible** — deterministic builds with pinned packages
- **SBOM generation** — produces Software Bill of Materials
- **Minimal images** — distroless-like, only what you specify

**Example (apko.yaml):**
```yaml
contents:
  repositories:
    - https://packages.wolfi.dev/os
  keyring:
    - https://packages.wolfi.dev/os/wolfi-signing.rsa.pub
  packages:
    - wolfi-baselayout
    - ca-certificates-bundle

entrypoint:
  command: /bin/sh
```

**Relevance to OCI factory:** apko is a declarative, package-based image builder. It's the most "factory-like" of the tools — you declare what packages you want, and it assembles the image. But it's APK-specific (Wolfi/Alpine), not a general-purpose Go-stdlib factory.

### 5. `google/go-containerregistry` (`crane`)

**Repository:** https://github.com/google/go-containerregistry

A Go library and CLI (`crane`) for manipulating OCI/Docker images programmatically.

**Key features:**
- **Go library** — `pkg/v1/mutate`, `pkg/v1/empty`, `pkg/v1/tarball` packages
- **`crane append`** — append tarball contents as a layer to an image
- **`crane mutate`** — change image config (entrypoint, env, labels)
- **`empty.Image`** — start from scratch
- **`mutate.Append`** / `mutate.AppendLayers` — Go API for appending layers
- **Registry operations** — push, pull, list, delete, copy

**Example (Go code, building from scratch):**
```go
import (
    "github.com/google/go-containerregistry/pkg/v1/empty"
    "github.com/google/go-containerregistry/pkg/v1/mutate"
    "github.com/google/go-containerregistry/pkg/v1/tarball"
)

// Start from scratch
img := empty.Image

// Create a layer from a tarball
layer, _ := tarball.LayerFromOpener(func() io.ReadCloser { ... })

// Append the layer
img, _ = mutate.AppendLayers(img, layer)

// Set config (entrypoint, env, etc.)
cfg, _ := img.ConfigFile()
cfg.Config.Entrypoint = []string{"/app/server"}
img, _ = mutate.Config(img, *cfg)
```

**Example (crane CLI):**
```bash
# Bundle a directory into an image, set entrypoint
crane mutate $(crane append -f <(tar -f - -c some-dir/) -t ${IMAGE}) --entrypoint=some-dir/entrypoint.sh
```

**Relevance to OCI factory:** This is the **most direct peer** to the gocker pattern. Both are Go libraries that construct OCI images programmatically. The difference:
- `go-containerregistry` is a full-featured library (~5K+ stars, production-grade, handles registries, multi-platform, signing)
- `gocker` is a minimal demo (~150 lines, stdlib-only, educational)

The OCI factory proposal already references `go-containerregistry` in `research/08-existing-go-oci-libraries.md`. This research confirms it's the right comparison.

### 6. `genuinetools/img`

**Repository:** https://github.com/genuinetools/img

A standalone, daemonless, unprivileged Dockerfile and OCI compatible container image builder.

**Key features:**
- **No Docker daemon** — uses BuildKit's DAG solver directly
- **Unprivileged** — can build without root
- **Dockerfile-compatible** — same UX as `docker build`
- **Cache-efficient** — uses BuildKit's caching

**Status:** Archived/unmaintained (last commit ~2022). Superseded by `buildx --builder` with remote builders.

**Relevance to OCI factory:** Historical interest only. Shows the daemonless approach is possible with BuildKit, but `img` is no longer maintained.

## Part 4: Comparison Matrix — All Image Building Approaches

| Tool | Daemon Required | Dockerfile Required | From Scratch | Language | Go Library | Production-Grade |
|------|----------------|---------------------|--------------|----------|------------|-----------------|
| **Docker build** | Yes | Yes | No (FROM scratch) | Any | No | Yes |
| **Ansible docker_image** | Yes | Yes | No | Any | No | Yes |
| **Ansible docker_image_build** | Yes (buildx) | Yes | No | Any | No | Yes |
| **Pulumi docker.Image** | Yes | Yes | No | Any | No | Yes |
| **Pulumi docker_build.Image** | Yes (buildx) | Yes | No | Any | No | Yes |
| **buildah** | No | Optional | Yes | Any | No (CLI) | Yes |
| **ko** | No | No | Yes (Go binaries) | Go only | No (CLI) | Yes |
| **jib** | No | No | Yes (Java apps) | Java only | No (Maven/Gradle) | Yes |
| **apko** | No | No (YAML manifest) | Yes (APK packages) | Any | No (CLI) | Yes |
| **go-containerregistry** | No | No | Yes (programmatic) | Go | **Yes** | Yes |
| **gocker (proposed)** | No | No | Yes (stdlib only) | Go | **Yes** | No (demo) |
| **img** | No | Yes | No | Any | No (CLI) | No (unmaintained) |

## Part 5: What This Means for the OCI Factory Proposal

### Ansible and Pulumi Are Not Peers

Neither Ansible nor Pulumi can build OCI images without Docker. They delegate all image building to the Docker daemon or buildx. This means:

- The OCI factory proposal has **no equivalent in Ansible or Pulumi**
- The OCI factory fills a gap that neither tool addresses
- The OCI factory is more comparable to `buildah`, `ko`, `jib`, `apko`, and `go-containerregistry`

### The True Peer: `google/go-containerregistry`

The OCI factory proposal's Part 3 (local OCI factory) should be compared to `google/go-containerregistry`, not to Ansible/Pulumi. The proposal already references this in `research/08-existing-go-oci-libraries.md`.

The key decision from the existing proposal:
- **Option A:** Use `google/go-containerregistry` (mature, feature-rich, handles registries)
- **Option B:** Use stdlib-only `gocker` pattern (minimal, no dependencies, educational)
- **Option C:** Use `buildah` CLI (daemonless, scriptable, but requires Linux namespaces)

The proposal recommends Option A or B. This research confirms that recommendation.

### What Ansible/Pulumi Teach Us (Indirectly)

Even though Ansible/Pulumi don't have an OCI factory, their image-building patterns offer lessons:

1. **Build context digest tracking (Pulumi):** Pulumi's `docker.Image` only rebuilds when the build context changes (computed via a digest). The OCI factory could adopt a similar pattern — only rebuild when input files change, tracked by a content hash.

2. **Registry caching (Pulumi):** Pulumi's `docker_build.Image` supports `cacheFrom`/`cacheTo` for registry-based layer caching. The OCI factory could support pushing layers to a registry cache for faster subsequent builds.

3. **Multi-platform builds (Pulumi/Ansible):** Both support `--platform` for multi-architecture images. The OCI factory should consider multi-platform support (at minimum, setting the OS/arch in the image config).

4. **`repoDigest` for tracking (Pulumi):** Pulumi exposes `repoDigest` as an immutable image identifier. The OCI factory should return the image digest (SHA256) as output, enabling users to track and pin specific builds.

5. **Build-on-preview (Pulumi):** Pulumi's `buildOnPreview` flag lets users skip building during preview/dry-run. The OCI factory should support a similar dry-run mode (log what would be built, don't actually create the image).

6. **Non-idempotent builds (Ansible):** Ansible's `docker_image_build` module is explicitly non-idempotent — it always rebuilds. The OCI factory should be the same: building an image is inherently non-idempotent (each build may produce a different result). `Check()` should always return `true` (or compare input digests).

### Updated Recommendation for the OCI Factory Proposal

Based on this research, the OCI factory proposal should:

1. **Reference `buildah`, `ko`, `jib`, `apko`, and `go-containerregistry` as peers** — not Ansible/Pulumi, which don't have equivalents.

2. **Consider `google/go-containerregistry` as the primary dependency** — it's the production-grade Go library for programmatic OCI image construction. The gocker pattern (stdlib-only) is educational but not production-grade.

3. **Adopt Pulumi's build context digest pattern** — only rebuild when inputs change, tracked by content hash.

4. **Support multi-platform image config** — at minimum, set OS/arch in the image config.

5. **Return the image digest as output** — for tracking and pinning (like Pulumi's `repoDigest`).

6. **Be non-idempotent** — building is inherently non-idempotent (like Ansible's `docker_image_build`). Use input digest comparison for "should I rebuild?" decisions.

7. **Support dry-run mode** — log what would be built without creating the image (like Pulumi's `buildOnPreview: false`).

## Summary Table: Ansible & Pulumi Image Building vs OCI Factory

| Aspect | Ansible | Pulumi | OCI Factory (proposed) |
|--------|---------|--------|------------------------|
| Build from Dockerfile | Yes (`docker_image`) | Yes (`docker.Image`) | No (programmatic) |
| Build from scratch | No | No | **Yes** |
| Docker daemon required | Yes | Yes | **No** |
| Daemonless | No | No | **Yes** |
| Go library | No (Python) | No (wraps Terraform provider) | **Yes** |
| Programmatic image construction | No | No | **Yes** |
| Multi-platform | Yes (buildx) | Yes (buildx) | Possible (set config) |
| Registry push | Yes | Yes | Possible (go-containerregistry) |
| Layer caching | Yes (buildx cache) | Yes (registry cache) | Possible (manual) |
| Build context digest | No | Yes (since v4) | **Recommended** |
| Image digest output | No | Yes (`repoDigest`) | **Recommended** |
| Dry-run / preview | Yes (check mode) | Yes (`buildOnPreview`) | **Recommended** |
| Idempotency | Non-idempotent (build) | Context-digest based | Non-idempotent (or input-digest) |

**Bottom line:** The OCI factory proposal fills a gap that neither Ansible nor Pulumi address. The true peers are `buildah`, `ko`, `jib`, `apko`, and especially `google/go-containerregistry`. The proposal should reference these tools, not Ansible/Pulumi, as the relevant comparisons for the "local OCI image factory" concept.
