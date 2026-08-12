# Source: OCI Image Format Specification

**Source URL:** https://github.com/opencontainers/image-spec
**Spec document:** https://github.com/opencontainers/image-spec/blob/main/spec.md
**Retrieved:** 2026-08-12 (via raw.githubusercontent.com — README.md, image-layout.md, manifest.md)

## Summary

The OCI Image Format Specification defines the structure of a container image: a content-addressed directory layout containing blobs (layers, config, manifest), an index file, and a layout descriptor. The spec is maintained by the Open Containers Initiative and licensed under Apache 2.0.

## Key Concepts

### OCI Image Layout
A directory structure for content-addressable blobs and location-addressable references. Can be transported via tar, zip, NFS, HTTP, FTP, rsync.

**Required structure:**
```
<image-layout>/
├── blobs/
│   └── <alg>/  (e.g., sha256/)
│       └── <encoded>  (content-addressed blob)
├── index.json  (REQUIRED — entry point, image index JSON)
└── oci-layout  (REQUIRED — JSON with imageLayoutVersion)
```

### Blob Naming
- Blobs are named by their content: `blobs/<alg>/<encoded>` MUST match digest `<alg>:<encoded>`
- Example: content of `blobs/sha256/da39a3ee...` MUST match `sha256:da39a3ee...`
- Self-verifying by construction: change a byte and the filename no longer matches
- Blobs are opaque — no schema, treated as arbitrary content

### oci-layout File
JSON marker for the base of the layout:
```json
{"imageLayoutVersion": "1.0.0"}
```
Media type: `application/vnd.oci.layout.header.v1+json`

### index.json File
REQUIRED entry point. An image index (multi-descriptor). Contains a `manifests` array with descriptors pointing to manifests or other indices. Each descriptor has `mediaType`, `size`, `digest`, optional `platform`, and optional `annotations` (e.g., `org.opencontainers.image.ref.name` for tags).

### Image Manifest (`application/vnd.oci.image.manifest.v1+json`)
For a single platform image. Required properties:
- `schemaVersion` — MUST be `2` for backward compatibility
- `mediaType` — `application/vnd.oci.image.manifest.v1+json`
- `config` — descriptor referencing the config blob (media type `application/vnd.oci.image.config.v1+json`)
- `layers` — array of descriptors, base layer at index 0, in stack order. Media types:
  - `application/vnd.oci.image.layer.v1.tar` (uncompressed)
  - `application/vnd.oci.image.layer.v1.tar+gzip` (compressed — "frequently used")
  - `application/vnd.oci.image.layer.v1.tar+zstd` (SHOULD be supported)
- Optional: `artifactType`, `subject`, `annotations`

### Config Blob (`application/vnd.oci.image.config.v1+json`)
Contains architecture, OS, root filesystem (type + DiffIDs), and runtime config (Cmd, Env, Entrypoint, etc.). The DiffIDs are the uncompressed layer digests, in order.

### Layer Ordering
- Base layer at `layers[0]`, subsequent layers in stack order
- Final filesystem = applying layers to an empty directory in order
- Layers use whiteout files (`.wh.<name>`) to delete entries from lower layers

## Spec Compliance Issues in the gocker Factory

Based on comparing the gocker source code against the spec:

1. **Config is embedded inline in the manifest** — the spec requires `config` to be a *descriptor* (reference by digest + size + mediaType), not the config struct itself. The gocker code puts the `Config` struct directly in the `Manifest.Config` field instead of a descriptor pointing to a separate config blob. (The code does write a separate config blob, but the manifest doesn't reference it by descriptor.)

2. **Missing mediaType on config and layer descriptors** — the spec says implementations MUST support specific media types on `config` and `layers[]` descriptors. The gocker manifest omits `mediaType` from these descriptors entirely.

3. **No compression** — layers are uncompressed (`application/vnd.oci.image.layer.v1.tar`). The spec says `+gzip` is "frequently used" and implementations SHOULD support `+zstd`. Uncompressed is technically valid but unusual.

4. **index.json is a raw array, not an image index object** — the spec says `index.json` MUST be an image index JSON object with `schemaVersion`, `mediaType`, and `manifests` array. The gocker code writes a bare JSON array `[...]` instead of `{"schemaVersion": 2, "mediaType": "...", "manifests": [...]}`.

5. **DiffID vs Digest** — for uncompressed layers, DiffID == blob digest. This is correct in gocker. For compressed layers, DiffID would be the digest of the *uncompressed* layer, while the blob digest is the digest of the *compressed* layer.

## Relevance to Ork

- The OCI Image Layout is a well-defined, stable format (v1.0.0) that Ork could produce or consume
- The spec is Apache 2.0 — compatible with Ork's AGPL-3.0
- Go types are available in `github.com/opencontainers/image-spec/specs-go` (but would add a dependency)
- A spec-compliant implementation would need to fix the gocker factory's issues (config descriptor, media types, proper index.json)
- The format's self-verifying property (content-addressed blobs) aligns with Ork's idempotency philosophy
