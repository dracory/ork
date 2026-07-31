# Implementation Plan: Functional Options on Run/Check (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Breaking change (major version bump candidate)
**Scope:** `types`, `ork` (node/inventory/group/command), all `skills/*`, tests, docs
**Compared to:** `plan-runnable-options-interface.md` (opts as interface), `plan-deepcopy.md` (deep-copy, non-breaking), `plan-serialize.md` (JSON serialize, non-breaking)

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

**Pass execution-time config as functional options directly to `Run`/`Check`.**
The skill applies them to a **local** `RunConfig` struct inside the method body
— never to its own fields. The shared skill instance is never mutated.

```go
// Skill contract:
type RunnableInterface interface {
    GetID() string
    GetDescription() string
    Check(opts ...RunOption) (bool, error)
    Run(opts ...RunOption) Result
}

// Call site:
result := skill.Run(
    types.WithContext(ctx),
    types.WithNodeConfig(n.cfg),
    types.WithDryRun(n.cfg.IsDryRunMode),
)
```

Inside the skill:

```go
func (u *UserCreate) Run(opts ...types.RunOption) types.Result {
    rc := types.ApplyOptions(opts)  // local, on the stack — no shared state
    cfg := rc.NodeConfig
    username := rc.Args["username"]
    // ...
}
```

**Why this is safe:** each goroutine calls `skill.Run(opts...)`. Inside `Run`,
`ApplyOptions(opts)` creates a `RunConfig` **local to that call's stack**. The
shared `skill` pointer is never written to. Concurrent goroutines each get
their own `RunConfig` — no shared mutable state, no race.

---

## 3. Core Design

### 3.1 `RunConfig` — the execution-time config struct (`types/run_config.go` — new file)

```go
package types

import (
	"context"
	"log/slog"
	"time"
)

// RunConfig carries everything a skill needs for a single execution against a
// single node. It is built fresh per call by ApplyOptions and lives on the
// call stack — it is never stored on the skill struct.
//
// Skills read fields directly (rc.NodeConfig, rc.Args["username"], etc.).
// No getters, no interface, no methods — just a plain struct.
type RunConfig struct {
	Ctx        context.Context
	NodeConfig NodeConfig
	Args       map[string]string
	DryRun     bool
	BecomeUser string
	Timeout    time.Duration
	Logger     *slog.Logger
}

// RunOption configures a RunConfig. New optional execution parameters can be
// added by introducing new RunOption constructors — no interface change needed.
type RunOption func(*RunConfig)

// ApplyOptions builds a RunConfig from options. Returns a zero-value
// RunConfig if no options are provided.
func ApplyOptions(opts ...RunOption) RunConfig {
	var rc RunConfig
	for _, opt := range opts {
		opt(&rc)
	}
	return rc
}
```

### 3.2 Built-in option constructors (`types/run_config.go`)

```go
func WithContext(ctx context.Context) RunOption {
	return func(rc *RunConfig) { rc.Ctx = ctx }
}

func WithNodeConfig(cfg NodeConfig) RunOption {
	return func(rc *RunConfig) { rc.NodeConfig = cfg }
}

func WithArgs(args map[string]string) RunOption {
	return func(rc *RunConfig) { rc.Args = args }
}

func WithArg(key, value string) RunOption {
	return func(rc *RunConfig) {
		if rc.Args == nil {
			rc.Args = make(map[string]string)
		}
		rc.Args[key] = value
	}
}

func WithDryRun(dryRun bool) RunOption {
	return func(rc *RunConfig) { rc.DryRun = dryRun }
}

func WithBecomeUser(user string) RunOption {
	return func(rc *RunConfig) { rc.BecomeUser = user }
}

func WithTimeout(d time.Duration) RunOption {
	return func(rc *RunConfig) { rc.Timeout = d }
}

func WithLogger(l *slog.Logger) RunOption {
	return func(rc *RunConfig) { rc.Logger = l }
}
```

### 3.3 `RunnableInterface` — new signature (`types/runnable_interface.go`)

```go
package types

// RunnableInterface is the contract every skill (built-in or third-party)
// implements. Execution-time config arrives via functional options at call
// time — the skill applies them to a local RunConfig, never to its own fields.
type RunnableInterface interface {
	GetID() string
	GetDescription() string
	Check(opts ...RunOption) (bool, error)
	Run(opts ...RunOption) Result
}
```

No `RunnableOptionsInterface`. No `RunnableOptions` struct. No setters on the
interface. No `ctx` as a separate param — it's just another option
(`WithContext(ctx)`).

### 3.4 `BaseSkill` rework (`types/base_skill.go`)

`BaseSkill` keeps **only construction-time** state: `id`, `description`.
Execution-time fields (`nodeCfg`, `args`, `dryRun`, `timeout`, `becomeUser`)
and **all their setters/with-ers are removed**.

```go
type BaseSkill struct {
	id          string
	description string
}

func NewBaseSkill() *BaseSkill { return &BaseSkill{} }

func (b *BaseSkill) GetID() string                      { return b.id }
func (b *BaseSkill) SetID(id string) *BaseSkill          { b.id = id; return b }
func (b *BaseSkill) GetDescription() string             { return b.description }
func (b *BaseSkill) SetDescription(d string) *BaseSkill  { b.description = d; return b }

// WithID / WithDescription remain as construction-time fluent helpers.
// WithNodeConfig / WithArgs / WithArg / WithDryRun / WithTimeout / WithBecomeUser
// are REMOVED — they were execution-time setters.

// Default Check/Run stubs now take the new signature:
func (b *BaseSkill) Check(opts ...RunOption) (bool, error) {
	return false, fmt.Errorf("Check() must be implemented by embedding type")
}
func (b *BaseSkill) Run(opts ...RunOption) Result {
	return Result{Error: fmt.Errorf("Run() must be implemented by embedding type")}
}
```

`BaseBecome` is no longer embedded in `BaseSkill` — `becomeUser` is now a field
on `RunConfig`, delivered via `WithBecomeUser` option. `BecomeInterface`
remains on `NodeInterface`/`RunnerInterface` (nodes still have a configurable
become-user that flows into the option at call time).

### 3.5 How a skill reads config — before and after

**Before:**
```go
func (u *UserCreate) Run() types.Result {
	cfg := u.GetNodeConfig()
	username := u.GetArg("username")
	if cfg.IsDryRunMode { ... }
	sshKey := u.GetArg("ssh-key")
	// ...
}
```

**After:**
```go
func (u *UserCreate) Run(opts ...types.RunOption) types.Result {
	rc := types.ApplyOptions(opts)
	cfg := rc.NodeConfig
	username := rc.Args["username"]
	if rc.DryRun { ... }
	sshKey := rc.Args["ssh-key"]
	// ...
}
```

The migration is mechanical: `u.GetNodeConfig()` → `rc.NodeConfig`,
`u.GetArg(k)` → `rc.Args[k]`, `u.IsDryRun()` → `rc.DryRun`,
`u.GetBecomeUser()` → `rc.BecomeUser`, `u.GetTimeout()` → `rc.Timeout`.

### 3.6 How the framework calls a skill — `node_implementation.go`

```go
func (n *nodeImplementation) Run(skill types.RunnableInterface, opts ...types.RunOption) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}

	// Build per-call options: node defaults first, caller opts last (they win).
	callOpts := n.buildRunOptions(opts...)
	result := skill.Run(callOpts...)

	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}

// buildRunOptions merges node-level defaults with caller-supplied options.
// Caller opts are appended last so they override node defaults.
func (n *nodeImplementation) buildRunOptions(opts ...types.RunOption) []types.RunOption {
	merged := []types.RunOption{
		types.WithNodeConfig(n.cfg),
		types.WithDryRun(n.cfg.IsDryRunMode),
	}
	if n.cfg.BecomeUser != "" {
		merged = append(merged, types.WithBecomeUser(n.cfg.BecomeUser))
	}
	if n.cfg.Logger != nil {
		merged = append(merged, types.WithLogger(n.cfg.Logger))
	}
	merged = append(merged, opts...) // caller opts win (applied last)
	return merged
}
```

`RunByID` is the same, except it looks up the skill from the registry first:

```go
func (n *nodeImplementation) RunByID(id string, opts ...types.RunOption) types.Results {
	// ... registry lookup unchanged ...
	result := skill.Run(n.buildRunOptions(opts...)...)
	// ... pack result ...
}
```

`Check` is the same pattern — and now sets full config (fixes the secondary bug):

```go
func (n *nodeImplementation) Check(skill types.RunnableInterface, opts ...types.RunOption) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}
	changed, err := skill.Check(n.buildRunOptions(opts...)...)
	results.Results[n.GetHost()] = types.Result{
		Changed: changed,
		Error:   err,
	}
	return results
}
```

### 3.7 `RunnerInterface` — new signatures

```go
type RunnerInterface interface {
	types.BecomeInterface
	RunCommand(cmd string) types.Results
	Run(runnable types.RunnableInterface, opts ...types.RunOption) types.Results
	RunByID(id string, opts ...types.RunOption) types.Results
	Check(runnable types.RunnableInterface, opts ...types.RunOption) types.Results
	GetLogger() *slog.Logger
	SetLogger(logger *slog.Logger) RunnerInterface
	SetDryRunMode(dryRun bool) RunnerInterface
	GetDryRunMode() bool
}
```

Note: `RunCommand` stays `(cmd string)` — it doesn't involve skills. `ctx` is
not added to `RunCommand` in this plan (it doesn't take options).

---

## 4. Concurrency Safety Analysis

### `Inventory.Run(skill, opts...)` — goroutine per node

```
goroutine 1: n1.Run(skill, opts...) → skill.Run(WithNodeConfig(n1.cfg), ...) → rc1 on stack → clone1.Run()
goroutine 2: n2.Run(skill, opts...) → skill.Run(WithNodeConfig(n2.cfg), ...) → rc2 on stack → clone2.Run()
```

- `skill` is read-only (never mutated by the framework or by Run).
- Each `skill.Run(opts...)` call creates a local `RunConfig` via `ApplyOptions`.
- Each goroutine's `RunConfig` is on its own stack — no sharing.
- The original `skill` struct is never written to.

### `Inventory.RunByID(id, opts...)` — goroutine per node

```
goroutine 1: n1.RunByID(id) → registry.FindByID(id) → skill.Run(WithNodeConfig(n1.cfg), ...)
goroutine 2: n2.RunByID(id) → registry.FindByID(id) → skill.Run(WithNodeConfig(n2.cfg), ...)
```

- `registry.FindByID` is RLock-protected — safe for concurrent reads.
- `skill` is the same singleton pointer, but it's **never mutated** — `Run`
  reads only `s.id`/`s.description` (immutable) and applies opts to a local
  `RunConfig`.
- No race, no shared mutable state.

### `Check` — same as `Run`

Now receives full config via opts (fixes the secondary bug where `Check` only
set `DryRun` and missed `NodeConfig`/`BecomeUser`).

---

## 5. What Changes

### 5.1 Core (must change)

| File | Change |
|------|--------|
| `types/run_config.go` | **NEW** — `RunConfig` struct, `RunOption` type, `ApplyOptions`, all `With*` constructors |
| `types/runnable_interface.go` | Replace `RunnableInterface` (new signature: `Run(opts ...RunOption)`, `Check(opts ...RunOption)`). Remove `RunnableOptions` struct, all setters, `BecomeInterface` embedding. |
| `types/base_skill.go` | Strip to `id`/`description` only; remove execution fields/setters/with-ers; new `Check`/`Run` stub signatures; drop `BaseBecome` embedding. |
| `types/become_interface.go` | Keep `BecomeInterface`/`BaseBecome` for node use; remove from `RunnableInterface`. |
| `node_implementation.go` | Add `buildRunOptions`; rewrite `Run`/`RunByID`/`Check` to new signatures. No `RunCommand` change. |
| `inventory_implementation.go` | Rewrite `Run`/`RunByID`/`Check` to thread opts through to `n.Run(skill, opts...)`. |
| `group_implementation.go` | Same as inventory. |
| `command_implementation.go` | `Run`/`Check` take new signature; read `command`/`chdir`/`required` from struct, everything else from `rc`. |
| `runner_interface.go` | `Run`/`RunByID`/`Check` signatures: `Run(runnable, opts ...RunOption)`, etc. |
| `node_interface.go` / `inventory_interface.go` | Embed updated `RunnerInterface` (no new methods). |
| `skill.go` | Update `NewSkill()` doc example. |
| `registry.go` (ork pkg) | No structural change. |

### 5.2 Skills (mechanical body migration)

~40 skill files across `skills/{apt,fail2ban,mariadb,ping,reboot,security,swap,ufw,user}`.

Per skill:

1. `Check()` → `Check(opts ...types.RunOption) (bool, error)`.
2. `Run()` → `Run(opts ...types.RunOption) types.Result`.
3. Add `rc := types.ApplyOptions(opts)` as first line.
4. `s.GetNodeConfig()` → `rc.NodeConfig`.
5. `s.GetArg(k)` → `rc.Args[k]` (with nil-map guard if needed, or use `types.GetArg(rc, k)` helper).
6. `s.IsDryRun()` / `cfg.IsDryRunMode` → `rc.DryRun`.
7. `s.GetBecomeUser()` → `rc.BecomeUser`.
8. `s.GetTimeout()` → `rc.Timeout`.
9. **Remove** per-skill `SetArgs`/`SetArg`/`SetID`/`SetDescription`/`SetTimeout`/`SetNodeConfig`/`SetDryRun` overrides (no longer on the interface).
10. **Remove** `WithNodeConfig`/`WithArgs`/`WithArg`/`WithDryRun`/`WithTimeout`/`WithBecomeUser` overrides that delegated to `BaseSkill` (those `BaseSkill` methods are gone). Keep constructor-specific `With*` (e.g. `WithRetries`).

### 5.3 Tests (must change)

| File | Change |
|------|--------|
| `skill_test.go` | Drop tests for removed `With*` execution setters; keep `WithID`/`WithDescription`. |
| `node_implementation_test.go` | Update `Run`/`RunByID`/`Check` call sites. |
| `inventory_implementation_test.go` | Same. |
| `group_implementation_test.go` | Same. |
| `command_implementation_test.go` | Same. |
| `registry_test.go` | Update direct `Run`/`Check` invocations. |
| `integration_test.go` | Update end-to-end call sites. |
| `playbook_test.go` | Update. |
| All `skills/**/*_test.go` | Update constructor-only usage; any direct `Run()`/`Check()` calls now need opts. |
| `skills/ping/ping_mock_test.go` | Update mock to new interface. |

### 5.4 Docs (must update)

`docs/skills.md`, `docs/advanced_usage.md`, `docs/idempotency.md`,
`docs/dry_run.md`, `docs/privilege_escalation.md`, `docs/commands.md`,
`docs/quick_start.md`, `docs/playbooks.md`, `docs/livewiki/**`, `README.md`,
`examples/example_playbook.go`, `cmd/ork/**`.

---

## 6. Arg-Access Helper

Skill bodies today read `u.GetArg("username")`. The new `RunConfig` exposes
`rc.Args` as a `map[string]string`. To avoid a nil-map guard in every skill,
add a small free helper in `types`:

```go
// GetArg returns rc.Args[key], or "" if absent. Safe on nil/empty args.
func GetArg(rc RunConfig, key string) string {
	if rc.Args == nil {
		return ""
	}
	return rc.Args[key]
}
```

Skill migration: `u.GetArg(k)` → `types.GetArg(rc, k)`.

---

## 7. Key Design Decisions

### 7.1 `ctx` as an option, not a separate param

Unlike `plan-runnable-options-interface.md` (which has `Run(ctx, opts)`), this
plan makes `ctx` just another option: `RunWithContext(ctx)`.

**Tradeoff — honest:**
- `go vet`'s `contextcheck` linter and the `context` linter expect
  `context.Context` as the **first parameter**. With `WithContext(ctx)` as an
  option, these tools **cannot trace context propagation** through the call
  chain. This is a real downside — you lose static cancellation-flow analysis.
- The benefit is uniformity: every execution-time parameter is an option, no
  special-casing `ctx`. The call site reads naturally:
  `skill.Run(WithContext(ctx), WithNodeConfig(cfg), WithArgs(args))`.

If the `contextcheck` linter loss is unacceptable, fall back to
`plan-runnable-options-interface.md` which keeps `ctx` as a first param.

### 7.2 `RunConfig` is a concrete struct, not an interface

Unlike `plan-runnable-options-interface.md` (which defines
`RunnableOptionsInterface` with getters/setters), this plan uses a plain
exported struct. Skills read fields directly (`rc.Args["username"]`) — no
method dispatch, no interface implementation, no mocking an interface.

**Tradeoff:**
- Simpler for skill authors (just read struct fields).
- Adding a new field to `RunConfig` is **non-breaking for readers** (existing
  skills that don't read the new field are unaffected). It's only breaking for
  **implementors of `RunnableInterface`** if the signature changes — but the
  signature is `Run(opts ...RunOption)`, which doesn't change when `RunConfig`
  gains a field. So adding a field to `RunConfig` + a new `With*` constructor
  is **fully non-breaking**.
- No interface to mock in tests — but `RunConfig` is a struct, so tests just
  construct one directly: `skill.Run(types.WithNodeConfig(testCfg))`.

### 7.3 Third-party extensibility — limited

`RunConfig` is a fixed struct in `types`. A third party cannot add a field to
it without modifying `ork`. They can:

1. Use the built-in fields (sufficient for most skills).
2. Store custom config on their skill struct at construction time
   (`NewMySkill().WithRetries(3)` — construction-time, immutable after
   registration).
3. Use a side-channel: define their own `RunOption` that stores into an
   `Extras map[string]any` field on `RunConfig` (if we add one — see below).

**Optional: `Extras` map for full third-party extensibility:**

```go
type RunConfig struct {
	Ctx        context.Context
	NodeConfig NodeConfig
	Args       map[string]string
	DryRun     bool
	BecomeUser string
	Timeout    time.Duration
	Logger     *slog.Logger
	Extras     map[string]any  // third-party extension point
}

func WithExtra(key string, value any) RunOption {
	return func(rc *RunConfig) {
		if rc.Extras == nil {
			rc.Extras = make(map[string]any)
		}
		rc.Extras[key] = value
	}
}
```

This is stringly-keyed (violates the "no stringly-keyed lookups" goal), but
it's opt-in — the common path uses typed fields, and `Extras` is only for
third-party extensions that genuinely need custom parameters. **Decision: defer
`Extras` — add it only when a real third-party use case emerges.**

### 7.4 Registry: keep instances (no factory change)

Same as `plan-runnable-options-interface.md` — the registry stores
`RunnableInterface` instances. Since `RunnableInterface` no longer has
execution setters and `Run` applies opts to a local `RunConfig`, the framework
never mutates a registered instance. The singleton-per-ID pattern is safe by
construction.

### 7.5 Dry-run source of truth

`BaseSkill.dryRun` is gone. The single source is `rc.DryRun`, populated by
`buildRunOptions` from `n.cfg.IsDryRunMode` (overridable by caller
`WithDryRun`). `NodeConfig.IsDryRunMode` remains for `ssh.Run`'s internal
dry-run handling. `buildRunOptions` sets `rc.DryRun = n.cfg.IsDryRunMode`, so
`rc.DryRun` and `rc.NodeConfig.IsDryRunMode` are always equal for a given call.

---

## 8. Phased Implementation Plan

### Phase 0 — Preparation (no behavior change)
- [ ] Create `types/run_config.go` with `RunConfig`, `RunOption`,
      `ApplyOptions`, all `With*` constructors, `GetArg` helper.
- [ ] Add unit tests for `run_config.go` (option precedence, `GetArg`
      nil-safety, `WithArg` on nil map).
- [ ] `go build ./...` must still pass (new file is additive).

### Phase 1 — Core type changes (breaks build; fix in same commit)
- [ ] Rewrite `types/runnable_interface.go`: new `RunnableInterface`
      (`Run(opts ...RunOption)`, `Check(opts ...RunOption)`). Remove
      `RunnableOptions` struct, all setters, `BecomeInterface` embedding.
- [ ] Rewrite `types/base_skill.go`: strip to `id`/`description`; new stub
      `Check`/`Run` signatures; remove execution setters/with-ers; drop
      `BaseBecome` embedding.
- [ ] Update `types/become_interface.go` doc comment.
- [ ] Update `runner_interface.go` signatures (§3.7).
- [ ] Update `node_interface.go` / `inventory_interface.go` (embed only).

### Phase 2 — Orchestration layer
- [ ] `node_implementation.go`: add `buildRunOptions`; rewrite `Run`/`RunByID`/
      `Check` to new signatures.
- [ ] `inventory_implementation.go`: rewrite `Run`/`RunByID`/`Check` to thread
      opts.
- [ ] `group_implementation.go`: same as inventory.
- [ ] `command_implementation.go`: rewrite `Check`/`Run` to new signature.
- [ ] `skill.go`: update doc example.
- [ ] `registry.go`: import sanity.

### Phase 3 — Skill bodies (one package per commit)
- [ ] `skills/ping` (simplest — template)
- [ ] `skills/apt` (5 files)
- [ ] `skills/ufw` (11 files)
- [ ] `skills/user` (5 files)
- [ ] `skills/mariadb` (14 files)
- [ ] `skills/security` (5 files)
- [ ] `skills/fail2ban` (2 files)
- [ ] `skills/swap` (3 files)
- [ ] `skills/reboot` (1 file)

Each commit: apply §5.2 checklist, run `go build ./skills/<pkg>` + tests.

### Phase 4 — Tests
- [ ] `skill_test.go`: drop removed-setter tests.
- [ ] `node_implementation_test.go` / `inventory_implementation_test.go` /
      `group_implementation_test.go` / `command_implementation_test.go`: update
      call sites; add **concurrency regression test** (N nodes,
      `SetMaxConcurrency(N)`, distinct per-node args, no leakage, run with
      `-race`).
- [ ] `registry_test.go` / `integration_test.go` / `playbook_test.go`: update.
- [ ] All `skills/**/*_test.go`: update.
- [ ] `skills/ping/ping_mock_test.go`: update mock.

### Phase 5 — Docs & examples
- [ ] `docs/skills.md`: new authoring guide (receive opts, apply to local
      `RunConfig`, read fields directly; how to add custom `RunOption`).
- [ ] `docs/advanced_usage.md`, `docs/quick_start.md`, `docs/commands.md`,
      `docs/dry_run.md`, `docs/idempotency.md`, `docs/privilege_escalation.md`,
      `docs/playbooks.md`: update call sites.
- [ ] `docs/livewiki/**`: update API reference.
- [ ] `README.md`, `examples/example_playbook.go`, `cmd/ork/**`: update.

### Phase 6 — Verification
- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...` (confirms the data race is gone).
- [ ] Manual: inventory with `SetMaxConcurrency > 1`, per-node args, no leakage.

---

## 9. Risks & Mitigations

| ID | Risk | Severity | Mitigation |
|----|------|----------|-----------|
| R1 | Breaking change for all external skill authors. | High | Major version bump; migration guide in `docs/skills.md`; per-skill change is mechanical (§5.2). |
| R2 | `contextcheck` linter can't trace `ctx` passed as option. | Medium | Document the tradeoff. If unacceptable, use `plan-runnable-options-interface.md` (ctx as first param). |
| R3 | Skill forgets `rc := types.ApplyOptions(opts)` → gets zero-value `RunConfig`, fails at SSH time. | Medium | The failure is loud (no NodeConfig → SSH connection fails with clear error). Add a lint rule or convention: first line of Run/Check is always `rc := types.ApplyOptions(opts)`. |
| R4 | Skill author mutates the skill struct inside Run (reads `s.someField` instead of `rc.SomeField`). | Medium | Document the contract: "read from `rc`, not from `s`". The built-in skills don't carry execution state after migration. For third-party skills, this is the same risk as `plan-runnable-options-interface.md`. |
| R5 | `BaseSkill` dropping `BaseBecome` breaks a skill that sets become-user at construction. | Low | Audit in Phase 3. Built-in skills don't do this. External skills pass `WithBecomeUser` as an option instead. |
| R6 | Third-party skill needs a custom execution param not in `RunConfig`. | Low | Use construction-time config on the skill struct, or defer `Extras` map (§7.3). |

---

## 10. Comparison: All Four Plans

| Aspect | This plan (func opts) | `plan-runnable-options-interface.md` | `plan-deepcopy.md` | `plan-serialize.md` |
|--------|----------------------|-------------------------------------|-------------------|-------------------|
| **API break** | Major (skill signatures) | Major (skill signatures) | None | None |
| **Files changed** | ~80+ | ~80+ | ~3 | ~5 |
| **`RunnableOptionsInterface`** | No (plain struct) | Yes (interface) | N/A | N/A |
| **`ctx` handling** | Option (`WithContext`) | First param (`ctx`) | N/A (no ctx) | N/A (no ctx) |
| **`contextcheck` linter** | **Broken** | Works | N/A | N/A |
| **Skill reads config from** | `rc` struct fields | `opts.GetX()` methods | `s.GetX()` (unchanged) | `s.GetX()` (unchanged) |
| **External dependency** | None | None | `brunoga/deep` (`unsafe`) | None (stdlib) |
| **Custom boilerplate** | 0 | 0 | 0 | ~50 lines |
| **Silent data loss risk** | None | None | None | **Yes** (R1 in serialize plan) |
| **Concurrency safety** | By construction (local RunConfig) | By construction (fresh opts) | By construction (clone) | By construction (clone) |
| **Check() bug fix** | Yes | Yes | Yes | Yes |
| **ctx future-proofing** | Yes (as option) | Yes (as param) | No | No |
| **Third-party extensibility** | Limited (defer Extras) | Limited (defer GetExtra) | Automatic | Automatic |
| **Performance** | Zero | Zero | ~10µs/call | ~50-200µs/call |
| **Migration effort** | Days | Days | Hours | Hours |

---

## 11. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now uses the same `NodeConfig`/`BecomeUser` as
  `Run` (secondary bug closed — verified by a dedicated test).
- No remaining `SetNodeConfig`/`SetArgs`/`SetDryRun`/`SetBecomeUser` calls on
  `RunnableInterface` anywhere in the tree (grep confirms zero hits).
- `docs/skills.md` documents the new authoring model.
- **Note:** `go vet -vettool=...contextcheck` may report context-flow
  warnings due to `ctx` being an option (R2). This is an accepted tradeoff.
