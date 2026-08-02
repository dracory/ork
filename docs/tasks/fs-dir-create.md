# Task: fs.DirCreate

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-dir-create`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Creating directories with proper ownership and permissions is a common operation
across all playbooks. Currently every project hardcodes `mkdir -p`, `chown`, and
`chmod` as separate raw SSH commands — verbose, error-prone, and not idempotent.

## Proposed Skill

```go
type DirCreate struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewDirCreate()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | Directory path to create |
| `owner` | no | "" | Owner (user:group) to set after creation |
| `mode` | no | "755" | Permissions (octal, e.g. "755", "700") |
| `parents` | no | "true" | Create parent directories if needed (mkdir -p) |

## Usage

```go
result := node.Run(fs.NewDirCreate().SetArgs(map[string]string{
    fs.ArgPath:   "/home/sinevia/gocarandvanrental.co.uk",
    fs.ArgOwner:  "sinevia:sinevia",
    fs.ArgMode:   "755",
})).FirstResult()
```

## Execution Flow

1. Validates `path` is non-empty
2. Checks if directory already exists (`test -d`)
3. If not exists, runs `mkdir -p <path>` (or `mkdir <path>` if parents=false)
4. If `mode` set, runs `chmod <mode> <path>`
5. If `owner` set, runs `chown <owner> <path>`
6. Returns `Changed=true` if directory was created, `Changed=false` if already existed

## Idempotency

- `Check()` returns `false` (no change needed) if directory already exists with
  correct ownership and permissions
- `Check()` returns `true` if directory doesn't exist or ownership/mode mismatch

## Replacement Targets

| Project | File | Lines saved |
|---------|------|-------------|
| GoCarAndVanRentalOrk | `main.go` runCaddyInstall (mkdir + chown) | ~15 |
| GoCarAndVanRentalOrk | `main.go` runProjectInstall (mkdir + chown + chmod x2) | ~30 |
