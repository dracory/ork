# Task: fs.ChangeOwner

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-change-owner`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Changing ownership of files and directories is one of the most common raw SSH
operations. Currently hardcoded as `chown user:group path` with no validation,
no recursive option, and no idempotency check.

## Proposed Skill

```go
type ChangeOwner struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewChangeOwner()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File or directory path |
| `owner` | yes | — | Owner in `user:group` format |
| `recursive` | no | "false" | Apply recursively (chown -R) |

## Usage

```go
result := node.Run(fs.NewChangeOwner().SetArgs(map[string]string{
    fs.ArgPath:      "/home/sinevia/caddy",
    fs.ArgOwner:     "sinevia:sinevia",
    fs.ArgRecursive: "true",
})).FirstResult()
```

## Execution Flow

1. Validates `path` and `owner` are non-empty
2. Validates `owner` matches `user:group` format
3. Checks current ownership with `stat -c '%U:%G' <path>`
4. If already correct, returns `Changed=false`
5. Runs `chown [-R] <owner> <path>`
6. Returns `Changed=true` if ownership was changed

## Idempotency

- `Check()` compares current owner against desired owner
- `Check()` returns `false` if already correct
- `Check()` returns `true` if ownership mismatch

## Replacement Targets

| Project | File | Lines saved |
|---------|------|-------------|
| GoCarAndVanRentalOrk | `main.go` runCaddyInstall (chown -R) | ~8 |
| GoCarAndVanRentalOrk | `main.go` runProjectInstall (chown x2) | ~16 |
| GoCarAndVanRentalOrk | `main.go` runCaddyRestart (chown Caddyfile) | ~8 |
