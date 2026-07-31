# Source: kkHAIKE/contextcheck — static analysis for context propagation

- **URL:** https://github.com/kkHAIKE/contextcheck
- **PkgDoc:** https://pkg.go.dev/github.com/kkHAIKE/contextcheck
- **Accessed:** 2026-07-31
- **Category:** Tooling / linter / context.Context

## Relevance to ork decision

`contextcheck` is the linter that checks whether functions properly propagate
`context.Context`. It is included in `golangci-lint`. This source determines
whether plan 2 (functional options with `WithContext(ctx)` as an option) would
pass linting — and the answer is **no**, it breaks the linter.

## Key excerpts

> `contextcheck` is a static analysis tool used to check whether a function
> uses a non-inherited context that could result in a broken call link.

### What it flags

```go
func call1(ctx context.Context) {
    ctx = getNewCtx(ctx)
    call2(ctx)                       // OK

    call2(context.Background())      // Non-inherited new context
    call3()                          // Function `call3` should pass ctx
    call4()                          // Function `call4->call3` should pass ctx
}
```

> To skip this check in some false-positive cases, you can add
> `// nolint: contextcheck` to the function declaration's comment.

## Implications for ork

- **Plan 1 (Opts Interface, `Run(ctx, opts)`):** `ctx` is the first parameter.
  The linter sees proper context propagation. **Passes `contextcheck`.**

- **Plan 2 (Functional Options, `Run(opts ...RunOption)` with
  `WithContext(ctx)`):** `ctx` is NOT a parameter — it's hidden inside an
  option function. The linter cannot trace context propagation through
  `opts[0].Apply(&cfg)`. **Breaks `contextcheck`** — every skill's `Run`
  would need `// nolint: contextcheck`, defeating the purpose of the linter.

- **Plans 3-7 (clone-based, no `ctx`):** no `ctx` at all, so `contextcheck`
  is N/A for now. But when we eventually add `ctx` (for cancellation,
  timeouts, tracing), we'll face the same choice: first-param (plan 1 style,
  linter works) vs option-based (plan 2 style, linter breaks).

This is a **strong argument for plan 1 over plan 2** if we go the
opts-interface route.
