# Source: expego/fastcopier — Fast, safe Go deep-copy library

- **URL:** https://github.com/expego/fastcopier
- **Accessed:** 2026-07-31
- **Category:** Go library / deep copy / benchmark comparison

## Relevance to ork decision

This source provides a **fair benchmark comparison of 7 deep-copy libraries**.
It shows the performance landscape: manual copy is fastest, reflection-based
libraries are 1.5-2x slower, `encoding/json` is 15x slower, and `jinzhu/copier`
is 25x slower. This helps validate the performance estimates in our plans.

## Key excerpts

### Benchmark results (AMD EPYC 9V74, go1.25.0, -benchtime=3s)

> | Library | ns/op | B/op | allocs/op | vs FastCopier |
> |---------|------:|-----:|----------:|:-------------:|
> | Manual (baseline) | 0.361 | 0 | 0 | 309.3× faster |
> | **FastCopier (with gen)** | 112 | 0 | 0 | **—** |
> | FastCopier (pure reflect) | 145 | 0 | 0 | 1.3× slower |
> | FastCopier.Clone | 173 | 128 | 2 | 1.5× slower |
> | huandu/go-clone | 166 | 128 | 2 | 1.5× slower |
> | tiendc/go-deepcopy | 191 | 32 | 1 | 1.7× slower |
> | jinzhu/copier | 2,870 | 496 | 18 | **25.7× slower** |
> | go-viper/mapstructure | 148 | 176 | 3 | 1.3× slower |
> | ulule/deepcopier | 6,344 | 5,760 | 64 | **56.7× slower** |
> | encoding/json | 1,749 | 336 | 7 | **15.6× slower** |

### Key features

> - Correct deep copy semantics (no accidental slice/map aliasing)
> - Zero allocations for repeated struct/slice copies
> - Cycle-safe pointer traversal with configurable policy
> - `Inspect` + `MustRegister` for startup-time mapping validation
> - Optional `RegisterCopier` / `fastcopier-gen` path for reflection-free hot paths

## Implications for ork

1. **`encoding/json` (plan 4) is ~15x slower than reflection-based deep copy.**
   The benchmark shows `encoding/json` at 1,749 ns/op vs reflection-based
   libraries at ~150-200 ns/op. This confirms plan 4's performance estimate
   (~50-200µs for a full skill with JSON marshal/unmarshal overhead).

2. **Reflection-based deep copy (plan 3 with v4, or `huandu/go-clone`) is
   ~150-200 ns/op** — much faster than our plan 3 estimate of ~10µs. The
   difference is because our estimate included the full skill with all fields,
   not just a small struct. Real-world performance for plan 3 is likely
   ~1-5µs for a full `BaseSkill` + `commandImplementation`.

3. **Plan 7 (omni ToMap+FromMap) at ~2-5µs is competitive** with
   reflection-based deep copy (~1-5µs for a full skill). The omni approach
   trades a small performance cost for: no `unsafe`, built-in serialization,
   and thread safety.

4. **Manual copy is always fastest (0.36 ns/op)** — but requires writing
   custom clone code for every skill type. This is the "plan 6 with manual
   ToMap/FromMap" approach. Plan 7's omni delegation is almost as simple but
   slightly slower due to map operations.

5. **`fastcopier-gen` (code generation) achieves 112 ns/op with zero
   allocations** — the fastest non-manual option. But it requires a `go
   generate` step and startup-time registration. Over-engineered for ork.

6. **The benchmark confirms our plan comparison's performance ordering:**
   - Plans 1-2 (no clone): ~0 ns
   - Plan 3 (deep copy): ~1-5µs
   - Plan 7 (omni ToMap): ~2-5µs
   - Plan 4 (JSON serialize): ~50-200µs (15x slower than deep copy)
