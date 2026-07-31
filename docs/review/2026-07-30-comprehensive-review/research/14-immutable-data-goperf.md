# Source: Immutable Data Sharing — Go Optimization Guide (goperf.dev)

- **URL:** https://goperf.dev/01-common-patterns/immutable-data/
- **Accessed:** 2026-07-31
- **Category:** Go performance / immutability / atomic.Pointer / config snapshots

## Relevance to ork decision

This is the Go performance team's guidance on immutable data sharing. It
shows the `atomic.Pointer[Config]` pattern for lock-free config reads — an
alternative to RWMutex that's even faster for read-heavy workloads. Relevant
to evaluating whether plan 7's RWMutex is the right primitive, or whether
`atomic.Pointer` would be better.

## Key excerpts

> A powerful alternative is immutable data sharing. Instead of protecting
> data with locks, you design your system so that shared data is never
> mutated after it's created. This minimizes contention and simplifies
> reasoning about your program.

> - **No locks needed**: Multiple goroutines can safely read immutable data
>   without synchronization.
> - **Easier reasoning**: If data can't change, you avoid entire classes of
>   race conditions.
> - **Copy-on-write optimizations**: You can create new versions of a
>   structure without altering the original.

### Defensive copies for maps

> Maps and slices in Go are reference types. Even if the Config struct isn't
> changed, someone could accidentally mutate a shared map. To prevent this,
> we make defensive copies:

```go
func NewConfig(logLevel string, timeout time.Duration, features map[string]bool) *Config {
    copiedFeatures := make(map[string]bool, len(features))
    for k, v := range features {
        copiedFeatures[k] = v
    }
    return &Config{
        LogLevel:  logLevel,
        Timeout:   timeout,
        Features:  copiedFeatures,
    }
}
```

### Atomic swapping

```go
var currentConfig atomic.Pointer[Config]

func LoadInitialConfig() {
    cfg := NewConfig("info", 5*time.Second, map[string]bool{"beta": true})
    currentConfig.Store(cfg)
}

func GetConfig() *Config {
    return currentConfig.Load()
}
```

> Now all goroutines can safely call `GetConfig()` with no locks. When the
> config is reloaded, you just `Store` a new immutable copy.

## Implications for ork

- **The ideal pattern for ork would be `atomic.Pointer[BaseSkill]`** — but
  this requires `BaseSkill` to be truly immutable (no setters). That's plans
  1-2's approach (no execution setters on the skill). Plans 3-7 keep setters,
  so they can't use `atomic.Pointer` alone — they need the clone.

- **Plan 7's `omni.Atom` uses `sync.RWMutex`, not `atomic.Pointer`.** This is
  the right choice because:
  1. The skill IS mutated (by setters) — it's not immutable.
  2. `atomic.Pointer` requires immutability; RWMutex allows controlled
     mutation.
  3. The clone-before-mutate pattern creates local immutability (the clone
     is never shared), which is the goperf.dev pattern applied per-goroutine.

- **Defensive copy of maps:** goperf.dev explicitly says to copy maps to
  prevent shared mutation. Plan 7's `ToMap()` does exactly this — it creates
  a new map copy. Plan 6 (map-storage) also does this. Plans 3-4 (deep
  copy / serialize) do this implicitly.

- **A future optimization:** if ork ever moves to immutable skills (plans 1-2
  style), we could replace the RWMutex + clone with `atomic.Pointer[Skill]`
  for lock-free reads. But that's a future direction, not the current fix.
