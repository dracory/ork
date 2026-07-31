# Implementation Plan: Map-Storage Runnable (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Non-breaking for skill bodies; internal restructuring of `BaseSkill`
**Scope:** `types/base_skill.go`, `types/runnable_interface.go`, `command_implementation.go`, `node_implementation.go`, tests, docs
**Inspiration:** [`dracory/ui`](https://github.com/dracory/ui) `Block` — state lives in a map, `ToMap()`/`FromMap()` are trivial
**Compared to:** `plan-map-backed.md` (Option A — map as projection), `plan-deepcopy.md`, `plan-serialize.md`

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

**Make `BaseSkill` store all state in a `map[string]any` (like `dracory/ui`
Block stores state in `parameters map[string]string`).** The map IS the
storage — not a projection. `ToMap()` returns a copy of the map. The framework
modifies the copy, creates a fresh instance from it via `FromMap()`, and runs
the fresh instance. The original shared skill is never mutated.

```go
// Framework per call — the ONLY change in node_implementation.go:
m := skill.ToMap()               // get a copy of the skill's state as a map
m["nodeConfig"] = n.cfg          // replace what you want
m["dryRun"] = n.cfg.IsDryRunMode // replace what you want
clone := cloneFromMap(skill, m)  // fresh instance of same concrete type, populated from map
result := clone.Run()            // run the clone, not the original
```

**Why `map[string]any` and not `map[string]string`:** you can store typed
values directly — `NodeConfig` structs, `bool`, `time.Duration`,
`map[string]string` — without string conversion. Type assertions recover the
original types. No `strconv`, no string parsing, no `time.Duration` ↔ nanoseconds
conversion. JSON serialization is still free (`json.Marshal(map[string]any)`
works natively).

---

## 3. Core Design

### 3.1 `BaseSkill` — map as storage (`types/base_skill.go`)

All typed fields are replaced by a single `data map[string]any`. Getters and
setters read/write the map. **Same signatures, same return types** — skill
bodies don't change.

```go
type BaseSkill struct {
	id          string
	description string
	data        map[string]any  // ← ALL execution-time state lives here
}

func NewBaseSkill() *BaseSkill {
	return &BaseSkill{
		data: make(map[string]any),
	}
}
```

Getters read from the map via type assertion (comma-ok form, no panic):

```go
func (b *BaseSkill) GetNodeConfig() NodeConfig {
	if v, ok := b.data["nodeConfig"].(NodeConfig); ok {
		return v
	}
	return NodeConfig{}
}

func (b *BaseSkill) SetNodeConfig(cfg NodeConfig) RunnableInterface {
	b.data["nodeConfig"] = cfg
	return b
}

func (b *BaseSkill) GetArg(key string) string {
	if args, ok := b.data["args"].(map[string]string); ok {
		return args[key]
	}
	return ""
}

func (b *BaseSkill) SetArg(key, value string) RunnableInterface {
	args, ok := b.data["args"].(map[string]string)
	if !ok || args == nil {
		args = make(map[string]string)
	}
	args[key] = value
	b.data["args"] = args
	return b
}

func (b *BaseSkill) GetArgs() map[string]string {
	if args, ok := b.data["args"].(map[string]string); ok {
		return args
	}
	return nil
}

func (b *BaseSkill) SetArgs(args map[string]string) RunnableInterface {
	b.data["args"] = args
	return b
}

func (b *BaseSkill) IsDryRun() bool {
	if v, ok := b.data["dryRun"].(bool); ok {
		return v
	}
	return false
}

func (b *BaseSkill) SetDryRun(dryRun bool) RunnableInterface {
	b.data["dryRun"] = dryRun
	return b
}

func (b *BaseSkill) GetTimeout() time.Duration {
	if v, ok := b.data["timeout"].(time.Duration); ok {
		return v
	}
	return 0
}

func (b *BaseSkill) SetTimeout(timeout time.Duration) RunnableInterface {
	b.data["timeout"] = timeout
	return b
}
```

`BaseBecome` is no longer embedded — `becomeUser` is a key in the data map:

```go
func (b *BaseSkill) GetBecomeUser() string {
	if v, ok := b.data["becomeUser"].(string); ok {
		return v
	}
	return ""
}

func (b *BaseSkill) SetBecomeUser(user string) RunnableInterface {
	b.data["becomeUser"] = user
	return b
}
```

`id` and `description` stay as typed fields (construction-time, immutable after
registration — not part of the execution-time data map).

### 3.2 `ToMap()` / `FromMap()` — trivial (the key advantage)

Because the map IS the storage, `ToMap`/`FromMap` are just map copies. No field
enumeration, no type conversion, no proxy struct:

```go
// ToMap returns a shallow copy of the skill's state as a map[string]any.
// Includes id and description (construction-time state).
// Callers may modify the returned map freely — it's a copy.
// Note: nested reference types (maps, slices) are shared, not deep-copied.
// Callers should replace top-level values, not mutate nested ones.
func (b *BaseSkill) ToMap() map[string]any {
	m := make(map[string]any, len(b.data)+2)
	for k, v := range b.data {
		m[k] = v
	}
	m["id"] = b.id
	m["description"] = b.description
	return m
}

// FromMap populates this BaseSkill's state from a map.
// Replaces all existing data. Does not create a new instance.
func (b *BaseSkill) FromMap(m map[string]any) {
	b.data = make(map[string]any, len(m))
	for k, v := range m {
		b.data[k] = v
	}
	if v, ok := m["id"].(string); ok {
		b.id = v
	}
	if v, ok := m["description"].(string); ok {
		b.description = v
	}
}
```

**~15 lines total.** Compare to `plan-map-backed.md` (Option A) which needed
~40 lines of field-by-field conversion, or `plan-serialize.md` which needed
~50 lines of custom `MarshalJSON`.

### 3.3 Why no per-type overrides needed

Because ALL state lives in the `data` map — including `commandImplementation`'s
extra fields — `ToMap`/`FromMap` on `BaseSkill` handle everything automatically.
A new field added to any skill is just a new key in the map. No override, no
proxy update, no silent data loss.

`commandImplementation` changes from typed fields to map keys:

```go
// BEFORE:
type commandImplementation struct {
	*types.BaseSkill
	command  string
	required bool
	chdir    string
}

// AFTER:
type commandImplementation struct {
	*types.BaseSkill
	// command, required, chdir now stored in BaseSkill.data map
}

func (c *commandImplementation) GetCommand() string {
	if v, ok := c.data["command"].(string); ok {
		return v
	}
	return ""
}

func (c *commandImplementation) SetCommand(cmd string) {
	c.data["command"] = cmd
}

// Similarly for required, chdir — just read/write the map.
```

No `ToMap`/`FromMap` override needed on `commandImplementation` — the
`BaseSkill` implementation handles all keys automatically.

### 3.4 `RunnableInterface` — add `ToMap`/`FromMap`

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

### 3.5 How the framework clones — `node_implementation.go`

```go
// cloneFromMap creates a fresh instance of the same concrete type as 'template',
// populated from the map. Uses reflect to create the concrete type.
func cloneFromMap(template types.RunnableInterface, m map[string]any) (types.RunnableInterface, error) {
	typ := reflect.TypeOf(template).Elem()
	clonePtr := reflect.New(typ)

	// Initialize the embedded *BaseSkill (reflect.New creates it as nil)
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

	// 1. Get the skill's state as a map (copy)
	m := skill.ToMap()

	// 2. Replace execution-time config
	m["nodeConfig"] = n.cfg
	m["dryRun"] = n.cfg.IsDryRunMode
	if m["becomeUser"] == nil || m["becomeUser"] == "" {
		m["becomeUser"] = n.cfg.BecomeUser
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

### 3.6 Free serialization (bonus)

```go
// Log the exact skill state (audit trail):
data, _ := json.Marshal(skill.ToMap())
logger.Info("skill execution", "state", string(data))

// Store in playbook files:
// {"id": "user-create", "args": {"username": "alice"}, "dryRun": true}

// Send to remote worker:
data, _ := json.Marshal(skill.ToMap())
// ... send over network ...
var m map[string]any
json.Unmarshal(data, &m)
clone, _ := cloneFromMap(skill, m)
clone.Run()
```

No custom `MarshalJSON`. `map[string]any` is natively JSON-serializable.

---

## 4. What Changes

### 4.1 Core (must change)

| File | Change |
|------|--------|
| `types/base_skill.go` | Replace typed fields with `data map[string]any`. Rewrite getters/setters to read/write the map (~80 lines, same signatures). Add `ToMap()`/`FromMap()` (~15 lines). Drop `BaseBecome` embedding. |
| `types/runnable_interface.go` | Add `ToMap() map[string]any` and `FromMap(m map[string]any)` to `RunnableInterface`. All existing methods unchanged. |
| `command_implementation.go` | Move `command`/`required`/`chdir` from typed fields to `data` map keys. Rewrite their getters/setters to read/write the map. **No `ToMap`/`FromMap` override needed.** |
| `node_implementation.go` | Add `cloneFromMap()` helper. Rewrite `Run`/`RunByID`/`Check` to `ToMap → modify → cloneFromMap → Run`. |

### 4.2 What does NOT change

| Component | Changes? |
|-----------|----------|
| `types/become_interface.go` | No (kept for NodeInterface) |
| `types/registry.go` | No |
| `runner_interface.go` | No |
| `node_interface.go` / `inventory_interface.go` | No |
| `inventory_implementation.go` | No (calls `n.Run(skill)` which now clones) |
| `group_implementation.go` | No |
| `registry.go` (ork pkg) | No |
| `skill.go` | No |
| All `skills/**/*.go` (40 files) | **No** — skill bodies unchanged. They call `u.GetNodeConfig()`, `u.GetArg()`, etc. — same signatures, now reading from the map internally. |
| All `skills/**/*_test.go` | **No** — signatures unchanged |
| All other tests | **No** — signatures unchanged |
| `go.mod` | No new dependency (uses stdlib `reflect`) |

**Total files changed: 4** + tests + docs.

### 4.3 Skill body — before and after (UNCHANGED)

```go
// BEFORE:
func (u *UserCreate) Run() types.Result {
    cfg := u.GetNodeConfig()
    username := u.GetArg("username")
    if cfg.IsDryRunMode { ... }
}

// AFTER (identical):
func (u *UserCreate) Run() types.Result {
    cfg := u.GetNodeConfig()       // now reads from data["nodeConfig"]
    username := u.GetArg("username") // now reads from data["args"]["username"]
    if cfg.IsDryRunMode { ... }
}
```

The skill body is **byte-for-byte identical**. The getter signatures are the
same. The return types are the same. The only difference is internal: getters
read from a map instead of typed fields.

---

## 5. Concurrency Safety Analysis

### `Inventory.Run(skill)` — goroutine per node

```
goroutine 1: n1.Run(skill) → m1 := skill.ToMap() → m1["nodeConfig"] = n1.cfg → clone1 := cloneFromMap(skill, m1) → clone1.Run()
goroutine 2: n2.Run(skill) → m2 := skill.ToMap() → m2["nodeConfig"] = n2.cfg → clone2 := cloneFromMap(skill, m2) → clone2.Run()
```

- `skill.ToMap()` reads the shared `skill` — concurrent reads of the `data` map
  are safe (no writes to the skill).
- Each goroutine gets its own map copy (`m1`, `m2`) — no shared map.
- Each goroutine creates its own clone — no shared instance.
- The original `skill` is never mutated.

### `Inventory.RunByID(id)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → m1 := skill.ToMap() → ... → clone1.Run()
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → m2 := skill.ToMap() → ... → clone2.Run()
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `skill.ToMap()` reads the shared singleton — safe (no writes to it).
- Each goroutine gets its own map copy and clone.

### `Check` — same as `Run`, also fixes the secondary bug

Now sets `nodeConfig`/`becomeUser`/`dryRun` in the map before cloning (fixes
the gap where `Check()` previously only set `DryRun`).

---

## 6. Why This Is Better Than Option A (`plan-map-backed.md`)

| Aspect | Option A (map as projection) | **Option B (map as storage — this plan)** |
|--------|---------------------------|------------------------------------------|
| `BaseSkill` fields | Typed fields (`nodeCfg NodeConfig`, `dryRun bool`, etc.) | Single `data map[string]any` |
| `ToMap()` | ~40 lines (field-by-field conversion to strings) | **~10 lines (map copy)** |
| `FromMap()` | ~40 lines (string parsing back to typed fields) | **~10 lines (map copy)** |
| Type conversion | `strconv.FormatBool` / `ParseBool`, etc. | **None — store typed values directly** |
| New field added to BaseSkill | Must update `ToMap`/`FromMap` or field is **silently lost** | **Automatic — it's just a new map key** |
| `commandImplementation` extra fields | Must override `ToMap`/`FromMap` (~10 lines) | **No override — fields are in the map** |
| Silent data loss risk | **Yes** (R2 — field not in ToMap) | **None** — all state is in the map |
| `unsafe` usage | No | No |
| JSON serialization | `json.Marshal(map[string]string)` — strings only | `json.Marshal(map[string]any)` — typed values |
| Skill body changes | No | No |

Option B eliminates the silent-data-loss risk entirely. Adding a field to
`BaseSkill` (or any skill) is just adding a new key to the map — `ToMap`/
`FromMap` automatically include it because they copy the entire map. No field
enumeration, no proxy to keep in sync.

---

## 7. Performance Analysis

`ToMap` + `cloneFromMap` per call:

- `ToMap`: map allocation + iteration (~500ns for ~10 entries)
- `reflect.New`: ~500ns
- `reflect.FieldByName` + `Set`: ~200ns
- `FromMap`: map allocation + iteration (~500ns)

Estimated cost: **~2-5µs per clone**.

For context:
- One SSH round-trip: **100-500 ms**
- Clone cost as fraction of SSH: **0.001%**

Comparable to `plan-map-backed.md` (Option A) and faster than
`plan-serialize.md` (~50-200µs JSON).

---

## 8. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | Adding `ToMap`/`FromMap` to interface breaks external implementors who don't embed `*BaseSkill` | Low | All built-in skills embed `*BaseSkill`. External skills following the documented pattern inherit automatically. |
| R2 | Type assertion fails (e.g. `data["nodeConfig"].(NodeConfig)` when value is wrong type) | Low | All getters use comma-ok form (`if v, ok := ...; ok`). Wrong type returns zero value, doesn't panic. |
| R3 | `reflect.FieldByName("BaseSkill")` fails if a skill doesn't name its embed `BaseSkill` | Low | All 40 built-in skills use `*types.BaseSkill` (anonymous embed → field name is `BaseSkill`). Guard in `cloneFromMap` returns error. |
| R4 | Shallow copy in `ToMap` — nested maps (e.g. `args`) are shared references | Low | Framework replaces top-level values (`m["args"] = newArgs`), doesn't mutate nested maps. Document the contract. If needed, deep-copy nested maps in `ToMap`. |
| R5 | `map[string]any` type assertions are not compile-time checked | Low | Getters encapsulate the assertions — skill bodies call `u.GetNodeConfig()` (returns typed `NodeConfig`), not `u.data["nodeConfig"].(NodeConfig)`. The assertion is in one place (the getter), not scattered. |
| R6 | `BaseSkill` internal restructuring breaks code that directly accesses fields (not via getters) | Low | All built-in code uses getters/setters. `commandImplementation` fields move to the map but are accessed via methods. Audit for direct field access in Phase 0. |
| R7 | **No silent data loss risk** — new fields are automatically in the map | None | This is the key advantage over Option A and `plan-serialize.md`. |

---

## 9. Phased Implementation Plan

### Phase 0 — Spike test (validate approach)

- [ ] Write a throwaway test: create a `*UserCreate`, set args + nodeConfig,
      call `ToMap()`, modify the map, call `cloneFromMap()`, verify the clone
      has the modified state and the original is unchanged.
- [ ] Verify `reflect.FieldByName("BaseSkill")` works for all skill types.
- [ ] Verify type assertions work for `NodeConfig`, `bool`, `time.Duration`,
      `map[string]string`.
- [ ] If the spike fails, fall back to `plan-deepcopy.md`.

### Phase 1 — Restructure `BaseSkill` to map-as-storage

- [ ] Replace typed fields with `data map[string]any` in `types/base_skill.go`.
- [ ] Rewrite all getters/setters to read/write the map (same signatures).
- [ ] Add `ToMap()` / `FromMap()`.
- [ ] Drop `BaseBecome` embedding; move `becomeUser` to the data map.
- [ ] Add `ToMap`/`FromMap` to `RunnableInterface` in
      `types/runnable_interface.go`.
- [ ] Add unit tests: verify all getters/setters work, verify `ToMap`/`FromMap`
      round-trip.
- [ ] `go build ./...`, `go test ./types/...`

### Phase 2 — Restructure `commandImplementation`

- [ ] Move `command`/`required`/`chdir` from typed fields to `data` map keys.
- [ ] Rewrite their getters/setters to read/write the map.
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
- [ ] Add a test verifying the original skill is unchanged after
      `ToMap → modify → cloneFromMap`.
- [ ] Run `go test -race ./...` — must be green.

### Phase 5 — Docs

- [ ] Update `docs/skills.md`: "BaseSkill stores all state in a
      `map[string]any`. Skills are cloneable via `ToMap()`/`FromMap()`. The
      framework clones each skill per call for concurrency safety. Use the
      getters/setters — they read/write the map internally."
- [ ] Document the `ToMap`/`FromMap` contract for skill authors.
- [ ] Update `docs/review/2026-07-30-comprehensive-review.md` marking
      Critical #1 as resolved.

### Phase 6 — Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] Manual: inventory with `SetMaxConcurrency > 1`, per-node args, no leakage.

---

## 10. Comparison: All Plans (revised)

| Aspect | This plan (map storage) | `plan-map-backed.md` (map projection) | `plan-deepcopy.md` | `plan-serialize.md` | `plan-runnable-options-interface.md` | `plan-functional-options.md` |
|--------|------------------------|--------------------------------------|-------------------|-------------------|-------------------------------------|---------------------|
| **Skill bodies change** | **No** | No | No | No | Yes | Yes |
| **Files changed** | **~4** | ~4 | ~3 | ~5 | ~80+ | ~80+ |
| **Clone mechanism** | `ToMap` → modify → `FromMap` | `ToMap` → `FromMap` (with conversion) | `deep.Copy` | JSON marshal/unmarshal | N/A (no shared state) | N/A (no shared state) |
| **Map type** | `map[string]any` (typed values) | `map[string]string` (strings) | N/A | N/A | N/A | N/A |
| **ToMap/FromMap complexity** | **~15 lines (map copy)** | ~80 lines (field conversion) | 0 | ~50 lines (MarshalJSON) | 0 | 0 |
| **Silent data loss risk** | **None** (new field = new key, automatic) | Yes (field not in ToMap) | None | Yes (proxy out of sync) | None | None |
| **Per-type overrides needed** | **None** | Yes (commandImplementation) | None | Yes (commandImplementation) | N/A | N/A |
| **`unsafe` usage** | No | No | Yes (in library) | No | No | No |
| **`reflect` usage** | Yes (in `cloneFromMap`) | Yes (in `cloneSkill`) | No (in library) | Yes (in `cloneSkill`) | No | No |
| **External dependency** | None | None | `brunoga/deep` | None | None | None |
| **Serialization** | **Free** (`json.Marshal(ToMap())`) | Free (`json.Marshal(map)`) | No | Yes (custom) | No | No |
| **Type safety in skill bodies** | **Yes** (getters return typed values) | Yes | Yes | Yes | Yes | Yes |
| **Performance** | ~2-5µs | ~2-5µs | ~10µs | ~50-200µs | ~0 | ~0 |
| **Concurrency safety** | By construction | By construction | By construction | By construction | By construction | By construction |
| **Check() bug fix** | Yes | Yes | Yes | Yes | Yes | Yes |
| **ctx future-proofing** | No | No | No | No | Yes | Yes |
| **Migration effort** | Hours | Hours | Hours | Hours | Days | Days |

---

## 11. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now sets `NodeConfig` and `BecomeUser` (secondary
  bug closed — verified by a dedicated test).
- All existing skill tests pass **without modification** (signatures unchanged).
- **`ToMap`/`FromMap` round-trip test passes** for `BaseSkill` and
  `commandImplementation`.
- **Original skill unchanged after clone**: verify that `ToMap → modify →
  cloneFromMap` does not mutate the original skill.
- **Serialization bonus:** `json.Marshal(skill.ToMap())` produces valid JSON.
- `docs/skills.md` documents the map-storage model and `ToMap`/`FromMap`
  contract.

---

## 12. Fallback

If Phase 0 (spike test) reveals that `reflect.FieldByName("BaseSkill")` doesn't
work reliably, or that the map-as-storage restructuring is too invasive, fall
back to `plan-deepcopy.md` (which handles cloning externally with zero
restructuring of `BaseSkill`).
