# Source: `docker container stop` and `docker container restart`

**Source URLs:**
- https://docs.docker.com/engine/reference/commandline/stop/
- https://docs.docker.com/engine/reference/commandline/restart/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## `docker container stop` (alias: `docker stop`)

**Description:** Stop one or more running containers

**Usage:** `docker container stop [OPTIONS] CONTAINER [CONTAINER...]`

### Behavior
- Main process inside the container receives `SIGTERM`
- After a grace period (default 10s on Linux, 30s on Windows), receives `SIGKILL`
- Grace period configurable via `--timeout` or `STOPSIGNAL` Dockerfile instruction

### Options
| Option | Description |
|--------|-------------|
| `-s, --signal` | Signal to send (default: SIGTERM, or image's STOPSIGNAL) |
| `-t, --timeout` | Seconds to wait before killing (default: 10 Linux, 30 Windows). `-1` = wait indefinitely |

### Idempotency
`docker stop` is idempotent — stopping an already-stopped container is a no-op (returns the container name with no error). For an Ork skill:
- `Check()`: is the container running? If yes, return `true` (needs stopping). If no, return `false`.
- `Run()`: execute `docker stop <name>`. If already stopped, Docker returns the name without error.

### Check pattern
```bash
docker ps -q -f name=^/<container-name>$
# Non-empty = running (needs stop)
# Empty = not running (no change needed)
```

## `docker container restart` (alias: `docker restart`)

**Description:** Restart one or more containers

**Usage:** `docker container restart [OPTIONS] CONTAINER [CONTAINER...]`

### Behavior
- Stops the container (SIGTERM → grace period → SIGKILL), then starts it again
- Works on both running and stopped containers
- If the container is stopped, it just starts it

### Options
| Option | Description |
|--------|-------------|
| `-s, --signal` | Signal to send (default: SIGTERM) |
| `-t, --timeout` | Seconds to wait before killing (default: 10 Linux, 30 Windows) |

### Idempotency
`docker restart` is NOT naturally idempotent — it always restarts, even if the container is already running. For an Ork skill:
- `Check()`: always returns `true` (restart is intentionally non-idempotent, like `caddy.Restart`)
- `Run()`: execute `docker restart <name>`

This matches the pattern used by `caddy.Restart` in the existing codebase (line 58 of `skills/caddy/restart.go`: `Check() always returns true since Restart is intentionally non-idempotent`).

## Relevance to Ork

- `docker-stop` skill: idempotent, `Check()` probes if container is running
- `docker-restart` skill: intentionally non-idempotent (always restarts), matches `caddy.Restart` pattern
- Both accept `--timeout` and `--signal` as optional args
- Both accept container name (or ID) as required arg
- Both should handle the case where the container doesn't exist (return error in `Run()`, return `false` in `Check()`)
