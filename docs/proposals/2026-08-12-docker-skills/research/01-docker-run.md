# Source: `docker container run` (alias: `docker run`)

**Source URL:** https://docs.docker.com/engine/reference/commandline/run/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## Summary

`docker run` creates and runs a new container from an image. It pulls the image if needed and starts the container. You can restart a stopped container with `docker start`. Use `docker ps -a` to view all containers including stopped ones.

## Usage

```
docker container run [OPTIONS] IMAGE [COMMAND] [ARG...]
```

## Key Options for Ork Skills

| Option | Description |
|--------|-------------|
| `--name` | Assign a name to the container (critical for idempotency) |
| `-d, --detach` | Run container in background and print container ID |
| `-p, --publish` | Publish a container's port(s) to the host (e.g., `-p 8080:80`) |
| `-P, --publish-all` | Publish all exposed ports to random ports |
| `-e, --env` | Set environment variables |
| `--env-file` | Read environment variables from a file |
| `-v, --volume` | Bind mount a volume |
| `--mount` | Attach a filesystem mount to the container |
| `--network` | Connect container to a network |
| `--restart` | Restart policy (`no`, `always`, `unless-stopped`, `on-failure`) |
| `-u, --user` | Username or UID |
| `-w, --workdir` | Working directory inside the container |
| `--entrypoint` | Overwrite the default ENTRYPOINT |
| `--rm` | Automatically remove container when it exits |
| `--pull` | Pull image before running (`always`, `missing`, `never`) |
| `--health-cmd` | Command to run for health check |
| `-i, --interactive` | Keep STDIN open |
| `-t, --tty` | Allocate a pseudo-TTY |
| `--privileged` | Give extended privileges |
| `--cap-add` / `--cap-drop` | Add/drop Linux capabilities |
| `--memory` / `--cpus` | Resource limits |
| `--label` | Set metadata on container |

## Idempotency Pattern for Ork

`docker run` is NOT idempotent by default — running it twice creates two containers. For an Ork skill, the `Check()` method must probe whether a container with the given `--name` already exists and is running:

```bash
# Check if container exists and is running
docker ps -q -f name=^/<container-name>$
# Non-empty output = running, empty = not running or doesn't exist
```

If the container exists but is stopped, `Check()` should detect that and `Run()` should `docker start` it instead of creating a new one.

If the container exists and is running, `Check()` returns `false` (no change needed).

If the container doesn't exist, `Check()` returns `true` and `Run()` executes `docker run`.

## Example Commands

```bash
# Run detached with name and port mapping
docker run --name myapp -d -p 8080:80 nginx:alpine

# Run with environment variables and volume
docker run --name myapp -d -e DEBUG=true -v /data:/app/data myapp:v1

# Run with restart policy
docker run --name myapp -d --restart unless-stopped myapp:v1

# Run with custom command
docker run --name myapp -d myapp:v1 /app/server --port 3000
```

## Relevance to Ork

- `docker-run` skill is the primary deployment skill
- Must be idempotent: check if container with `--name` already exists and is running
- If stopped, `docker start` instead of `docker run`
- If running with different config (ports, env, image), optionally recreate (with `--force` flag)
- All user-supplied values (name, image, ports, env, command) must be shell-escaped via `skills.ShellEscapeArg`
- Should default to `--detach` mode (background) since Ork skills are non-interactive
- Should default to `--restart unless-stopped` for production deployments (configurable)
