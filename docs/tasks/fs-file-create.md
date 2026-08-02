# Task: fs.FileCreate

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-file-create`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Creating files with specific content, ownership, and permissions is common
(e.g. writing config files, authorized_keys). Currently done with raw
`echo > file`, `chown`, `chmod` commands.

## Proposed Skill

```go
type FileCreate struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewFileCreate()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File path to create |
| `content` | no | "" | Content to write (empty creates empty file) |
| `owner` | no | "" | Owner (user:group) to set |
| `mode` | no | "644" | Permissions (octal, e.g. "644", "600") |
| `overwrite` | no | "false" | Overwrite if file already exists |

## Usage

```go
result := node.Run(fs.NewFileCreate().SetArgs(map[string]string{
    fs.ArgPath:    "/home/sinevia/caddy/Caddyfile",
    fs.ArgContent: caddyfileContent,
    fs.ArgOwner:   "sinevia:sinevia",
    fs.ArgMode:    "644",
})).FirstResult()
```

## Execution Flow

1. Validates `path` is non-empty
2. Checks if file already exists
3. If exists and `overwrite=false`, returns `Changed=false`
4. Writes content via heredoc or `echo`
5. If `mode` set, runs `chmod <mode> <path>`
6. If `owner` set, runs `chown <owner> <path>`
7. Returns `Changed=true` if file was created/overwritten

## Idempotency

- `Check()` returns `false` if file exists with matching content and permissions
- `Check()` returns `true` if file doesn't exist or content/mode mismatch
