# Source: [BUG] Systematic data races in dubbo-go — 44 race conditions

- **URL:** https://github.com/apache/dubbo-go/issues/3247
- **Accessed:** 2026-07-31
- **Category:** Real-world bug report / registry pattern / concurrent map access

## Relevance to ork decision

This is a real-world bug report from Apache dubbo-go that describes the
**exact same pattern as ork's bug**: 20+ global maps accessed without
synchronization, causing `fatal error: concurrent map read and map write`.
The recommended fix is a generic `Registry[T]` container with built-in
`sync.RWMutex` — which is essentially what `omni.Atom` provides (plan 7).

## Key excerpts

> A comprehensive audit of the dubbo-go codebase reveals **44 data race
> conditions** across the framework. The most critical is a **systematic
> pattern** in the `common/extension/` package where 20+ global maps are
> accessed without any synchronization, plus a confirmed bug where
> `SetProviderService` uses the **wrong lock variable**. These issues will
> cause `fatal error: concurrent map read and map write` crashes in
> production.

### The exact ork pattern

> **Worst cases**: `GetDefaultConfigReader()` and `GetRouterFactories()`
> **return the internal map reference directly** — callers iterate the
> returned map without any lock, while `Set*` functions write to the same
> map concurrently. This is guaranteed to crash.

This is ork's pattern: `GetArgs()` returns the internal `map[string]string`
reference, and `SetArgs()` writes to it. Concurrent calls = crash.

### The recommended fix — generic Registry with RWMutex

> **Recommended fix**: Introduce a generic `Registry[T]` container with
> built-in `sync.RWMutex`:

```go
type Registry[T any] struct {
    mu    sync.RWMutex
    items map[string]T
}

func (r *Registry[T]) Set(name string, v T) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[name] = v
}

func (r *Registry[T]) Get(name string) (T, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.items[name]
    return v, ok
}

func (r *Registry[T]) Snapshot() map[string]T {
    r.mu.RLock()
    defer r.mu.RUnlock()
    m := make(map[string]T, len(r.items))
    for k, v := range r.items {
        m[k] = v
    }
    return m
}
```

> This would fix all 20+ files with a single refactor.

### The wrong-lock bug (Category 2)

> `SetProviderService` uses `conLock` (consumer lock) to protect writes to
> `providerServices`, but `loadProvider()` uses `proLock.RLock()` to protect
> reads of the same map. **Two different locks protecting the same map = no
> protection at all.**

## Implications for ork

1. **ork's bug is not unique** — Apache dubbo-go had the same systematic
   pattern (20+ unprotected maps) in production. This validates the Critical
   severity rating.

2. **The recommended fix is a `Registry[T]` with `sync.RWMutex`** — which is
   exactly what `omni.Atom` provides. `omni.Atom` is a `Registry[string]`
   (key-value store) with `sync.RWMutex` built in. Plan 7 adopts this fix
   directly.

3. **The `Snapshot()` method** in the recommended fix is exactly `ToMap()` in
   omni — it copies the map under read lock and returns a safe snapshot. This
   is the clone-before-mutate pattern, implemented with the mutex as
   defense-in-depth.

4. **The wrong-lock bug** is a cautionary tale: if ork adds mutexes manually
   (without omni), a developer could accidentally use the wrong mutex for a
   field. Plan 7's `omni.Atom` has a single mutex protecting all properties —
   no possibility of wrong-lock bugs.

5. **"Fix all 20+ files with a single refactor"** — the dubbo-go fix is a
   single generic container. Plan 7 is similar: a single `omni.Atom` in
   `BaseSkill` fixes all skill types at once. No per-skill changes needed.
