# Task: fs.FileDelete

**Date:** 2026-07-31
**Status:** Draft
**Skill ID:** `fs-file-delete`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Deleting a single file is a common operation. Distinct from `fs.Remove`
which handles both files and directories recursively. `FileDelete` is
specifically for a single file, non-recursive, with idempotency.

## Proposed Skill

```go
type FileDelete struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewFileDelete()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File path to delete |

## Usage

```go
result := node.Run(fs.NewFileDelete().SetArg(fs.ArgPath, "/tmp/composer-installer.php")).FirstResult()
```

## Execution Flow

1. Validates `path` is non-empty
2. Checks if file exists (`test -f`)
3. If not exists, returns `Changed=false`
4. Runs `rm -f <path>`
5. Returns `Changed=true` if file was deleted

## Idempotency

- `Check()` returns `false` if file doesn't exist
- `Check()` returns `true` if file exists and needs deletion
