# Source: CONCURRENCY_GUIDELINES.md — GoCodeAlone/modular framework

- **URL:** https://github.com/GoCodeAlone/modular/blob/main/CONCURRENCY_GUIDELINES.md
- **Accessed:** 2026-07-31
- **Category:** Go concurrency patterns / framework design guidelines

## Relevance to ork decision

This is a real-world framework's codified concurrency guidelines. It
explicitly lists the patterns we're choosing between: RWMutex for
read-heavy/write-light, atomic pointer swap for snapshots, and **defensive
deep copy for caller-provided config maps**. This directly validates the
clone-before-mutate approach (plans 3-7) and the RWMutex defense-in-depth
(plan 7).

## Key excerpts

### Core principles

> 1. **Safety First**: Code MUST pass `go test -race` across core, modules,
>    examples, and CLI. Any race is a release blocker.
> 2. **Clarity Over Cleverness**: Prefer simple, easily audited
>    synchronization over intricate lock-free or channel gymnastics unless a
>    measurable performance need is proven.
> 3. **Immutability by Construction**: When feasible, construct immutable
>    snapshots (config, slices, maps, request bodies) and share read-only.
> 4. **Encapsulation**: Internal goroutines own their state; external callers
>    interact via explicit update / retrieval APIs instead of mutating shared
>    maps or slices directly.
> 5. **Minimize Lock Scope**: Hold locks only around mutation or snapshot
>    creation—never across blocking I/O or user callbacks.

### Primitive selection table

> | Concern | Preferred Primitive | Rationale |
> |---------|---------------------|-----------|
> | Multiple readers, infrequent writers | `sync.RWMutex` | Cheap uncontended reads; explicit write exclusion |
> | Single-owner background goroutine publishing snapshots | Atomic pointer swap to immutable struct/map | Zero-copy read, no per-read locking |
> | Bounded append-only event capture in tests | Mutex around slice | Simplicity |
> | Parallel fan-out needing shared input body | Pre-buffer into `[]byte` + pass slice | Eliminates per-goroutine `*http.Request` body races |
> | **Config maps provided by caller** | **Defensive deep copy under lock** | **Prevents external mutation races** |

### Snapshot-then-fan-out pattern

```go
func (s *Subject) notify(evt Event) {
    // Snapshot under read lock, then fan out without holding lock
    s.mu.RLock()
    observers := make([]Observer, len(s.observers))
    copy(observers, s.observers)
    s.mu.RUnlock()

    for _, o := range observers {
        o.OnEvent(evt)
    }
}
```

## Implications for ork

1. **"Config maps provided by caller → Defensive deep copy under lock"** —
   this is exactly the ork pattern. The framework provides a skill (config
   map), and multiple goroutines (nodes) need to use it. The guideline says:
   **defensive deep copy**. This validates plans 3-7 (clone-based).

2. **"Multiple readers, infrequent writers → sync.RWMutex"** — ork's skill
   is read by N node goroutines and written by the framework once (before
   execution). This is the RWMutex sweet spot. Validates plan 7's use of
   `omni.Atom` (which has a built-in RWMutex).

3. **"Immutability by Construction"** — the strongest pattern. Plans 1-2
   achieve this (no shared mutable state). Plans 3-7 approximate it (clone
   to create local immutability). Plan 7 gets closest to the ideal with both
   clone + mutex.

4. **"Minimize Lock Scope — never across blocking I/O"** — important for plan
   7: the RWMutex in `omni.Atom` is held only during `Get`/`Set` (microseconds),
   never during `Run()` (which does SSH I/O). The clone happens before `Run()`,
   so no lock is held during network operations. Plan 7 satisfies this
   guideline.

5. **The snapshot-then-fan-out pattern** is exactly what ork should do:
   ```
   RLock → copy skill state → RUnlock → fan out to N goroutines, each with its own copy
   ```
   Plan 7's `ToMap()` is the "snapshot under read lock" step. The clone is
   the "copy" step. `Run()` is the "fan out without holding lock" step.
