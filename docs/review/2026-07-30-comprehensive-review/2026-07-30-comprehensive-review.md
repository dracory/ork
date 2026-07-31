# Code Review: `ork`

**Date:** 2026-07-30
**Scope:** Full repository review (161 Go files across `ork` core, `ssh`, `vault`, `types`, `cmd/ork`, and `skills/*`)
**Tooling:** `go build ./...`, `go vet ./...`, and `go test ./...` all pass cleanly.

## Summary

The codebase is well-documented and disciplined: fluent APIs, consistent skill boilerplate, Argon2id + AES-256-GCM vault crypto, and careful shell/SQL escaping in most skills. Test coverage is broad (every skill has a `_test.go`). The issues below are ordered by severity with concrete file/line references.

---

## Critical

### 1. ~~Shared, mutable skill instances make concurrent `Inventory`/`Group` execution unsafe~~ ✅ Fixed

**Status:** Fixed. `Run()`, `RunByID()`, and `Check()` in `node_implementation.go` now clone the skill via `cloneFromMap()` (ToMap/FromMap round-trip) before mutating any state. Each goroutine gets its own isolated clone — the original shared instance is never mutated. `BaseSkill`/`BasePlaybook` state is stored in `omni.Atom` (thread-safe) with `NodeConfig` protected by `sync.RWMutex`.

`Inventory.Run()` / `RunByID()` / `Check()` spin up one goroutine per node (`inventory_implementation.go:170-223`) and call `skill.SetNodeConfig(...)`, `skill.SetArgs(...)`, `skill.Run()` on **the same `RunnableInterface` instance** for every node when:

- The caller passes a single skill instance to `inventory.Run(skill)` (the *recommended*, non-deprecated API), or
- `RunByID(id)` is used — it fetches the skill from the **global singleton registry** (`node_implementation.go:508-545`, `registry.FindByID(id)`), which returns the same pointer every time.

`types.BaseSkill` (`types/base_skill.go:34-114`) has **no synchronization** — `nodeCfg`, `args` (a plain `map[string]string`), and `dryRun` are mutated directly. When `SetMaxConcurrency` is set above 1 (an explicitly documented, supported feature), multiple goroutines will:

- Race on `b.nodeCfg` / `b.args`, causing one node's command to silently execute against another node's host/args, and
- Potentially trigger Go's **fatal, non-recoverable** "concurrent map writes" / "concurrent map read and map write" runtime error via `SetArg` / `GetArg` — this crashes the whole process and is **not** caught by the `recover()` in the goroutine wrapper (that only catches regular panics, not the fatal runtime error class).

This affects the primary, documented usage pattern, not an edge case.

**Fix options:**
- Have the registry / `Run` / `RunByID` clone a fresh skill instance per node (e.g., add a `Clone()` / factory method to `RunnableInterface`, or store constructor funcs in the registry instead of instances), or
- Make `BaseSkill` immutable per invocation (e.g., pass config/args as a parameter to `Run` / `Check` rather than storing them as mutable fields).

*(Default `Inventory` concurrency is 1, so this is latent unless a caller opts into concurrency — but the API explicitly advertises `SetMaxConcurrency` for this purpose, so it's a real trap.)*

## Low / Style

- ~~`RunnerInterface.RunByID` is marked `// Deprecated: Use Run() instead` (`runner_interface.go:23-25`), but `Run()` has the identical shared-instance hazard described in finding #1 — the deprecation doesn't actually steer users away from the underlying problem.~~ ✅ Fixed — `RunByID` deprecation comment updated to note both methods are concurrency-safe via cloning; the deprecation is now purely about API simplicity.

---

## Positives worth calling out

- `vault/crypto.go`: solid design — Argon2id KDF, random salt+nonce per encryption, AES-256-GCM, magic-number + length validation before decrypt, base64+line-wrapping for portability.
- `skills/mariadb/functions.go` and `skills/apt/install.go` / `skills/user/*`: careful, well-documented layered escaping (SQL string vs. identifier vs. shell single/double quote) — a good pattern that should be the template applied to the ufw/swap gaps above.
- `inventoryImplementation`'s concurrency plumbing itself (semaphore + `sync.WaitGroup` + `recover()` + mutex-protected results map) is well structured — the one hazard is the shared skill instance, not the orchestration code.
- Consistent fluent-API boilerplate and extensive package/function-level GoDoc across all skills; test coverage is broad (every skill has a `_test.go`) and `go test ./...` is green.
