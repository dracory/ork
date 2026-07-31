# Source: Why Concurrent Map Writes Crash Go Programs: Root Cause and Fix — KruN

- **URL:** https://krun.pro/go-concurrent-map-writes/
- **Accessed:** 2026-07-31
- **Category:** Concurrency / Go runtime internals / map implementation

## Relevance to ork decision

Explains the Go runtime internals: WHY concurrent map writes are fatal (not
panic). The `hashWriting` flag in `hmap` + `runtime.throw()` (not
`runtime.panic()`) means `recover()` is structurally impossible. Also notes
that a mutex on a **value receiver** copies the lock — a subtle trap relevant
to ork's interface design.

## Key excerpts

> Concurrent map writes in Go are one of the most common causes of unexpected
> process crashes in production systems. Unlike typical runtime panics, this
> error is not recoverable and cannot be handled with `recover()`.

> The message `fatal error: concurrent map writes` appears when the internal
> state of a Go map becomes unsafe due to unsynchronized access. Instead of
> triggering a standard panic, the runtime calls `throw()`, which stops the
> entire process to prevent memory corruption and undefined behavior.

### Root cause in the runtime

> Go maps use a `hashWriting` flag in the `hmap` struct; if two goroutines
> trip it simultaneously, the runtime calls `throw()` — process-level fatal,
> not goroutine-level panic.
>
> `recover()` does not intercept `throw()`. Wrapping map writes in
> `defer/recover` does nothing against this error.

### The value-receiver mutex trap

> A mutex on a value receiver copies the lock — the map is still unprotected.
> You need a pointer receiver or an external lock around the map.

### Primitive recommendation

> `sync.Map` eliminates the error structurally but trades off read
> performance; for write-heavy workloads, `sync.RWMutex` with a pointer
> receiver wins.

## Implications for ork

1. **Value-receiver trap:** ork's `BaseSkill` methods use pointer receivers
   (`func (b *BaseSkill) SetArgs(...)`), so this specific trap doesn't apply.
   But if any skill accidentally uses a value receiver for a method that
   touches the map, the mutex (if we add one) would be copied and useless.
   Plan 7's `omni.Atom` uses pointer receivers throughout — safe.

2. **`sync.RWMutex` with pointer receiver wins over `sync.Map`** for
   write-heavy workloads. ork's skill execution writes (SetArgs,
   SetNodeConfig) happen once per node call — moderately write-heavy. Plan 7's
   `omni.Atom` uses `sync.RWMutex` on a pointer receiver — the recommended
   approach per this source.

3. **Structural elimination vs. defense-in-depth:** `sync.Map` eliminates the
   error *structurally* (no `hashWriting` flag to trip). `sync.RWMutex`
   prevents it *by discipline* (all access goes through Lock/Unlock). Plan 7
   uses RWMutex — strong, but not as structural as sync.Map. However,
   `sync.Map` has poor type safety (`any` keys/values) and is optimized for
   read-heavy/stable-key workloads, not ork's pattern. RWMutex is the right
   choice for ork.
