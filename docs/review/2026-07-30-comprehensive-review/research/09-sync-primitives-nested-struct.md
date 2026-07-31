# Source: A Pattern for Synchronization Primitives in Complex Go Structures — jen20.dev

- **URL:** https://jen20.dev/post/a-pattern-for-sync-primitives-in-complex-go-structs/
- **Author:** James Nugent
- **Published:** 2025-03-01
- **Accessed:** 2026-07-31
- **Category:** Go patterns / sync.RWMutex / struct design

## Relevance to ork decision

This is the canonical pattern for protecting multiple fields in a Go struct
with mutexes. It's directly relevant to how `BaseSkill` should be structured
if we add mutex protection (plan 7's approach, or a hypothetical "plan 8: add
RWMutex to BaseSkill without omni").

The pattern: **embed the mutex in a nested struct alongside the fields it
protects**, so it's visually obvious which lock guards which fields.

## Key excerpts

> It's common in Go to have a `struct` that contains many pieces of state,
> some of which need protecting by different synchronization primitives.

> While it's possible that Go will one day (like Rust) have generic versions
> of `sync.Mutex` and `sync.RWMutex` which prevent such misuse, there is a
> useful pattern [...] use a nested struct with an embedded field which is
> the synchronization primitive and the fields it is supposed to protect.

### The pattern

```go
type Raft struct {
    lastContact struct {
        sync.RWMutex
        time time.Time
    }

    leader struct {
        sync.RWMutex
        addr ServerAddress
        id   ServerID
    }

    observers struct {
        sync.RWMutex
        value map[uint64]*Observer
    }
}
```

> Written like this, it's obvious which fields are protected by which
> synchronization primitive, and the code reads nicely at call sites:

```go
func getLastContactTime() time.Time {
    raft.lastContact.RLock()
    defer raft.lastContact.RUnlock()
    return raft.lastContact.time
}
```

## Implications for ork

- **If we were to add mutexes to `BaseSkill` directly** (without omni), this
  nested-struct pattern would be the idiomatic way:
  ```go
  type BaseSkill struct {
      config struct {
          sync.RWMutex
          nodeConfig NodeConfig
          dryRun     bool
          becomeUser string
      }
      args struct {
          sync.RWMutex
          data map[string]string
      }
  }
  ```
  This is verbose and error-prone — every getter/setter must remember to
  Lock/Unlock the right nested struct.

- **Plan 7 (omni) avoids this entirely:** `omni.Atom` already implements this
  pattern internally (one `sync.RWMutex` protecting `properties` and
  `children`). We get the jen20.dev pattern for free, tested, without writing
  it ourselves.

- **Plans 3-6 (clone-based):** don't need this pattern because they don't
  share the struct — each goroutine has its own clone. But they also don't
  get the defense-in-depth benefit.

- **This pattern is an argument FOR plan 7** (or a hypothetical "add RWMutex
  to BaseSkill" plan): it shows the idiomatic Go way to protect shared state,
  and omni implements it for us.
