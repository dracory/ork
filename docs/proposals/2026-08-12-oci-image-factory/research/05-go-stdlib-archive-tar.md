# Source: Go `archive/tar` Package

**Source URL:** https://pkg.go.dev/archive/tar
**Go version:** go1.26.5 (latest)
**License:** BSD-3-Clause
**Retrieved:** 2026-08-12 (via pkg.go.dev)

## Summary

Package `tar` implements access to tar archives. It aims to cover most variations of the format, including those produced by GNU and BSD tar tools. Supports streaming read/write. Part of Go's standard library — no external dependencies.

## Key Types and Functions

### Constants (Type Flags)
```go
TypeReg    = '0'  // Regular file
TypeLink   = '1'  // Hard link
TypeSymlink = '2' // Symbolic link
TypeChar   = '3'  // Character device
TypeBlock  = '4'  // Block device
TypeDir    = '5'  // Directory
TypeFifo   = '6'  // FIFO node
TypeXHeader = 'x' // PAX extended header (per-file)
TypeXGlobalHeader = 'g' // PAX global header
```

### type Header
Represents a single tar header. Key fields:
- `Name` — name of file entry
- `Mode` — permission and mode bits
- `Uid`, `Gid` — numeric user/group ID
- `Uname`, `Gname` — user/group name
- `Size` — size in bytes
- `ModTime` — modification time
- `Typeflag` — type of entry (TypeReg, TypeDir, etc.)
- `Linkname` — target name of link
- `Format` — tar format (USTAR, PAX, GNU)

### func FileInfoHeader(fi fs.FileInfo, link string) (*Header, error)
Creates a Header from an `os.FileInfo`. This is the standard way to build headers when walking a directory. The `link` parameter is the target name for symlinks (empty for non-links).

### type Writer
```go
func NewWriter(w io.Writer) *Writer
func (tw *Writer) WriteHeader(hdr *Header) error
func (tw *Writer) Write(b []byte) (int, error)
func (tw *Writer) Close() error
func (tw *Writer) AddFS(fsys fs.FS) error  // Added in Go 1.23
func (tw *Writer) Flush() error
```

`AddFS` (Go 1.23+) writes all files from an `fs.FS` to the archive — could simplify layer creation.

### type Reader
```go
func NewReader(r io.Reader) *Reader
func (tr *Reader) Next() (*Header, error)
func (tr *Reader) Read(b []byte) (int, error)
```

Streaming reader — `Next()` advances to the next file, `Read()` reads the current file's content.

### Variables (Errors)
```go
ErrHeader          // invalid tar header
ErrWriteTooLong    // write too long
ErrFieldTooLong    // header field too long
ErrWriteAfterClose // write after close
ErrInsecurePath    // insecure file path
```

## Usage Pattern (from gocker and standard examples)

```go
func CreateTarball(sourceDir, outputPath string) error {
    f, _ := os.Create(outputPath)
    defer f.Close()
    tw := tar.NewWriter(f)
    defer tw.Close()

    return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
        if err != nil { return err }

        rel, _ := filepath.Rel(sourceDir, path)
        rel = filepath.ToSlash(rel)
        if rel == "." { return nil }

        header, _ := tar.FileInfoHeader(info, "")
        header.Name = rel
        // CRITICAL: set mode explicitly for executable bits
        header.Mode = int64(info.Mode().Perm())

        if info.IsDir() { header.Name += "/" }

        tw.WriteHeader(header)

        if info.IsDir() { return nil }

        in, _ := os.Open(path)
        defer in.Close()
        _, err = io.Copy(tw, in)
        return err
    })
}
```

## Key Observations for Ork

1. **Stdlib only** — no external dependencies, aligns with Ork's minimal-dependency philosophy
2. **Streaming** — `io.Copy(tw, in)` never loads full files into memory; scales to large layers
3. **Permission preservation is critical** — if a binary isn't marked executable in the tar, the container will fail at runtime with "permission denied". The gocker code hardcodes 0755; a proper implementation should use `info.Mode().Perm()`.
4. **Cross-platform** — works identically on Linux, macOS, Windows (important for Ork, which is developed on Windows per the project structure)
5. **`AddFS` (Go 1.23+)** — could simplify layer creation from an `fs.FS` but Ork requires Go 1.26 so this is available
6. **28,261 known importers** — this is one of the most widely used Go packages; well-tested and stable
7. **BSD-3-Clause license** — compatible with Ork's AGPL-3.0

## Relevance to Ork

- `archive/tar` is already available in Go's stdlib (no new dependency)
- Useful for any file-transfer or file-bundling skill, not just OCI image building
- Could be used to create tarballs for `docker import`, backup skills, or file distribution
- The `ErrInsecurePath` error is relevant for Ork's shell-injection prevention philosophy — tar extraction should validate paths to prevent directory traversal
