# Source: Can recover() catch concurrent map writes? — Moment For Technology

- **URL:** https://cloud.mo4tech.com/can-recover-be-used-to-recover-an-error-generated-when-go-writes-to-map-concurrently.html
- **Accessed:** 2026-07-31
- **Category:** Concurrency / fatal errors / recover() limitations

## Relevance to ork decision

This source is critical: it proves that `recover()` **cannot** catch the
concurrent-map-write error. ork's node goroutines use `defer recover()` to
isolate skill failures — but this defense is **useless** against the map race.
The whole process dies. This is why the bug is Critical severity, not just
High.

## Key excerpts

> In some cases, we use sync.map, map+sync.Mutex, or map+sync.RWMutex to avoid
> exceptions caused by concurrent map writing. If an exception occurs while
> writing to the map, can I recover it by using recover()?

> When we run this code, we get the following error message:
>
>     fatal error: concurrent map writes
>     goroutine 7 [running]:
>     runtime.throw({0x2c7b35?, 0x2bc2c0?})
>     <- C:/Program Files/Go/src/runtime/panic.go:992 +0x76

> Call `runtime.throw()` and end the program at once.

### Three types of errors in Go

> There are three types of errors in Go: error, panic and fatal error.
>
> - **Error** is what we often call an error, usually passed through the
>   function return value, and needs to be handled using `if err != nil`.
> - **Panic** is what we sometimes call an exception [...] This type of error
>   can be caught using `recover`.
> - **Fatal error** is a serious error triggered by the system, which is
>   usually related to system resources. [...] It is severe because the
>   program cannot recover from such errors.
>
> Because fatal errors are unrecoverable, when they occur, they cause the
> entire process to exit, which in turn affects all current requests.

> Before go 1.6, concurrent reading [of maps] was not checked.

## Implications for ork

- **ork's `recover()` in node goroutines is useless against this bug.** The
  `fatal error: concurrent map writes` bypasses `recover()` entirely and kills
  the whole ork process — taking down all in-flight node executions on all
  hosts, not just the one that triggered the race.
- This elevates the bug from "one skill fails" to "entire ork process dies" —
  confirming the **Critical** severity rating in the comprehensive review.
- **Plan 7 (omni):** the `sync.RWMutex` inside `omni.Atom` structurally
  prevents the runtime from ever reaching the `throw()` path. Even if the
  clone-before-mutate pattern has a bug, the mutex is a second safety net.
- **Plans 3-6 (clone-based, no mutex):** rely solely on the clone pattern. If
  a future bug reintroduces shared mutation, the process dies with no
  fallback.
- **Plans 1-2 (opts-based):** eliminate shared mutable state entirely — the
  strongest structural fix, but the most invasive (breaking change).
