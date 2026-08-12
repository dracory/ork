# Source: `docker image pull` and `docker image tag`

**Source URLs:**
- https://docs.docker.com/engine/reference/commandline/pull/
- https://docs.docker.com/engine/reference/commandline/tag/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## `docker image pull` (alias: `docker pull`)

**Description:** Download an image from a registry

**Usage:** `docker image pull [OPTIONS] NAME[:TAG|@DIGEST]`

### Options
| Option | Description |
|--------|-------------|
| `-a, --all-tags` | Download all tagged images in the repository |
| `--platform` | Set platform if server is multi-platform capable |
| `-q, --quiet` | Suppress verbose output |

### Behavior
- Downloads an image from a registry (Docker Hub by default)
- If no tag provided, defaults to `:latest`
- Can pull by digest for immutable pinning: `docker pull ubuntu@sha256:...`
- Can pull from other registries: `docker pull myregistry.local:5000/myimage`
- Layers are reused — pulling `debian:bookworm` after `debian:latest` only pulls metadata if layers are shared

### Idempotency
`docker pull` is idempotent — if the image is already pulled with the same digest, it reports "Image is up to date" and exits 0. For an Ork skill:
- `Check()`: does the image exist locally? If yes, return `false` (already pulled). If no, return `true`.
- `Run()`: execute `docker pull <image>`. If already up to date, Docker exits 0.

### Check pattern
```bash
docker image inspect <image> >/dev/null 2>&1
# Exit 0 = exists (no pull needed)
# Non-zero = doesn't exist (needs pull)
```

### Note on `:latest` tag
If the tag is `:latest`, the image may exist locally but be outdated. The `Check()` method could optionally always return `true` for `:latest` to force a pull (matching Docker's behavior of always checking for updates on `latest`). This should be configurable via an arg (e.g., `always-pull: "true"`).

## `docker image tag` (alias: `docker tag`)

**Description:** Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE

**Usage:** `docker image tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]`

### Behavior
- Creates an alias (tag) pointing from target to source image
- Does NOT create a copy — both tags point to the same image layers
- If target tag already exists, it's overwritten
- Can tag for a private registry: `docker tag myapp:v1 myregistry:5000/myapp:v1`

### Image reference format
```
[HOST[:PORT]/]NAMESPACE/REPOSITORY[:TAG]
```
- Host defaults to `docker.io`
- Namespace defaults to `library` (Docker Official Images)
- Tag defaults to `latest`

### Idempotency
`docker tag` is idempotent — tagging an image that's already tagged with the same target is a no-op. For an Ork skill:
- `Check()`: does the target tag exist and point to the same source image? If yes, return `false`. If no, return `true`.
- `Run()`: execute `docker tag <source> <target>`.

### Check pattern
```bash
# Check if target tag exists and matches source image ID
source_id=$(docker image inspect <source> --format '{{.Id}}' 2>/dev/null)
target_id=$(docker image inspect <target> --format '{{.Id}}' 2>/dev/null)
[ "$source_id" = "$target_id" ]
# Exit 0 = tags match (no change needed)
# Non-zero = tags don't match or target doesn't exist (needs tagging)
```

## Relevance to Ork

- `docker-pull` skill: idempotent, `Check()` probes if image exists locally
- `docker-tag` skill: idempotent, `Check()` probes if target tag points to same image as source
- `docker-pull` should accept `--platform` and `--all-tags` as optional args
- `docker-tag` takes two required args: source image and target tag
- Both should handle registry authentication (Docker login is a prerequisite, not handled by these skills)
- `docker-pull` with `:latest` should optionally force-pull even if image exists (configurable via arg)
- All image names must be shell-escaped via `skills.ShellEscapeArg`
