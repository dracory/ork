# Source: Go `crypto/sha256` Package

**Source URL:** https://pkg.go.dev/crypto/sha256
**Go version:** go1.26.5 (latest)
**License:** BSD-3-Clause
**Retrieved:** 2026-08-12 (via pkg.go.dev)

## Summary

Package `sha256` implements the SHA-224 and SHA-256 hash algorithms as defined in FIPS 180-4. Part of Go's standard library. Used for content-addressing in OCI images (every blob is named by its SHA-256 digest).

## Constants
```go
const BlockSize = 64   // block size in bytes
const Size = 32        // SHA-256 checksum size in bytes
const Size224 = 28     // SHA-224 checksum size in bytes
```

## Functions

### func Sum256(data []byte) [Size]byte
Returns the SHA-256 checksum of the data. Simplest API — one-shot hashing of a byte slice.
```go
sum := sha256.Sum256([]byte("hello world\n"))
fmt.Printf("%x", sum)
// Output: a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
```

### func New() hash.Hash
Returns a new `hash.Hash` computing SHA-256. Supports incremental hashing via `Write()`. Implements `encoding.BinaryMarshaler`/`BinaryUnmarshaler` for state save/restore.
```go
h := sha256.New()
h.Write([]byte("hello world\n"))
fmt.Printf("%x", h.Sum(nil))
```

### func Sum224(data []byte) [Size224]byte
SHA-224 variant. Rarely used in OCI contexts.

### func New224() hash.Hash
SHA-224 incremental variant.

## Usage in OCI Image Building

The OCI format uses SHA-256 digests as content addresses. Every blob (layer, config, manifest) is stored at `blobs/sha256/<hex-digest>` and referenced by `sha256:<hex-digest>`.

### Pattern from gocker:
```go
data, _ := os.ReadFile(tmpPath)  // read entire tarball
h := sha256.Sum256(data)
digest := "sha256:" + hex.EncodeToString(h[:])
blobPath := filepath.Join(blobDir, hex.EncodeToString(h[:]))
os.WriteFile(blobPath, data, 0644)
```

### Better Pattern (Streaming, for Large Layers):
```go
f, _ := os.Open(layerPath)
defer f.Close()
h := sha256.New()
io.Copy(h, f)  // stream, don't load into memory
digest := "sha256:" + hex.EncodeToString(h.Sum(nil))
```

## Key Observations

1. **143,338 known importers** — one of the most widely used Go packages
2. **Stdlib only** — no external dependencies
3. **FIPS 180-4 compliant** — meets cryptographic standards
4. **Two APIs**: `Sum256` (one-shot, simple) and `New()` (incremental, streaming)
5. **`[Size]byte` return type** — fixed-size array, not slice; prevents accidental modification
6. **BSD-3-Clause license** — compatible with Ork's AGPL-3.0

## Relevance to Ork

- Already available in Go's stdlib (no new dependency)
- Essential for OCI image building (content-addressing)
- Could also be used for file integrity verification skills (e.g., verifying remote file checksums)
- The streaming `New()` + `io.Copy` pattern is important for large files — Ork should use this instead of `Sum256(os.ReadFile(...))` to avoid loading entire layers into memory
- Ork already uses `golang.org/x/crypto` for SSH; `crypto/sha256` is a separate stdlib package
