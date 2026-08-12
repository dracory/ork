# Source: gocker Source Code (builder.go, main.go, experiment.sh, README.md)

**Source URL:** https://github.com/cheikh2shift/go-snippets/tree/main/gocker
**Author:** Cheikh Seck
**License:** MIT
**Retrieved:** 2026-08-12 (via raw.githubusercontent.com)

## Summary

The `gocker` project is a minimal OCI-compliant container image builder built entirely with Go's standard library. Zero external dependencies. It consists of four files: `builder.go` (library), `main.go` (CLI), `experiment.sh` (end-to-end test), and `README.md`.

## File: builder.go (Full Source Retrieved)

### Imports (stdlib only)
```go
import (
    "archive/tar"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
)
```

### Types
- `Layer` — `Digest string`, `Size int64`, `DiffID string` (JSON-tagged `diffIds`)
- `Config` — `Architecture`, `OS`, `RootFS` (with `Type` and `DiffIDs []string`)
- `Manifest` — `SchemaVersion int`, `MediaType string`, `Config Config`, `Layers []Layer`
- `Builder` — `outputDir string`, `verbose bool`, `layers []Layer`, `config Config`

### Key Implementation Details

**CreateTarball:**
- Resolves source dir to absolute path
- Creates output file, wraps in `tar.NewWriter`
- Walks source with `filepath.Walk`
- For each entry: builds tar header with `tar.FileInfoHeader`, sets relative slash-separated name, sets `header.Mode = 0755` (hardcoded — note: the blog post says it preserves `info.Mode().Perm()` but the actual code uses 0755 for all entries)
- Directories get trailing `/` in name, header only
- Regular files: streams bytes via `io.Copy(tw, in)`

**AddLayerFromDir:**
- Creates temp file, tars source dir into it
- Reads entire tarball into memory (`os.ReadFile`)
- Computes `sha256.Sum256`, builds `sha256:<hex>` digest
- Writes tarball to `blobs/sha256/<hex>` in output dir
- For uncompressed layers: `DiffID = Digest` (same value)
- Appends `Layer` to `b.layers`

**Build:**
- Appends all layer DiffIDs to `config.RootFS.DiffIDs`
- Sets `config.RootFS.Type = "layers"` (via SetDiffIDs, though Build doesn't call it — potential bug)
- Marshals manifest with `json.MarshalIndent`
- Computes manifest digest, writes manifest blob
- Marshals config, computes config digest, writes config blob
- Writes `index.json` — array with one descriptor pointing at manifest
- Writes `oci-layout` — `{"imageLayoutVersion":"1.0.0"}`

### Notable Issues in the Code
1. **Hardcoded 0755 mode** for all entries — the blog claims `info.Mode().Perm()` but the actual code uses `0755` for everything. This loses original permissions.
2. **Config not properly separated** — the manifest's `Config` field embeds the config struct inline rather than referencing it by descriptor (as the OCI spec requires). The blog acknowledges this as a limitation.
3. **No compression** — layers are uncompressed tar. The OCI spec says layers "frequently use the `+gzip` types."
4. **No media types on config/layer descriptors** — the manifest's `config` and `layers` entries lack `mediaType` fields, which the OCI spec says implementations MUST support.
5. **Reads entire tarball into memory** — `os.ReadFile(tmpPath)` loads the whole layer. Fine for small layers, problematic for large ones. The blog's `io.Copy` streaming is only for the tar creation, not the blob storage.
6. **No error checking** on several `os.WriteFile` and `os.MkdirAll` calls.

## File: main.go (Full Source Retrieved)

Simple CLI with `flag` package:
- `--source` (required) — directory to build into a layer
- `--output` (default "oci-image") — output directory for OCI layout
- `--tar` (optional) — if set, creates a flat tarball instead of OCI layout
- `--verbose` — enables digest/size logging

Two paths:
- `--tar` set → `CreateTarball(source, tarOut)` only (flat tar for `docker import`)
- `--tar` not set → `NewBuilder(output, verbose)` → `AddLayerFromDir(source)` → `Build()`

## File: experiment.sh (Full Source Retrieved)

7-step end-to-end test:
1. Pre-flight checks (Go + Docker installed)
2. Create sample files in `./layer-content/`
3. Cross-compile server binary (`GOOS=linux GOARCH=amd64`)
4. Build OCI image layout (`go run . --source ./layer-content --output ./oci-image --verbose`)
5. Verify OCI layout structure (`find` + `cat index.json`)
6. Import into Docker: build flat tar → `docker import` → `docker run` → `curl` test
7. Round-trip: `docker save` → `docker rmi` → `docker load` → `docker run` → `curl` test
8. Cleanup

Uses `MSYS_NO_PATHCONV=1` for Windows Git Bash compatibility.

## File: README.md (Full Source Retrieved)

Key design decisions documented:
- **Why Pure Go?** No Python, no platform-specific tar flags. Works identically on Linux, macOS, Windows.
- **Why Cross-Compile?** Binary built on any host can be packaged and run on Linux.
- **Why OCI Layout?** Simple directory format, inspectable with `cat`/`jq`, loadable with `docker load`, pushable with `skopeo copy`.
- **Why Zero Dependencies?** No `go mod download`, no supply chain risk, fast compilation, statically linked binaries.

## Relevance to Ork

- The code is ~150 lines of Go, stdlib only — could be vendored into Ork with minimal footprint
- The `CreateTarball` function is independently useful for any file-transfer skill
- The OCI layout builder is a local build tool, not a remote SSH operation
- The code has real bugs (hardcoded permissions, inline config, no media types) that would need fixing for spec compliance
- MIT license is compatible with Ork's AGPL-3.0
