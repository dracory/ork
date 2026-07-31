# Implementation Plan: Omni-Backed Runnable (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Non-breaking for skill bodies; `BaseSkill` internally restructured to use `omni.Atom`
**Scope:** `types/base_skill.go`, `types/runnable_interface.go`, `command_implementation.go`, `node_implementation.go`, `go.mod`, tests, docs
**Library:** [`github.com/dracory/omni`](https://github.com/dracory/omni) — composable, thread-safe, serializable atoms
**Compared to:** `plan-map-storage.md`, `plan-map-backed.md`, `plan-deepcopy.md`, `plan-serialize.md`

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

**Make `BaseSkill` store all state inside an `omni.Atom`.** The Atom provides:
- `map[string]string` properties (all state lives here)
- **Thread-safe by default** (`sync.RWMutex` built into every Atom)
- **Free serialization** — `ToMap()`, `ToJSON()`, `ToGob()` built-in
- **Free deserialization** — `NewAtomFromMap()`, `NewAtomFromJSON()`, `FromGob()` built-in
- **Free cloning** — serialize to map, create new Atom from map
- **Hierarchical** — children support (future: skill groups, nested configs)
- **Functional options** — `WithID`, `WithProperties`, `WithType`, `WithChildren`

The framework clones the skill via the Atom's serialization before mutating.
The original shared instance is never touched. All existing signatures stay
the same — `Run()`, `Check()`, `SetNodeConfig()`, `SetArgs()`, etc.

```go
// Framework per call — the ONLY change in node_implementation.go:
m := skill.ToMap()               // omni.Atom.ToMap() — free, built-in
m["properties"].(map[string]string)["nodeConfig"] = serializeNodeConfig(n.cfg)
m["properties"].(map[string]string)["dryRun"] = "true"
clone := cloneFromMap(skill, m)  // fresh instance, populated from map
result := clone.Run()            // run the clone, not the original
```

---

## 3. Why omni

`omni` is a battle-tested library that provides exactly what we need:

| Need | omni provides | Custom code needed |
|------|--------------|-------------------|
| Map-backed state storage | `Atom.properties map[string]string` | None |
| Thread-safe access | `sync.RWMutex` in every Atom | None |
| `ToMap()` | `Atom.ToMap() map[string]any` | None |
| `FromMap()` | `NewAtomFromMap(m) (AtomInterface, error)` | None |
| JSON serialization | `Atom.ToJSON() / NewAtomFromJSON()` | None |
| Gob serialization | `Atom.ToGob() / FromGob()` | None |
| Functional options | `WithID, WithProperties, WithType, WithChildren` | None |
| Children/hierarchy | `ChildAdd, ChildrenGet, ChildrenFindByType` | None |
| Property get/set | `Get(key), Set(key, value), Has(key), Remove(key)` | None |
| Bulk property ops | `GetAll(), SetAll(map)` | None |
| Memory profiling | `MemoryUsage()` | None |

**Zero custom serialization code.** Compare:
- `plan-map-storage.md`: ~15 lines for `ToMap`/`FromMap`
- `plan-map-backed.md`: ~80 lines for `ToMap`/`FromMap` with string conversion
- `plan-serialize.md`: ~50 lines for custom `MarshalJSON`
- **This plan: 0 lines** — omni provides everything

---

## 4. Core Design

### 4.1 `BaseSkill` — backed by `omni.Atom` (`types/base_skill.go`)

`BaseSkill` embeds an `omni.Atom` internally. All state lives in the Atom's
properties. Getters and setters delegate to the Atom. **Same signatures, same
return types** — skill bodies don't change.

```go
package types

import (
	"strconv"
	"time"

	"github.com/dracory/omni"
)

type BaseSkill struct {
	atom omni.AtomInterface  // ALL state lives here
}

func NewBaseSkill() *BaseSkill {
	return &BaseSkill{
		atom: omni.NewAtom("skill"),
	}
}

// === ID / Description (delegated to Atom) ===

func (b *BaseSkill) GetID() string {
	return b.atom.GetID()
}

func (b *BaseSkill) SetID(id string) RunnableInterface {
	b.atom.SetID(id)
	return b
}

func (b *BaseSkill) GetDescription() string {
	return b.atom.Get("description")
}

func (b *BaseSkill) SetDescription(d string) RunnableInterface {
	b.atom.Set("description", d)
	return b
}

// === NodeConfig (stored as serialized string in Atom properties) ===

func (b *BaseSkill) GetNodeConfig() NodeConfig {
	s := b.atom.Get("nodeConfig")
	if s == "" {
		return NodeConfig{}
	}
	return deserializeNodeConfig(s)
}

func (b *BaseSkill) SetNodeConfig(cfg NodeConfig) RunnableInterface {
	b.atom.Set("nodeConfig", serializeNodeConfig(cfg))
	return b
}

// === Args (stored as Atom properties with "arg_" prefix) ===

func (b *BaseSkill) GetArg(key string) string {
	return b.atom.Get("arg_" + key)
}

func (b *BaseSkill) SetArg(key, value string) RunnableInterface {
	b.atom.Set("arg_"+key, value)
	return b
}

func (b *BaseSkill) GetArgs() map[string]string {
	all := b.atom.GetAll()
	args := make(map[string]string)
	for k, v := range all {
		if len(k) > 4 && k[:4] == "arg_" {
			args[k[4:]] = v
		}
	}
	return args
}

func (b *BaseSkill) SetArgs(args map[string]string) RunnableInterface {
	for k, v := range args {
		b.atom.Set("arg_"+k, v)
	}
	return b
}

// === DryRun (stored as string "true"/"false") ===

func (b *BaseSkill) IsDryRun() bool {
	return b.atom.Get("dryRun") == "true"
}

func (b *BaseSkill) SetDryRun(dryRun bool) RunnableInterface {
	b.atom.Set("dryRun", strconv.FormatBool(dryRun))
	return b
}

// === BecomeUser (stored as string) ===

func (b *BaseSkill) GetBecomeUser() string {
	return b.atom.Get("becomeUser")
}

func (b *BaseSkill) SetBecomeUser(user string) RunnableInterface {
	b.atom.Set("becomeUser", user)
	return b
}

// === Timeout (stored as nanoseconds string) ===

func (b *BaseSkill) GetTimeout() time.Duration {
	s := b.atom.Get("timeout")
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return time.Duration(n)
}

func (b *BaseSkill) SetTimeout(timeout time.Duration) RunnableInterface {
	b.atom.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	return b
}
```

`BaseBecome` is no longer embedded — `becomeUser` is an Atom property.

### 4.2 `ToMap()` / `FromMap()` — free from omni

`BaseSkill` exposes `ToMap`/`FromMap` by delegating to the Atom:

```go
// ToMap returns the skill's state as a map (delegates to omni.Atom.ToMap)
func (b *BaseSkill) ToMap() map[string]any {
	return b.atom.ToMap()
}

// FromMap populates the skill's state from a map (delegates to omni.Atom)
func (b *BaseSkill) FromMap(m map[string]any) {
	atom, err := omni.NewAtomFromMap(m)
	if err != nil {
		// fallback: create empty atom
		b.atom = omni.NewAtom("skill")
		return
	}
	b.atom = atom
}
```

**~10 lines.** And the actual work is done by omni's tested implementation.

### 4.3 `NodeConfig` serialization helpers

Since `omni.Atom` stores `map[string]string` (strings only), `NodeConfig` (a
struct) must be serialized to a string. Two options:

**Option A: JSON string** (simplest):
```go
func serializeNodeConfig(cfg NodeConfig) string {
	data, _ := json.Marshal(cfg)
	return string(data)
}

func deserializeNodeConfig(s string) NodeConfig {
	var cfg NodeConfig
	json.Unmarshal([]byte(s), &cfg)
	return cfg
}
```

**Option B: Flatten to prefixed keys** (more omni-idiomatic):
```go
func serializeNodeConfig(cfg NodeConfig) string {
	// Store each NodeConfig field as a separate Atom property with "cfg_" prefix
	// This is more granular but requires SetAll/GetAll manipulation
}
```

**Decision: Option A (JSON string).** It's simpler, handles all NodeConfig
fields automatically (including future additions), and the JSON string is
still human-readable in the Atom's properties. The NodeConfig is one property
value among many — the rest of the state (args, dryRun, becomeUser, timeout)
is stored as individual Atom properties for direct access.

### 4.4 `RunnableInterface` — add `ToMap`/`FromMap`

```go
type RunnableInterface interface {
	// === EXISTING (unchanged signatures) ===
	BecomeInterface
	GetID() string
	SetID(id string) RunnableInterface
	GetDescription() string
	SetDescription(description string) RunnableInterface
	GetNodeConfig() NodeConfig
	SetNodeConfig(cfg NodeConfig) RunnableInterface
	GetArg(key string) string
	SetArg(key, value string) RunnableInterface
	GetArgs() map[string]string
	SetArgs(args map[string]string) RunnableInterface
	IsDryRun() bool
	SetDryRun(dryRun bool) RunnableInterface
	GetTimeout() time.Duration
	SetTimeout(timeout time.Duration) RunnableInterface
	Check() (bool, error)
	Run() Result

	// === NEW (serialization / cloning) ===
	ToMap() map[string]any
	FromMap(m map[string]any)
}
```

`Run()` and `Check()` signatures are **unchanged**. All setters are
**unchanged**. Only two methods are added.

### 4.5 `commandImplementation` — no override needed

`commandImplementation`'s extra fields (`command`, `required`, `chdir`) become
Atom properties. No `ToMap`/`FromMap` override needed — the Atom handles all
properties automatically:

```go
type commandImplementation struct {
	*types.BaseSkill
	// command, required, chdir now stored as Atom properties
}

func (c *commandImplementation) GetCommand() string {
	return c.atom.Get("command")
}

func (c *commandImplementation) SetCommand(cmd string) {
	c.atom.Set("command", cmd)
}

func (c *commandImplementation) GetRequired() bool {
	return c.atom.Get("required") == "true"
}

func (c *commandImplementation) SetRequired(r bool) {
	c.atom.Set("required", strconv.FormatBool(r))
}

func (c *commandImplementation) GetChdir() string {
	return c.atom.Get("chdir")
}

func (c *commandImplementation) SetChdir(d string) {
	c.atom.Set("chdir", d)
}
```

No `ToMap`/`FromMap` override — `BaseSkill.ToMap()` delegates to the Atom,
which includes `command`/`required`/`chdir` automatically.

### 4.6 How the framework clones — `node_implementation.go`

```go
func cloneFromMap(template types.RunnableInterface, m map[string]any) (types.RunnableInterface, error) {
	typ := reflect.TypeOf(template).Elem()
	clonePtr := reflect.New(typ)

	// Initialize the embedded *BaseSkill
	bsField := clonePtr.Elem().FieldByName("BaseSkill")
	if !bsField.IsValid() {
		return nil, fmt.Errorf("cloneFromMap: no embedded BaseSkill in %s", typ.Name())
	}
	bsField.Set(reflect.ValueOf(types.NewBaseSkill()))

	clone := clonePtr.Interface().(types.RunnableInterface)
	clone.FromMap(m)
	return clone, nil
}

func (n *nodeImplementation) Run(skill types.RunnableInterface) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	// 1. Get the skill's state as a map (via omni.Atom.ToMap)
	m := skill.ToMap()

	// 2. Replace execution-time config in the properties
	if props, ok := m["properties"].(map[string]string); ok {
		props["nodeConfig"] = serializeNodeConfig(n.cfg)
		props["dryRun"] = strconv.FormatBool(n.cfg.IsDryRunMode)
		if props["becomeUser"] == "" {
			props["becomeUser"] = n.cfg.BecomeUser
		}
		m["properties"] = props
	}

	// 3. Create a fresh instance from the modified map
	clone, err := cloneFromMap(skill, m)
	if err != nil {
		results.Results[n.GetHost()] = types.Result{
			Changed: false,
			Message: fmt.Sprintf("failed to clone skill: %v", err),
			Error:   err,
		}
		return results
	}

	// 4. Run the clone
	result := clone.Run()

	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}
```

`RunByID` and `Check` follow the same pattern. `Check` now also sets
`nodeConfig`/`becomeUser` (fixes the secondary bug).

### 4.7 Free serialization — all formats (bonus)

Because omni provides JSON, Gob, and Map serialization, you get all three for free:

```go
// JSON (for playbooks, audit logs, remote workers):
jsonStr, _ := skill.ToJSON()  // omni.Atom.ToJSON()
// {"id":"user-create","type":"skill","properties":{"arg_username":"alice","dryRun":"true"}}

// Gob (for binary storage, network transmission):
data, _ := skill.atom.ToGob()

// Map (for in-memory cloning, inspection):
m := skill.ToMap()
```

---

## 5. What Changes

### 5.1 Core (must change)

| File | Change |
|------|--------|
| `go.mod` | Add `github.com/dracory/omni` dependency |
| `types/base_skill.go` | Replace typed fields with `atom omni.AtomInterface`. Rewrite getters/setters to delegate to Atom (~80 lines, same signatures). Add `ToMap()`/`FromMap()` (~10 lines). Add `serializeNodeConfig`/`deserializeNodeConfig` helpers (~10 lines). Drop `BaseBecome` embedding. |
| `types/runnable_interface.go` | Add `ToMap() map[string]any` and `FromMap(m map[string]any)` to `RunnableInterface`. All existing methods unchanged. |
| `command_implementation.go` | Move `command`/`required`/`chdir` from typed fields to Atom properties. Rewrite their getters/setters. **No `ToMap`/`FromMap` override needed.** |
| `node_implementation.go` | Add `cloneFromMap()` helper. Rewrite `Run`/`RunByID`/`Check` to `ToMap → modify → cloneFromMap → Run`. |

### 5.2 What does NOT change

| Component | Changes? |
|-----------|----------|
| `types/become_interface.go` | No (kept for NodeInterface) |
| `types/registry.go` | No |
| `runner_interface.go` | No |
| `node_interface.go` / `inventory_interface.go` | No |
| `inventory_implementation.go` | No |
| `group_implementation.go` | No |
| `registry.go` (ork pkg) | No |
| `skill.go` | No |
| All `skills/**/*.go` (40 files) | **No** — skill bodies unchanged |
| All `skills/**/*_test.go` | **No** — signatures unchanged |
| All other tests | **No** — signatures unchanged |

**Total files changed: 5** (`go.mod`, `types/base_skill.go`,
`types/runnable_interface.go`, `command_implementation.go`,
`node_implementation.go`) + tests + docs.

### 5.3 Skill body — before and after (UNCHANGED)

```go
// BEFORE:
func (u *UserCreate) Run() types.Result {
    cfg := u.GetNodeConfig()
    username := u.GetArg("username")
    if cfg.IsDryRunMode { ... }
}

// AFTER (identical):
func (u *UserCreate) Run() types.Result {
    cfg := u.GetNodeConfig()        // now reads from Atom property
    username := u.GetArg("username") // now reads from Atom property
    if cfg.IsDryRunMode { ... }
}
```

---

## 6. Concurrency Safety Analysis

### `Inventory.Run(skill)` — goroutine per node

```
goroutine 1: n1.Run(skill) → m1 := skill.ToMap() → modify m1 → clone1 := cloneFromMap(skill, m1) → clone1.Run()
goroutine 2: n2.Run(skill) → m2 := skill.ToMap() → modify m2 → clone2 := cloneFromMap(skill, m2) → clone2.Run()
```

- `skill.ToMap()` reads the shared `skill` — omni's `ToMap()` acquires `RLock`,
  so concurrent reads are safe.
- Each goroutine gets its own map copy (`m1`, `m2`) — no shared map.
- Each goroutine creates its own clone — no shared instance.
- The original `skill` is never mutated.

### `Inventory.RunByID(id)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → m1 := skill.ToMap() → ... → clone1.Run()
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → m2 := skill.ToMap() → ... → clone2.Run()
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `skill.ToMap()` uses omni's RLock — safe for concurrent reads.
- Each goroutine gets its own map copy and clone.

### Bonus: omni's built-in thread safety

Even if a bug somehow allowed concurrent access to the skill's Atom, omni's
`sync.RWMutex` would prevent the fatal "concurrent map read/write" panic.
The mutex makes concurrent reads safe and serializes writes. This is a
**defense-in-depth** layer on top of the clone-before-mutate pattern.

### `Check` — same as `Run`, also fixes the secondary bug

Now sets `nodeConfig`/`becomeUser`/`dryRun` in the map before cloning (fixes
the gap where `Check()` previously only set `DryRun`).

---

## 7. Why This Is Better Than `plan-map-storage.md`

| Aspect | `plan-map-storage.md` (raw map) | **This plan (omni-backed)** |
|--------|-------------------------------|---------------------------|
| State storage | `map[string]any` (raw) | `omni.Atom` (managed, thread-safe) |
| Thread safety | None (raw map) | **`sync.RWMutex` built-in** |
| `ToMap`/`FromMap` | ~15 lines custom | **0 lines (omni provides)** |
| JSON serialization | `json.Marshal(map)` (manual) | **`atom.ToJSON()` (built-in)** |
| Gob serialization | Not available | **`atom.ToGob()` (built-in)** |
| Children/hierarchy | Not available | **Built-in (future: skill groups)** |
| Functional options | Not available | **`WithID, WithProperties, WithType`** |
| Memory profiling | Not available | **`atom.MemoryUsage()`** |
| Tested serialization | No (custom code) | **Yes (omni has tests + benchmarks)** |
| External dependency | None | `github.com/dracory/omni` (zero transitive deps) |
| Defense-in-depth | Clone-only | **Clone + mutex (double protection)** |

The omni-backed approach gives you everything `plan-map-storage.md` does, plus
thread safety, tested serialization, hierarchy support, and memory profiling —
for the cost of one dependency (which itself has zero dependencies).

---

## 8. Performance Analysis

`ToMap` + `cloneFromMap` per call:

- `omni.Atom.ToMap()`: RLock + map copy (~500ns for ~15 properties)
- `reflect.New`: ~500ns
- `reflect.FieldByName` + `Set`: ~200ns
- `omni.NewAtomFromMap()`: map iteration + Atom construction (~1µs)

Estimated cost: **~2-5µs per clone**.

omni benchmarks (from README):
```
BenchmarkAtom_ToJSON-8    2000000    750 ns/op
BenchmarkAtom_ToGob-8     1000000    1200 ns/op
```

For context:
- One SSH round-trip: **100-500 ms**
- Clone cost as fraction of SSH: **0.001%**

Comparable to `plan-map-storage.md` and faster than `plan-serialize.md`.

---

## 9. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | Adding `ToMap`/`FromMap` to interface breaks external implementors who don't embed `*BaseSkill` | Low | All built-in skills embed `*BaseSkill`. External skills following the documented pattern inherit automatically. |
| R2 | `omni.Atom` stores `map[string]string` — `NodeConfig` struct must be serialized to string | Low | JSON string serialization (~10 lines). Human-readable. Handles future NodeConfig fields automatically. |
| R3 | `reflect.FieldByName("BaseSkill")` fails if a skill doesn't name its embed `BaseSkill` | Low | All 40 built-in skills use `*types.BaseSkill`. Guard in `cloneFromMap` returns error. |
| R4 | `omni` library has a bug or breaking change | Low | omni is MIT/AGPL licensed, zero dependencies, small codebase. Can vendor if needed. |
| R5 | `omni.Atom`'s `sync.RWMutex` adds lock overhead to every getter/setter | Low | RLock is ~25ns. Getters are called ~10x per skill execution. Total: ~250ns — negligible vs SSH. |
| R6 | New `BaseSkill` field not stored as Atom property → lost during clone | Low | All state goes through getters/setters which write to the Atom. If a new field is added with a getter/setter, it's automatically in the Atom. If someone adds a raw struct field bypassing the Atom, that's a code review issue. |
| R7 | **No silent data loss risk** — all state is in the Atom, `ToMap`/`FromMap` are automatic | None | Key advantage. |

---

## 10. Phased Implementation Plan

### Phase 0 — Spike test (validate approach)

- [ ] Add `github.com/dracory/omni` to `go.mod`.
- [ ] Write a throwaway test: create a `*BaseSkill` backed by `omni.Atom`,
      set args + nodeConfig, call `ToMap()`, create a fresh instance via
      `cloneFromMap()`, verify the clone has the same state.
- [ ] Verify `omni.NewAtomFromMap()` correctly reconstructs the Atom.
- [ ] Verify `reflect.FieldByName("BaseSkill")` works for all skill types.
- [ ] If the spike fails, fall back to `plan-map-storage.md`.

### Phase 1 — Restructure `BaseSkill` to use `omni.Atom`

- [ ] Add `github.com/dracory/omni` dependency.
- [ ] Replace typed fields with `atom omni.AtomInterface` in `types/base_skill.go`.
- [ ] Rewrite all getters/setters to delegate to Atom (same signatures).
- [ ] Add `ToMap()` / `FromMap()` (delegating to Atom).
- [ ] Add `serializeNodeConfig` / `deserializeNodeConfig` helpers.
- [ ] Drop `BaseBecome` embedding; move `becomeUser` to Atom property.
- [ ] Add `ToMap`/`FromMap` to `RunnableInterface`.
- [ ] Add unit tests: verify all getters/setters, `ToMap`/`FromMap` round-trip.
- [ ] `go build ./...`, `go test ./types/...`

### Phase 2 — Restructure `commandImplementation`

- [ ] Move `command`/`required`/`chdir` to Atom properties.
- [ ] Rewrite their getters/setters.
- [ ] No `ToMap`/`FromMap` override needed.
- [ ] Add round-trip test for `commandImplementation`.
- [ ] `go build ./...`, `go test ./...`

### Phase 3 — `cloneFromMap` + framework integration

- [ ] Add `cloneFromMap()` helper to `node_implementation.go`.
- [ ] Rewrite `nodeImplementation.Run` to `ToMap → modify → cloneFromMap → Run`.
- [ ] Rewrite `nodeImplementation.RunByID` similarly.
- [ ] Rewrite `nodeImplementation.Check` to clone + set full config (fixes
      secondary bug).
- [ ] `go build ./...`

### Phase 4 — Tests

- [ ] Add a **concurrency regression test**: N nodes, `SetMaxConcurrency(N)`,
      distinct per-node args, assert no cross-node arg leakage. Run with
      `-race`.
- [ ] Add a test verifying `Check()` now propagates `NodeConfig` and
      `BecomeUser` (verifies the secondary bug fix).
- [ ] Add a test verifying `cloneFromMap` failure is handled gracefully.
- [ ] Add a test verifying the original skill is unchanged after clone.
- [ ] Add a test verifying JSON serialization round-trip (`ToJSON` →
      `NewAtomFromJSON` → `FromMap`).
- [ ] Run `go test -race ./...` — must be green.

### Phase 5 — Docs

- [ ] Update `docs/skills.md`: "BaseSkill is backed by `omni.Atom`. All state
      is stored as Atom properties. Skills are cloneable via `ToMap()`/
      `FromMap()`. The framework clones each skill per call for concurrency
      safety. Use the getters/setters — they delegate to the Atom internally."
- [ ] Document the `ToMap`/`FromMap` contract for skill authors.
- [ ] Document the JSON/Gob serialization bonus features.
- [ ] Update `docs/review/2026-07-30-comprehensive-review.md` marking
      Critical #1 as resolved.

### Phase 6 — Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] Manual: inventory with `SetMaxConcurrency > 1`, per-node args, no leakage.

---

## 11. Comparison: All Plans (revised)

| Aspect | This plan (omni) | `plan-map-storage.md` | `plan-map-backed.md` | `plan-deepcopy.md` | `plan-serialize.md` | `plan-runnable-options-interface.md` | `plan-functional-options.md` |
|--------|-----------------|---------------------|---------------------|-------------------|-------------------|-------------------------------------|---------------------|
| **Skill bodies change** | **No** | No | No | No | No | Yes | Yes |
| **Files changed** | **~5** | ~4 | ~4 | ~3 | ~5 | ~80+ | ~80+ |
| **State storage** | `omni.Atom` (managed) | `map[string]any` (raw) | typed fields | typed fields | typed fields | typed fields | `RunConfig` struct |
| **Thread-safe by default** | **Yes (RWMutex)** | No | No | No | No | N/A | N/A |
| **ToMap/FromMap code** | **0 lines (omni)** | ~15 lines | ~80 lines | 0 | ~50 lines | 0 | 0 |
| **JSON serialization** | **Built-in** | Free | Free | No | Custom | No | No |
| **Gob serialization** | **Built-in** | No | No | No | No | No | No |
| **Children/hierarchy** | **Built-in** | No | No | No | No | No | No |
| **Silent data loss risk** | **None** | None | Yes | None | Yes | None | None |
| **Per-type overrides** | **None** | None | Yes | None | Yes | N/A | N/A |
| **`unsafe` usage** | No | No | No | Yes | No | No | No |
| **External dependency** | `dracory/omni` (zero deps) | None | None | `brunoga/deep` | None | None | None |
| **Defense-in-depth** | **Clone + mutex** | Clone only | Clone only | Clone only | Clone only | By construction | By construction |
| **Performance** | ~2-5µs | ~2-5µs | ~2-5µs | ~10µs | ~50-200µs | ~0 | ~0 |
| **Concurrency safety** | By construction + mutex | By construction | By construction | By construction | By construction | By construction | By construction |
| **Check() bug fix** | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| **ctx future-proofing** | No | No | No | No | No | Yes | Yes |
| **Migration effort** | Hours | Hours | Hours | Hours | Hours | Days | Days |

---

## 12. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now sets `NodeConfig` and `BecomeUser` (secondary
  bug closed — verified by a dedicated test).
- All existing skill tests pass **without modification** (signatures unchanged).
- **`ToMap`/`FromMap` round-trip test passes** for `BaseSkill` and
  `commandImplementation`.
- **JSON serialization round-trip:** `atom.ToJSON()` → `omni.NewAtomFromJSON()`
  → `FromMap()` produces the same skill state.
- **Original skill unchanged after clone**: verify that `ToMap → modify →
  cloneFromMap` does not mutate the original skill.
- `docs/skills.md` documents the omni-backed model and `ToMap`/`FromMap`
  contract.

---

## 13. Fallback

If Phase 0 (spike test) reveals that `omni.Atom`'s `map[string]string` property
storage is too limiting (e.g. `NodeConfig` serialization is too fragile), or
that `omni.NewAtomFromMap()` doesn't reconstruct correctly, fall back to
`plan-map-storage.md` (which uses a raw `map[string]any` with no external
dependency, at the cost of writing ~15 lines of `ToMap`/`FromMap` manually).
