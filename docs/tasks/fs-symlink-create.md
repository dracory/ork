# Task: fs.SymlinkCreate

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-symlink-create`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Creating symlinks is needed for making binaries available system-wide
(e.g. pm2, caddy). Currently hardcoded as `ln -sf target linkname`.

## Proposed Skill

```go
type SymlinkCreate struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewSymlinkCreate()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `target` | yes | — | Path the symlink points to |
| `link` | yes | — | Path of the symlink itself |

## Usage

```go
result := node.Run(fs.NewSymlinkCreate().SetArgs(map[string]string{
    fs.ArgTarget: "/home/sinevia/node_modules/.bin/pm2",
    fs.ArgLink:   "/usr/local/bin/pm2",
})).FirstResult()
```

## Execution Flow

1. Validates `target` and `link` are non-empty
2. Checks if symlink already exists and points to correct target
3. If correct, returns `Changed=false`
4. Runs `ln -sf <target> <link>`
5. Returns `Changed=true` if symlink was created or updated

## Idempotency

- `Check()` reads current symlink target with `readlink -f <link>`
- `Check()` returns `false` if symlink already points to correct target
- `Check()` returns `true` if symlink missing or points elsewhere
