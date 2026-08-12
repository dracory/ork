# Source: Docker Engine Installation on Ubuntu

**Source URL:** https://docs.docker.com/engine/install/ubuntu/
**Retrieved:** 2026-08-12 (via docs.docker.com)

## Summary

Docker Engine can be installed on Ubuntu via Docker's official `apt` repository. The installation involves adding Docker's GPG key, adding the repository to apt sources, and installing the Docker packages. Supported Ubuntu versions: 26.04 LTS, 24.04 LTS, 22.04 LTS. Supported architectures: amd64, armhf, arm64, s390x, ppc64le.

## Prerequisites

### OS Requirements
- 64-bit Ubuntu (26.04, 24.04, or 22.04 LTS)
- x86_64 (amd64), armhf, arm64, s390x, or ppc64le

### Uninstall Conflicting Packages
Before installing Docker, remove unofficial packages:
```bash
sudo apt remove $(dpkg --get-selections docker.io docker-compose docker-compose-v2 docker-doc docker-buildx podman-docker containerd runc | cut -f1)
```

### Firewall Considerations
- Docker bypasses ufw/firewalld rules for exposed ports
- Docker is only compatible with `iptables-nft` and `iptables-legacy` (not `nft` directly)

## Installation via apt Repository (Recommended for Ork)

### Step 1: Add Docker's GPG key and repository
```bash
# Add Docker's official GPG key:
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
```

### Step 2: Install Docker packages
```bash
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

### Step 3: Verify installation
```bash
sudo systemctl status docker
sudo docker run hello-world
```

## Packages Installed

| Package | Description |
|---------|-------------|
| `docker-ce` | Docker Engine daemon |
| `docker-ce-cli` | Docker CLI client |
| `containerd.io` | Container runtime (bundled) |
| `docker-buildx-plugin` | BuildKit-based build support |
| `docker-compose-plugin` | Docker Compose v2 |

## Post-Installation (Linux postinstall)

The `docker` group exists but contains no users by default. To allow non-root users to run Docker:
```bash
sudo usermod -aG docker $USER
# User must log out and back in for group change to take effect
```

## Idempotency for Ork

A `docker-install` skill should:
- `Check()`: is `docker-ce` already installed? Use `dpkg-query -W -f='${Status}' docker-ce` or `docker --version`
- `Run()`: if not installed, run the full installation sequence (remove conflicts → add repo → apt update → install)
- If already installed, return `Changed: false`

### Check pattern
```bash
# Check if docker is installed and the daemon is running
docker --version >/dev/null 2>&1 && docker info >/dev/null 2>&1
# Exit 0 = installed and running
# Non-zero = not installed or daemon not running
```

Or more precisely:
```bash
dpkg-query -W -f='${Status}' docker-ce 2>/dev/null | grep -q "install ok installed"
# Exit 0 = package installed
# Non-zero = not installed
```

## Relevance to Ork

- `docker-install` skill is the prerequisite for all other Docker skills
- Should use Ork's existing `apt` skill patterns (see `skills/apt/`) for package management
- Must run with `BecomeUser: "root"` (apt install requires root)
- Should use `DEBIAN_FRONTEND=noninteractive` and `DpkgConfOptions` (matching existing apt skills)
- Should handle the case where old conflicting packages are installed (remove them first)
- Should verify the daemon is running after installation (`systemctl is-active docker`)
- Should optionally add the SSH user to the `docker` group (configurable via arg)
- The installation is multi-step (add repo → apt update → install) — should be a single `Run()` that executes all steps
- Can leverage Ork's existing `systemctl` skills for daemon management
