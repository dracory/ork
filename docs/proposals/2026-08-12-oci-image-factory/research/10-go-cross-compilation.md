# Source: Go Cross-Compilation

**Source URL:** https://go.dev/doc/install/source#environment (referenced in blog post)
**Retrieved:** 2026-08-12 (knowledge from Go documentation + blog post context)

## Summary

Go's toolchain supports cross-compilation natively via two environment variables: `GOOS` (target operating system) and `GOARCH` (target architecture). This is a key enabler for the OCI image factory pattern — you can build a Linux binary on any host (Windows, macOS) and package it into a container image.

## How It Works

```bash
# Build a Linux/amd64 binary on any host
GOOS=linux GOARCH=amd64 go build -o ./app/server ./cmd/server/main.go

# Other common targets
GOOS=linux GOARCH=arm64 go build -o ./app/server-arm64 ./cmd/server
GOOS=linux GOARCH=arm   go build -o ./app/server-arm ./cmd/server
GOOS=windows GOARCH=amd64 go build -o ./app.exe ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o ./app-mac ./cmd/server
```

On Windows (PowerShell):
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o ./app/server ./cmd/server/main.go
```

## Why This Matters for the Factory

The gocker experiment's key proof point: a binary built on **Windows** was packaged into an OCI image and ran successfully inside a **Linux** container. This works because:

1. Go produces statically-linked binaries by default (no shared library dependencies)
2. Cross-compilation is built into the toolchain (no cross-compiler needed, unlike C/C++)
3. The resulting binary is a Linux ELF file regardless of the build host
4. The OCI image just wraps the binary in a tarball — no OS is needed in the image

This is why the gocker factory can produce an 8MB image: it's just the Go binary + two sample files. No base OS, no package manager, no shell — just the binary.

## Supported GOOS/GOARCH Combinations

Common combinations relevant for container images:
| GOOS | GOARCH | Architecture |
|------|--------|-------------|
| linux | amd64 | x86-64 (most common for servers) |
| linux | arm64 | ARM 64-bit (AWS Graviton, RPi 4) |
| linux | arm | ARM 32-bit |
| linux | ppc64le | PowerPC 64-bit LE |
| linux | s390x | IBM System Z |

Full list: `go tool dist list`

## CGO Considerations

- **Default:** CGO is disabled when cross-compiling (`CGO_ENABLED=0`)
- This means Go's `net` and `os/user` packages use pure-Go implementations
- For most server applications this is fine
- If CGO is needed, cross-compilation becomes much harder (requires a C cross-compiler)

## Relevance to Ork

1. **Not directly needed for Ork skills** — Ork skills run on remote servers via SSH, they don't build binaries. Cross-compilation is a build-time concern, not a deployment concern.

2. **Relevant for a potential `ork oci build` CLI command** — if Ork adds a local OCI factory CLI, it could optionally cross-compile a Go binary before packaging it. This would be a convenience feature, not a core skill.

3. **Relevant for the proposal's scope** — the factory pattern (build locally → package → deploy remotely) requires cross-compilation to be useful. Without it, you'd need a Linux build host to produce Linux container images.

4. **Ork is developed on Windows** (per the project structure) — cross-compilation is essential for testing Linux container images during development.

5. **Not a dependency** — cross-compilation is a Go toolchain feature, not a library. No new dependencies needed.

## Conclusion

Cross-compilation is a **prerequisite capability** for the OCI factory pattern to be useful, but it's already available in Go's toolchain. No action needed — just documentation in the proposal about how the factory would use it.
