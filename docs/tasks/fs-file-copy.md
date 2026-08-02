# Task: fs.FileCopy

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-file-copy`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Copying files on the remote server (not upload from local) is needed for
operations like copying root's authorized_keys to a new user, backing up
configs, etc. Currently done with raw `cp` commands.

## Proposed Skill

```go
type FileCopy struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewFileCopy()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `src` | yes | — | Source file path |
| `dst` | yes | — | Destination file path |
| `force` | no | "false" | Overwrite destination if it exists |

## Usage

```go
result := node.Run(fs.NewFileCopy().SetArgs(map[string]string{
    fs.ArgSrc:   "/root/.ssh/authorized_keys",
    fs.ArgDst:   "/home/sinevia/.ssh/authorized_keys",
    fs.ArgForce: "true",
})).FirstResult()
```

## Execution Flow

1. Validates `src` and `dst` are non-empty
2. Checks if `src` exists
3. If `src` doesn't exist, returns error
4. If `dst` exists and `force=false`, returns `Changed=false`
5. Runs `cp <src> <dst>`
6. Returns `Changed=true` if file was copied

## Idempotency

- `Check()` returns `false` if `dst` exists and content matches `src`
  (compared with `diff` or `cmp`)
- `Check()` returns `true` if `dst` doesn't exist or content differs
