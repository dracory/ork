# Task: fs.Rename

**Date:** 2026-07-31
**Status:** Draft
**Skill ID:** `fs-rename`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Renaming/moving files and directories is needed for atomic config updates
(e.g. write temp file then rename), backup operations, and reorganization.
Currently done with raw `mv` commands.

## Proposed Skill

```go
type Rename struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewRename()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `src` | yes | — | Source path |
| `dst` | yes | — | Destination path |
| `force` | no | "false" | Overwrite destination if it exists |

## Usage

```go
result := node.Run(fs.NewRename().SetArgs(map[string]string{
    fs.ArgSrc:   "/etc/ssh/sshd_config.bak",
    fs.ArgDst:   "/etc/ssh/sshd_config",
    fs.ArgForce: "true",
})).FirstResult()
```

## Execution Flow

1. Validates `src` and `dst` are non-empty
2. Checks if `src` exists
3. If `src` doesn't exist, returns error
4. If `dst` exists and `force=false`, returns error
5. Runs `mv [-f] <src> <dst>`
6. Returns `Changed=true` if rename succeeded

## Idempotency

- `Check()` returns `false` if `src` doesn't exist and `dst` exists
  (already renamed)
- `Check()` returns `true` if `src` exists (needs rename)
