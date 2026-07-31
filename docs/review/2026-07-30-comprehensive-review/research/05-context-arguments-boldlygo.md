# Source: Context arguments — Boldly Go

- **URL:** https://boldlygo.tech/archive/2025-04-02-context-arguments/
- **Author:** Boldly Go
- **Published:** 2025-04-02
- **Accessed:** 2026-07-31
- **Category:** Go conventions / context.Context

## Relevance to ork decision

Explains *why* `context.Context` must be the first parameter and why the
`revive` linter's `context-as-argument` rule enforces this. Reinforces that
plan 2's approach (ctx as a functional option) violates a widely-enforced Go
convention.

## Key excerpts

> The Context should be the first parameter, typically named ctx:
>
>     func DoSomething(ctx context.Context, arg Arg) error {
>         // ... use ctx ...
>     }
>
> This short sentence includes two "rules". Let's talk about both, why they
> exist, and when to (possibly) violate them.

### Why first parameter?

> Why this is a rule is probably a lot less important than following the rule.
> [...] since variadic variables may only be the last in the list of function
> arguments, this leaves the only option for a common argument to be the first
> one.

### The revive linter

> The revive linter, also included in golangci-lint, has a `context-as-argument`
> rule, which will warn you when you violate this convention.

### Name your contexts `ctx`

> It's great to know, at a glance, that an argument represents a
> `context.Context` value.

## Implications for ork

- **Plan 1 (`Run(ctx, opts)`):** follows both rules — `ctx` is first, named
  `ctx`. Passes `revive` and `contextcheck`.

- **Plan 2 (`Run(opts ...RunOption)` with `WithContext(ctx)`):** violates the
  `context-as-argument` rule — `ctx` is not a parameter at all. Would trigger
  `revive` warnings and breaks `contextcheck`.

- **Plans 3-7:** no `ctx` yet. When added, must follow this convention to
  avoid linter issues — which means plan 1's signature style is the
  future-proof choice.
