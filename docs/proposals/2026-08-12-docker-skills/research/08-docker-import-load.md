# Source: `docker image import` and `docker image load`

**Source URLs:**
- `docker import`: https://docs.docker.com/engine/reference/commandline/import/
- `docker load`: https://docs.docker.com/engine/reference/commandline/load/
**Retrieved:** 2026-08-12 (via docs.docker.com)
**Cross-reference:** See `docs/proposals/2026-08-12-oci-image-factory/research/07-docker-import-load.md` for the full analysis

## Summary

Docker provides two commands for getting images into the local daemon from tar archives. They serve different purposes and accept different input formats.

## `docker image import` (alias: `docker import`)

**Usage:** `docker image import [OPTIONS] file|URL|- [REPOSITORY[:TAG]]`

Takes a **flat filesystem tarball** (not an image format — just a tar of files). Docker untars it relative to `/` (root). Creates a new image with a single layer. **No metadata** (no CMD, ENTRYPOINT, ENV).

### Options
| Option | Description |
|--------|-------------|
| `-c, --change` | Apply Dockerfile instruction (CMD, ENTRYPOINT, ENV, EXPOSE, LABEL, USER, VOLUME, WORKDIR) |
| `-m, --message` | Set commit message |
| `--platform` | Set platform (os/arch/variant) |

### Idempotency
`docker import` always creates a new image (new SHA). For an Ork skill:
- `Check()`: does the target image name:tag already exist? If yes, return `false`. If no, return `true`.
- `Run()`: execute `docker import <tarball> <image:tag>`.

### Check pattern
```bash
docker image inspect <image:tag> >/dev/null 2>&1
# Exit 0 = exists (no import needed)
# Non-zero = doesn't exist (needs import)
```

## `docker image load` (alias: `docker load`)

**Usage:** `docker image load [OPTIONS]`

Takes a **Docker save tarball** (proper image format with manifest.json, layers, config, tags). Restores both images AND tags. Preserves all metadata.

### Options
| Option | Description |
|--------|-------------|
| `-i, --input` | Read from tar archive file, instead of STDIN |
| `--platform` | Load only the given platform(s) — API 1.48+ |
| `-q, --quiet` | Suppress the load output |

### Idempotency
`docker load` is idempotent — loading an image that already exists with the same digest is a no-op. For an Ork skill:
- `Check()`: does the image (by tag or digest) already exist? If yes, return `false`. If no, return `true`.
- `Run()`: execute `docker load -i <tarball>`.

## Critical Difference

| Aspect | `docker import` | `docker load` |
|--------|----------------|---------------|
| Input | Flat filesystem tarball | Docker save tarball (image format) |
| Metadata | None (bare filesystem) | Full (CMD, ENV, layers, tags, history) |
| Layers | Single layer | Multiple layers preserved |
| Tags | Must specify explicitly | Restored automatically |
| Running | Must specify command or use `--change` | Can `docker run` directly |

## Relevance to Ork

- `docker-import` skill: takes a remote tarball path + image name:tag, runs `docker import`
- `docker-load` skill: takes a remote tarball path, runs `docker load -i`
- Both require the tarball to already be on the remote server (file transfer is a separate concern — see SFTP proposal in the OCI factory proposal)
- `docker-import` should support `--change` flags for adding CMD/ENTRYPOINT/ENV at import time
- Both are idempotent: `Check()` probes if the image already exists
- All paths and image names must be shell-escaped via `skills.ShellEscapeArg`
