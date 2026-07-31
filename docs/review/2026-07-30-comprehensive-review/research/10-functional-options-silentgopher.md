# Source: Functional Options in Go: The Pattern Behind Clean Service Constructors — SilentGopher

- **URL:** https://www.silentgopher.dev/en/posts/functional-options/
- **Author:** @SilentGopher
- **Published:** 2026-04-20
- **Accessed:** 2026-07-31
- **Category:** Go patterns / functional options / API design

## Relevance to ork decision

This is a thorough modern explanation of the Functional Options pattern
(plan 2's approach). Critically, it distinguishes **constructor-level options**
(service configuration) from **method-level options** (per-call variation) —
and argues that per-call data should go through `context.Context`, not through
method options. This is an argument **against** plan 2's design and **for**
plan 1 (or for keeping the current signature and cloning).

## Key excerpts

> If your function signature has grown to five parameters, the sixth one
> doesn't make the code harder to read — it makes it dangerous.

### The pattern

```go
type UserServiceOptions struct {
    crmClient      CRMClient
    defaultLocale  string
    includeDeleted bool
}

type UserServiceOption func(*UserServiceOptions)

func WithCRMEnrichment(client CRMClient) UserServiceOption {
    return func(opts *UserServiceOptions) {
        opts.crmClient = client
    }
}
```

### Two levels: constructor vs. method

> You can apply functional options at two different levels:
> 1. **Constructor-level**: configures how the service *works* (dependency,
>    behavior toggle)
> 2. **Method-level**: configures how *this specific call* behaves

> Use constructor options for: dependencies, feature flags, timeouts, default
> behaviors.
> Use method options for: caller-specific data, request-scoped tokens,
> per-call overrides.

### The critical insight — per-request data belongs in context, not options

> **Step 2:** The supplier token travels through `context.Context` — where
> per-request data belongs.

The article's refactored example moves a per-request token OUT of method
options and INTO `context.Context`:

```go
// Middleware sets the token in context:
ctx = auth.WithSupplierToken(ctx, r.Header.Get("X-Supplier-Token"))

// Service reads it from context:
func (s *ProductService) Upsert(ctx context.Context, product model.Product) error {
    token, ok := auth.SupplierTokenFromContext(ctx)
    // ...
}
```

> **What feels wrong?** The option bundles two things:
> 1. A **behavior toggle** (`SyncFromSupplier: true`) — this is
>    *configuration*
> 2. A **session credential** (`SupplierToken`) — this is *per-request data*
>
> They have different lifecycles and don't belong in the same option.

### Why not a Config struct?

> `Port: 0` could mean two completely different things:
> - "I didn't set it, use the default (8080)"
> - "I want port 0 so the OS chooses a free port"
>
> Those are indistinguishable.

> Functional options sidestep both problems: the variadic signature makes the
> default case require zero arguments, and options compose safely without
> sharing internal state.

## Implications for ork

- **Plan 2 (Functional Options, `Run(opts ...RunOption)`):** this article
  actually argues AGAINST putting per-call data (NodeConfig, args, dryRun) in
  method-level options. Those are per-request data, and the article says
  per-request data belongs in `context.Context` or an explicit parameter —
  not in `RunOption` functions.

- **Plan 1 (Opts Interface, `Run(ctx, opts)`):** aligns with this article's
  refactored approach — per-call data in an explicit parameter (the `opts`
  interface), `ctx` as first param. This is the article's recommended
  pattern.

- **Plans 3-7 (clone-based):** sidestep the debate entirely — per-call data
  stays in the cloned skill's fields, accessed via the existing getters.
  No new parameter type needed. But also no `ctx` propagation.

- **The article's core lesson:** "per-request data and configuration have
  different lifecycles — don't bundle them." ork's `NodeConfig`/`args`/`dryRun`
  are per-request data. Plans 1 and 3-7 keep them separate from configuration
  (id, description). Plan 2 bundles them into `RunOption` functions, which
  the article considers an anti-pattern.
