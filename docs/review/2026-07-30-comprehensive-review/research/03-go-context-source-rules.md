# Source: context package source — rules for Context use

- **URL:** https://go.dev/src/context/context.go (lines 33-53)
- **Author:** The Go Authors
- **Accessed:** 2026-07-31
- **Category:** Go official docs / context.Context conventions

## Relevance to ork decision

The `context` package source code itself documents the **official rules** for
using `Context`. These rules are enforced by linters like `contextcheck` and
`revive (context-as-argument)`. They directly inform whether plans 1-2 (which
add `ctx` to the signature) are idiomatic Go.

## Key excerpts (from the source code comments)

> Programs that use Contexts should follow these rules to keep interfaces
> consistent across packages and enable static analysis tools to check context
> propagation:
>
> Do not store Contexts inside a struct type; instead, pass a Context
> explicitly to each function that needs it. This is discussed further in
> https://go.dev/blog/context-and-structs. The Context should be the first
> parameter, typically named ctx:
>
>     func DoSomething(ctx context.Context, arg Arg) error {
>         // ... use ctx ...
>     }
>
> Do not pass a nil Context, even if a function permits it. Pass context.TODO
> if you are unsure about which Context to use.
>
> Use context Values only for request-scoped data that transits processes and
> APIs, not for passing optional parameters to functions.
>
> The same Context may be passed to functions running in different goroutines;
> Contexts are safe for simultaneous use by multiple goroutines.

## Implications for ork

1. **"Do not store Contexts inside a struct type"** — this rule means if we
   ever want `ctx` support, it MUST be a parameter to `Run`/`Check`, not a
   field on `BaseSkill`. Plans 3-7 (which keep `BaseSkill` as a shared struct)
   would violate this rule if we tried to add `ctx` as a field later. Plans 1-2
   follow it from the start.

2. **"The Context should be the first parameter"** — this is the convention
   plan 1 (`Run(ctx, opts)`) follows. Plan 2 (`Run(opts ...RunOption)` with
   `WithContext(ctx)` as an option) does NOT follow this convention, which is
   why the `contextcheck` linter breaks for plan 2.

3. **"Use context Values only for request-scoped data... not for passing
   optional parameters"** — this means we should NOT use `context.WithValue`
   to pass `NodeConfig`/`args`/`dryRun`. They should be explicit parameters or
   a dedicated options struct. Both plans 1 and 2 satisfy this.

4. **"Contexts are safe for simultaneous use by multiple goroutines"** —
   confirms that a per-call `opts` object (if immutable) is safe to share,
   but a per-call `opts` that gets mutated is NOT safe unless cloned.
