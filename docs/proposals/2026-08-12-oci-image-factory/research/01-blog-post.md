# Source: "Building a Docker Image Factory From Scratch in Go"

**Source URL:** Cheikh Seck's blog post (Jul 30, 2026) — delivered via user message
**Author:** Cheikh Seck
**Source code:** https://github.com/cheikh2shift/go-snippets/tree/main/gocker
**Retrieved:** 2026-08-12

## Summary

The blog post demonstrates building a Docker/OCI image factory in Go using **only the standard library** — no Docker daemon, no BuildKit, no third-party packages. The core insight is that an OCI image is just a directory of content-addressed files (tarballs + JSON manifests), so any program that can compute SHA-256 and write files can manufacture a valid image.

## Key Claims

1. **A Docker image builder is just a program.** Given a tarball and a few JSON blobs, any code that can compute SHA-256 and write files can produce a valid OCI image.
2. **The format is simple enough to implement from scratch.** Blobs, a manifest, and a layout file — the entire format fits in a few hundred lines of Go.
3. **Small image size is a side effect of ownership.** The factory adds nothing you didn't ask for — no base OS, no cache. The example produces an 8MB image for a Go server.

## Architecture of the Factory

The factory (`builder.go`) defines four types and three functions:

### Types
- `Layer` — one content-addressed tarball (Digest, Size, DiffID)
- `Config` — image architecture, OS, and root filesystem (DiffIDs)
- `Manifest` — ties config and layers together (SchemaVersion, MediaType, Config, Layers)
- `Builder` — accumulates layers and knows where to write output

### Functions
1. **`CreateTarball(sourceDir, outputPath)`** — walks a directory, writes a tar archive with `archive/tar`. Streams file bytes via `io.Copy` (never loads fully into memory). Sets file mode explicitly (critical: if the binary isn't marked executable, the container fails at runtime with "permission denied").
2. **`AddLayerFromDir(sourceDir)`** — tars the directory to a temp file, reads bytes, computes `sha256:<hex>` digest, writes the tarball into `blobs/sha256/<hexDigest>`. For uncompressed layers, DiffID == blob digest.
3. **`Build()`** — appends layer DiffIDs to config's rootfs, marshals the manifest, writes the manifest blob, config blob, `index.json` (pointing at the manifest by digest), and `oci-layout` descriptor.

### CLI (`main.go`)
Two modes:
- `--source <dir> --output <dir>` — builds full OCI layout (blobs + manifest + index.json)
- `--source <dir> --tar <output.tar>` — creates a flat filesystem tarball for `docker import`

### Experiment (`experiment.sh`)
End-to-end test:
1. Cross-compile Go server binary (`GOOS=linux GOARCH=amd64`)
2. Build OCI image layout
3. Verify layout structure
4. `docker import` the flat tarball → `docker run` → `curl` verify
5. `docker save` → `docker rmi` → `docker load` → `docker run` (round-trip integrity test)

Result: 8.36MB image, round-trip passed, binary built on Windows ran inside Linux container.

## Limitations Acknowledged by Author

- Config is embedded inline rather than separated per the stricter OCI reading
- No multi-platform manifest list (no linux/arm64 alongside amd64)
- Layers aren't compressed (stdlib `compress/flate` would handle that)
- Stops at local layout + flat import tar — doesn't push to a registry (would need OCI Distribution API bearer-token auth + blob upload flow)

## Relevance to Ork

The factory is a **build-time tool** (runs locally, produces OCI artifacts). Ork skills run **on remote servers via SSH**. These execution contexts don't naturally overlap. However:
- The factory's output (flat tarball) is consumed by `docker import`, which IS a remote-server operation Ork could automate
- The factory could be a local library/CLI companion to Ork's remote deployment skills
- The blog validates that stdlib-only image building is feasible and produces tiny images

## Links Referenced in the Post

- OCI Image Specification: https://github.com/opencontainers/image-spec
- OCI Distribution Specification: https://github.com/opencontainers/distribution-spec
- Go archive/tar: https://pkg.go.dev/archive/tar
- Go crypto/sha256: https://pkg.go.dev/crypto/sha256
- Docker import reference: https://docs.docker.com/engine/reference/commandline/import/
- Go cross-compilation docs: https://go.dev/doc/install/source#environment
- Full source: https://github.com/cheikh2shift/go-snippets/tree/main/gocker
