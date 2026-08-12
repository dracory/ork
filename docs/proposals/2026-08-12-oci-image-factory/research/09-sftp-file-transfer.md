# Source: SFTP File Transfer in Go (`github.com/pkg/sftp`)

**Sources:**
- https://github.com/pkg/sftp (repository)
- https://pkg.go.dev/github.com/pkg/sftp (package docs)
- https://github.com/pkg/sftp/blob/master/client.go (source)
- https://github.com/pkg/sftp/blob/master/README.md
**Retrieved:** 2026-08-12 (via web search + webfetch)
**Corrected:** 2026-08-12 (initial draft incorrectly claimed Ork has no file transfer; Ork has `file_create` and `Command.Stdin` for text, but lacks streaming binary transfer)

## Summary

`github.com/pkg/sftp` is the de facto Go library for SFTP (SSH File Transfer Protocol) support. It works on top of `golang.org/x/crypto/ssh` (which Ork already uses). It provides a `Client` type that supports file system operations on remote SSH servers — upload, download, mkdir, remove, rename, stat, etc. ~2K GitHub stars, actively maintained.

## Why This Matters for Ork

Ork's `ssh` package currently supports **command execution** (`ssh.Run`, `Client.Run`) plus two text-oriented file-writing mechanisms:

1. **`fs.NewFileCreate()`** — writes content to a remote file via `printf '%s' <escaped_content> > <path>` over SSH. Idempotent, with content/mode/owner checking. Works well for **text files** (configs, scripts).
2. **`types.Command.Stdin`** — a `string` field on `Command` that gets piped to the remote command's stdin. So `cat > /remote/file` with `Stdin: <content>` works, including under `BecomeUser`/`BecomePassword` (the `becomeWriter` state machine handles stdin delivery after escalation).

What Ork **lacks** for the Docker deployment use case:
- **SFTP/SCP** — no proper file-transfer protocol. The two mechanisms above are command-execution-based, not file-transfer-based.
- **Streaming transfer for large binary files** — `Stdin` is a `string` (loaded entirely into memory), not an `io.Reader` stream. Fine for a 2KB config file, impractical for a 500MB Docker image tarball.
- **Binary-safe upload** — `file_create` uses `printf '%s'` with shell-escaped content, which breaks on null bytes. `Stdin` as a Go string *can* technically hold binary data (Go strings allow null bytes), but the shell-escaping path and memory limits make it fragile for large binaries.

This is the gap for any Docker image deployment workflow:
- To `docker import` a tarball on a remote server, you first need to **get the (binary, potentially large) tarball to the server**
- To `docker load` a Docker save archive, you need to **transfer the (binary, potentially large) archive to the server**
- Ork's existing `file_create` and `Stdin` mechanisms work for text but are not suitable for large binary tarballs

## pkg/sftp Overview

### Client Creation
```go
import (
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

// Create SFTP client from an existing *ssh.Client
sftpClient, err := sftp.NewClient(sshClient)
```

### Key Methods
```go
// File operations
func (c *Client) Create(path string) (*File, error)     // create/open for writing
func (c *Client) Open(path string) (*File, error)        // open for reading
func (c *Client) OpenFile(path string, flags int) (*File, error)

// Directory operations
func (c *Client) Mkdir(path string) error
func (c *Client) MkdirAll(path string) error
func (c *Client) Remove(path string) error
func (c *Client) RemoveAll(path string) error

// Metadata
func (c *Client) Stat(path string) (os.FileInfo, error)
func (c *Client) Lstat(path string) (os.FileInfo, error)
func (c *Client) Rename(oldname, newname string) error

// Directory listing
func (c *Client) ReadDir(path string) ([]os.FileInfo, error)
func (c *Client) Walk(root string) *Walk                  // recursive walk

// File transfer
func (c *Client) Put(src, dst string) error              // convenience: local → remote
func (c *Client) Get(dst, src string) error              // convenience: remote → local
```

### File Read/Write
```go
// Upload a file
f, err := sftpClient.Create("/remote/path/file.tar")
defer f.Close()
io.Copy(f, localFile)

// Download a file
f, err := sftpClient.Open("/remote/path/file.tar")
defer f.Close()
io.Copy(localFile, f)
```

### Concurrent Operations
The `Client` is safe for concurrent use from multiple goroutines. Supports configurable max packet size, concurrent reads, and concurrent writes (off by default due to complexity).

### Dependencies
```
github.com/kr/fs           v0.1.0
github.com/stretchr/testify v1.11.1
golang.org/x/crypto        v0.54.0  (Ork already has this)
golang.org/x/sys           v0.47.0  (Ork already has this)
```

Only one truly new dependency: `github.com/kr/fs` (tiny filesystem-walking utility).

## Integration with Ork's SSH Package

Ork's `ssh.Client` wraps `*ssh.Client` from `golang.org/x/crypto/ssh`. The `pkg/sftp` library creates an SFTP client from the same `*ssh.Client`. Integration would be straightforward:

```go
// In ssh/sftp.go (new file)
package ssh

import (
    sftp "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

// SFTPClient wraps an SFTP session over an SSH connection.
type SFTPClient struct {
    client *sftp.Client
}

// NewSFTPClient creates an SFTP client from an existing SSH client.
func NewSFTPClient(sshClient *ssh.Client) (*SFTPClient, error) {
    c, err := sftp.NewClient(sshClient)
    if err != nil {
        return nil, err
    }
    return &SFTPClient{client: c}, nil
}

// UploadFile copies a local file to the remote server.
func (s *SFTPClient) UploadFile(localPath, remotePath string) error {
    localFile, err := os.Open(localPath)
    if err != nil { return err }
    defer localFile.Close()

    remoteFile, err := s.client.Create(remotePath)
    if err != nil { return err }
    defer remoteFile.Close()

    _, err = io.Copy(remoteFile, localFile)
    return err
}

func (s *SFTPClient) Close() error { return s.client.Close() }
```

### Ork's Current SSH Client Structure
From `ssh/ssh.go`:
```go
type Client struct {
    host              string
    port              string
    user              string
    keyPath           string
    kexAlgorithms     []string
    hostKeyAlgorithms []string
    client            *ssh.Client  // ← this is what sftp.NewClient needs
}
```

The `client` field is the raw `*ssh.Client` — exactly what `sftp.NewClient()` expects. Adding SFTP support would require:
1. Adding `github.com/pkg/sftp` to `go.mod`
2. Adding `github.com/kr/fs` (transitive dep)
3. New file `ssh/sftp.go` with `SFTPClient` type
4. Method on `ssh.Client` to expose the SFTP client: `func (c *Client) SFTP() (*SFTPClient, error)`

## Alternative: Stdin Piping via Existing `Command.Stdin`

Ork already has a `Stdin string` field on `types.Command` that pipes content to the remote command's stdin. This means you can already do:
```go
cmd := types.Command{
    Command: "cat > /tmp/remote.tar",
    Stdin:   tarballContent,  // string loaded into memory
}
ssh.Run(cfg, cmd)
```

This works today, without any new dependencies. However, it has significant drawbacks for the Docker use case:
- **`Stdin` is a `string`** — the entire tarball must fit in memory as a Go string
- **No streaming** — a 500MB Docker image tarball would require 500MB of memory just for the stdin string
- **No progress tracking** — no way to report upload progress
- **No integrity verification** — no checksum confirmation after transfer
- **Binary fragility** — while Go strings can hold null bytes, the surrounding command execution and shell context make large binary transfers fragile

For small text files (configs, scripts), `file_create` and `Stdin` are sufficient. For large binary tarballs (Docker images), SFTP is the proper solution because it streams via `io.Copy` without loading the entire file into memory.

## Relevance to Ork

1. **Essential for Docker image deployment** — to `docker import` or `docker load` on a remote server, the (binary, potentially large) tarball must get there first. Ork's existing `file_create` (text via `printf`) and `Command.Stdin` (string in memory) work for configs but not for large binaries. SFTP is the clean, streaming solution.
2. **Minimal new dependencies** — only `github.com/pkg/sftp` + `github.com/kr/fs`. Ork already has `golang.org/x/crypto` and `golang.org/x/sys`.
3. **Aligns with existing architecture** — works on top of Ork's existing SSH client, no new connection management needed.
4. **Useful beyond Docker** — streaming binary file transfer is a general-purpose capability that many Ork skills could benefit from (log retrieval, backup upload/download, certificate deployment, etc.).
5. **Concurrent-safe** — SFTP client can be used from multiple goroutines, aligns with Ork's inventory parallel execution.
6. **License:** MIT (checked via pkg.go.dev) — compatible with Ork's AGPL-3.0.

## Risk Assessment

- **Low risk:** `pkg/sftp` is mature (2K stars), well-tested, and widely used
- **Low risk:** Integration is additive — doesn't change existing SSH command execution
- **Medium risk:** Requires SFTP subsystem to be enabled on the remote SSH server (most servers have it, but some hardened configurations disable it)
- **Low risk:** Dependency footprint is tiny (`github.com/kr/fs` is ~100 lines)
