# Source: sync: RWMutex scales poorly with CPU count — golang/go#17973

- **URL:** https://github.com/golang/go/issues/17973
- **Accessed:** 2026-07-31
- **Category:** Go runtime / sync.RWMutex performance / known issue

## Relevance to ork decision

This is a known Go issue: `sync.RWMutex` scales poorly with CPU count under
heavy read contention. Since plan 7 uses `omni.Atom` which has a
`sync.RWMutex`, this is a potential performance concern. However, the issue
is about **high-contention** scenarios — ork's usage pattern is low-contention.

## Key excerpts

> `sync.RWMutex` scales poorly with CPU count

> It may be difficult to apply the algorithm described in that paper to our
> existing `sync.RWMutex` type.

> general application code can work around the problem (in part) by using
> per-goroutine or per-goroutine-pool caches rather than global caches shared
> throughout the process.

> The bigger issue is that `sync.RWMutex` is used fairly extensively within
> the standard library for package-level locks

## Implications for ork

1. **The scaling issue is about HIGH read contention** — many goroutines
   all trying to `RLock` the same mutex simultaneously on many CPU cores.
   ork's pattern is:
   - N node goroutines (typically 1-50, not thousands)
   - Each calls `ToMap()` once (one `RLock`), then works on its own clone
   - The `RLock` is held for microseconds (map copy), not during I/O

   This is **low contention**. The scaling issue doesn't apply.

2. **ork's clone-before-mutate pattern naturally reduces contention.** Each
   goroutine only touches the shared Atom's mutex once (during `ToMap()`),
   then works on its private clone. The mutex is never held during the
   long-running `Run()` (which does SSH I/O). This is the "minimize lock
   scope" principle from source 12 (modular concurrency guidelines).

3. **If ork ever scales to thousands of concurrent nodes**, the RWMutex could
   become a bottleneck. The fix would be:
   - Per-goroutine caching (as the issue suggests)
   - `atomic.Pointer[Atom]` for lock-free reads (source 14, goperf.dev)
   - Sharded locks (like `goshard`, source from earlier search)

   But this is a future optimization, not a current concern. ork's typical
   workload is 1-50 nodes, not 1000+.

4. **Plan 6 (raw map, no mutex) doesn't have this issue** — but it also
   doesn't have defense-in-depth. The tradeoff is: plan 7 has a theoretical
   scaling ceiling (very high), plan 6 has no scaling ceiling but no safety
   net.

5. **The issue confirms that `sync.RWMutex` is the right choice for ork's
   current scale.** The Go team hasn't deprecated it; they just acknowledge
   it doesn't scale to thousands of cores. ork is nowhere near that scale.
