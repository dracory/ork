# Implementation Plan: Per-Call Execution Options (Concurrency-Safe Skill Execution)

**Date:** 2026-07-30
**Status:** Draft
**Tracking review:** `docs/review/2026-07-30-comprehensive-review.md` (Critical finding #1)
**Type:** Breaking change (major version bump candidate)
**Scope:** `types`, `ork` (node/inventory/group/command/registry), all `skills/*`, tests, docs

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

Three call paths are affected:

| Path | Mutation site | Shared-instance risk |
|------|---------------|----------------------|
| `Inventory.Run(skill)` → `n.Run(skill)` (goroutine/node) | `nodeImplementation.Run` | Same skill pointer per node goroutine |
| `Inventory.RunByID(id)` → `n.RunByID(id)` (goroutine/node) | `nodeImplementation.RunByID` after `registry.FindByID(id)` | Registry returns same singleton pointer to every concurrent caller |
| `Group.Run` | `nodeImplementation.Run` | Currently safe (sequential) but not future-proof |

Secondary bug (also fixed by this redesign): `nodeImplementation.Check()` never
calls `SetNodeConfig`/`SetBecomeUser` before `skill.Run()`, unlike `Run`/`RunByID`.

---

## 2. Goals

1. Eliminate the shared-mutable-state hazard **by construction**, not by convention.
   The fix is **per-call construction**: each skill invocation gets a fresh
   `RunnableOptionsInterface` instance, so there is no shared mutable state to
   race on. The interface may have setters — that is not a correctness hazard
   once the framework never shares an opts instance across goroutines.
2. Keep the public API **statically typed** (no `context.Value`-style stringly-keyed lookups).
3. Allow **third-party skill authors** to add their own execution-time options without
   changes to `ork` (via `RunOption` constructors — see §3.3).
4. Preserve today's **fluent, chainable construction-time API**
   (`NewPing().WithRetries(3)`) — only the *execution-time* configuration model changes.
5. Close the `Check()` config-propagation gap as a side effect.

> Note: "interface stability over time" is intentionally **not** a goal. Adding a
> new method to `RunnableOptionsInterface` later is a normal breaking change
> handled by a major version bump, the same as any other public Go interface.
> The plan does not impose a "frozen interface" constraint.

---

## 3. Target Design

### 3.1 New core interfaces (`types/runnable_interface.go`)

```go
package types

import (
	"context"
	"log/slog"
	"time"
)

// RunnableOptionsInterface carries everything a skill needs for a single
// execution against a single node. It is constructed fresh per call by the
// framework (see §5.1) — each skill invocation receives its own instance, so
// there is no shared mutable state to race on.
//
// The interface has both getters and setters. Setters are safe because the
// framework never shares an opts instance across goroutines. A skill may
// mutate the opts it receives (e.g. to override args before forwarding to a
// sub-skill) — that mutation is local to the skill's own call and does not
// leak to other concurrent invocations, which hold distinct instances.
//
// New optional execution parameters may be added as new getters/setters in a
// future major version (normal Go breaking-change process) or, without a
// breaking change, via new RunOption constructors (see §3.3).
type RunnableOptionsInterface interface {
	GetNodeConfig() NodeConfig
	SetNodeConfig(cfg NodeConfig) RunnableOptionsInterface

	GetArgs() map[string]string
	SetArgs(args map[string]string) RunnableOptionsInterface

	IsDryRun() bool
	SetDryRun(dryRun bool) RunnableOptionsInterface

	GetBecomeUser() string
	SetBecomeUser(user string) RunnableOptionsInterface

	GetTimeout() time.Duration
	SetTimeout(d time.Duration) RunnableOptionsInterface

	GetLogger() *slog.Logger
	SetLogger(l *slog.Logger) RunnableOptionsInterface
}

// RunnableInterface is the contract every skill (built-in or third-party)
// implements. Note: no execution-time setters. Configuration arrives only
// through the options parameter at call time.
type RunnableInterface interface {
	GetID() string
	GetDescription() string
	Check(ctx context.Context, opts RunnableOptionsInterface) (bool, error)
	Run(ctx context.Context, opts RunnableOptionsInterface) Result
}
```

`ctx` stays a separate, first positional parameter (not a method on `opts`),
per standard Go idiom (`net/http`, `database/sql`, `grpc`) — keeps cancellation
discoverable in the signature and keeps `go vet`'s `contextcheck` linter working.

### 3.2 Concrete options (`types/runnable_options.go` — new file)

```go
package types

import (
	"log/slog"
	"time"
)

// runnableOptions is the sole concrete implementation of RunnableOptionsInterface.
// It is constructed fresh per call by the framework (see §5.1) and may be
// mutated by the skill that receives it via its Set* methods. Fields are
// unexported so external code must go through the interface.
type runnableOptions struct {
	nodeConfig  NodeConfig
	args        map[string]string
	dryRun      bool
	becomeUser  string
	timeout     time.Duration
	logger      *slog.Logger
}

// Enforce interface implementation at compile time.
var _ RunnableOptionsInterface = (*runnableOptions)(nil)

func (o *runnableOptions) GetNodeConfig() NodeConfig              { return o.nodeConfig }
func (o *runnableOptions) SetNodeConfig(cfg NodeConfig) RunnableOptionsInterface { o.nodeConfig = cfg; return o }

func (o *runnableOptions) GetArgs() map[string]string            { return o.args }
func (o *runnableOptions) SetArgs(args map[string]string) RunnableOptionsInterface { o.args = args; return o }

func (o *runnableOptions) IsDryRun() bool                        { return o.dryRun }
func (o *runnableOptions) SetDryRun(dryRun bool) RunnableOptionsInterface { o.dryRun = dryRun; return o }

func (o *runnableOptions) GetBecomeUser() string                { return o.becomeUser }
func (o *runnableOptions) SetBecomeUser(user string) RunnableOptionsInterface { o.becomeUser = user; return o }

func (o *runnableOptions) GetTimeout() time.Duration            { return o.timeout }
func (o *runnableOptions) SetTimeout(d time.Duration) RunnableOptionsInterface { o.timeout = d; return o }

func (o *runnableOptions) GetLogger() *slog.Logger               { return o.logger }
func (o *runnableOptions) SetLogger(l *slog.Logger) RunnableOptionsInterface { o.logger = l; return o }
```

### 3.3 Functional `RunOption` extension point (`types/runnable_options.go`)

```go
// RunOption configures a single execution. It is a convenience for building
// opts at call sites without listing every setter; it is not the only way to
// configure opts (the Set* methods on the interface work too).
//
// New optional execution parameters can be added either:
//   - as a new RunOption constructor here (non-breaking for callers), or
//   - as a new getter/setter pair on RunnableOptionsInterface (breaking for
//     implementors, handled by a major version bump).
type RunOption func(*runnableOptions)

// NewRunnableOptions builds a RunnableOptionsInterface from options.
func NewRunnableOptions(opts ...RunOption) RunnableOptionsInterface {
	o := &runnableOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Built-in RunOption constructors:
func WithNodeConfig(cfg NodeConfig) RunOption      { return func(o *runnableOptions) { o.nodeConfig = cfg } }
func WithArgs(args map[string]string) RunOption   { return func(o *runnableOptions) { o.args = args } }
func WithDryRun(dryRun bool) RunOption            { return func(o *runnableOptions) { o.dryRun = dryRun } }
func WithBecomeUser(user string) RunOption       { return func(o *runnableOptions) { o.becomeUser = user } }
func WithTimeout(d time.Duration) RunOption       { return func(o *runnableOptions) { o.timeout = d } }
func WithLogger(l *slog.Logger) RunOption        { return func(o *runnableOptions) { o.logger = l } }
```

Third parties extend by defining their own `RunOption`-returning constructors
in their own package — no `ork` change required (goal #4).

### 3.4 Arg-access helper

Skill bodies today read `u.GetArg("username")`. The interface exposes
`GetArgs() map[string]string` (and a `SetArgs`). To keep skill bodies ergonomic and
avoid a per-skill `args["username"]` + nil-map guard, add a small free helper
in `types`:

```go
// GetArg returns opts.GetArgs()[key], or "" if absent.
// Safe on a nil/empty args map.
func GetArg(opts RunnableOptionsInterface, key string) string {
	if opts == nil {
		return ""
	}
	args := opts.GetArgs()
	if args == nil {
		return ""
	}
	return args[key]
}
```

Migration of a skill body is then a mechanical `u.GetArg(k)` → `types.GetArg(opts, k)`
and `u.GetNodeConfig()` → `opts.GetNodeConfig()`.

### 3.5 `BaseSkill` rework (`types/base_skill.go`)

`BaseSkill` keeps **only construction-time, immutable** state: `id`,
`description`, and any skill-specific builder config. Execution-time fields
(`nodeCfg`, `args`, `dryRun`, `timeout`, `becomeUser`) and **all their
setters/with-ers are removed** from `BaseSkill`.

```go
type BaseSkill struct {
	id          string
	description string
}

func NewBaseSkill() *BaseSkill { return &BaseSkill{} }

func (b *BaseSkill) GetID() string                 { return b.id }
func (b *BaseSkill) SetID(id string) *BaseSkill     { b.id = id; return b }          // construction-time
func (b *BaseSkill) GetDescription() string        { return b.description }
func (b *BaseSkill) SetDescription(d string) *BaseSkill { b.description = d; return b }

// WithID / WithDescription remain as construction-time fluent helpers.
// WithNodeConfig / WithArgs / WithDryRun / WithTimeout / WithBecomeUser / WithArg
// are REMOVED (they were execution-time setters).

// Default Check/Run stubs now take the new signature:
func (b *BaseSkill) Check(ctx context.Context, opts RunnableOptionsInterface) (bool, error) {
	return false, fmt.Errorf("Check() must be implemented by embedding type")
}
func (b *BaseSkill) Run(ctx context.Context, opts RunnableOptionsInterface) Result {
	return Result{Error: fmt.Errorf("Run() must be implemented by embedding type")}
}
```

`BaseBecome` stays as-is (it is construction-time only and no longer carries
execution-time become state; become-user now lives in `runnableOptions`).
`BaseSkill` no longer embeds `BaseBecome` for *execution* state, but may still
embed it if any skill relies on a construction-time become default. **Decision:**
drop `BaseBecome` embedding from `BaseSkill` (become-user is purely an execution
option now). Audit skills for construction-time `WithBecomeUser` usage — none of
the built-in skills set become-user at construction time (they rely on node
propagation), so this is safe. See §7 risk R3.

### 3.6 `BecomeInterface`

`BecomeInterface` (the `SetBecomeUser/GetBecomeUser` pair) is **removed from
`RunnableInterface`**. It remains on `NodeInterface`/`RunnerInterface` (nodes
still have a configurable become-user that flows into `runnableOptions`).
Skills no longer implement it.

---

## 4. Migration Impact (file inventory)

### 4.1 Core (must change)

| File | Change |
|------|--------|
| `types/runnable_interface.go` | Replace `RunnableInterface` + `RunnableOptions` with new design (§3.1, §3.3). `RunnableInterface` loses execution setters; `RunnableOptionsInterface` gains getters + setters. |
| `types/base_skill.go` | Strip to `id`/`description` only; remove execution fields/setters; new `Check`/`Run` stub signatures. |
| `types/become_interface.go` | Keep `BecomeInterface`/`BaseBecome` for node use; remove from `RunnableInterface`. |
| `types/runnable_options.go` | **NEW** — `runnableOptions`, `RunOption`, constructors, `NewRunnableOptions`, `GetArg` helper. |
| `types/registry.go` | No structural change needed (see §5.7 decision). |
| `node_implementation.go` | `Run`/`RunByID`/`Check` build `RunnableOptionsInterface` per call and pass to `skill.Run(ctx, opts)`/`Check(ctx, opts)`. No more `skill.SetXxx`. |
| `inventory_implementation.go` | `Run`/`RunByID`/`Check` pass `ctx` through; build nothing on the skill. |
| `group_implementation.go` | Same as inventory. |
| `command_implementation.go` | `commandImplementation.Run`/`Check` take new signature; read `command`/`chdir` from struct, everything else from `opts`. |
| `runner_interface.go` | `Run`/`RunByID`/`Check` signatures: `Run(ctx, runnable, opts...)`, `RunByID(ctx, id, opts...)`, `Check(ctx, runnable, opts...)`. Deprecation note on `RunByID` updated. |
| `node_interface.go` / `inventory_interface.go` | Embed updated `RunnerInterface` (no field changes). |
| `registry.go` (ork pkg) | `NewDefaultRegistry` unchanged (still registers instances). |
| `skill.go` | `NewSkill()` returns `*BaseSkill` (now slimmer). Update doc example. |

### 4.2 Skills (mechanical body migration)

~40 skill files across `skills/{apt,fail2ban,mariadb,ping,reboot,security,swap,ufw,user}`.

Per skill, the changes are:

1. `Check()` → `Check(ctx context.Context, opts types.RunnableOptionsInterface) (bool, error)`.
2. `Run()` → `Run(ctx context.Context, opts types.RunnableOptionsInterface) types.Result`.
3. Replace `s.GetNodeConfig()` → `opts.GetNodeConfig()`.
4. Replace `s.GetArg(k)` → `types.GetArg(opts, k)`.
5. Replace `s.IsDryRun()` / `cfg.IsDryRunMode` checks → `opts.IsDryRun()` (note: `cfg.IsDryRunMode` is also still on `NodeConfig`; prefer `opts.IsDryRun()` as the single source of truth — see §5.4 resolution rule).
6. Replace `s.GetBecomeUser()` → `opts.GetBecomeUser()`.
7. Replace `s.GetTimeout()` → `opts.GetTimeout()`.
8. **Remove** per-skill `SetArgs`/`SetArg`/`SetID`/`SetDescription`/`SetTimeout`/`SetNodeConfig`/`SetDryRun` overrides (they no longer exist on the interface). Keep `NewXxx()` constructors (construction-time fluent config stays).
9. Remove `WithNodeConfig`/`WithArgs`/`WithArg`/`WithDryRun`/`WithTimeout`/`WithBecomeUser` overrides that just delegated to `BaseSkill` (those `BaseSkill` methods are gone). Constructor-specific `With*` (e.g. a hypothetical `WithRetries`) stays.

### 4.3 Tests (must change)

| File | Change |
|------|--------|
| `skill_test.go` | Drop tests for removed `With*` execution setters; keep `WithID`/`WithDescription`. |
| `node_implementation_test.go` | Update `Run`/`RunByID`/`Check` call signatures; pass `ctx` + opts. |
| `inventory_implementation_test.go` | Same. |
| `group_implementation_test.go` | Same. |
| `command_implementation_test.go` | Same. |
| `registry_test.go` | Update any direct `Run`/`Check` invocations. |
| `integration_test.go` | Update end-to-end call sites. |
| `playbook_test.go` | Update. |
| All `skills/**/*_test.go` | Update constructor-only usage; any direct `Run()`/`Check()` calls now need `ctx`+opts. Many tests construct skills and assert on getters — those getter calls for execution state are removed and must be rewritten to assert via opts. |
| `skills/ping/ping_mock_test.go` | Update mock to new interface. |

### 4.4 Docs (must update)

`docs/skills.md`, `docs/advanced_usage.md`, `docs/idempotency.md`,
`docs/dry_run.md`, `docs/privilege_escalation.md`, `docs/commands.md`,
`docs/quick_start.md`, `docs/playbooks.md`, `docs/livewiki/**`,
`docs/proposals/**` (historical — leave as-is, they are dated proposals),
`README.md`, `examples/example_playbook.go`, `cmd/ork/**` (CLI call sites).

---

## 5. Detailed Design Decisions

### 5.1 Where options are built (the single mutation point)

The **only** place execution options are assembled is inside
`nodeImplementation`, right before invoking the skill. This replaces the
current "mutate the shared skill" pattern with "build a fresh opts value per call".

`node_implementation.go`:

```go
func (n *nodeImplementation) buildOpts(opts ...types.RunOption) types.RunnableOptionsInterface {
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
	merged = append(merged, opts...) // caller-supplied opts win (applied last)
	return types.NewRunnableOptions(merged...)
}

func (n *nodeImplementation) Run(ctx context.Context, skill types.RunnableInterface, opts ...types.RunOption) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}
	result := skill.Run(ctx, n.buildOpts(opts...))
	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed, Message: result.Message,
		Details: result.Details, Error: result.Error,
	}
	return results
}
```

Note: caller-supplied `opts` are appended **last** so they override node
defaults (matches today's `RunByID` precedence: skill opts > node args).

### 5.2 `RunByID` — registry still returns a shared instance, and that's now safe

Because `RunnableInterface` no longer has setters and the framework no longer
mutates the instance, the registry handing back the **same singleton pointer**
to concurrent callers is safe **at the framework level**. The instance now
holds only immutable construction config (`id`, `description`, and any
skill-specific builder state the author chose to freeze after `NewXxx()`).

`nodeImplementation.RunByID`:

```go
func (n *nodeImplementation) RunByID(ctx context.Context, id string, opts ...types.RunOption) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}
	reg, err := GetGlobalSkillRegistry()
	if err != nil { /* ... */ }
	skill, ok := reg.FindByID(id)
	if !ok { /* ... */ }
	result := skill.Run(ctx, n.buildOpts(opts...))
	results.Results[n.GetHost()] = /* ... */
	return results
}
```

**Author responsibility (documented):** skill structs must not carry mutable
*execution* state — execution state now lives in `RunnableOptionsInterface`,
which the framework builds fresh per call. Construction-time mutable state on
the skill struct is fine **only if** the author guarantees it is not mutated
after registration (the framework never mutates a registered instance, but a
skill that mutates its own struct fields during `Run` would reintroduce a
shared-state hazard across concurrent `RunByID` calls). We document this in
`docs/skills.md`.

**Optional hardening (deferred, §8):** offer an opt-in `Cloneable` interface
(`Clone() RunnableInterface`) that `RunByID` checks for and invokes per call,
for skills that genuinely need per-call private mutable state. Not required
for the built-in skills (none carry mutable execution state after this
migration).

### 5.3 `Check` config-propagation gap (secondary bug) — fixed for free

Today `nodeImplementation.Check()` only calls `SetDryRun` and never
`SetNodeConfig`/`SetBecomeUser`. In the new design, `Check` and `Run` both
receive the same `opts` (built via `buildOpts`), so `Check` automatically gets
the correct `NodeConfig` and `BecomeUser`. No separate fix needed.

```go
func (n *nodeImplementation) Check(ctx context.Context, skill types.RunnableInterface, opts ...types.RunOption) types.Results {
	results := types.Results{Results: make(map[string]types.Result)}
	changed, err := skill.Check(ctx, n.buildOpts(opts...))
	results.Results[n.GetHost()] = types.Result{
		Changed: changed,
		Message: fmt.Sprintf("check: changed=%v", changed),
		Error:   err,
	}
	return results
}
```

### 5.4 Dry-run source-of-truth resolution

Today dry-run lives in two places: `BaseSkill.dryRun` (via `SetDryRun`) and
`NodeConfig.IsDryRunMode`. After migration, `BaseSkill.dryRun` is gone. The
single source is `opts.IsDryRun()`, which `buildOpts` populates from
`n.cfg.IsDryRunMode` (overridable by caller `WithDryRun`). Skill bodies that
read `cfg.IsDryRunMode` should switch to `opts.IsDryRun()` for consistency,
**but** `cfg` (=`opts.GetNodeConfig()`) still carries `IsDryRunMode` for
`ssh.Run`'s internal dry-run handling. Resolution:

- `ssh.Run` keeps reading `cfg.IsDryRunMode` (it operates on `NodeConfig`).
- `buildOpts` sets `opts.dryRun = n.cfg.IsDryRunMode`, so `opts.IsDryRun()` and
  `cfg.IsDryRunMode` are always equal for a given call. Skill bodies may use
  either; we standardize on `opts.IsDryRun()` in skill logic and leave
  `cfg.IsDryRunMode` for `ssh.Run` plumbing.

### 5.5 `ctx` threading

All `Run`/`Check`/`RunCommand` paths gain a `ctx context.Context` first param.
For this change, `ctx` is accepted and forwarded into skill methods but is
**not yet plumbed into `ssh.Run`** (which currently takes only `NodeConfig` +
`Command`). Adding a ctx-aware `ssh.RunContext` is a follow-up (§9). For now,
skills that don't need ctx simply accept it and ignore it; the signature
change is the important part for future-proofing and cancellation.

`RunCommand` (which doesn't touch skills) also gains `ctx` for API
consistency, but may ignore it until `ssh.RunContext` lands.

### 5.6 `RunnerInterface` new signatures

```go
type RunnerInterface interface {
	types.BecomeInterface
	RunCommand(ctx context.Context, cmd string) types.Results
	Run(ctx context.Context, runnable types.RunnableInterface, opts ...types.RunOption) types.Results
	RunByID(ctx context.Context, id string, opts ...types.RunOption) types.Results
	Check(ctx context.Context, runnable types.RunnableInterface, opts ...types.RunOption) types.Results
	GetLogger() *slog.Logger
	SetLogger(logger *slog.Logger) RunnerInterface
	SetDryRunMode(dryRun bool) RunnerInterface
	GetDryRunMode() bool
}
```

### 5.7 Registry: keep instances (no factory change)

Decision: **keep** `types.Registry` storing `RunnableInterface` instances.
Rationale: `RunnableInterface` (the skill contract) no longer has execution
setters, so the framework never mutates a registered instance. The
concurrency hazard is eliminated by construction at the framework level
(goal #1): each call builds a fresh `RunnableOptionsInterface` and passes it
by value-semantics to the skill, so concurrent calls never share mutable
state. Switching to `func() RunnableInterface` factories would be a larger,
riskier change (every `NewXxx()` constructor signature/register call site
changes) for no safety gain. The factory approach remains documented as a
future option (§8) if per-call private skill state is ever needed.

---

## 6. Phased Implementation Plan

### Phase 0 — Preparation (no behavior change)
- [ ] Create `types/runnable_options.go` with `runnableOptions`, `RunOption`,
      `NewRunnableOptions`, all `With*` constructors, `GetArg` helper.
- [ ] Add compile-time `var _ RunnableOptionsInterface = (*runnableOptions)(nil)`.
- [ ] Add unit tests for `runnable_options.go` (immutability, option precedence,
      `GetArg` nil-safety).
- [ ] `go build ./...` must still pass (new file is additive).

### Phase 1 — Core type changes (breaks build; fix in same commit)
- [ ] Rewrite `types/runnable_interface.go`: new `RunnableInterface` (no
      setters, `ctx`+`opts` signatures), remove old `RunnableOptions` struct.
- [ ] Rewrite `types/base_skill.go`: strip to `id`/`description`; new stub
      `Check`/`Run` signatures; remove execution setters/with-ers; drop
      `BaseBecome` embedding.
- [ ] Update `types/become_interface.go` doc comment (no longer part of
      `RunnableInterface`).
- [ ] Update `runner_interface.go` signatures (§5.6).
- [ ] Update `node_interface.go` / `inventory_interface.go` (embed only;
      no new methods, but doc comments).

### Phase 2 — Orchestration layer (node/inventory/group/command)
- [ ] `node_implementation.go`: add `buildOpts`; rewrite `Run`/`RunByID`/`Check`/
      `RunCommand` to new signatures (§5.1, §5.3, §5.5).
- [ ] `inventory_implementation.go`: rewrite `Run`/`RunByID`/`Check`/`RunCommand`
      to thread `ctx` and call `n.Run(ctx, skill, opts...)`. No skill mutation.
- [ ] `group_implementation.go`: same as inventory.
- [ ] `command_implementation.go`: rewrite `Check`/`Run` to new signature; read
      `command`/`chdir`/`required` from struct, everything else from `opts`.
- [ ] `skill.go`: update `NewSkill()` doc example.
- [ ] `registry.go` (ork pkg): no change beyond import sanity.

### Phase 3 — Skill bodies (mechanical, one package per commit)
Commit per skill package to keep diffs reviewable:
- [ ] `skills/ping` (simplest — do first as the template)
- [ ] `skills/apt` (5 files)
- [ ] `skills/ufw` (11 files)
- [ ] `skills/user` (5 files)
- [ ] `skills/mariadb` (14 files)
- [ ] `skills/security` (5 files)
- [ ] `skills/fail2ban` (2 files)
- [ ] `skills/swap` (3 files)
- [ ] `skills/reboot` (1 file)

Each commit: apply §4.2 checklist, run `go build ./skills/<pkg>` + that
package's tests.

### Phase 4 — Tests
- [ ] `skill_test.go`: drop removed-setter tests; keep `WithID`/`WithDescription`.
- [ ] `node_implementation_test.go` / `inventory_implementation_test.go` /
      `group_implementation_test.go` / `command_implementation_test.go`: update
      call signatures; add a **concurrency regression test** that runs an
      arg-heavy skill across N nodes with `SetMaxConcurrency(N)` and asserts no
      cross-node arg leakage (this is the test that would have caught the
      original bug).
- [ ] `registry_test.go` / `integration_test.go` / `playbook_test.go`: update.
- [ ] All `skills/**/*_test.go`: update; replace direct `Run()`/`Check()` calls
      with `Run(ctx, opts)`/`Check(ctx, opts)`; replace execution-getter
      assertions with opts-based assertions.
- [ ] `skills/ping/ping_mock_test.go`: update mock interface.

### Phase 5 — Docs & examples
- [ ] `docs/skills.md`: new authoring guide (no setters; receive `opts`;
      immutability contract; how to add custom `RunOption`).
- [ ] `docs/advanced_usage.md`, `docs/quick_start.md`, `docs/commands.md`,
      `docs/dry_run.md`, `docs/idempotency.md`, `docs/privilege_escalation.md`,
      `docs/playbooks.md`: update call sites.
- [ ] `docs/livewiki/**`: regenerate/update API reference.
- [ ] `README.md`, `examples/example_playbook.go`, `cmd/ork/**`: update.
- [ ] Add `docs/proposals/implemented/2026-07-30-per-call-options.md`
      summarizing the shipped design (per `docs/proposals/README.md` convention).

### Phase 6 — Verification
- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...` (incl. the new concurrency regression test, ideally run
      with `-race`).
- [ ] `go test -race ./...` to confirm the data race is gone.
- [ ] Manual: run an example inventory with `SetMaxConcurrency > 1` and
      per-node args; confirm no leakage.

---

## 7. Risks & Mitigations

| ID | Risk | Mitigation |
|----|------|-----------|
| R1 | Breaking change for all external skill authors. | Bump major version; provide a migration guide in `docs/skills.md`; the per-skill change is mechanical (§4.2 checklist). |
| R2 | Skill authors storing mutable execution state on their struct. | Document the immutability contract; offer optional `Cloneable` (§9) later. The framework itself no longer mutates instances, so the framework-induced hazard is gone. |
| R3 | `BaseSkill` dropping `BaseBecome` embedding breaks a skill that sets become-user at construction. | Audit in Phase 3 (grep `WithBecomeUser\|SetBecomeUser` on skill structs, not nodes). Built-in skills don't do this. If any external skill does, they pass `WithBecomeUser` as a `RunOption` instead. |
| R4 | `ctx` not yet plumbed into `ssh.Run` — cancellation doesn't actually cancel in-flight SSH. | Accepted for this change; tracked as follow-up (§9). Signature is future-proofed. |
| R5 | `RunCommand` gaining `ctx` is an API break for users who call it directly. | Same major-version bump covers it; `RunCommand` is on `RunnerInterface` and the break is consistent with `Run`/`Check`. |
| R6 | Test count is large (~40 skill test files) — slow migration. | Per-package commits (Phase 3); many tests only construct skills and don't call `Run` directly, so changes are minimal. |
| R7 | `RunnableOptions` (old struct) is referenced by `RunByID(id, opts ...types.RunnableOptions)` everywhere. | Replaced by `...types.RunOption` variadic. Grep-and-replace in Phase 2/3. |

---

## 8. Out of Scope (follow-ups)

- **`ssh.RunContext`** — ctx-aware SSH execution with cancellation/timeout
  enforcement. This change accepts `ctx` but doesn't wire it through `ssh`.
- **`Cloneable` interface** for per-call skill cloning (only if a skill needs
  private mutable per-call state).
- **Timeout enforcement** — `opts.GetTimeout()` is available but not yet
  enforced by `ssh.Run`. Skills may opt to enforce it themselves.
- **Registry factory mode** (`func() RunnableInterface`) — not needed while
  skills carry no mutable execution state (the framework no longer mutates
  registered instances).
- **Deprecation/removal of `RunByID`** — keep the existing deprecation note;
  removal is a separate decision.

---

## 9. Open Questions

1. Should `RunCommand` also take `...RunOption` (e.g. for per-call become-user),
   or stay `(ctx, cmd string)`? **Proposal:** keep it `(ctx, cmd)` — commands
   don't have the args/options richness of skills; node-level config suffices.
2. Do we want a `DefaultRunnableOptions()` convenience that pre-fills logger
   from `slog.Default()`? **Proposal:** yes, `buildOpts` already falls back to
   `slog.Default()` via `NodeConfig.GetLoggerOrDefault()`; no separate default
   needed.
3. Should `Result` gain a `Context`-derived cancellation indicator? **Defer** —
   not needed until `ssh.RunContext` lands.

---

## 10. Acceptance Criteria

- `go build ./...`, `go vet ./...`, `go test ./...` all green.
- `go test -race ./...` green (the original data race is gone).
- New concurrency regression test passes: N nodes, `SetMaxConcurrency(N)`,
  distinct per-node args, no cross-node leakage, no fatal map-write panic.
- `nodeImplementation.Check` now uses the same `NodeConfig`/`BecomeUser` as
  `Run` (secondary bug closed — verified by a dedicated test).
- No remaining `SetNodeConfig`/`SetArgs`/`SetDryRun`/`SetBecomeUser` calls on
  `RunnableInterface` (the skill contract) anywhere in the tree — those setters
  now live only on `RunnableOptionsInterface` (grep confirms zero hits on
  skill instances; the framework builds opts via `RunOption`/setters on
  `RunnableOptionsInterface`).
- `docs/skills.md` documents the new authoring model and the immutability
  contract.
