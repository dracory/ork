# Source: OCI Distribution Specification

**Source URL:** https://github.com/opencontainers/distribution-spec
**Spec document:** https://github.com/opencontainers/distribution-spec/blob/main/spec.md
**Retrieved:** 2026-08-12 (via raw.githubusercontent.com — README.md)

## Summary

The OCI Distribution Specification defines an API protocol to facilitate and standardize the distribution of content — primarily OCI container images. It defines the push/pull API that registries (Docker Hub, GitHub Container Registry, Google Artifact Registry, etc.) implement. Licensed under Apache 2.0.

## Relationship to Other OCI Specs

- **OCI Image Format Spec** — defines what an OCI Image IS (manifest, layers, config)
- **OCI Distribution Spec** (this) — defines how images are PUSHED TO and PULLED FROM registries
- **OCI Runtime Spec** — defines how to RUN a container from an unpacked image

The Distribution Spec is designed generically enough to distribute any content type, not just container images — the manifest format just needs to reference blobs.

## Key API Operations

The distribution spec defines these core endpoints (based on the Docker Registry HTTP API V2):

### Blob Operations
- `GET /v2/<name>/blobs/<digest>` — pull a blob (layer, config, manifest)
- `HEAD /v2/<name>/blobs/<digest>` — check blob existence
- `POST /v2/<name>/blobs/uploads/` — initiate blob upload (returns upload URL)
- `PUT /v2/<name>/blobs/uploads/<reference>` — complete blob upload with digest
- `DELETE /v2/<name>/blobs/<digest>` — delete a blob (if enabled)

### Manifest Operations
- `GET /v2/<name>/manifests/<reference>` — pull a manifest (by tag or digest)
- `HEAD /v2/<name>/manifests/<reference>` — check manifest existence
- `PUT /v2/<name>/manifests/<reference>` — push a manifest (by tag or digest)
- `DELETE /v2/<name>/manifests/<reference>` — delete a manifest (if enabled)

### Other
- `GET /v2/` — version check (returns 200 if V2 supported)
- `GET /v2/_catalog` — list repositories (if enabled)
- `GET /v2/<name>/tags/list` — list tags for a repository

### Authentication
Uses HTTP bearer-token auth challenges. The registry returns a `WWW-Authenticate` header pointing to a token service. The client fetches a token and includes it in subsequent requests.

### Upload Flow
1. Client POSTs to initiate upload → gets upload location
2. Client PUTs blob content to the upload location with `digest` query param
3. Registry verifies the blob matches the digest
4. For monolithic upload: single PUT with all content
5. For chunked upload: multiple PATCH requests then a final PUT

### Push Order (Critical)
When pushing an image:
1. Upload all layer blobs first
2. Upload the config blob
3. Upload the manifest last (it references layers and config by digest)

This ordering is required because the manifest references the other blobs, and the registry must verify they exist.

## Relevance to Ork

- **Not needed for the initial proposal.** The gocker factory stops at a local OCI layout. Pushing to a registry is a future enhancement.
- **If Ork ever adds registry push/pull skills**, this spec defines the API to implement. The `google/go-containerregistry` library already implements it (see research file 06).
- **Authentication complexity.** Bearer-token auth, token services, and various registry-specific auth flows (Docker Hub, GCR, GHCR, ECR) make this non-trivial. Using `go-containerregistry` would be far simpler than implementing from scratch.
- **Apache 2.0 license** — compatible with Ork's AGPL-3.0.
- **The spec is generic** — can distribute any content, not just images. Could be relevant if Ork ever distributes skills or playbooks as OCI artifacts.
