# Implementation Plan: Map-Backed Serializable Runnable (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Non-breaking change (internal only)
**Scope:** `types/base_skill.go`, `types/runnable_interface.go`, `command_implementation.go`, `node_implementation.go`, tests, docs
**Inspiration:** [`dracory/ui`](https://github.com/dracory/ui) `BlockInterface` — `ToMap()`/`FromMap()` serialization, all state round-trips through a map
**Compared to:** `plan-runnable-options-interface.md`, `plan-functional-options.md`, `plan-deepcopy.md`, `plan-serialize.md`

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

**Add `ToMap()` / `FromMap()` to the skill (the `dracory/ui` Block pattern).
The framework clones the skill via `FromMap(ToMap())` before mutating it.**
The original shared instance is never touched. All existing signatures stay
the same — `Run()`, `Check()`, `SetNodeConfig()`, `SetArgs()`, etc. are
unchanged.

```go
// Framework per call (the ONLY change in node_implementation.go):
clone := skill.FromMap(skill.ToMap())  // fresh instance, same state
clone.SetNodeConfig(n.cfg)             // mutate the CLONE, not the original
clone.SetDryRun(n.cfg.IsDryRunMode)
result := clone.Run()                  // run the clone
```

This is the same safety mechanism as `plan-deepcopy.md` and `plan-serialize.md`
— clone before mutate — but uses a `map[string]string` as the intermediate
representation instead of `deep.Copy` (reflection + unsafe) or JSON
marshaling. The map approach is inspired by `dracory/ui`'s `Block`, which
serializes all state through `ToMap()` / `NewBlockFromMap()`.

---

## 3. Why Map Instead of deep.Copy or JSON

| Concern | `deep.Copy` | JSON serialize | **Map (this plan)** |
|---------|------------|---------------|-------------------|
| Unexported fields | Handled via `unsafe` | Needs custom `MarshalJSON` (~50 lines) | **Handled via existing getters/setters** (no unsafe, no proxy) |
| External dependency | `brunoga/deep` | None (stdlib) | **None (stdlib)** |
| Extra skill fields (beyond BaseSkill) | Automatic | Needs custom MarshalJSON per type | Needs `ToMap`/`FromMap` override per type (but simpler than JSON) |
| Serialization reusable for other features | No | Yes (JSON string) | **Yes** (`map[string]string` — can `json.Marshal` the map for free) |
| Performance | ~10µs (reflection) | ~50-200µs (JSON) | **~1-5µs** (map allocation + field copies) |
| `unsafe` usage | Yes (in library) | No | **No** |

The map approach hits a sweet spot: no `unsafe`, no external dependency, no
custom JSON proxy structs, and the `map[string]string` is reusable for
logging/auditing/playbooks. The cost is adding `ToMap`/`FromMap` methods to
`BaseSkill` (and overriding them on `commandImplementation` for its extra
fields).

---

## 4. Core Design

### 4.1 `ToMap()` / `FromMap()` on `RunnableInterface`

Add two methods to `RunnableInterface`. These are the only additions — all
existing methods stay:

```go
type RunnableInterface interface {
	// === EXISTING (unchanged) ===
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

	// === NEW (serialization) ===
	// ToMap serializes the skill's execution-time state to a map[string]string.
	// FromMap creates a fresh instance of the same concrete type, populated
	// from the map. The framework uses these to clone the skill before
	// mutating it, eliminating the shared-state race.
	ToMap() map[string]string
	FromMap(m map[string]string) RunnableInterface
}
```

**Note:** Adding methods to an interface is technically a breaking change for
external implementors who don't have `ToMap`/`FromMap`. However:
- `Run()`, `Check()`, and all setters are **unchanged** — skill bodies don't change.
- All built-in skills embed `*BaseSkill`, which provides default `ToMap`/`FromMap`
  implementations. Skills with no extra fields beyond `*BaseSkill` (all 40
  built-in skills except `commandImplementation`) inherit these automatically.
- Only `commandImplementation` needs to override (it has 3 extra fields).

### 4.2 `ToMap()` / `FromMap()` on `BaseSkill` (`types/base_skill.go`)

`BaseSkill` implements `ToMap`/`FromMap` using its existing getters/setters.
No `unsafe`, no reflection, no proxy struct — just read/write fields through
the existing methods:

```go
// ToMap serializes BaseSkill's execution-time state to a map[string]string.
// Construction-time fields (id, description) are included so FromMap can
// create a fully populated clone.
func (b *BaseSkill) ToMap() map[string]string {
	m := map[string]string{
		"id":          b.id,
		"description": b.description,
		"dryRun":      strconv.FormatBool(b.dryRun),
		"becomeUser":  b.GetBecomeUser(),
		"timeout":     strconv.FormatInt(int64(b.timeout), 10),
	}
	// NodeConfig fields
	m["sshHost"] = b.nodeCfg.SSHHost
	m["sshPort"] = b.nodeCfg.SSHPort
	m["sshLogin"] = b.nodeCfg.SSHLogin
	m["sshKey"] = b.nodeCfg.SSHKey
	m["rootUser"] = b.nodeCfg.RootUser
	m["nonRootUser"] = b.nodeCfg.NonRootUser
	m["dbPort"] = b.nodeCfg.DBPort
	m["dbRootPassword"] = b.nodeCfg.DBRootPassword
	m["chdir"] = b.nodeCfg.Chdir
	m["isDryRunMode"] = strconv.FormatBool(b.nodeCfg.IsDryRunMode)
	// NodeConfig.Args → "arg_*" keys
	for k, v := range b.nodeCfg.Args {
		m["arg_"+k] = v
	}
	// Skill-level args → "sarg_*" keys (separate namespace from node args)
	for k, v := range b.args {
		m["sarg_"+k] = v
	}
	// KexAlgorithms / HostKeyAlgorithms (slices) → comma-separated
	m["kexAlgorithms"] = strings.Join(b.nodeCfg.KexAlgorithms, ",")
	m["hostKeyAlgorithms"] = strings.Join(b.nodeCfg.HostKeyAlgorithms, ",")
	return m
}

// FromMap creates a fresh BaseSkill from a map and returns it as RunnableInterface.
// This is used by the framework to clone a skill before mutating it.
func (b *BaseSkill) FromMap(m map[string]string) RunnableInterface {
	clone := NewBaseSkill()
	clone.SetID(m["id"])
	clone.SetDescription(m["description"])
	clone.SetDryRun(strToBool(m["dryRun"]))
	clone.SetBecomeUser(m["becomeUser"])
	clone.SetTimeout(time.Duration(strToInt64(m["timeout"])))

	cfg := NodeConfig{
		SSHHost:       m["sshHost"],
		SSHPort:       m["sshPort"],
		SSHLogin:      m["sshLogin"],
		SSHKey:        m["sshKey"],
		RootUser:      m["rootUser"],
		NonRootUser:   m["nonRootUser"],
		DBPort:        m["dbPort"],
		DBRootPassword: m["dbRootPassword"],
		Chdir:         m["chdir"],
		IsDryRunMode:  strToBool(m["isDryRunMode"]),
		KexAlgorithms: strToSlice(m["kexAlgorithms"]),
		HostKeyAlgorithms: strToSlice(m["hostKeyAlgorithms"]),
	}
	// Reconstruct NodeConfig.Args from "arg_*" keys
	cfg.Args = make(map[string]string)
	for k, v := range m {
		if strings.HasPrefix(k, "arg_") {
			cfg.Args[k[4:]] = v
		}
	}
	clone.SetNodeConfig(cfg)

	// Reconstruct skill-level args from "sarg_*" keys
	skillArgs := make(map[string]string)
	for k, v := range m {
		if strings.HasPrefix(k, "sarg_") {
			skillArgs[k[5:]] = v
		}
	}
	clone.SetArgs(skillArgs)

	return clone
}
```

**~40 lines.** No proxy struct, no `MarshalJSON`, no `unsafe`. Just read/write
through the existing getters/setters.

### 4.3 The `FromMap` return-type problem

`BaseSkill.FromMap` returns `*BaseSkill` as `RunnableInterface`. But the
framework needs a clone of the **concrete type** (e.g. `*UserCreate`), not a
bare `*BaseSkill` — because `*BaseSkill.Run()` is the stub that returns an
error. The concrete skill's `Run()` is what actually executes.

Two solutions:

#### Solution A: Each skill overrides `FromMap` (cleanest, ~5 lines per skill)

```go
// In skills/user/create.go:
func (u *UserCreate) FromMap(m map[string]string) types.RunnableInterface {
	clone := NewUserCreate()                    // creates *UserCreate with fresh *BaseSkill
	clone.BaseSkill = u.BaseSkill.FromMap(m).(*types.BaseSkill)  // populate BaseSkill from map
	return clone
}
```

This is ~3 lines per skill × 40 skills = ~120 lines of mechanical boilerplate.
Each skill's `FromMap` calls its own constructor + delegates to
`BaseSkill.FromMap` for the base fields. Since all 40 skills only embed
`*BaseSkill` (no extra fields), the delegation is identical in every skill.

**Could be reduced with a helper:**

```go
// In types/base_skill.go:
// CloneFromMap creates a new instance of the same concrete type as 'template',
// populated from the map. Uses reflect to create the concrete type, then
// calls BaseSkill.FromMap to populate the embedded BaseSkill.
func CloneFromMap(template RunnableInterface, m map[string]string) RunnableInterface {
	typ := reflect.TypeOf(template).Elem()
	clonePtr := reflect.New(typ)
	clone := clonePtr.Interface().(RunnableInterface)
	// All skills embed *BaseSkill — find it and populate from map
	// ... (via reflect field lookup or interface assertion)
	return clone
}
```

But this reintroduces reflection. **Decision: Solution B.**

#### Solution B: `reflect` in the framework, `ToMap`/`FromMap` only on `BaseSkill` (zero per-skill code)

The framework uses `reflect` to create the concrete type, then calls
`BaseSkill.FromMap` on the embedded `*BaseSkill`:

```go
// In node_implementation.go:
func cloneSkill(skill types.RunnableInterface) (types.RunnableInterface, error) {
	data := skill.ToMap()  // promoted from BaseSkill — works for all skills

	// reflect to create the concrete type (e.g. *user.UserCreate)
	typ := reflect.TypeOf(skill).Elem()
	clonePtr := reflect.New(typ)
	clone := clonePtr.Interface().(types.RunnableInterface)

	// Populate the embedded BaseSkill from the map.
	// All skills embed *BaseSkill as the first (anonymous) field.
	clone.FromMap(data)  // promoted from BaseSkill — populates the embedded BaseSkill

	return clone, nil
}
```

Wait — `clone.FromMap(data)` would call `BaseSkill.FromMap` which returns a
**new** `*BaseSkill`, not populate the existing embedded one. We need
`FromMap` to populate in-place, or we need a different approach.

**Revised `FromMap` — populate in-place (not return new):**

```go
// FromMap populates this BaseSkill's fields from the map.
// Does NOT create a new instance — modifies in place.
func (b *BaseSkill) FromMap(m map[string]string) {
	b.id = m["id"]
	b.description = m["description"]
	b.dryRun = strToBool(m["dryRun"])
	b.SetBecomeUser(m["becomeUser"])
	b.timeout = time.Duration(strToInt64(m["timeout"]))
	// ... NodeConfig, args, etc. (same as above but writing to b, not a new clone)
}
```

Then the interface method becomes:

```go
type RunnableInterface interface {
	// ... existing methods unchanged ...
	ToMap() map[string]string
	FromMap(m map[string]string)  // populates in place, no return
}
```

And the framework:

```go
func cloneSkill(skill types.RunnableInterface) (types.RunnableInterface, error) {
	data := skill.ToMap()
	typ := reflect.TypeOf(skill).Elem()
	clonePtr := reflect.New(typ)
	clone := clonePtr.Interface().(types.RunnableInterface)
	clone.FromMap(data)  // populates the embedded BaseSkill in place
	return clone, nil
}
```

This works because `reflect.New(typ)` creates a zero-value `*UserCreate` with
a nil `*BaseSkill` embed. We need to initialize the embed first:

```go
func cloneSkill(skill types.RunnableInterface) (types.RunnableInterface, error) {
	data := skill.ToMap()
	typ := reflect.TypeOf(skill).Elem()
	clonePtr := reflect.New(typ)
	clone := clonePtr.Interface().(types.RunnableInterface)

	// The concrete skill embeds *BaseSkill. reflect.New created it as nil.
	// We need to set it to a fresh BaseSkill before calling FromMap.
	// Use reflect to find the embedded *BaseSkill field and set it.
	v := clonePtr.Elem()
	bsField := v.FieldByName("BaseSkill")  // the embedded *BaseSkill
	if !bsField.IsValid() {
		return nil, fmt.Errorf("cloneSkill: no embedded BaseSkill in %s", typ.Name())
	}
	bsField.Set(reflect.ValueOf(types.NewBaseSkill()))

	clone.FromMap(data)  // now populates the fresh BaseSkill in place
	return clone, nil
}
```

**This uses `reflect` but NOT `unsafe`.** `reflect.New`, `reflect.ValueOf`,
`FieldByName`, and `Set` are all safe, standard operations. The only
requirement is that the skill embeds `*BaseSkill` as a field named `BaseSkill`
(which all 40 skills do — it's the documented pattern).

#### Solution C: `FromMap` returns `RunnableInterface`, each skill overrides (no reflect)

```go
// Interface:
FromMap(m map[string]string) RunnableInterface

// BaseSkill:
func (b *BaseSkill) FromMap(m map[string]string) RunnableInterface {
	clone := NewBaseSkill()
	clone.populateFromMap(m)
	return clone
}

// Each skill (e.g. UserCreate):
func (u *UserCreate) FromMap(m map[string]string) types.RunnableInterface {
	clone := NewUserCreate()
	clone.BaseSkill = types.NewBaseSkill()
	clone.BaseSkill.populateFromMap(m)
	return clone
}
```

~3 lines per skill, no reflect, no unsafe. But ~120 lines of boilerplate.

**Decision: Solution B (reflect, zero per-skill code).** The reflect usage is
contained in one function (`cloneSkill`), uses only safe operations, and
eliminates 120 lines of mechanical boilerplate across 40 skills. If an
organization policy prohibits reflect, fall back to Solution C.

### 4.4 `commandImplementation` override

`commandImplementation` has 3 extra fields (`command`, `required`, `chdir`)
beyond `*BaseSkill`. It overrides `ToMap`/`FromMap`:

```go
// ToMap includes BaseSkill fields + commandImplementation's extra fields.
func (c *commandImplementation) ToMap() map[string]string {
	m := c.BaseSkill.ToMap()
	m["command"] = c.command
	m["required"] = strconv.FormatBool(c.required)
	m["chdir"] = c.chdir
	return m
}

// FromMap populates BaseSkill fields + commandImplementation's extra fields.
func (c *commandImplementation) FromMap(m map[string]string) {
	c.BaseSkill.FromMap(m)
	c.command = m["command"]
	c.required = strToBool(m["required"])
	c.chdir = m["chdir"]
}
```

~10 lines. The `cloneSkill` function handles this automatically — since
`commandImplementation` overrides `FromMap`, the `clone.FromMap(data)` call
dispatches to the override via the interface.

### 4.5 How the framework clones — `node_implementation.go`

```go
func (n *nodeImplementation) Run(skill types.RunnableInterface) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	// Clone the skill so concurrent goroutines each get an isolated instance.
	clone, err := cloneSkill(skill)
	if err != nil {
		results.Results[n.GetHost()] = types.Result{
			Changed: false,
			Message: fmt.Sprintf("failed to clone skill: %v", err),
			Error:   err,
		}
		return results
	}

	// Mutate the CLONE — same setters as before, but on a private copy.
	clone.SetNodeConfig(n.cfg)
	clone.SetDryRun(n.cfg.IsDryRunMode)
	if clone.GetBecomeUser() == "" {
		clone.SetBecomeUser(n.cfg.BecomeUser)
	}
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
`NodeConfig`/`BecomeUser` (fixes the secondary bug).

### 4.6 Free serialization (the bonus)

Because `ToMap()` returns a `map[string]string`, serialization to JSON is free:

```go
// Log the exact skill state before execution (audit trail):
m := skill.ToMap()
data, _ := json.Marshal(m)
logger.Info("skill execution", "state", string(data))

// Store skill definitions in playbook files:
// { "id": "user-create", "sarg_username": "alice", "sarg_shell": "/bin/bash" }

// Send skill state to a remote worker:
data, _ := json.Marshal(skill.ToMap())
// ... send over network ...
var m map[string]string
json.Unmarshal(data, &m)
clone, _ := cloneSkill(skill)  // or NewUserCreate().FromMap(m)
clone.FromMap(m)
clone.Run()
```

No custom `MarshalJSON`, no proxy struct. The `map[string]string` IS the
serializable form.

---

## 5. What Changes

### 5.1 Core (must change)

| File | Change |
|------|--------|
| `types/runnable_interface.go` | Add `ToMap() map[string]string` and `FromMap(m map[string]string)` to `RunnableInterface`. All existing methods unchanged. |
| `types/base_skill.go` | Add `ToMap()` / `FromMap()` implementations (~40 lines). Add `populateFromMap` helper. Add small `strToBool`/`strToInt64`/`strToSlice` helpers. No existing code changed. |
| `command_implementation.go` | Override `ToMap`/`FromMap` to include `command`/`required`/`chdir` (~10 lines). No existing code changed. |
| `node_implementation.go` | Add `cloneSkill()` helper. Rewrite `Run`/`RunByID`/`Check` to clone before mutating. Same setter calls, just on the clone. |
| `go.mod` | No new dependency (uses stdlib `reflect`, `strconv`, `strings`). |

### 5.2 What does NOT change

| Component | Changes? |
|-----------|----------|
| `types/become_interface.go` | No |
| `types/registry.go` | No |
| `runner_interface.go` | No |
| `node_interface.go` / `inventory_interface.go` | No |
| `inventory_implementation.go` | No (goroutines already call `n.Run(skill)` which now clones) |
| `group_implementation.go` | No (sequential; calls `n.Run(skill)` which now clones) |
| `registry.go` (ork pkg) | No |
| `skill.go` | No |
| All `skills/**/*.go` (40 files) | **No** — skill bodies unchanged, no overrides needed |
| All `skills/**/*_test.go` | **No** — signatures unchanged |
| `skill_test.go` / `node_implementation_test.go` / etc. | **No** (signatures unchanged) |
| `docs/skills.md` / `docs/quick_start.md` / etc. | Minimal (add a note about ToMap/FromMap) |

**Total files changed: 4** (`types/runnable_interface.go`, `types/base_skill.go`,
`command_implementation.go`, `node_implementation.go`) + tests + docs.

### 5.3 Interface change — is it really non-breaking?

Adding `ToMap`/`FromMap` to `RunnableInterface` is technically a breaking
change for **external implementors** who don't embed `*BaseSkill`. However:

- All built-in skills embed `*BaseSkill` and inherit `ToMap`/`FromMap`
  automatically — zero changes needed.
- `commandImplementation` is the only type with extra fields; it overrides
  (~10 lines).
- External skills that embed `*BaseSkill` also inherit automatically.
- External skills that **don't** embed `*BaseSkill` (rare, non-idiomatic) must
  add `ToMap`/`FromMap` — but this is a much smaller change than rewriting
  `Run`/`Check` signatures.

**Compared to plans 1, 2, 5:** those rewrite `Run`/`Check` signatures on every
skill. This plan adds two methods that are inherited via embedding. The skill
**bodies** — the `Run()` and `Check()` implementations — are completely
unchanged.

---

## 6. Concurrency Safety Analysis

Identical to `plan-deepcopy.md` — the safety comes from "clone before mutate":

### `Inventory.Run(skill)` — goroutine per node

```
goroutine 1: n1.Run(skill) → cloneSkill(skill) → clone1.SetNodeConfig(n1.cfg) → clone1.Run()
goroutine 2: n2.Run(skill) → cloneSkill(skill) → clone2.SetNodeConfig(n2.cfg) → clone2.Run()
```

- `skill.ToMap()` reads the shared `skill` — concurrent reads are safe (no writes).
- Each goroutine mutates its own clone — no shared mutable state.
- The original `skill` is never mutated by the framework.

### `Inventory.RunByID(id)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → cloneSkill(skill) → clone1.Run()
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → cloneSkill(skill) → clone2.Run()
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `skill.ToMap()` reads the shared singleton — safe (no writes to it).
- Each goroutine mutates its own clone.

### `Check` — same as `Run`, also fixes the secondary bug

Now clones + sets full config (`NodeConfig`, `BecomeUser`, `DryRun`), closing
the gap where `Check()` previously only set `DryRun`.

---

## 7. Performance Analysis

`ToMap` + `FromMap` per clone:

- `ToMap`: ~15 string assignments + 2 map iterations (args) + 2 `strings.Join`
- `FromMap`: `reflect.New` (~500ns) + ~15 string assignments + 2 map iterations
- Map allocation: ~200-500ns

Estimated cost: **~2-5µs per clone**.

For context:
- One SSH round-trip: **100-500 ms**
- Clone cost as fraction of SSH: **0.001%**

**Faster than JSON serialize** (~50-200µs) and comparable to `deep.Copy`
(~10µs, but `deep` uses reflection more heavily + `unsafe`).

---

## 8. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | Adding `ToMap`/`FromMap` to interface breaks external implementors who don't embed `*BaseSkill` | Low | All built-in skills embed `*BaseSkill`. External skills following the documented pattern inherit automatically. Non-embedding skills are rare and non-idiomatic. |
| R2 | `BaseSkill` gains a new field → `ToMap`/`FromMap` must be updated or the field is **silently lost** during clone | **Medium** | Same risk as `plan-serialize.md`. Mitigation: add a round-trip test that sets all fields, calls `ToMap` → `FromMap`, and asserts all fields survive. If a field is added without updating `ToMap`/`FromMap`, the test fails. |
| R3 | `reflect.FieldByName("BaseSkill")` fails if a skill doesn't name its embed `BaseSkill` | Low | All 40 built-in skills use `*types.BaseSkill` (anonymous embed → field name is `BaseSkill`). Add a guard in `cloneSkill` that returns an error if the field isn't found. |
| R4 | `commandImplementation` (or future types with extra fields) forgets to override `ToMap`/`FromMap` → extra fields lost | Medium | Same risk as serialize plan. Mitigation: round-trip test for `commandImplementation`. Document the override requirement. |
| R5 | `reflect` usage in `cloneSkill` | Low | `reflect.New`, `FieldByName`, `Set` are safe, standard operations. No `unsafe`. Contained in one function. |
| R6 | Slices (`KexAlgorithms`, `HostKeyAlgorithms`) serialized as comma-separated strings — empty string vs empty slice ambiguity | Low | `strToSlice("")` returns `nil` (or empty slice). Document the convention. These fields are rarely set. |

---

## 9. Phased Implementation Plan

### Phase 0 — Spike test (validate approach)

- [ ] Write a throwaway test: create a `*UserCreate`, set args + nodeConfig,
      call `ToMap()`, create a fresh instance via `reflect.New`, call
      `FromMap()`, verify the clone has the same state but is a different
      pointer.
- [ ] Verify `reflect.FieldByName("BaseSkill")` works for all skill types.
- [ ] Verify mutating the clone doesn't affect the original.
- [ ] If the spike fails, fall back to `plan-deepcopy.md`.

### Phase 1 — `ToMap`/`FromMap` on `BaseSkill`

- [ ] Add `ToMap()` / `FromMap()` / `populateFromMap` to `types/base_skill.go`.
- [ ] Add `strToBool`/`strToInt64`/`strToSlice` helpers.
- [ ] Add `ToMap()` / `FromMap()` to `RunnableInterface` in
      `types/runnable_interface.go`.
- [ ] Add a **round-trip test**: set all BaseSkill fields, `ToMap` → `FromMap`,
      assert all fields survive. **This is the guard against R2.**
- [ ] `go build ./...`, `go test ./types/...`

### Phase 2 — `commandImplementation` override

- [ ] Override `ToMap`/`FromMap` on `commandImplementation`.
- [ ] Add a round-trip test for `commandImplementation`.
- [ ] `go build ./...`, `go test ./...`

### Phase 3 — `cloneSkill` + framework integration

- [ ] Add `cloneSkill()` helper to `node_implementation.go`.
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
- [ ] Add a test verifying `cloneSkill` failure is handled gracefully.
- [ ] Run `go test -race ./...` — must be green.

### Phase 5 — Docs

- [ ] Add a note in `docs/skills.md`: "Skills are cloneable via
      `ToMap()`/`FromMap()`. The framework clones each skill per call for
      concurrency safety. If your skill has fields beyond `*BaseSkill`,
      override `ToMap`/`FromMap` to include them."
- [ ] Update `docs/review/2026-07-30-comprehensive-review.md` or add a
      follow-up note marking Critical #1 as resolved.

### Phase 6 — Verification

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...` (the critical check)
- [ ] Manual: inventory with `SetMaxConcurrency > 1`, per-node args, no leakage.

---

## 10. Comparison: All Five Plans (revised)

| Aspect | This plan (map clone) | `plan-deepcopy.md` | `plan-serialize.md` | `plan-runnable-options-interface.md` | `plan-functional-options.md` |
|--------|----------------------|-------------------|-------------------|-------------------------------------|---------------------|
| **API break** | Minimal (add 2 methods to interface) | None | None | Major (rewrite Run/Check) | Major (rewrite Run/Check) |
| **Skill bodies change** | **No** | **No** | **No** | Yes (rewrite Run/Check) | Yes (rewrite Run/Check) |
| **Files changed** | ~4 | ~3 | ~5 | ~80+ | ~80+ |
| **Clone mechanism** | `ToMap`/`FromMap` (map) | `deep.Copy` (reflection+unsafe) | JSON marshal/unmarshal | N/A (no shared state) | N/A (no shared state) |
| **External dependency** | None | `brunoga/deep` (unsafe) | None (stdlib) | None | None |
| **`unsafe` usage** | No | Yes (in library) | No | No | No |
| **`reflect` usage** | Yes (in `cloneSkill`, safe ops only) | No (in library) | Yes (in `cloneSkill`) | No | No |
| **Custom boilerplate** | ~50 lines (ToMap/FromMap on BaseSkill + commandImpl) | 0 | ~50 lines (MarshalJSON) | 0 | 0 |
| **Silent data loss risk** | Medium (R2 — new field not in ToMap) | None | Yes (proxy out of sync) | None | None |
| **Serialization reusable** | **Yes** (map → JSON for free) | No | Yes (JSON string) | No | No |
| **Third-party extensibility** | Automatic (embed BaseSkill) | Automatic | Automatic (with custom MarshalJSON) | Limited | Limited |
| **Performance** | ~2-5µs/call | ~10µs/call | ~50-200µs/call | ~0 | ~0 |
| **Concurrency safety** | By construction (clone per call) | By construction (clone per call) | By construction (clone per call) | By construction (fresh opts) | By construction (local RunConfig) |
| **Check() bug fix** | Yes | Yes | Yes | Yes | Yes |
| **ctx future-proofing** | No (needs break later) | No | No | Yes | Yes |
| **Migration effort** | Hours | Hours | Hours | Days | Days |

---

## 11. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now sets `NodeConfig` and `BecomeUser` (secondary
  bug closed — verified by a dedicated test).
- All existing skill tests pass **without modification** (signatures unchanged).
- **BaseSkill round-trip test passes** — all fields survive `ToMap` → `FromMap`
  (guards against R2).
- **`commandImplementation` round-trip test passes** — extra fields survive
  (guards against R4).
- `docs/skills.md` documents the `ToMap`/`FromMap` contract for skill authors.
- **Serialization bonus:** `json.Marshal(skill.ToMap())` produces valid JSON
  that can be round-tripped back via `json.Unmarshal` → `FromMap`.

---

## 12. Fallback

If Phase 0 (spike test) reveals that `reflect.FieldByName("BaseSkill")` doesn't
work reliably across all skill types, or if the `ToMap`/`FromMap` field-mapping
proves too fragile, fall back to `plan-deepcopy.md` (which handles all fields
automatically via `unsafe`, with zero custom code).
