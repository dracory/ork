# Source: golang-design/reflect — Generic DeepCopy for Go (proposal 51520)

- **URL:** https://github.com/golang-design/reflect
- **Accessed:** 2026-07-31
- **Category:** Go library / deep copy / Go proposal

## Relevance to ork decision

This is the external implementation of Go proposal #51520 for a standard
library `DeepCopy` generic function. It's relevant because:
1. It shows what a future stdlib deep copy might look like.
2. It documents the **stateful object problem** — `sync.Mutex`, `os.File`,
   `net.Conn` should NOT be deep copied, which is a concern for ork if
   `BaseSkill` ever gains a mutex.

## Key excerpts

> Package reflect provides a generic `DeepCopy` for Go, the external
> implementation of proposal go.dev/issue/51520.

> `DeepCopy` copies `src` to a freshly allocated value of the same type and
> returns it:
> - Numbers, bools and strings are copied and have a different underlying
>   memory address.
> - Slices and arrays are deeply copied, including their elements.
> - Maps are deeply copied for all keys and values.
> - Pointers are deeply copied for the pointed value, and the copy points to
>   the deeply copied value.
> - Structs are deeply copied for all fields, including unexported ones.
> - Interfaces are deeply copied if the underlying type can be deeply copied.

### The stateful object problem

> Stateful objects (`sync.Mutex`, `os.File`, `net.Conn`, `js.Value`, ...)
> and singletons should usually not be copied by their memory representation.
> Use the following options to override the default behavior per type:

```go
dst := reflect.DeepCopy(src,
    // Keep singleton shared by reference instead
    reflect.RetainType[*Config](),
    // Reset a stateful value to its zero value (e.g. an unlocked mutex).
    reflect.ZeroType[sync.Mutex](),
)
```

### Options

> - `DisallowCopyUnexported()` skips unexported struct fields instead of
>   copying them.
> - `DisallowCopyCircular()` panics on circular structures instead of
>   handling them.
> - `DisallowCopyBidirectionalChan()` keeps the source channel instead of
>   creating a new one.
> - `DisallowType[T]()` panics if a value of type T is encountered.

## Implications for ork

1. **The stateful object problem is critical for plan 7.** If `BaseSkill`
   stores state in `omni.Atom` (which contains a `sync.RWMutex`), then deep
   copying the `BaseSkill` struct directly would copy the mutex — which is
   a `go vet` error and a potential deadlock source. **This is why plan 7
   uses ToMap/FromMap (serialization) instead of struct deep copy** — the
   map doesn't include the mutex, so the clone gets a fresh Atom with a
   fresh mutex.

2. **Plan 3 (brunoga/deep) has the same problem.** If `BaseSkill` ever gets
   a mutex (for defense-in-depth), `deep.Copy()` would copy it. The v4
   library doesn't have `ZeroType[sync.Mutex]()` like this proposal does.
   Plan 3 would need to exclude the mutex manually.

3. **This proposal is not yet in the stdlib** — it's issue #51520, still
   under discussion. ork can't rely on it. But it shows the Go community
   recognizes the need for a standard deep copy solution.

4. **`DisallowCopyUnexported()` is relevant to ork.** ork's `BaseSkill` has
   unexported fields. A deep copy library that copies unexported fields
   needs `unsafe` (to bypass Go's visibility rules). A library that skips
   them loses data. **Plan 7 avoids this entirely** — ToMap/FromMap only
   copies what's in the map (exported via getters), no `unsafe` needed.

5. **The proposal confirms that serialization-based cloning (plan 7's
   approach) is safer than struct-based deep copy (plan 3)** for objects
   with mutexes or other stateful fields.
