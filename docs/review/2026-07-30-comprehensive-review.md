# Code Review: `ork`

**Date:** 2026-07-30
**Scope:** Full repository review (161 Go files across `ork` core, `ssh`, `vault`, `types`, `cmd/ork`, and `skills/*`)
**Tooling:** `go build ./...`, `go vet ./...`, and `go test ./...` all pass cleanly.

## Summary

The codebase is well-documented and disciplined: fluent APIs, consistent skill boilerplate, Argon2id + AES-256-GCM vault crypto, and careful shell/SQL escaping in most skills. Test coverage is broad (every skill has a `_test.go`). The issues below are ordered by severity with concrete file/line references.

---

## Critical

### 1. Shared, mutable skill instances make concurrent `Inventory`/`Group` execution unsafe — ☐ OPEN

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

---

## High

### 2. Nine fully-implemented skills are unreachable via the registry / `RunByID` — ☐ OPEN

`skills/constants.go` defines IDs and every skill has full docs/tests, but `NewDefaultRegistry()` (`registry.go:68-108`) never registers:

- `user.NewUserAddToGroup` (`user-add-to-group`)
- `mariadb.NewPurge` (referenced by `skills/mariadb/purge.go`)
- `ufw.NewAllow`, `ufw.NewDeny`, `ufw.NewDelete`, `ufw.NewEnable`, `ufw.NewDisable`, `ufw.NewReset`, `ufw.NewDefault`

Anyone calling `node.RunByID("ufw-allow")` (exactly as instructed in that skill's own doc comment) gets `"skill 'ufw-allow' not found in registry"`. This is a straightforward oversight — add these to the `skills` slice in `NewDefaultRegistry`.

### 3. Inconsistent shell-argument escaping — real injection vectors — ✅ FIXED

Most skills are careful (see `skills.ShellEscapeArg`, `apt.shellEscapePackages`, the MariaDB SQL/shell double-escaping helpers in `skills/mariadb/functions.go`), but a few aren't:

- **`skills/ufw/allow.go:81`** and **`skills/ufw/deny.go:72`**: the optional `comment` arg is interpolated unescaped: `cmdStr += fmt.Sprintf(" comment '%s'", comment)`. A value like `x'; rm -rf / #` breaks out of the quotes and injects arbitrary shell commands.
- **`skills/swap/create.go`** and **`skills/swap/delete.go`**: `swapFilePath` (arg `swapfile-path`, user-settable, default `/swapfile`) is used raw, unescaped, in multiple commands — `dd if=/dev/zero of=%s ...`, `chmod 600 %s`, `mkswap %s`, inside a single-quoted `grep` pattern, and in `echo '%s none swap sw 0 0' | tee -a /etc/fstab` (`create.go:169-173`). A path containing a single quote or shell metacharacters injects commands into `/etc/fstab` and beyond.

Given the codebase already has `skills.ShellEscapeArg` for exactly this purpose, these two spots should use it (and reject/validate the port and swap path more strictly).

### 4. Widespread stale documentation: `--playbook=<id> --arg=key=value` CLI does not exist — ✅ FIXED

Every skill file's doc comment (~40 occurrences, e.g. `skills/user/add_to_group.go:48-51`) documents usage as `go run . --playbook=user-add-to-group --arg=username=...`, but `cmd/ork/main.go` only implements `vault` and `help` subcommands — there is no `--playbook` / `--arg` flag parsing anywhere in the repo (`grep` for `"playbook"`, `ArgPlaybook`, `flag.Parse` in skills/cmd returns nothing beyond these doc comments). This will mislead every user who reads GoDoc and tries to run a skill from the CLI as documented. Either implement this CLI runner or fix the docs to reflect the actual (Go API) usage.

---

## Medium

### 5. `groupImplementation` has inconsistent locking — ✅ FIXED

`groupImplementation` has a `sync.RWMutex` (`g.mu`) but it only guards `dryRunMode` / `becomeUser`. `AddNode` / `GetNodes` mutate/read `g.nodes` (and `SetArg` / `GetArg` / `GetArgs` touch `g.args`) with **no lock at all** (`group_implementation.go:37-73`). If a group is mutated (`AddNode`) concurrently with reads (e.g. from an `Inventory` that has this group and is running nodes concurrently, or from another goroutine), this is a data race. Either the mutex should also protect `nodes` / `args`, or document that groups must be fully built before use across goroutines (and this should be enforced/asserted, since `inventoryImplementation.GetNodes()` calls `group.GetNodes()` concurrently with the app potentially still calling `AddNode`).

### 6. `Group` runs sequentially while `Inventory` runs concurrently — inconsistent semantics — ✅ FIXED (documented)

`groupImplementation.RunCommand` / `Run` / `RunByID` / `Check` all loop over nodes sequentially (`group_implementation.go:92-104`), whereas `inventoryImplementation`'s equivalents spawn a goroutine per node with a semaphore (`inventory_implementation.go:112-167`). A user building a `Group` and calling `.Run()` directly gets no concurrency at all, while putting the same nodes into an `Inventory` does. This asymmetry isn't documented and is a likely source of surprise/performance bugs; consider giving `Group` the same concurrency model (which would also make findings #1 and #5 more urgent) or explicitly documenting that only `Inventory` parallelizes.

### 7. `CommandInterface.WithTimeout(timeout interface{})` breaks type safety — ✅ FIXED

`command_implementation.go:217-223` takes `interface{}` and silently no-ops if the argument isn't a `time.Duration`:

```go
func (c *commandImplementation) WithTimeout(timeout interface{}) CommandInterface {
	if td, ok := timeout.(time.Duration); ok {
		c.BaseSkill.SetTimeout(td)
	}
	return c
}
```

This is the only `interface{}`-typed setter in an otherwise strongly-typed fluent API and silently swallows caller mistakes (e.g. passing an `int` seconds value) instead of failing to compile or erroring. It should just take `time.Duration` like every other `SetTimeout` / `WithTimeout` in the codebase.

---

## Low / Style

- ✅ FIXED: `ssh/functions.go`'s `Run()` now escapes `becomeUser` / `chdir` with `ShellEscapeArg` (`ssh/functions.go:140-145`).
- ☐ OPEN: `RunnerInterface.RunByID` is marked `// Deprecated: Use Run() instead` (`runner_interface.go:23-25`), but `Run()` has the identical shared-instance hazard described in finding #1 — the deprecation doesn't actually steer users away from the underlying problem.

---

## Positives worth calling out

- `vault/crypto.go`: solid design — Argon2id KDF, random salt+nonce per encryption, AES-256-GCM, magic-number + length validation before decrypt, base64+line-wrapping for portability.
- `skills/mariadb/functions.go` and `skills/apt/install.go` / `skills/user/*`: careful, well-documented layered escaping (SQL string vs. identifier vs. shell single/double quote) — a good pattern that should be the template applied to the ufw/swap gaps above.
- `inventoryImplementation`'s concurrency plumbing itself (semaphore + `sync.WaitGroup` + `recover()` + mutex-protected results map) is well structured — the one hazard is the shared skill instance, not the orchestration code.
- Consistent fluent-API boilerplate and extensive package/function-level GoDoc across all skills; test coverage is broad (every skill has a `_test.go`) and `go test ./...` is green.
