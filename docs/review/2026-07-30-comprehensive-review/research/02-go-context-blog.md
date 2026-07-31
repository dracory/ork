# Source: Go Concurrency Patterns: Context — The Go Blog

- **URL:** https://go.dev/blog/context
- **Author:** Sameer Ajmani (Go team)
- **Published:** 29 July 2014
- **Accessed:** 2026-07-31
- **Category:** Go official docs / context.Context design

## Relevance to ork decision

This is the original article introducing `context.Context`. It establishes the
Go convention that **request-scoped data should be passed explicitly per call,
not stored in a shared struct**. This is the philosophical foundation for
plans 1 and 2 (pass `RunnableOptionsInterface` / `RunOption` per call).

The article also confirms `Context` is "safe for simultaneous use by multiple
goroutines" — the same property we need for per-call config objects.

## Key excerpts

> At Google, we require that Go programmers pass a `Context` parameter as the
> first argument to every function on the call path between incoming and
> outgoing requests. This allows Go code developed by many different teams to
> interoperate well.

> A `Context` is safe for simultaneous use by multiple goroutines. Code can
> pass a single `Context` to any number of goroutines and cancel that
> `Context` to signal all of them.

> A `Context` does *not* have a `Cancel` method for the same reason the
> `Done` channel is receive-only: the function receiving a cancellation
> signal is usually not the one that sends the signal.

### Derived contexts form a tree

> The `context` package provides functions to *derive* new `Context` values
> from existing ones. These values form a tree: when a `Context` is canceled,
> all `Contexts` derived from it are also canceled.

## Implications for ork

- **Plans 1-2 (opts-based):** align with Go's official guidance — pass
  per-call config explicitly, don't store it in the shared skill. This is the
  "Go standard" approach for request-scoped data.
- **Plans 3-7 (clone-based):** don't follow this convention — they keep the
  shared mutable struct and clone it. This works for fixing the race, but
  doesn't give us `ctx` propagation for free. A future migration to add
  `context.Context` would require another breaking change.
- **The article's core principle** ("pass request-scoped data explicitly, don't
  store it in a struct") is an argument *for* plans 1-2 and *against* the
  shared-mutable-struct model that plans 3-7 preserve.
