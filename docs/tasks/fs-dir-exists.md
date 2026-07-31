# Task: fs.DirExists

**Date:** 2026-07-31
**Status:** Draft
**Skill ID:** `fs-dir-exists`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Checking if a directory exists is a prerequisite for many playbooks.
Currently done with raw `test -d` commands.

## Proposed Skill

```go
type DirExists struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewDirExists()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | Directory path to check |

## Usage

```go
result := node.Run(fs.NewDirExists().SetArg(fs.ArgPath, "/home/sinevia/caddy")).FirstResult()
if result.Details["exists"] == "true" {
    fmt.Println("Directory exists")
}
```

## Execution Flow

1. Validates `path` is non-empty
2. Runs `test -d <path>`
3. Returns `Changed=false` (no modification), with `Details["exists"]` = "true" or "false"

## Idempotency

This is a read-only check. `Check()` always returns `false` (no changes needed).
`Run()` reports existence in `Result.Details`.
