# Research: Concurrency-Safe Skill Execution in ork

**Date:** 2026-07-31
**Purpose:** Inform the decision between 7 plans for fixing the data race in
ork's `Inventory.Run` / `RunByID` / `Check` (Critical finding #1 in the
2026-07-30 comprehensive review).

## How to use this research

Each file in this folder covers **one source** (web page, library, bug report,
or official Go documentation). Each file has:
- **Source metadata** (URL, author, date accessed)
- **Relevance to ork decision** (why this source matters)
- **Key excerpts** (the important parts, quoted)
- **Implications for ork** (how this affects the plan comparison)

Read the files in order — they're numbered to build understanding from
fundamentals (what is a data race?) to specifics (which library should we
use?).

## Source index

### Go official documentation (the rules)

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 01 | [01-go-race-detector.md](01-go-race-detector.md) | Go Race Detector | ork's pattern is the textbook "unprotected global variable" race. Fix: mutex or copy. |
| 02 | [02-go-context-blog.md](02-go-context-blog.md) | Go Context blog post | Per-call data should be passed explicitly, not stored in shared structs. |
| 03 | [03-go-context-source-rules.md](03-go-context-source-rules.md) | context.Context source rules | "Do not store Contexts in a struct." ctx must be first param. |
| 04 | [04-contextcheck-linter.md](04-contextcheck-linter.md) | contextcheck linter | Plan 2 (ctx as option) breaks this linter. Plan 1 (ctx as first param) passes. |
| 05 | [05-context-arguments-boldlygo.md](05-context-arguments-boldlygo.md) | Context arguments convention | revive linter enforces ctx-as-first-param. Plan 2 violates it. |

### The fatal crash (why this is Critical, not just High)

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 06 | [06-concurrent-map-fatal-stackoverflow.md](06-concurrent-map-fatal-stackoverflow.md) | concurrent map fatal error | Fix: sync.RWMutex or sync.Map. Test with -race. |
| 07 | [07-concurrent-map-fatal-unrecoverable.md](07-concurrent-map-fatal-unrecoverable.md) | recover() cannot catch fatal errors | ork's recover() is useless against this crash. Whole process dies. |
| 08 | [08-concurrent-map-writes-root-cause.md](08-concurrent-map-writes-root-cause.md) | Root cause: hashWriting flag + throw() | RWMutex with pointer receiver is the recommended fix. |

### Patterns and real-world frameworks

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 09 | [09-sync-primitives-nested-struct.md](09-sync-primitives-nested-struct.md) | Nested struct + mutex pattern | The idiomatic way to protect struct fields. omni.Atom does this for us. |
| 10 | [10-functional-options-silentgopher.md](10-functional-options-silentgopher.md) | Functional Options pattern | Per-request data belongs in context/params, NOT in method options. Argues against plan 2. |
| 11 | [11-dave-cheney-first-class-functions.md](11-dave-cheney-first-class-functions.md) | Dave Cheney's original FO talk | Functional options (plan 2) and interface options (plan 1) are equivalent. Plan 1 is better for ork (linter, type safety). |
| 12 | [12-modular-concurrency-guidelines.md](12-modular-concurrency-guidelines.md) | Modular framework concurrency rules | "Config maps from caller → defensive deep copy under lock." Validates plans 3-7. |
| 13 | [13-textile-go-thread-safety.md](13-textile-go-thread-safety.md) | textile-go: RWMutex + deep clone + COW | Real-world framework using plan 7's exact architecture (RWMutex + clone + copy-on-write). |
| 14 | [14-immutable-data-goperf.md](14-immutable-data-goperf.md) | Go perf guide: immutable data sharing | atomic.Pointer for lock-free reads (future optimization). Defensive copy of maps. |

### Deep copy libraries (for plans 3-4)

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 15 | [15-brunoga-deep-v5.md](15-brunoga-deep-v5.md) | brunoga/deep v5 | **v5 changed the API** — plan 3 was written for v4. v5 is over-engineered (CRDTs, patches). |
| 16 | [16-fastcopier-benchmark.md](16-fastcopier-benchmark.md) | Deep copy library benchmarks | JSON serialize (plan 4) is 15x slower than reflection deep copy. Confirms our performance estimates. |
| 17 | [17-golang-design-reflect-deepcopy.md](17-golang-design-reflect-deepcopy.md) | Go proposal #51520: DeepCopy | Stateful objects (mutexes) should not be deep copied. Plan 7's serialization avoids this. |

### Real-world bug reports (same pattern as ork)

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 18 | [18-dubbo-go-registry-races.md](18-dubbo-go-registry-races.md) | Apache dubbo-go: 44 data races | Same pattern as ork. Fix: generic Registry[T] with RWMutex = omni.Atom. |

### The proposed library (plan 7)

| # | File | Topic | Key takeaway for ork |
|---|------|-------|----------------------|
| 19 | [19-omni-library.md](19-omni-library.md) | dracory/omni source code analysis | What omni provides, thread safety, serialization, license concerns. |
| 20 | [20-rwmutex-scaling-issue.md](20-rwmutex-scaling-issue.md) | sync.RWMutex scaling issue | RWMutex scales poorly at 1000+ cores. ork is 1-50 nodes — not a concern. |

## Summary of findings

### The bug is well-understood and well-documented

ork's pattern (shared mutable struct with map fields, accessed by multiple
goroutines without synchronization) is the **textbook data race** described in
Go's official race detector documentation (source 01). It produces a
**fatal, unrecoverable crash** (sources 07, 08) that ork's `recover()` cannot
catch. Apache dubbo-go had the same pattern and fixed it with a generic
RWMutex-protected registry (source 18) — which is what `omni.Atom` provides.

### The fix options form a spectrum

```
Most invasive ←————————————————————→ Least invasive
Plan 1 (opts interface)  Plan 3 (deep copy)  Plan 7 (omni)
Plan 2 (func opts)       Plan 4 (serialize)  Plan 6 (map storage)
                         Plan 5 (map clone)
```

- **Plans 1-2 (breaking, opts-based):** eliminate shared mutable state — the
  "Go standard" approach per sources 02-05, 10-11. But require rewriting every
  skill's `Run`/`Check` signature. Plan 1 is better than plan 2 (linter
  compatibility, type safety).
- **Plans 3-4 (breaking, clone-based):** clone the shared struct before
  mutating. Plan 3's library (brunoga/deep) changed its API in v5 (source 15).
  Plan 4 (JSON serialize) is 15x slower (source 16).
- **Plans 5-6 (non-breaking, map-based):** store state in a map, clone via
  map copy. Simple, no external dependency, no silent data loss (plan 6).
- **Plan 7 (non-breaking, omni):** store state in `omni.Atom`, clone via
  `ToMap()`/`FromMap()`. Gets RWMutex defense-in-depth + free serialization.
  One concern: license ambiguity (MIT vs AGPLv3 — source 19).

### Key findings that affect the decision

1. **Plan 2 is weakened** by sources 04, 05, 10: the `contextcheck` and
   `revive` linters break, and the Functional Options pattern is meant for
   construction, not per-call invocation. Per-request data belongs in
   `context.Context` or explicit params, not in method options.

2. **Plan 3 is weakened** by source 15: `brunoga/deep` v5 changed its API
   dramatically. Plan 3 was written for v4. v5 is over-engineered for ork.

3. **Plan 7 is strengthened** by sources 09, 12, 13, 18:
   - Source 09: the nested-struct+mutex pattern is idiomatic Go; omni does it
     for us.
   - Source 12: the modular framework guidelines explicitly recommend
     "defensive deep copy under lock" for caller-provided config maps.
   - Source 13: textile-go uses the exact same architecture (RWMutex + deep
     clone + copy-on-write) in production.
   - Source 18: Apache dubbo-go fixed the same bug with a generic
     RWMutex-protected registry — which is what omni.Atom is.

4. **Plan 7's license concern is RESOLVED** (source 19): omni is AGPLv3,
   ork is also AGPLv3, both from the same `dracory` organization. AGPLv3 is
   compatible with AGPLv3 — no license issue. ork can use omni without
   restriction.

5. **The `sync.RWMutex` scaling concern** (source 20) doesn't apply to ork's
   scale (1-50 nodes, not 1000+ cores).

### Recommendation (preliminary, pending license verification)

Based on this research:

1. **Plan 7 (omni-backed) is the recommended non-breaking option.** License
   confirmed compatible (both AGPLv3, same author). It combines
   clone-before-mutate with RWMutex defense-in-depth, free serialization,
   and is validated by real-world frameworks (textile-go, dubbo-go's
   recommended fix).

2. **Plan 6 (map-storage) is the fallback** if we want zero external
   dependencies. Same architecture (map-backed state, ToMap/FromMap clone)
   without the RWMutex defense-in-depth.

3. **If we're willing to do a major version bump:** Plan 1 (opts interface)
   is the most idiomatic Go approach (per sources 02-05, 10-11), gives us
   `ctx` propagation for free, and eliminates shared mutable state entirely.
   But it requires rewriting all 40+ skills.

**Plan 2 is not recommended** — it breaks linters (sources 04-05) and the
Functional Options pattern is not designed for per-call data (source 10).

**Plan 3 is not recommended** — the library it depends on (brunoga/deep)
changed its API in v5 (source 15), and v4 uses `unsafe`.

**Plan 4 is not recommended** — JSON serialization is 15x slower than
alternatives (source 16) and has silent-data-loss risk.
