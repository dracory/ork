# Source: Existing Go OCI Libraries (go-containerregistry, etc.)

**Sources:**
- https://github.com/google/go-containerregistry (main library)
- https://pkg.go.dev/github.com/google/go-containerregistry (Go package docs)
- https://github.com/google/go-containerregistry/tree/main/pkg/v1/remote (remote package)
- https://ahmet.im/blog/building-container-images-in-go/index.html (tutorial blog)
**Retrieved:** 2026-08-12 (via web search + webfetch)

## Summary

There are several mature Go libraries for working with OCI/container images. The most prominent is `google/go-containerregistry`, which provides a complete library and CLI tools (`crane`, `gcrane`) for building, mutating, pushing, and pulling container images. Other tools include `containers/image` (Red Hat, used by `skopeo`), `oras-project/oras` (Microsoft), and `regclient` (community).

## google/go-containerregistry

**Repository:** https://github.com/google/go-containerregistry
**License:** Apache 2.0
**Stars:** ~3k+
**Status:** Actively maintained, widely used (Kubernetes ecosystem, ko, crane, etc.)

### Design Philosophy
Defines immutable interfaces for container image resources, backed by multiple mediums:
- `v1.Image` — represents a single-platform image
- `v1.Layer` — represents a single layer
- `v1.ImageIndex` — represents a multi-platform image index

### Image Backends (read)
- `remote.Image` — from a registry
- `tarball.Image` — from a Docker save tarball
- `daemon.Image` — from the local Docker daemon
- `layout.Image` — from an OCI Image Layout on disk
- `random.Image` — generated random image (for testing)

### Image Backends (write)
- `remote.Write` — push to a registry
- `tarball.Write` — write to a Docker save tarball
- `daemon.Write` — load into local Docker daemon
- `layout.AppendImage` — write to an OCI Image Layout on disk

### Key Packages
- `pkg/v1` — core types (Image, Layer, ImageIndex, Manifest, Config)
- `pkg/v1/remote` — registry client (push/pull, implements OCI Distribution Spec)
- `pkg/v1/tarball` — Docker save format read/write
- `pkg/v1/layout` — OCI Image Layout read/write
- `pkg/v1/daemon` — Docker daemon integration
- `pkg/v1/mutate` — mutate images (add layers, change config, etc.)
- `pkg/v1/partial` — build images from minimal subsets
- `pkg/authn` — authentication (keychain, basic, bearer)
- `pkg/name` — reference parsing (e.g., `gcr.io/my-project/my-image:tag`)
- `pkg/v1/stream` — streaming layer (for memory-efficient layer upload)

### CLI Tools
- `crane` — general-purpose CLI for managing images
  - `crane pull` / `crane push` — pull/push images
  - `crane cp` — copy images between registries
  - `crane append` — append layers to an image
  - `crane mutate` — modify image config
- `gcrane` — GCR-specific extensions

### Usage Example (from ahmet.im blog)
```go
import (
    "github.com/google/go-containerregistry/pkg/v1/remote"
    "github.com/google/go-containerregistry/pkg/authn"
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/mutate"
    "github.com/google/go-containerregistry/pkg/v1/tarball"
)

// Pull base image
img, err := crane.Pull("nginx")

// Create a layer that deletes /usr/share/nginx/html
deleteMap := map[string][]byte{
    "usr/share/nginx/.wh.html": []byte{},
}
deleteLayer, _ := crane.Layer(deleteMap)

// Create a layer with static content
staticLayer, _ := crane.Layer(contentMap)

// Append layers to the image
img, _ = mutate.AppendLayers(img, deleteLayer, staticLayer)

// Push to registry
crane.Push(img, "gcr.io/my-project/my-nginx:tag")
```

### Push Ordering (Critical)
When pushing to a registry:
1. Upload all layer blobs first
2. Upload the config blob
3. Upload the manifest last

This is because the manifest references the other blobs by digest, and the registry must verify they exist.

## Other Go OCI Tools

### containers/image (Red Hat)
- Used by `skopeo`
- Supports copying between registries, daemon, OCI layout, Docker archive
- `skopeo copy oci-archive:$path docker://$image` — bridge OCI to Docker

### oras-project/oras (Microsoft)
- Focused on OCI artifacts (not just container images)
- `oras cp --from-oci-layout $path $image` — copy from OCI layout to registry

### regclient (community)
- `regctl image import $image $file` — import OCI tar to Docker
- `regctl image copy ocidir://$dir $image` — copy OCI dir to registry

## go-containerregistry vs gocker (stdlib factory)

| Aspect | go-containerregistry | gocker (stdlib) |
|--------|---------------------|-----------------|
| **Dependencies** | Many (transitive deps) | Zero (stdlib only) |
| **Spec compliance** | Full OCI + Docker | Partial (inline config, missing media types) |
| **Registry push/pull** | Yes (OCI Distribution Spec) | No |
| **Multi-platform** | Yes (ImageIndex) | No |
| **Compression** | Yes (gzip, zstd) | No |
| **Docker daemon integration** | Yes (daemon.Write) | No (uses docker import via CLI) |
| **Image mutation** | Yes (mutate package) | No |
| **Code size** | Large (full library) | ~150 lines |
| **Maintenance** | Actively maintained by Google | Single-author proof of concept |
| **License** | Apache 2.0 | MIT |

## Relevance to Ork

### Option A: Use go-containerregistry
- **Pros:** Full spec compliance, registry push/pull, Docker daemon integration, actively maintained, handles auth/compression/multi-platform
- **Cons:** Adds a significant dependency tree to Ork (which currently has minimal deps), may be overkill for simple `docker import`/`docker load` skills
- **Best for:** If Ork wants to push images to registries or do complex image manipulation

### Option B: Use gocker's stdlib approach
- **Pros:** Zero dependencies, tiny code, aligns with Ork's minimal-dep philosophy
- **Cons:** Not spec-compliant (needs fixes), no registry push, no compression, proof-of-concept quality
- **Best for:** If Ork wants a local OCI layout builder with minimal footprint

### Option C: Neither — just use `docker` CLI commands via SSH
- **Pros:** Simplest, no new Go code needed, Docker handles everything
- **Cons:** Requires Docker installed on the control machine (for building) or the remote machine (for loading)
- **Best for:** Docker management skills (import, load, run, stop, etc.) that just shell out to `docker` via SSH

### Recommendation
For Ork's SSH-automation use case, **Option C is the best starting point** — Docker skills that shell out to `docker` via SSH. The factory (Option B) is a separate local tool that doesn't need to be integrated into Ork's skill system. If registry push/pull is ever needed, **Option A** (go-containerregistry) would be the right choice, but that's a much bigger commitment.
