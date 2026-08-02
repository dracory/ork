// Package caddy provides skills for managing the Caddy web server when
// installed via the official apt repository on Debian/Ubuntu.
//
// It covers the four operations needed to provision and operate Caddy:
//   - Install: add the official Caddy apt repo, install the package, and
//     prepare the log directory (Install).
//   - Harden: apply a systemd drop-in override with sandboxing directives
//     (Harden).
//   - Restart: upload a local Caddyfile, validate it, and reload the
//     systemd unit with a restart fallback (Restart).
//   - Status: show the systemd unit status (Status, read-only).
//
// All skills target the systemd service named "caddy" (the unit shipped by
// the official apt package) and use the standard paths /etc/caddy/Caddyfile
// and /var/log/caddy. These can be overridden via the argument constants
// below where it makes sense.
//
// Usage:
//
//	node.Run(caddy.NewInstall())
//	node.Run(caddy.NewHarden())
//	node.Run(caddy.NewRestart().SetCaddyfilePath("webserver/Caddyfile"))
//	node.Run(caddy.NewStatus())
package caddy

// Argument key constants for use with GetArg/SetArg.
const (
	// ArgCaddyfilePath is the local filesystem path to the Caddyfile that
	// Restart uploads to the remote server. Defaults to DefaultCaddyfilePath
	// when not set. Only used by Restart.
	ArgCaddyfilePath = "caddyfile-path"

	// ArgCaddyfileRemotePath is the remote path where Restart writes the
	// uploaded Caddyfile. Defaults to DefaultCaddyfileRemotePath when not set.
	// Only used by Restart.
	ArgCaddyfileRemotePath = "caddyfile-remote-path"

	// ArgOverrideDir is the systemd drop-in directory for the Caddy unit
	// override. Defaults to DefaultOverrideDir when not set. Only used by
	// Harden.
	ArgOverrideDir = "override-dir"

	// ArgProtectHome controls the systemd ProtectHome directive in the Harden
	// override. Accepted values: "true", "false", "read-only". Defaults to
	// DefaultProtectHome ("true"). Set to "false" if the web root is under
	// /home, otherwise Caddy cannot read it. Only used by Harden.
	ArgProtectHome = "protect-home"

	// ArgReadWritePaths is a space-separated list of paths Caddy may write to,
	// mapped to the systemd ReadWritePaths directive in the Harden override.
	// Defaults to DefaultReadWritePaths. Add custom data/cache directories
	// here. Only used by Harden.
	ArgReadWritePaths = "read-write-paths"

	// ArgMemoryDenyWriteExecute controls the systemd
	// MemoryDenyWriteExecute directive in the Harden override. Accepted
	// values: "true", "false". Defaults to DefaultMemoryDenyWriteExecute
	// ("true"). Set to "false" if a Caddy plugin needs W^X memory (rare).
	// Only used by Harden.
	ArgMemoryDenyWriteExecute = "memory-deny-write-execute"

	// ArgProtectSystem controls the systemd ProtectSystem directive in the
	// Harden override. Accepted values: "strict", "full", "true". Defaults
	// to DefaultProtectSystem ("strict"). Set to "full" if Caddy manages
	// ACME certificates itself and needs to write to /etc/letsencrypt.
	// Only used by Harden.
	ArgProtectSystem = "protect-system"
)

// Default values for Caddy paths and service name. These match the layout
// created by the official Caddy apt package.
const (
	// DefaultCaddyService is the systemd unit name shipped by the Caddy apt
	// package.
	DefaultCaddyService = "caddy"

	// DefaultCaddyUser is the service user created by the Caddy apt package.
	// It has no shell and no login.
	DefaultCaddyUser = "caddy"

	// DefaultCaddyfilePath is the default local path used by Restart when
	// ArgCaddyfilePath is not set. Callers typically override this with their
	// project-specific location (e.g. "webserver/Caddyfile").
	DefaultCaddyfilePath = "Caddyfile"

	// DefaultCaddyfileRemotePath is the remote path the Caddy apt package
	// reads its configuration from.
	DefaultCaddyfileRemotePath = "/etc/caddy/Caddyfile"

	// DefaultCaddyLogDir is where Caddy writes its structured JSON access
	// logs. Created by Install with the caddy user as owner.
	DefaultCaddyLogDir = "/var/log/caddy"

	// DefaultCaddyDataDir is where Caddy stores its data and cache. Used by
	// Harden as a default ReadWritePaths entry.
	DefaultCaddyDataDir = "/var/lib/caddy"

	// DefaultOverrideDir is the systemd drop-in directory for the Caddy unit
	// override. Harden writes override.conf here.
	DefaultOverrideDir = "/etc/systemd/system/caddy.service.d"

	// DefaultProtectHome is the default value for the systemd ProtectHome
	// directive in the Harden override. "true" hides /home, /root, and
	// /run/user from the caddy process — safe when the web root is under
	// /var/www (the apt package's default).
	DefaultProtectHome = "true"

	// DefaultReadWritePaths is the default value for the systemd
	// ReadWritePaths directive in the Harden override. Lists the directories
	// Caddy may write to under an otherwise read-only filesystem
	// (ProtectSystem=strict).
	DefaultReadWritePaths = "/var/lib/caddy /var/log/caddy"

	// DefaultMemoryDenyWriteExecute is the default value for the systemd
	// MemoryDenyWriteExecute directive in the Harden override. "true" blocks
	// W^X memory mappings — a hardening measure that very rarely conflicts
	// with Caddy plugins.
	DefaultMemoryDenyWriteExecute = "true"

	// DefaultProtectSystem is the default value for the systemd ProtectSystem
	// directive in the Harden override. "strict" makes the entire filesystem
	// read-only except for ReadWritePaths. Use "full" if Caddy needs to write
	// to /etc (e.g. ACME certificate storage under /etc/letsencrypt).
	DefaultProtectSystem = "strict"
)
