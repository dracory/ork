# Source: `docker container rm` and `docker image rm`

**Source URLs:**
- https://docs.docker.com/engine/reference/commandline/rm/
- https://docs.docker.com/engine/reference/commandline/rmi/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## `docker container rm` (alias: `docker rm`)

**Description:** Remove one or more containers

**Usage:** `docker container rm [OPTIONS] CONTAINER [CONTAINER...]`

### Options
| Option | Description |
|--------|-------------|
| `-f, --force` | Force removal of a running container (uses SIGKILL) |
| `-l, --link` | Remove the specified link |
| `-v, --volumes` | Remove anonymous volumes associated with the container |

### Behavior
- Removes a stopped container
- Cannot remove a running container without `--force`
- `--force` sends SIGKILL first, then removes
- `--volumes` removes anonymous volumes (named volumes are NOT removed)

### Idempotency
`docker rm` on a non-existent container returns an error. For an Ork skill:
- `Check()`: does the container exist? If yes, return `true` (needs removal). If no, return `false`.
- `Run()`: execute `docker rm <name>`. If the container doesn't exist, return an error OR treat as already-removed (idempotent success).

### Check pattern
```bash
docker ps -a -q -f name=^/<container-name>$
# Non-empty = exists (needs removal)
# Empty = doesn't exist (no change needed)
```

## `docker image rm` (alias: `docker rmi`)

**Description:** Remove one or more images

**Usage:** `docker image rm [OPTIONS] IMAGE [IMAGE...]`

### Options
| Option | Description |
|--------|-------------|
| `-f, --force` | Force removal of the image |
| `--no-prune` | Do not delete untagged parents |
| `--platform` | Remove only the given platform variant (API 1.50+) |

### Behavior
- Removes (un-tags) one or more images from the host
- If an image has multiple tags, removing by tag only removes the tag
- Cannot remove an image of a running container without `--force`
- Digest references are removed automatically when removing by tag

### Idempotency
`docker rmi` on a non-existent image returns an error. For an Ork skill:
- `Check()`: does the image exist? If yes, return `true`. If no, return `false`.
- `Run()`: execute `docker rmi <image>`. If the image doesn't exist, return error OR treat as already-removed.

### Check pattern
```bash
docker image inspect <image> >/dev/null 2>&1
# Exit 0 = exists (needs removal)
# Non-zero = doesn't exist (no change needed)
```

## Relevance to Ork

- `docker-rm` skill: idempotent, `Check()` probes if container exists (use `docker ps -a`)
- `docker-rmi` skill: idempotent, `Check()` probes if image exists (use `docker image inspect`)
- Both accept `--force` as an optional arg for removing running containers/in-use images
- `docker-rm` should accept `--volumes` as an optional arg
- Both should handle non-existent resources gracefully (return `Changed: false` if already gone, not an error)
- All image/container names must be shell-escaped via `skills.ShellEscapeArg`
