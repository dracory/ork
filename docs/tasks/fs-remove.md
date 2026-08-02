# Task: fs.Remove

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-remove`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Removing files/directories is needed for cleanup and uninstall operations.
Currently hardcoded as `rm -rf <path>` with no idempotency.

## Proposed Skill

```go
type Remove struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewRemove()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File or directory path to remove |
| `recursive` | no | "false" | Remove recursively (rm -r) |
| `force` | no | "false" | Force remove (rm -f, ignore errors) |

## Usage

```go
result := node.Run(fs.NewRemove().SetArgs(map[string]string{
    fs.ArgPath:      "/home/sinevia/.pm2",
    fs.ArgRecursive: "true",
    fs.ArgForce:     "true",
})).FirstResult()
```

## Execution Flow

1. Validates `path` is non-empty
2. Checks if path exists
3. If not exists, returns `Changed=false`
4. Runs `rm [-r] [-f] <path>`
5. Returns `Changed=true` if path was removed

## Idempotency

- `Check()` returns `false` if path doesn't exist
- `Check()` returns `true` if path exists and needs removal
