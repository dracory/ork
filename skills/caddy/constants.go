// Package caddy provides skills for managing the Caddy web server when
// installed via the official apt repository on Debian/Ubuntu.
//
// It covers the three operations needed to provision and operate Caddy:
//   - Install: add the official Caddy apt repo, install the package, and
//     prepare the log directory (Install).
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
)
