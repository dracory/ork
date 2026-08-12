# Source: `docker container exec`

**Source URL:** https://docs.docker.com/engine/reference/commandline/exec/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## Summary

`docker exec` runs a new command in a running container. The command only runs while the container's primary process (PID 1) is running. The command runs in the default working directory of the container. The command must be an executable — chained or quoted commands don't work directly.

## Usage

```
docker container exec [OPTIONS] CONTAINER COMMAND [ARG...]
```

## Key Options

| Option | Description |
|--------|-------------|
| `-d, --detach` | Detached mode: run command in the background |
| `-e, --env` | Set environment variables |
| `--env-file` | Read environment variables from a file |
| `-i, --interactive` | Keep STDIN open even if not attached |
| `--privileged` | Give extended privileges to the command |
| `-t, --tty` | Allocate a pseudo-TTY |
| `-u, --user` | Username or UID (format: `<name|uid>[:<group|gid>]`) |
| `-w, --workdir` | Working directory inside the container |

## Important Notes

1. **Command must be an executable** — not a shell expression:
   - Works: `docker exec -it my_container sh -c "echo a && echo b"`
   - Doesn't work: `docker exec -it my_container "echo a && echo b"`

2. **Container must be running** — `docker exec` fails on stopped containers:
   ```
   Error response from daemon: Container <name> is not running
   ```

3. **Container must not be paused** — `docker exec` fails on paused containers:
   ```
   Error response from daemon: Container <name> is paused, unpause the container before exec
   ```

4. **Not restarted on container restart** — the exec command is a one-shot; it doesn't persist across container restarts.

## Idempotency

`docker exec` is NOT idempotent — it runs a command every time. For an Ork skill:
- `Check()`: always returns `true` (exec is intentionally non-idempotent, like a command execution)
- `Run()`: execute `docker exec <options> <container> <command>`

This is similar to how `caddy.Restart` works — `Check()` always returns `true` because the operation is intentionally run every time.

## Example Commands

```bash
# Run a command in a container
docker exec myapp ls /app

# Run with custom working directory
docker exec -w /app myapp ./server --check

# Run with environment variables
docker exec -e DEBUG=true myapp env

# Run a shell command (must use sh -c for chaining)
docker exec myapp sh -c "echo hello && ls /tmp"

# Run as a specific user
docker exec -u www-data myapp ls /var/www
```

## Relevance to Ork

- `docker-exec` skill: intentionally non-idempotent (always runs the command)
- Takes required args: container name, command (and optional args)
- Takes optional args: user, workdir, env, detach
- Must validate the container is running before executing (return error if not)
- The command must be shell-escaped properly — the command inside the container is a new shell context
- Useful for running health checks, migrations, config reloads, etc. inside containers
- Should support `Stdin` via `types.Command.Stdin` for piping data into the container command
