# Source: Docker `import` and `load` Commands

**Source URLs:**
- `docker import`: https://docs.docker.com/engine/reference/commandline/import/
- `docker load`: https://docs.docker.com/engine/reference/commandline/load/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## Summary

Docker provides two commands for getting images into the local daemon from tar archives: `docker import` and `docker load`. They serve different purposes and accept different input formats. Understanding the difference is critical for any image-factory-to-Docker bridge.

## `docker image import` (alias: `docker import`)

**Purpose:** Import the contents from a tarball to create a filesystem image.

**Usage:** `docker image import [OPTIONS] file|URL|- [REPOSITORY[:TAG]]`

### What it does
- Takes a **flat filesystem tarball** (not an image format — just a tar of files)
- Docker untars it relative to `/` (root)
- Creates a new image with a single layer containing those files
- **No metadata** (no CMD, ENTRYPOINT, ENV, etc.) — the imported image is a bare filesystem

### Options
| Option | Description |
|--------|-------------|
| `-c, --change` | Apply Dockerfile instruction to the created image (CMD, ENTRYPOINT, ENV, EXPOSE, LABEL, USER, VOLUME, WORKDIR, etc.) |
| `-m, --message` | Set commit message for imported image |
| `--platform` | Set platform (os/arch/variant) for the imported image |

### Examples
```bash
# Import from local file
docker import /path/to/exampleimage.tgz

# Import from pipe/STDIN
cat exampleimage.tgz | docker import - exampleimagelocal:new

# Import from remote URL
docker import https://example.com/exampleimage.tgz

# Import with metadata (CMD, ENV)
docker import --change "ENV DEBUG=true" --change "CMD [\"/app/server\"]" ./rootfs.tgz exampleimagelocal:new

# Import from a local directory (via tar pipe)
sudo tar -c . | docker import - exampleimagedir
```

### Key Note on Permissions
> "Note the `sudo` in this example — you must preserve the ownership of the files (especially root ownership) during the archiving with tar. If you are not root (or the sudo command) when you tar, then the ownerships might not get preserved."

### Key Note on Running
An imported image has no metadata (CMD, ENTRYPOINT). You must specify the command explicitly:
```bash
docker run -d exampleimagelocal:new /app/server
```
Or use `--change` to set CMD/ENTRYPOINT at import time.

## `docker image load` (alias: `docker load`)

**Purpose:** Load an image from a tar archive or STDIN.

**Usage:** `docker image load [OPTIONS]`

### What it does
- Takes a **Docker save tarball** (a proper image format with manifest.json, layer tarballs, config, tags)
- Restores both images AND tags
- Preserves all metadata (CMD, ENTRYPOINT, ENV, layers, history)

### Options
| Option | Description |
|--------|-------------|
| `-i, --input` | Read from tar archive file, instead of STDIN |
| `--platform` | Load only the given platform(s) — API 1.48+ |
| `-q, --quiet` | Suppress the load output |

### Examples
```bash
# Load from STDIN
docker load < busybox.tar.gz

# Load from file
docker load --input fedora.tar

# Load specific platform
docker image load -i image.tar --platform=linux/amd64
```

## Critical Difference: `import` vs `load`

| Aspect | `docker import` | `docker load` |
|--------|----------------|---------------|
| **Input** | Flat filesystem tarball | Docker save tarball (image format) |
| **Metadata** | None (bare filesystem) | Full (CMD, ENV, layers, tags, history) |
| **Layers** | Single layer | Multiple layers preserved |
| **Tags** | Must specify explicitly | Restored automatically |
| **Use case** | Import arbitrary filesystem as image | Restore saved images |
| **Companion command** | `docker export` (container → tarball) | `docker save` (image → tarball) |
| **Running** | Must specify command or use `--change` | Can `docker run` directly |

## OCI Layout vs Docker Save Format

**Important finding from Stack Overflow research:**
- `docker load` expects the **Docker save format** (its own tarball with `manifest.json`), NOT the OCI Image Layout directly
- Loading an OCI tarball with `docker load` can fail with: `open /var/lib/docker/tmp/docker-import-.../blobs/json: no such file or directory`
- Docker Engine is adding support for loading OCI Layout directly, but it's still experimental (as of the research date)
- Tools like `crane`, `skopeo`, `oras`, `regclient` can bridge OCI Layout to Docker:
  ```bash
  crane push $path $image
  skopeo copy oci-archive:$path docker://$image
  regctl image import $image $file
  ```

## What the gocker Factory Does

The gocker factory produces TWO outputs:
1. **OCI Image Layout** (`--output`) — content-addressed blobs + manifest + index.json (the "proper" OCI format)
2. **Flat filesystem tarball** (`--tar`) — just a tar of files (for `docker import`)

The experiment uses the flat tarball + `docker import` path because `docker load` doesn't directly accept OCI Layout. This is a pragmatic bridge, not a spec-compliant one.

## Relevance to Ork

- **`docker import` is the simplest remote-server operation** — just transfer a flat tarball and run `docker import`. No special image format needed.
- **`docker load` requires Docker save format** — more complex, but preserves metadata (CMD, ENTRYPOINT). Better UX for end users.
- **Ork skills should support both** — `docker-import` skill (flat tarball → image) and `docker-load` skill (Docker save tarball → image).
- **The `--change` flag on `docker import`** is powerful — lets you add CMD/ENTRYPOINT/ENV at import time, making the imported image directly runnable. An Ork skill should expose this.
- **Permission preservation** is critical — the tarball must preserve Unix permissions (especially executable bit on binaries). This is a known gotcha that Ork skills should handle correctly.
- **`docker load` from OCI Layout** is not reliably supported — Ork should either use `docker import` (flat tarball) or convert OCI Layout to Docker save format first.
