# Implementation Plan: Deep-Copy Per Call (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Non-breaking change (internal only)
**Scope:** `node_implementation.go`, `go.mod`, tests, docs
**Library:** [`github.com/brunoga/deep`](https://pkg.go.dev/github.com/brunoga/deep)
**Compared to:** `plan.md` (opts-param approach — major breaking change)

---

## 1. Problem Statement (recap)

`Inventory.Run/RunByID/Check` and `Node.Run/RunByID/Check` mutate a **shared**
`RunnableInterface` instance once per node goroutine via `SetNodeConfig`,
`SetArgs`, `SetDryRun`, `SetBecomeUser`. `types.BaseSkill` stores these as
unsynchronized fields (incl. a plain `map[string]string`), so under
`SetMaxConcurrency > 1`:

- one node's args/config can leak into another node's execution (logic corruption), and
- Go's **fatal, non-recoverable** "concurrent map read/write" runtime panic can
  crash the whole process (not caught by the goroutine `recover()`).

Secondary bug: `nodeImplementation.Check()` never calls
`SetNodeConfig`/`SetBecomeUser` before `skill.Run()`, unlike `Run`/`RunByID`.

---

## 2. Approach

**Clone the skill before mutating it.** Each call to `nodeImplementation.Run`,
`RunByID`, or `Check` deep-copies the skill instance, then mutates and runs the
**clone**. The original shared instance is never touched by the framework.

```go
// Before (racy):
skill.SetNodeConfig(n.cfg)
skill.SetDryRun(n.cfg.IsDryRunMode)
result := skill.Run()

// After (safe):
cloned, err := deep.Copy(skill)
if err != nil { /* return error result */ }
cloned.SetNodeConfig(n.cfg)
cloned.SetDryRun(n.cfg.IsDryRunMode)
result := cloned.Run()
```

This is a **purely internal change**. No public interface changes, no skill
signature changes, no skill body changes, no test signature changes.

---

## 3. Why `github.com/brunoga/deep`

| Requirement | `deep` support |
|-------------|----------------|
| Deep-copy structs with pointer fields (`*BaseSkill`) | Yes — recursively copies through pointers |
| Copy unexported fields (`BaseSkill.id`, `.args`, etc.) | Yes — uses `unsafe` to read/write unexported fields (confirmed in README) |
| Copy maps (`args map[string]string`) | Yes — creates new map with copied entries |
| Generic API (`deep.Copy[T](src) (T, error)`) | Yes — returns typed result |
| Custom copier for non-copyable types (future) | Yes — `Copier[T]` interface for opt-in custom behavior |
| Cycle detection | Yes |
| License | Apache-2.0 |
| Go module stability | v1.3.1, tagged, stable |

### Why not a typed `Clone()` method per skill?

A typed `Clone()` on `BaseSkill` + each concrete skill would avoid reflection,
but requires ~5 lines of boilerplate per skill × 40 skills = ~200 lines of
mechanical code, plus every third-party skill author must remember to implement
it. `deep.Copy` is zero-boilerplate and works automatically.

### Why not `plan.md` (opts-param approach)?

The opts-param plan is a **major breaking change**: every skill signature,
every skill body, every test, every doc changes. This plan achieves the same
safety with **zero API breakage**. The tradeoff is a reflection+unsafe
dependency and no `ctx` future-proofing (see §7).

---

## 4. What Changes

### 4.1 `go.mod`

Add dependency:

```sh
go get github.com/brunoga/deep
```

### 4.2 `node_implementation.go` (the only code change)

Three methods change: `Run`, `RunByID`, `Check`. Each gains a `deep.Copy`
call before mutating the skill. `RunCommand` is unchanged (it doesn't touch
skills).

#### `Run`

```go
func (n *nodeImplementation) Run(skill types.RunnableInterface) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	// Deep-copy the skill so concurrent goroutines each get an isolated instance.
	// The original shared instance (from caller or registry) is never mutated.
	cloned, err := deep.Copy(skill)
	if err != nil {
		results.Results[n.GetHost()] = types.Result{
			Changed: false,
			Message: fmt.Sprintf("failed to clone skill: %v", err),
			Error:   fmt.Errorf("failed to clone skill: %w", err),
		}
		return results
	}

	cloned.SetNodeConfig(n.cfg)
	cloned.SetDryRun(n.cfg.IsDryRunMode)
	if cloned.GetBecomeUser() == "" {
		cloned.SetBecomeUser(n.cfg.BecomeUser)
	}
	result := cloned.Run()

	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}
```

#### `RunByID`

```go
func (n *nodeImplementation) RunByID(id string, opts ...types.RunnableOptions) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	registry, err := GetGlobalSkillRegistry()
	if err != nil { /* unchanged error handling */ }

	skill, ok := registry.FindByID(id)
	if !ok { /* unchanged error handling */ }

	// Clone the registry singleton — each concurrent caller gets its own copy.
	cloned, err := deep.Copy(skill)
	if err != nil { /* return error result */ }

	cloned.SetNodeConfig(n.cfg)
	cloned.SetDryRun(n.cfg.IsDryRunMode)
	if cloned.GetBecomeUser() == "" {
		cloned.SetBecomeUser(n.cfg.BecomeUser)
	}
	if len(opts) > 0 {
		cloned.SetArgs(opts[0].Args)
		cloned.SetDryRun(opts[0].DryRun)
		cloned.SetTimeout(opts[0].Timeout)
	}

	result := cloned.Run()
	results.Results[n.GetHost()] = types.Result{ /* unchanged */ }
	return results
}
```

#### `Check` (also fixes the secondary bug — now sets NodeConfig/BecomeUser)

```go
func (n *nodeImplementation) Check(skill types.RunnableInterface) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	cloned, err := deep.Copy(skill)
	if err != nil { /* return error result */ }

	// FIX: previously Check() only set DryRun, missing NodeConfig and BecomeUser.
	// Now it sets the same config as Run(), closing the gap.
	cloned.SetNodeConfig(n.cfg)
	cloned.SetDryRun(n.cfg.IsDryRunMode)
	if cloned.GetBecomeUser() == "" {
		cloned.SetBecomeUser(n.cfg.BecomeUser)
	}

	result := cloned.Run()
	results.Results[n.GetHost()] = types.Result{ /* unchanged */ }
	return results
}
```

### 4.3 What does NOT change

| Component | Changes? |
|-----------|----------|
| `types/runnable_interface.go` | No |
| `types/base_skill.go` | No |
| `types/become_interface.go` | No |
| `types/registry.go` | No |
| `runner_interface.go` | No |
| `node_interface.go` / `inventory_interface.go` | No |
| `inventory_implementation.go` | No (goroutines already call `n.Run(skill)` which now clones) |
| `group_implementation.go` | No (sequential; calls `n.Run(skill)` which now clones) |
| `command_implementation.go` | No |
| `registry.go` (ork pkg) | No |
| `skill.go` | No |
| All `skills/**/*.go` (40 files) | No |
| All `skills/**/*_test.go` | No |
| `skill_test.go` / `node_implementation_test.go` / etc. | No (signatures unchanged) |
| `docs/skills.md` / `docs/quick_start.md` / etc. | Minimal (add a note about clone-per-call) |

**Total files changed: 2** (`go.mod`, `node_implementation.go`) + tests + docs.

---

## 5. Concurrency Safety Analysis

### `Inventory.Run(skill)` — goroutine per node

```
goroutine 1: n1.Run(skill) → deep.Copy(skill) → clone1.SetNodeConfig(n1.cfg) → clone1.Run()
goroutine 2: n2.Run(skill) → deep.Copy(skill) → clone2.SetNodeConfig(n2.cfg) → clone2.Run()
```

- `deep.Copy(skill)` reads the shared `skill` — concurrent reads are safe (no writes).
- Each goroutine mutates its own clone — no shared mutable state.
- The original `skill` is never mutated by the framework.

### `Inventory.RunByID(id)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → deep.Copy(skill) → clone1.Run()
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → deep.Copy(skill) → clone2.Run()
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `deep.Copy(skill)` reads the shared singleton — safe (no writes to it).
- Each goroutine mutates its own clone.

### `Group.Run` — sequential

Each `node.Run(skill)` call clones, mutates, and runs the clone. Sequential, so
even without cloning it would be safe — but cloning makes it future-proof if
Group ever becomes concurrent.

### `Check` — same as `Run`

Now clones + sets full config (fixes the secondary bug).

---

## 6. Performance Analysis

`deep.Copy` uses reflection + `unsafe`. Cost per clone:

- `BaseSkill` has 6 fields (2 strings, 1 struct, 1 map, 1 bool, 1 Duration)
- Concrete skill adds 0-3 fields (strings/bools)
- Map copy: `args` typically has 0-5 entries

Estimated cost: **~5-20 µs per clone** (reflection overhead on a small struct).

For context:
- One SSH round-trip: **100-500 ms**
- Clone cost as fraction of SSH: **0.004%**

Even at `SetMaxConcurrency(100)` across 1000 nodes (1000 clones), total clone
cost is ~20ms vs ~100s of SSH I/O. Negligible.

---

## 7. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | `deep` uses `unsafe` for unexported fields — could break on Go runtime changes | Low | Library is maintained (v1.3.1, 29 importers); `unsafe` usage for unexported fields is a stable pattern; pin version in `go.mod` |
| R2 | Third-party skill with non-copyable fields (channels, funcs, `sync.Mutex`, `*ssh.Client`) causes `deep.Copy` to fail | Medium | `deep.Copy` returns an error (not a panic) — handled gracefully, returns failed Result. Skills with such fields can implement `deep.Copier[T]` for custom behavior. Document this in `docs/skills.md`. |
| R3 | `deep.Copy` on an interface value (`RunnableInterface`) — does it copy the concrete type? | Medium | Library README says it handles interfaces. **Phase 0 verifies this with a spike test before committing to the approach.** If it doesn't work, fall back to `plan.md` (opts-param). |
| R4 | No `ctx context.Context` future-proofing — adding cancellation later still requires a signature break | Low | Accepted tradeoff. This plan prioritizes zero-break now over future-proofing. Cancellation can be added later as a separate breaking change if needed. |
| R5 | Organization policy prohibits `unsafe` in dependencies | Low | Check policy. If blocked, fall back to typed `Clone()` per skill (no `unsafe`, no external dep, but ~200 lines of boilerplate). |
| R6 | `deep` library is overkill — it's a full diff/patch/sync library, we only need `Copy` | Low | We only import `deep.Copy`. The rest of the library is unused. Go's module system trims unused code from binary via dead-code elimination. |

---

## 8. Phased Implementation Plan

### Phase 0 — Spike test (validate approach before committing)

- [ ] `go get github.com/brunoga/deep`
- [ ] Write a throwaway test: create a `*UserCreate`, set args, `deep.Copy` it,
      verify the clone has the same args but is a different pointer, and that
      mutating the clone's args doesn't affect the original.
- [ ] Verify `deep.Copy` works when the skill is passed as `RunnableInterface`
      (interface value, not concrete type) — this is risk R3.
- [ ] If the spike fails, stop and fall back to `plan.md`.

### Phase 1 — Implement the fix

- [ ] Add `github.com/brunoga/deep` to `go.mod`.
- [ ] Add `import "github.com/brunoga/deep"` to `node_implementation.go`.
- [ ] Rewrite `nodeImplementation.Run` (§4.2).
- [ ] Rewrite `nodeImplementation.RunByID` (§4.2).
- [ ] Rewrite `nodeImplementation.Check` (§4.2 — also fixes secondary bug).
- [ ] `go build ./...`

### Phase 2 — Tests

- [ ] Add a **concurrency regression test**: N nodes, `SetMaxConcurrency(N)`,
      distinct per-node args, assert no cross-node arg leakage. Run with
      `-race`. This is the test that would have caught the original bug.
- [ ] Add a test verifying `Check()` now propagates `NodeConfig` and
      `BecomeUser` (verifies the secondary bug fix).
- [ ] Add a test verifying `deep.Copy` failure is handled gracefully (return
      error result, not panic).
- [ ] Run `go test -race ./...` — must be green.

### Phase 3 — Docs

- [ ] Add a note in `docs/skills.md`: "The framework deep-copies each skill
      instance per call. Skill structs must be deep-copyable (no channels,
      funcs, or sync primitives as fields). If your skill has non-copyable
      fields, implement `deep.Copier[T]`."
- [ ] Update `docs/review/2026-07-30-comprehensive-review.md` or add a
      follow-up note marking Critical #1 as resolved.

### Phase 4 — Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...` (the critical check)
- [ ] Manual: run an example inventory with `SetMaxConcurrency > 1` and
      per-node args; confirm no leakage.

---

## 9. Comparison: This Plan vs `plan.md` (opts-param)

| Aspect | This plan (deep-copy) | `plan.md` (opts-param) |
|--------|----------------------|----------------------|
| **API break** | None | Major (every skill signature changes) |
| **Files changed** | ~3 (go.mod, node_impl, tests) | ~80+ (types, all skills, all tests, all docs) |
| **Skill author impact** | None | Must rewrite every skill's Run/Check |
| **Third-party skill impact** | None (automatic) | Must rewrite (breaking) |
| **Concurrency safety** | By construction (clone per call) | By construction (fresh opts per call) |
| **Check() bug fix** | Yes (side effect) | Yes (side effect) |
| **ctx/cancellation future-proofing** | No (would need signature break later) | Yes (ctx in signature now) |
| **Performance cost** | ~10µs reflection per call (negligible vs SSH) | Zero |
| **External dependency** | `github.com/brunoga/deep` (unsafe) | None |
| **Risk: unexported fields** | Handled by `deep` via `unsafe` | N/A |
| **Risk: non-copyable skill fields** | `deep.Copy` error (graceful) | N/A |
| **Migration effort** | Hours | Days |

---

## 10. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now sets `NodeConfig` and `BecomeUser` (secondary
  bug closed — verified by a dedicated test).
- No public API changes — all existing tests pass without modification.
- `docs/skills.md` documents the deep-copy contract for skill authors.

---

## 11. Fallback

If Phase 0 (spike test) reveals that `deep.Copy` cannot handle interface-typed
values or unexported fields in practice, fall back to `plan.md` (opts-param
approach). The spike costs 15 minutes; the fallback is the other plan.
