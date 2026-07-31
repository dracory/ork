# Source: Golang fatal error: concurrent map read and map write — StackOverflow

- **URL:** https://stackoverflow.com/questions/45585589/golang-fatal-error-concurrent-map-read-and-map-write
- **Accessed:** 2026-07-31
- **Category:** Concurrency / fatal errors / map safety

## Relevance to ork decision

This is the exact error ork can produce: `fatal error: concurrent map read and
map write`. The answers confirm the two standard fixes (`sync.RWMutex` or
`sync.Map`) and recommend testing with `-race`.

## Key excerpts

> I'm writing minecraft server in Go, when server is being stressed by 2000+
> connections I get this crash:
>
>     fatal error: concurrent map read and map write

### Recommended fixes

> Generally speaking you have a few options. Here are two of them:
>
> **sync.RWMutex** — Control access to the map with `sync.RWMutex{}`. Use this
> option if you have single reads and writes, not loops over the map.
>
> **sync.Map** — Use a `sync.Map{}` instead of a normal map. This map is
> already taking care of race issues but may be slower depending on your
> usage. `sync.Map`'s main advantage lies with for loops.

> You should test your server with `-race` option and then eliminate all the
> race conditions it throws.

## Implications for ork

- ork's `BaseSkill` has a `map[string]string` (args) that is read/written
  concurrently across node goroutines → exactly this crash.
- **Plan 7 (omni):** the `omni.Atom` uses `sync.RWMutex` internally — this is
  fix #1 from this answer, applied automatically.
- **Plans 3-6 (clone-based):** avoid the crash by never sharing the map —
  each goroutine gets its own clone. But if a bug ever reintroduces sharing,
  the crash returns (no defense-in-depth).
- **Plan 7's defense-in-depth:** even if a bug shares the Atom, the RWMutex
  prevents the fatal crash — it degrades to slow (serialized) access instead
  of a process kill.
