# Task: fs.FileExists

**Date:** 2026-07-31
**Status:** Draft
**Skill ID:** `fs-file-exists`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Checking if a file exists is a prerequisite for many playbooks.
Currently done with raw `test -f` commands.

## Proposed Skill

```go
type FileExists struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewFileExists()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File path to check |

## Usage

```go
result := node.Run(fs.NewFileExists().SetArg(fs.ArgPath, "/home/sinevia/caddy/caddy")).FirstResult()
if result.Details["exists"] == "true" {
    fmt.Println("Caddy binary exists")
}
```

## Execution Flow

1. Validates `path` is non-empty
2. Runs `test -f <path>`
3. Returns `Changed=false` (read-only), with `Details["exists"]` = "true" or "false"

## Idempotency

Read-only check. `Check()` always returns `false`.
