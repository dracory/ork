# Task: fs.ChangeMode

**Date:** 2026-07-31
**Status:** Completed
**Skill ID:** `fs-change-mode`
**Package:** `github.com/dracory/ork/skills/fs`

## Problem

Changing file/directory permissions is common. Currently hardcoded as
`chmod <mode> <path>` with no idempotency check.

## Proposed Skill

```go
type ChangeMode struct {
    *types.BaseSkill
}
```

**Constructor:** `fs.NewChangeMode()`

## Arguments

| Arg | Required | Default | Description |
|-----|----------|---------|-------------|
| `path` | yes | — | File or directory path |
| `mode` | yes | — | Permissions (octal, e.g. "755", "600") |
| `recursive` | no | "false" | Apply recursively (chmod -R) |

## Usage

```go
result := node.Run(fs.NewChangeMode().SetArgs(map[string]string{
    fs.ArgPath:  "/home/sinevia/.ssh",
    fs.ArgMode:  "700",
})).FirstResult()
```

## Execution Flow

1. Validates `path` and `mode` are non-empty
2. Validates `mode` is a valid octal string (3 or 4 digits)
3. Checks current mode with `stat -c '%a' <path>`
4. If already correct, returns `Changed=false`
5. Runs `chmod [-R] <mode> <path>`
6. Returns `Changed=true` if mode was changed

## Idempotency

- `Check()` compares current mode against desired mode
- `Check()` returns `false` if already correct
- `Check()` returns `true` if mode mismatch
