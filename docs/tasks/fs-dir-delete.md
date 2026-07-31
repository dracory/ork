# Task: fs.DirDelete

**Date:** 2026-07-31
**Status:** Draft
**Skill ID:** `fs-dir-delete`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Deleting a directory (and its contents) is a common cleanup operation.
Distinct from `fs.Remove` which is generic. `DirDelete` validates the
path is a directory before removing, preventing accidental file deletion.

## Proposed Skill

```go
type DirDelete struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewDirDelete()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | Directory path to delete |
| `recursive` | no | "true" | Remove recursively (rm -r) |

## Usage

```go
result := node.Run(fs.NewDirDelete().SetArgs(map[string]string{
    fs.ArgPath:      "/home/sinevia/.pm2",
    fs.ArgRecursive: "true",
})).FirstResult()
```

## Execution Flow

1. Validates `path` is non-empty
2. Checks if path exists and is a directory (`test -d`)
3. If not exists or not a directory, returns `Changed=false`
4. Runs `rm -r <path>` (or `rmdir <path>` if recursive=false)
5. Returns `Changed=true` if directory was deleted

## Idempotency

- `Check()` returns `false` if directory doesn't exist
- `Check()` returns `true` if directory exists
