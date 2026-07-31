# Source: Data Race Detector — The Go Programming Language

- **URL:** https://go.dev/doc/articles/race_detector
- **Author:** The Go Authors
- **Accessed:** 2026-07-31
- **Category:** Go official docs / Concurrency safety

## Relevance to ork decision

This is the canonical Go documentation on data races. It defines what a data
race is, how the `-race` detector works, and shows the **exact pattern we have
in ork**: unprotected global/shared state accessed from multiple goroutines.

The "Unprotected global variable" section is a near-verbatim description of
ork's `BaseSkill` problem: a shared map/struct accessed by `RegisterService`
(setters) and `LookupService` (getters) from multiple goroutines without a
mutex. The prescribed fix is a `sync.Mutex` (or `sync.RWMutex`).

## Key excerpts

> A data race occurs when two goroutines access the same variable concurrently
> and at least one of the accesses is a write.

### Unprotected global variable (the ork pattern)

```go
var service map[string]net.Addr

func RegisterService(name string, addr net.Addr) {
    service[name] = addr          // write
}

func LookupService(name string) net.Addr {
    return service[name]          // read
}
```

> If the following code is called from several goroutines, it leads to races on
> the `service` map. Concurrent reads and writes of the same map are not safe.

**Fix prescribed by Go docs:**

```go
var (
    service   map[string]net.Addr
    serviceMu sync.Mutex
)

func RegisterService(name string, addr net.Addr) {
    serviceMu.Lock()
    defer serviceMu.Unlock()
    service[name] = addr
}

func LookupService(name string) net.Addr {
    serviceMu.Lock()
    defer serviceMu.Unlock()
    return service[name]
}
```

### Primitive unprotected variable

> Data races can happen on variables of primitive types as well (`bool`, `int`,
> `int64`, etc.) [...] Even such "innocent" data races can lead to
> memory corruption or invalid program behavior.

This confirms that ork's `SetDryRun(bool)` and `SetNodeConfig(struct)` on a
shared `BaseSkill` are both racy — not just the `map[string]string` args.

### Race on loop counter / accidentally shared variable

These sections confirm that even closure-captured variables and loop variables
race if not copied per-goroutine. This validates the **clone-before-mutate**
approach: each goroutine must have its own copy of the skill state.

## Implications for each plan

- **Plans 3-7 (clone-based):** directly implement the "make a copy" fix shown
  in the "Race on loop counter" section (`go func(j int)` — pass a copy).
- **Plan 7 (omni):** adds the `sync.RWMutex` from the "Unprotected global
  variable" fix as **defense-in-depth** on top of the clone.
- **Plans 1-2 (opts-based):** eliminate the shared mutable state entirely —
  the strongest fix per Go's own guidance ("make the code safe, protect the
  accesses with a mutex" or restructure so there's no shared write).
