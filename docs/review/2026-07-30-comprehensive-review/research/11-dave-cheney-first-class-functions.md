# Source: Do not fear first class functions — Dave Cheney

- **URL:** https://dave.cheney.net/2016/11/13/do-not-fear-first-class-functions
- **Author:** Dave Cheney
- **Published:** 2016-11-13
- **Accessed:** 2026-07-31
- **Category:** Go patterns / functional options / original source

## Relevance to ork decision

This is Dave Cheney's original dotGo talk that popularized the Functional
Options pattern. It's the foundational source for plan 2. Dave also shows the
interface-based alternative (which is plan 1's approach) and notes they're
equivalent — the choice is stylistic.

## Key excerpts

### The functional options pattern

```go
type Config struct{ ... }
type Terrain struct {
    config Config
}

func NewTerrain(options ...func(*Config)) *Terrain {
    var t Terrain
    for _, option := range options {
        option(&t.config)
    }
    return &t
}
```

### The interface equivalent (plan 1's approach)

> Another way to think about what is going on here is to try to rewrite the
> functional option pattern using an interface.

```go
type Option interface {
    Apply(*Config)
}

func NewTerrain(options ...Option) *Terrain {
    var config Config
    for _, option := range options {
        option.Apply(&config)
    }
    // ...
}
```

> Whenever we call `NewTerrain` we pass in one or more values that implement
> the `Option` interface. Inside `NewTerrain`, just as before, we range over
> the slice of options and call the `Apply` method on each.

> This doesn't look too different to the previous example. Rather than ranging
> over a slice of functions and calling them, we range over a slice of
> interface values and call a method on each.

## Implications for ork

- **Dave Cheney himself shows that functional options (plan 2) and interface
  options (plan 1) are equivalent** — one uses `func(*Config)`, the other uses
  an `Option` interface with `Apply(*Config)`. The choice is stylistic, not
  functional.

- **For ork, plan 1 is better than plan 2** because:
  1. Plan 1's interface can carry typed getters (`opts.GetNodeConfig()`) —
     more type-safe than plan 2's struct fields accessed via `Apply`.
  2. Plan 1 works with the `contextcheck` linter (ctx is first param); plan 2
     breaks it (ctx hidden in an option).
  3. Plan 1's interface is mockable in tests; plan 2's function values are
     harder to mock.

- **However, both plans 1 and 2 are breaking changes** — they require
  rewriting every skill's `Run()` and `Check()` signatures. This is why plans
  3-7 (non-breaking) exist as alternatives.

- **Dave's article is about construction, not per-call invocation.** The
  pattern was designed for `NewX()` constructors, not for `Run()` methods
  called millions of times. Applying it to `Run()` (plan 2) stretches the
  pattern beyond its original intent — per-call data really belongs in
  explicit parameters or `context.Context`, per source 10.
