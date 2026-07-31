# Source: brunoga/deep v5 — High-Performance Type-Safe Synchronization Toolkit

- **URL:** https://github.com/brunoga/deep
- **PkgDoc:** https://pkg.go.dev/github.com/brunoga/deep/v5
- **Accessed:** 2026-07-31
- **Category:** Go library / deep copy / cloning

## Relevance to ork decision

`brunoga/deep` is the library proposed in `plan-deepcopy.md` (plan 3). This
source reveals that **deep v5 has changed dramatically** from what the plan
assumed. v5 is no longer a simple deep-copy library — it's a "synchronization
toolkit" with code generation, patches, CRDTs, and HLC clocks. The simple
`deep.Copy()` API of v4 has been replaced by a patch-based model.

This is a **critical finding**: plan 3 was written assuming `deep.Copy(x)`
from v4. v5's API is fundamentally different.

## Key excerpts

> `deep` is a comprehensive Go library for comparing, cloning, and
> synchronizing complex data structures. Deep introduces a revolutionary
> architecture centered on **Code Generation** and **Type-Safe Selectors**,
> delivering up to **15x** performance improvements over traditional
> reflection-based libraries.

### Performance (v5 generated vs v4 reflection)

> | Operation | v4 (Reflection) | Deep (Generated) | Speedup |
> | :--- | :--- | :--- | :--- |
> | **Apply Patch** | 726 ns/op | **50 ns/op** | **14.5x** |
> | **Diff + Apply** | 2,391 ns/op | **270 ns/op** | **8.8x** |
> | **Clone** | 1,872 ns/op | **290 ns/op** | **6.4x** |
> | **Equality** | 202 ns/op | **84 ns/op** | **2.4x** |

### Architecture change

> v4 used a **Recursive Tree Patch** model. Every field was a nested patch
> object. While flexible, this caused high memory allocations and made
> serialization difficult.
>
> Deep uses a **Flat Operation Model**. A patch is a simple slice of
> `Operations`. This makes patches:
> 1. Portable: Trivially serializable to any format.
> 2. Fast: Iterating a slice is much faster than traversing a tree.

## Implications for ork

1. **Plan 3 needs revision.** `plan-deepcopy.md` was written assuming
   `deep.Copy(x)` from v4. v5's API is patch-based, not a simple `Copy`
   function. The plan's code examples may not work with v5.

2. **v5 requires code generation** (`deep-gen`) for maximum performance. Without
   generated code, it falls back to reflection (slower). This adds build
   complexity — a `go generate` step.

3. **v5 is over-engineered for ork's needs.** ork just needs "clone this
   struct." v5 provides diffing, patching, CRDTs, HLC clocks, JSON Patch
   export — none of which ork needs. This is dependency bloat.

4. **The `unsafe` concern from plan 3 remains.** Even in v5, the reflection
   fallback path uses `unsafe` to access unexported fields. ork's `BaseSkill`
   has unexported fields — deep would need `unsafe` to copy them.

5. **Plan 7 (omni) is simpler:** omni is a small, focused library (one file,
   ~500 lines) that does exactly what ork needs (map-backed state with
   serialization). No code generation, no CRDTs, no patches.

6. **If plan 3 is still desired, it should use v4 (`github.com/brunoga/deep`
   v4), not v5.** But v4 is slower (1,872 ns/op for clone) and uses
   `unsafe`. Plan 7's ToMap+FromMap (~2-5µs) is comparable and has no
   `unsafe`.
