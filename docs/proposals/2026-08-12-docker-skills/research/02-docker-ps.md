# Source: `docker container ls` (alias: `docker ps`)

**Source URL:** https://docs.docker.com/engine/reference/commandline/ps/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## Summary

`docker ps` lists containers. By default shows only running containers. Use `-a` to show all containers including stopped ones. Supports filtering by name, status, label, exit code, etc.

## Usage

```
docker container ls [OPTIONS]
docker ps [OPTIONS]
```

## Key Options

| Option | Description |
|--------|-------------|
| `-a, --all` | Show all containers (default shows just running) |
| `-f, --filter` | Filter output based on conditions |
| `--format` | Format output (table, json, or Go template) |
| `-n, --last` | Show n last created containers |
| `-l, --latest` | Show the latest created container |
| `--no-trunc` | Don't truncate output |
| `-q, --quiet` | Only display container IDs |
| `-s, --size` | Display total file sizes |

## Filters (for idempotency checks)

| Filter | Description |
|--------|-------------|
| `id` | Container's ID |
| `name` | Container's name |
| `label` | Key or key=value label |
| `exited` | Exit code (use with `--all`) |
| `status` | `created`, `restarting`, `running`, `removing`, `paused`, `exited`, `dead` |
| `ancestor` | Image name/tag/id/digest |
| `before` / `since` | Created before/after a given container |
| `volume` | Mounted volume |
| `network` | Connected network |
| `publish` / `expose` | Published/exposed port |
| `health` | `starting`, `healthy`, `unhealthy`, `none` |

## Idempotency Patterns for Ork

### Check if a container is running (by name)
```bash
docker ps -q -f name=^/<exact-name>$
# Non-empty output = running
# Empty output = not running or doesn't exist
```

The `^/<name>$` anchors the name filter to prevent substring matches (e.g., `name=web` would match `webapp` without anchoring).

### Check if a container exists (running or stopped)
```bash
docker ps -a -q -f name=^/<exact-name>$
# Non-empty output = exists (may be stopped)
# Empty output = doesn't exist
```

### Check container status
```bash
docker ps -a -f name=^/<exact-name>$ --format '{{.Status}}'
# Output: "Up 2 hours", "Exited (0) 5 minutes ago", etc.
```

### List all containers as JSON
```bash
docker ps -a --format json
# One JSON object per line with all container details
```

## Relevance to Ork

- `docker-ps` skill is a **read-only** skill (check-only, never changes state)
- Used by other Docker skills' `Check()` methods to determine container state
- The `-q` (quiet) + `-f name=^/<name>$` pattern is the standard idempotency probe
- `--format json` is useful for structured output in `Result.Details`
- The `name` filter does substring matching by default — must anchor with `^/...$` for exact matches
