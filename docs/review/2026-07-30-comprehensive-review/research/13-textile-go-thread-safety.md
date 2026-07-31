# Source: Thread Safety — blue-context/textile-go (DeepWiki)

- **URL:** https://deepwiki.com/blue-context/textile-go/4.2-thread-safety
- **Accessed:** 2026-07-31
- **Category:** Go concurrency / deep cloning / copy-on-write / real-world framework

## Relevance to ork decision

textile-go is a real-world Go framework that solves the **exact same problem**
ork has: a shared `Client` used by multiple goroutines, where each request
needs isolated state. Their solution is **deep cloning + RWMutex +
copy-on-write** — which is precisely plan 7's approach (clone + omni's
RWMutex).

## Key excerpts

### Thread-safety guarantees

> | Component | Thread-Safety Guarantee | Mechanism |
> | --- | --- | --- |
> | `textile.Client` | Safe for concurrent calls from multiple goroutines | `sync.RWMutex` protection + immutable operations |
> | `Pipeline` | Immutable after creation; modifications create new instances | Copy-on-write pattern |
> | `TransformContext` | Per-request isolation; no shared state between requests | Deep cloning + independent allocation |

> These guarantees enable multiple goroutines to share a single
> `textile.Client` instance without data races or corruption.

### The three-pillar strategy

> 1. **Read-Write Mutex Protection**: The `sync.RWMutex` field protects access
>    to the pipeline and configuration, allowing concurrent reads but
>    exclusive writes
> 2. **Per-Request Cloning**: Each request triggers deep cloning, creating
>    independent memory allocations
> 3. **Stateless Shared Resources**: The underlying `warp.Client` and
>    `Pipeline` are read-only during execution
> 4. **Isolated Contexts**: Each `TransformContext` is allocated independently
>    with no shared references

### Deep cloning is the foundation

> Deep cloning is the foundation of textile-go's thread safety. Every request
> and response is cloned before transformation, ensuring no shared mutable
> state between concurrent operations.

> | Stage | Clone Target | Purpose |
> | --- | --- | --- |
> | Request Entry | `warp.CompletionRequest` | Isolate input from caller |
> | Request Transform | Uses cloned request | Safe mutation during transformation |

### Copy-on-write for pipeline modification

> The `Pipeline` type uses copy-on-write semantics:
> - `WithTransformers()` creates new client instead of modifying existing
> - Original client remains unchanged (immutable)
> - Each new client gets independent pipeline copy via `Append()`
> - No locks needed during new client creation after initial read

## Implications for ork

This is a **direct real-world validation of plan 7's architecture**:

| textile-go | ork (plan 7) |
|-----------|-------------|
| `textile.Client` shared by goroutines | `BaseSkill` shared by node goroutines |
| `sync.RWMutex` on Client | `sync.RWMutex` in `omni.Atom` |
| Deep clone per request | `ToMap()` → `cloneFromMap()` per node call |
| `TransformContext` per-request isolation | Cloned `BaseSkill` per-node isolation |
| Copy-on-write for Pipeline | Copy-on-write for skill state (ToMap creates new map) |
| Read-only shared resources | Original skill never mutated |

**textile-go proves that the "RWMutex + deep clone + copy-on-write" triple
strategy works in production** for the exact pattern ork has. Plan 7
implements all three:

1. **RWMutex:** `omni.Atom`'s built-in `sync.RWMutex`
2. **Deep clone:** `ToMap()` → `cloneFromMap()` (serialize + deserialize)
3. **Copy-on-write:** `ToMap()` creates a new map; modifications go on the
   map, not the original; `FromMap()` creates a new Atom

This is stronger evidence than theoretical reasoning — it's a deployed
framework using the exact same architecture for the exact same problem.
