# Source: dracory/omni — Universal Go module for composable primitives

- **URL:** https://github.com/dracory/omni
- **README:** https://raw.githubusercontent.com/dracory/omni/main/README.md
- **Source (atom.go):** https://raw.githubusercontent.com/dracory/omni/main/atom.go
- **Source (interfaces.go):** https://raw.githubusercontent.com/dracory/omni/main/interfaces.go
- **Source (constructors.go):** https://raw.githubusercontent.com/dracory/omni/main/constructors.go
- **Accessed:** 2026-07-31
- **Category:** Go library / the library proposed in plan 7

## Relevance to ork decision

This IS the library proposed in plan 7. This file documents what omni
provides, based on reading the actual source code (not just the README).

## What omni provides

### Atom — the core type

```go
type Atom struct {
    id         string
    atomType   string
    properties map[string]string
    children   []AtomInterface
    mu         sync.RWMutex  // <-- built-in thread safety
}
```

- **`map[string]string` properties** — all state stored as strings
- **`sync.RWMutex`** — every Atom is thread-safe by default
- **`children []AtomInterface`** — hierarchical (tree) structure
- **`id` and `atomType`** — special properties (not in the properties map)

### AtomInterface — the contract

```go
type AtomInterface interface {
    GetID() string
    SetID(id string) AtomInterface
    GetType() string
    SetType(atomType string) AtomInterface

    // Property access (string key-value)
    Get(key string) string
    Has(key string) bool
    Remove(key string) AtomInterface
    Set(key, value string) AtomInterface
    GetAll() map[string]string
    SetAll(properties map[string]string) AtomInterface

    // Children management
    ChildAdd(child AtomInterface) AtomInterface
    ChildDeleteByID(id string) AtomInterface
    ChildFindByID(id string) AtomInterface
    ChildrenAdd(children []AtomInterface) AtomInterface
    ChildrenFindByType(atomType string) []AtomInterface
    ChildrenGet() []AtomInterface
    ChildrenSet(children []AtomInterface) AtomInterface
    ChildrenLength() int

    RecursiveFindByID(id string) AtomInterface

    // Serialization (ALL built-in, no custom code needed)
    ToMap() map[string]any
    ToJSON() (string, error)
    ToJSONPretty() (string, error)
    ToGob() ([]byte, error)

    MemoryUsage() int
}
```

### Functional options (constructors)

```go
omni.NewAtom("skill",
    omni.WithID("user-create"),
    omni.WithProperties(map[string]string{
        "description": "Create a user",
        "dryRun":      "true",
    }),
    omni.WithChildren(childAtom1, childAtom2),
)
```

### Serialization — all built-in

- **`ToMap() map[string]any`** — returns `{"id":..., "type":..., "properties":{...}, "children":[...]}`
- **`ToJSON() (string, error)`** — JSON string
- **`ToGob() ([]byte, error)`** — binary gob encoding
- **`NewAtomFromMap(m map[string]any)`** — reconstruct from map
- **`NewAtomFromJSON(s string)`** — reconstruct from JSON
- **`FromGob(data []byte)`** — reconstruct from gob

### Thread safety — every method is protected

Every getter uses `mu.RLock()` / `mu.RUnlock()`:
```go
func (a *Atom) Get(key string) string {
    a.mu.RLock()
    defer a.mu.RUnlock()
    if a.properties == nil { return "" }
    return a.properties[key]
}
```

Every setter uses `mu.Lock()` / `mu.Unlock()`:
```go
func (a *Atom) Set(key, value string) AtomInterface {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.properties == nil {
        a.properties = make(map[string]string)
    }
    a.properties[key] = value
    return a
}
```

`ToMap()` copies under read lock (defensive copy):
```go
func (a *Atom) ToMap() map[string]interface{} {
    a.mu.RLock()
    defer a.mu.RUnlock()
    props := make(map[string]string, len(a.properties))
    for k, v := range a.properties {
        if k != "id" && k != "type" {
            props[k] = v
        }
    }
    // ... build result map with id, type, children
    return result
}
```

### Dependencies

- **Zero external dependencies** — uses only Go stdlib (`sync`, `encoding/json`,
  `encoding/gob`, `bytes`, `fmt`)
- **One internal dependency:** `github.com/dracory/uid` (for ID generation)
- **Go 1.24+** required
- **License:** MIT (README says MIT, LICENSE file says AGPLv3 — needs
  verification before use)

### Benchmarks (from README)

```
BenchmarkAtom_Get-8           1000000000   0.000001 ns/op
BenchmarkAtom_Set-8           500000000    0.000003 ns/op
BenchmarkAtom_ToJSON-8        2000000      750 ns/op
BenchmarkAtom_ToGob-8         1000000      1200 ns/op
```

## Implications for ork

### What plan 7 gets for free

| Need | omni provides | Custom code needed |
|------|--------------|-------------------|
| Map-backed state | `Atom.properties map[string]string` | None |
| Thread-safe access | `sync.RWMutex` in every Atom | None |
| `ToMap()` | `Atom.ToMap() map[string]any` | None |
| `FromMap()` | `NewAtomFromMap(m)` | None |
| JSON serialization | `Atom.ToJSON() / NewAtomFromJSON()` | None |
| Gob serialization | `Atom.ToGob() / FromGob()` | None |
| Functional options | `WithID, WithProperties, WithType, WithChildren` | None |
| Children/hierarchy | `ChildAdd, ChildrenGet, ChildrenFindByType` | None |
| Property get/set | `Get(key), Set(key, value), Has(key), Remove(key)` | None |
| Bulk property ops | `GetAll(), SetAll(map)` | None |
| Memory profiling | `MemoryUsage()` | None |
| Defensive copy on read | `ToMap()` copies under RLock | None |

### Concerns

1. **`map[string]string` only** — `NodeConfig` (a struct) must be JSON-serialized
   to a string before storing. This is a minor inconvenience (~10 lines of
   helper code) but not a blocker.

2. **License — RESOLVED.** omni is AGPLv3 (confirmed from LICENSE file).
   ork is also AGPLv3 (confirmed from ork's LICENSE file). Both are from the
   same `dracory` GitHub organization and reference the same commercial
   license contact (lesichkov.co.uk). **AGPLv3 is compatible with AGPLv3 —
   no license issue.** ork can use omni without restriction. (The README's
   mention of "MIT" was an error in omni's README; the LICENSE file is
   authoritative.)

3. **Small library** — 1 star, 0 forks, 36 commits. Not widely battle-tested.
   But the codebase is small (~500 lines), zero-dependency, and can be vendored
   or audited easily.

4. **Go 1.24+ required** — ork must verify it's on Go 1.24+.

5. **`ToMap()` returns `map[string]any`** with nested structure
   (`{"id":..., "type":..., "properties":{...}, "children":[...]}`). The
   framework's clone code must navigate this structure to modify properties.
   Slightly more complex than plan 6's flat `map[string]any`.
