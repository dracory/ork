# Implementation Plan: Serialize/Unserialize Per Call (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Non-breaking change (internal only)
**Scope:** `types/base_skill.go`, `command_implementation.go`, `node_implementation.go`, tests, docs
**Compared to:** `plan-deepcopy.md` (deep-copy via `brunoga/deep`), `plan-runnable-options-interface.md` (opts-param, major break)

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

**Serialize the skill to JSON bytes, unserialize into a fresh instance, then
mutate and run the fresh instance.** The original shared instance is never
touched by the framework.

```go
// Before (racy):
skill.SetNodeConfig(n.cfg)
skill.SetDryRun(n.cfg.IsDryRunMode)
result := skill.Run()

// After (safe):
data, err := json.Marshal(skill)
clone := newCloneOfSameType(skill) // via reflect, see §4.3
json.Unmarshal(data, clone)
clone.SetNodeConfig(n.cfg)
clone.SetDryRun(n.cfg.IsDryRunMode)
result := clone.Run()
```

This is a **purely internal change**. No public interface changes, no skill
signature changes, no skill body changes.

---

## 3. The Critical Blocker: Unexported Fields

### The problem

`BaseSkill` has **all-unexported fields**:

```go
type BaseSkill struct {
    BaseBecome
    id          string         // unexported
    description string         // unexported
    nodeCfg     NodeConfig     // unexported
    args        map[string]string  // unexported
    dryRun      bool           // unexported
    timeout     time.Duration  // unexported
}
```

`BaseBecome` also has an unexported field:

```go
type BaseBecome struct {
    becomeUser string  // unexported
}
```

Both `encoding/json` and `encoding/gob` **skip unexported fields**. Without
custom serialization:

```go
skill := types.NewBaseSkill().SetID("test").SetArgs(map[string]string{"k": "v"})
data, _ := json.Marshal(skill)
fmt.Println(string(data))
// Output: {}    ←  EMPTY. All state lost.
```

This is a **hard blocker**. The approach is dead unless we add custom
serialization.

### The workaround: custom `MarshalJSON`/`UnmarshalJSON` on `BaseSkill`

Add custom JSON marshaling to `BaseSkill` that uses the existing getters/setters
to serialize/unserialize the unexported fields through an exported intermediate
struct:

```go
// baseSkillJSON is the exported proxy used for JSON serialization.
// It mirrors BaseSkill's fields with JSON tags.
type baseSkillJSON struct {
    ID          string            `json:"id"`
    Description string            `json:"description"`
    NodeConfig  NodeConfig        `json:"nodeConfig"`
    Args        map[string]string `json:"args"`
    DryRun      bool              `json:"dryRun"`
    Timeout     time.Duration     `json:"timeout"`
    BecomeUser  string            `json:"becomeUser"`
}

// MarshalJSON implements json.Marshaler.
func (b *BaseSkill) MarshalJSON() ([]byte, error) {
    return json.Marshal(baseSkillJSON{
        ID:          b.id,
        Description: b.description,
        NodeConfig:  b.nodeCfg,
        Args:        b.args,
        DryRun:      b.dryRun,
        Timeout:     b.timeout,
        BecomeUser:  b.GetBecomeUser(),
    })
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *BaseSkill) UnmarshalJSON(data []byte) error {
    var proxy baseSkillJSON
    if err := json.Unmarshal(data, &proxy); err != nil {
        return err
    }
    b.id = proxy.ID
    b.description = proxy.Description
    b.nodeCfg = proxy.NodeConfig
    b.args = proxy.Args
    b.dryRun = proxy.DryRun
    b.timeout = proxy.Timeout
    b.SetBecomeUser(proxy.BecomeUser)
    return nil
}
```

This is **~30 lines in one place** (`types/base_skill.go`). Since every skill
struct embeds `*BaseSkill` and adds **no extra fields** (verified across all 40
skills), this single custom marshaler covers all skills automatically via Go's
field promotion.

### The `commandImplementation` complication

`commandImplementation` embeds `*BaseSkill` but adds 3 extra unexported fields:

```go
type commandImplementation struct {
    *types.BaseSkill
    command  string  // unexported — NOT captured by BaseSkill.MarshalJSON
    required bool    // unexported
    chdir    string  // unexported
}
```

These fields are lost during serialization because `BaseSkill.MarshalJSON`
only serializes `BaseSkill`'s fields, not the embedding struct's extra fields.
`commandImplementation` needs its own custom marshaler (~15 lines):

```go
type commandJSON struct {
    *baseSkillJSON  // embed the proxy (but we can't access it directly)
    Command  string `json:"command"`
    Required bool   `json:"required"`
    Chdir    string `json:"chdir"`
}

func (c *commandImplementation) MarshalJSON() ([]byte, error) {
    return json.Marshal(commandJSON{
        Command:  c.command,
        Required: c.required,
        Chdir:    c.chdir,
        // BaseSkill fields: need to serialize separately and merge...
    })
}
```

**Problem:** Go's `MarshalJSON` promotion is shadowed — if `commandImplementation`
defines `MarshalJSON`, it **replaces** `BaseSkill.MarshalJSON`, not extends it.
So `commandImplementation.MarshalJSON` must serialize **both** the BaseSkill
fields and its own fields. This means either:

1. Duplicating the BaseSkill field serialization in `commandJSON` (fragile —
   if BaseSkill gains a field, `commandJSON` must be updated), or
2. Two-step serialization: marshal BaseSkill to JSON, marshal command fields to
   JSON, merge the two JSON objects (ugly, error-prone), or
3. Having `commandJSON` embed `baseSkillJSON` and populate it from the BaseSkill
   getters:

```go
type commandJSON struct {
    baseSkillJSON
    Command  string `json:"command"`
    Required bool   `json:"required"`
    Chdir    string `json:"chdir"`
}

func (c *commandImplementation) MarshalJSON() ([]byte, error) {
    bs := c.BaseSkill
    return json.Marshal(commandJSON{
        baseSkillJSON: baseSkillJSON{
            ID:          bs.id,
            Description: bs.description,
            NodeConfig:  bs.nodeCfg,
            Args:        bs.args,
            DryRun:      bs.dryRun,
            Timeout:     bs.timeout,
            BecomeUser:  bs.GetBecomeUser(),
        },
        Command:  c.command,
        Required: c.required,
        Chdir:    c.chdir,
    })
}
```

This works but **couples `commandImplementation` to `BaseSkill`'s field list** —
if `BaseSkill` gains a field, `commandJSON` must be updated or the field is
silently lost during clone. This is a maintenance hazard.

**Mitigation:** Add a method `BaseSkill.exportState() baseSkillJSON` and
`BaseSkill.importState(baseSkillJSON)` so `commandImplementation` (and any
future type with extra fields) can delegate:

```go
// In base_skill.go:
func (b *BaseSkill) exportState() baseSkillJSON {
    return baseSkillJSON{
        ID: b.id, Description: b.description, NodeConfig: b.nodeCfg,
        Args: b.args, DryRun: b.dryRun, Timeout: b.timeout,
        BecomeUser: b.GetBecomeUser(),
    }
}
func (b *BaseSkill) importState(s baseSkillJSON) {
    b.id = s.ID; b.description = s.Description; b.nodeCfg = s.NodeConfig
    b.args = s.Args; b.dryRun = s.DryRun; b.timeout = s.Timeout
    b.SetBecomeUser(s.BecomeUser)
}

// In command_implementation.go:
func (c *commandImplementation) MarshalJSON() ([]byte, error) {
    return json.Marshal(commandJSON{
        baseSkillJSON: c.BaseSkill.exportState(),
        Command:  c.command,
        Required: c.required,
        Chdir:    c.chdir,
    })
}
func (c *commandImplementation) UnmarshalJSON(data []byte) error {
    var proxy commandJSON
    if err := json.Unmarshal(data, &proxy); err != nil {
        return err
    }
    c.BaseSkill.importState(proxy.baseSkillJSON)
    c.command = proxy.Command
    c.required = proxy.Required
    c.chdir = proxy.Chdir
    return nil
}
```

This reduces the coupling: `commandImplementation` calls `exportState`/
`importState` and doesn't need to know the individual BaseSkill fields. If
BaseSkill gains a field, only `exportState`/`importState` need updating.

**Total custom serialization code: ~50 lines** (`baseSkillJSON` struct +
`MarshalJSON`/`UnmarshalJSON` on BaseSkill + `exportState`/`importState` +
`commandJSON` struct + `MarshalJSON`/`UnmarshalJSON` on commandImplementation).

---

## 4. The Interface-Typing Problem

### The problem

`json.Unmarshal` needs a concrete target type. But `nodeImplementation.Run`
receives `skill types.RunnableInterface` — an interface. You can't do:

```go
var clone types.RunnableInterface
json.Unmarshal(data, &clone)  // ERROR: cannot unmarshal into interface
```

### The solution: `reflect` to discover the concrete type

```go
// cloneSkill serializes skill to JSON and deserializes into a fresh
// instance of the same concrete type. No factory registry needed.
func cloneSkill(skill types.RunnableInterface) (types.RunnableInterface, error) {
    data, err := json.Marshal(skill)
    if err != nil {
        return nil, fmt.Errorf("cloneSkill: marshal failed: %w", err)
    }

    // reflect.TypeOf(skill) returns the concrete type (e.g., *user.UserCreate)
    // .Elem() dereferences the pointer to get the struct type (e.g., user.UserCreate)
    // reflect.New() creates a new zero-value instance of that struct type
    // .Interface() converts back to an any, which we assert to RunnableInterface
    typ := reflect.TypeOf(skill).Elem()
    clonePtr := reflect.New(typ)
    clone := clonePtr.Interface().(types.RunnableInterface)

    if err := json.Unmarshal(data, clone); err != nil {
        return nil, fmt.Errorf("cloneSkill: unmarshal failed: %w", err)
    }
    return clone, nil
}
```

This uses `reflect` but **not `unsafe`** — a key difference from the `deep-copy`
plan. `reflect.TypeOf` and `reflect.New` are safe, standard operations.

---

## 5. What Changes

### 5.1 `types/base_skill.go`

- Add `baseSkillJSON` struct (exported-field proxy with JSON tags)
- Add `MarshalJSON()` / `UnmarshalJSON()` methods on `*BaseSkill`
- Add `exportState()` / `importState()` helper methods (for use by types with
  extra fields, like `commandImplementation`)
- Add `import "encoding/json"` to imports

~35 lines added. No existing code changed.

### 5.2 `command_implementation.go`

- Add `commandJSON` struct (embeds `baseSkillJSON` + 3 extra fields)
- Add `MarshalJSON()` / `UnmarshalJSON()` methods on `*commandImplementation`
- Add `import "encoding/json"` to imports

~20 lines added. No existing code changed.

### 5.3 `node_implementation.go`

- Add `import "encoding/json"` and `import "reflect"`
- Add `cloneSkill()` helper function (§4)
- Rewrite `Run`, `RunByID`, `Check` to clone before mutating (same pattern as
  `plan-deepcopy.md` but using `cloneSkill` instead of `deep.Copy`)

~30 lines changed across 3 methods. Same structure as the deep-copy plan.

### 5.4 What does NOT change

| Component | Changes? |
|-----------|----------|
| `types/runnable_interface.go` | No |
| `types/become_interface.go` | No |
| `types/registry.go` | No |
| `runner_interface.go` | No |
| `node_interface.go` / `inventory_interface.go` | No |
| `inventory_implementation.go` | No |
| `group_implementation.go` | No |
| `registry.go` (ork pkg) | No |
| `skill.go` | No |
| All `skills/**/*.go` (40 files) | No |
| All `skills/**/*_test.go` | No |
| All other tests | No (signatures unchanged) |
| `go.mod` | No new dependency (uses stdlib `encoding/json` + `reflect`) |

**Total files changed: 3** (`types/base_skill.go`, `command_implementation.go`,
`node_implementation.go`) + tests + docs. **No external dependency.**

---

## 6. JSON vs Gob

| Aspect | JSON | Gob |
|--------|------|-----|
| Unexported fields | Skipped → needs custom marshaler | Skipped → needs custom marshaler |
| `time.Duration` | Marshals as int64 nanoseconds (round-trips correctly) | Native int64 support |
| Type registration | Not needed (uses `reflect` for concrete type) | Requires `gob.Register` for each of 40 skill types |
| Debuggability | Human-readable (can log/inspect serialized form) | Binary (opaque) |
| Performance | ~50-200µs per marshal+unmarshal | ~20-80µs per marshal+unmarshal |
| External deps | None (stdlib) | None (stdlib) |
| Maps | Handles `map[string]string` natively | Handles `map[string]string` natively |

**Decision: JSON.** Gob's performance advantage is irrelevant (both are
negligible vs SSH). Gob's 40 `gob.Register` calls are unnecessary boilerplate.
JSON's debuggability is a bonus — you can log the serialized skill state for
auditing. JSON uses `reflect` for type discovery (same as the clone approach),
avoiding the registration requirement.

---

## 7. Concurrency Safety Analysis

Identical to `plan-deepcopy.md` — the safety comes from "clone before mutate",
not from the cloning mechanism:

### `Inventory.Run(skill)` — goroutine per node

```
goroutine 1: n1.Run(skill) → cloneSkill(skill) → clone1.SetNodeConfig(n1.cfg) → clone1.Run()
goroutine 2: n2.Run(skill) → cloneSkill(skill) → clone2.SetNodeConfig(n2.cfg) → clone2.Run()
```

- `json.Marshal(skill)` reads the shared `skill` — concurrent reads are safe.
- Each goroutine mutates its own clone — no shared mutable state.
- The original `skill` is never mutated by the framework.

### `Inventory.RunByID(id)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → cloneSkill(skill) → clone1.Run()
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → cloneSkill(skill) → clone2.Run()
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `json.Marshal(skill)` reads the shared singleton — safe (no writes to it).
- Each goroutine mutates its own clone.

### `Check` — same as `Run`, also fixes the secondary bug

Now clones + sets full config (`NodeConfig`, `BecomeUser`, `DryRun`), closing
the gap where `Check()` previously only set `DryRun`.

---

## 8. Performance Analysis

JSON marshal + unmarshal per clone:

- `BaseSkill` has 7 serializable fields (2 strings, 1 struct, 1 map, 1 bool,
  1 Duration, 1 string from BaseBecome)
- Map: `args` typically has 0-5 entries
- `NodeConfig` has ~15 fields (all exported, all strings/bools/slices)

Estimated cost: **~50-200µs per clone** (JSON reflection + allocation).

For context:
- One SSH round-trip: **100-500 ms**
- Clone cost as fraction of SSH: **0.04%**

Even at `SetMaxConcurrency(100)` across 1000 nodes (1000 clones), total clone
cost is ~200ms vs ~100s of SSH I/O. Negligible.

**~5-10x slower than `deep.Copy`** (~10µs), but both are negligible vs SSH.

---

## 9. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | `BaseSkill` gains a new field → `baseSkillJSON` / `exportState` / `importState` must be updated or the field is **silently lost** during clone | **High** | This is the biggest risk. Mitigation: add a unit test that marshals + unmarshals a `BaseSkill` with all fields set and asserts all fields round-trip. If a field is added without updating the proxy, the test fails. |
| R2 | Third-party skill with extra fields (beyond `*BaseSkill`) loses those fields during clone | **High** | `BaseSkill.MarshalJSON` is promoted to the embedding type, so it serializes only BaseSkill fields. A third-party skill with extra fields must implement its own `MarshalJSON`/`UnmarshalJSON`. Document this requirement in `docs/skills.md`. This is **more error-prone than deep-copy** (which handles extra fields automatically). |
| R3 | `reflect.TypeOf(skill).Elem()` panics if `skill` is not a pointer | Low | All skills are pointer types (`*UserCreate`, etc.). Add a guard: if `typ.Kind() != reflect.Ptr`, return an error instead of panicking. |
| R4 | `time.Duration` serializes as nanoseconds (large int) — less readable in logs | Low | Acceptable. If readability matters, add a custom `MarshalJSON` on a `Duration` wrapper type. Not needed for correctness. |
| R5 | `NodeConfig.Args` map is shared by reference after unmarshal (JSON creates a new map, so this is actually safe) | None | JSON unmarshal creates a new map — no aliasing. |
| R6 | Non-serializable fields (channels, funcs, `sync.Mutex`) cause `json.Marshal` to error | Low | Same constraint as deep-copy. `json.Marshal` returns an error (not a panic) — handled gracefully. |
| R7 | `commandImplementation` custom marshaler gets out of sync with `BaseSkill` fields | Medium | Mitigated by `exportState`/`importState` delegation (§3). The command marshaler calls these helpers and doesn't list BaseSkill fields individually. |

---

## 10. Phased Implementation Plan

### Phase 0 — Spike test (validate approach before committing)

- [ ] Write a throwaway test: create a `*UserCreate`, set args + nodeConfig,
      `json.Marshal` it, `reflect.New` + `json.Unmarshal` into a clone, verify
      the clone has the same args but is a different pointer.
- [ ] Verify `reflect.TypeOf(skill).Elem()` + `reflect.New` works when `skill`
      is passed as `RunnableInterface` (interface value).
- [ ] Verify that mutating the clone's args doesn't affect the original.
- [ ] If the spike fails, stop and fall back to `plan-deepcopy.md`.

### Phase 1 — Custom serialization on `BaseSkill`

- [ ] Add `baseSkillJSON` struct to `types/base_skill.go`.
- [ ] Add `MarshalJSON` / `UnmarshalJSON` on `*BaseSkill`.
- [ ] Add `exportState` / `importState` helpers.
- [ ] Add a round-trip test: set all fields, marshal, unmarshal, assert all
      fields survive. **This is the guard against R1.**
- [ ] `go build ./...`, `go test ./types/...`

### Phase 2 — Custom serialization on `commandImplementation`

- [ ] Add `commandJSON` struct to `command_implementation.go`.
- [ ] Add `MarshalJSON` / `UnmarshalJSON` on `*commandImplementation` using
      `exportState` / `importState` delegation.
- [ ] Add a round-trip test for `commandImplementation`.
- [ ] `go build ./...`, `go test ./...`

### Phase 3 — Implement the clone in `node_implementation.go`

- [ ] Add `cloneSkill()` helper function (§4).
- [ ] Rewrite `nodeImplementation.Run` to clone before mutating.
- [ ] Rewrite `nodeImplementation.RunByID` to clone before mutating.
- [ ] Rewrite `nodeImplementation.Check` to clone + set full config (fixes
      secondary bug).
- [ ] `go build ./...`

### Phase 4 — Tests

- [ ] Add a **concurrency regression test**: N nodes, `SetMaxConcurrency(N)`,
      distinct per-node args, assert no cross-node arg leakage. Run with
      `-race`.
- [ ] Add a test verifying `Check()` now propagates `NodeConfig` and
      `BecomeUser` (verifies the secondary bug fix).
- [ ] Add a test verifying `cloneSkill` failure (marshal/unmarshal error) is
      handled gracefully (return error result, not panic).
- [ ] Run `go test -race ./...` — must be green.

### Phase 5 — Docs

- [ ] Add a note in `docs/skills.md`: "The framework clones each skill instance
      per call via JSON serialization. Skill structs must be JSON-serializable.
      If your skill has extra fields beyond `*BaseSkill`, you must implement
      `MarshalJSON`/`UnmarshalJSON` using `BaseSkill.exportState()`/
      `importState()`."
- [ ] Update `docs/review/2026-07-30-comprehensive-review.md` or add a
      follow-up note marking Critical #1 as resolved.

### Phase 6 — Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...` (the critical check)
- [ ] Manual: run an example inventory with `SetMaxConcurrency > 1` and
      per-node args; confirm no leakage.

---

## 11. Comparison: All Three Plans

| Aspect | This plan (serialize) | `plan-deepcopy.md` | `plan-runnable-options-interface.md` |
|--------|----------------------|-------------------|-------------------------------------|
| **API break** | None | None | Major (every skill signature changes) |
| **Files changed** | ~5 (base_skill, command, node_impl, tests, docs) | ~3 (go.mod, node_impl, tests) | ~80+ (types, all skills, all tests, all docs) |
| **External dependency** | None (stdlib `encoding/json` + `reflect`) | `github.com/brunoga/deep` (uses `unsafe`) | None |
| **Custom boilerplate** | ~50 lines (MarshalJSON on BaseSkill + commandImpl) | 0 lines | 0 lines (but 80+ files rewritten) |
| **Handles extra skill fields automatically** | **No** — skill author must add custom MarshalJSON | **Yes** — deep.Copy handles any field | N/A (no fields to copy) |
| **Risk: new BaseSkill field forgotten** | **High** — silent data loss if proxy not updated | None — automatic | None |
| **Risk: third-party skill with extra fields** | **High** — fields silently lost | Low — deep.Copy handles them | N/A |
| **Concurrency safety** | By construction (clone per call) | By construction (clone per call) | By construction (fresh opts per call) |
| **Check() bug fix** | Yes (side effect) | Yes (side effect) | Yes (side effect) |
| **Performance** | ~50-200µs per clone (JSON) | ~10µs per clone (reflection) | Zero |
| **`unsafe` usage** | No | Yes (in `deep` library) | No |
| **Serialized form reusable** | **Yes** — can log/audit/cache skill state | No | No |
| **ctx/cancellation future-proofing** | No | No | Yes |
| **Migration effort** | Hours | Hours | Days |

---

## 12. The Honest Assessment

This plan works but has **two significant disadvantages** vs `plan-deepcopy.md`:

1. **Silent data loss risk (R1, R2):** If `BaseSkill` gains a field and the
   `baseSkillJSON` proxy isn't updated, that field is **silently lost** during
   every clone. `deep.Copy` has no such risk — it copies all fields
   automatically. Similarly, a third-party skill with extra fields must
   implement custom `MarshalJSON` or those fields are silently lost. `deep.Copy`
   handles extra fields with zero skill-author action.

2. **More boilerplate:** ~50 lines of custom serialization code vs zero for
   `deep.Copy`. The `exportState`/`importState` pattern reduces but doesn't
   eliminate the coupling between `BaseSkill` and its JSON proxy.

**The one unique advantage:** the serialized JSON form is reusable for other
purposes — logging skill state before execution, auditing, caching, network
transmission to remote workers, debugging. Neither `deep.Copy` nor the
opts-param approach produces a serializable form. If you envision future
features like "log the exact skill config that ran on each node" or "send skill
definitions to remote workers," this plan's serialization infrastructure is a
foundation for those features.

**Recommendation:** If the only goal is fixing the concurrency bug,
`plan-deepcopy.md` is simpler and safer (no silent-data-loss risk). If you
want the serialized form for future logging/auditing/caching features, this
plan is the better foundation — but accept the R1/R2 maintenance burden and
add the round-trip guard test early.

---

## 13. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now sets `NodeConfig` and `BecomeUser` (secondary
  bug closed — verified by a dedicated test).
- No public API changes — all existing tests pass without modification.
- **BaseSkill round-trip test passes** — all fields survive marshal + unmarshal
  (guards against R1).
- `docs/skills.md` documents the JSON-serialization contract for skill authors
  (guards against R2).

---

## 14. Fallback

If Phase 0 (spike test) reveals that `reflect.TypeOf(skill).Elem()` +
`reflect.New` doesn't work for interface-typed skill values, or if the
custom `MarshalJSON` approach proves too fragile during Phase 1, fall back to
`plan-deepcopy.md` (which uses `deep.Copy` and handles unexported fields
automatically via `unsafe`, with zero custom serialization code).
