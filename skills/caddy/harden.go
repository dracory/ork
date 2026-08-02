package caddy

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/types"
)

// Harden applies security hardening to the Caddy systemd unit via a drop-in
// override. It writes a sandboxing profile (ProtectSystem=strict,
// ProtectHome, PrivateTmp, NoNewPrivileges, capability bounding, etc.) to
// the unit's override directory, reloads systemd, and restarts Caddy so the
// directives take effect.
//
// The opinionated directives are exposed as args so callers can adapt the
// profile to their layout:
//   - protect-home: "true" (default) | "false" | "read-only". Set to "false"
//     if the web root lives under /home, otherwise Caddy cannot read it.
//   - protect-system: "strict" (default) | "full" | "true". Set to "full"
//     if Caddy manages ACME certificates and needs to write to /etc/letsencrypt.
//   - read-write-paths: space-separated paths Caddy may write to (default
//     "/var/lib/caddy /var/log/caddy"). Add custom data/cache directories.
//   - memory-deny-write-execute: "true" (default) | "false". Set to "false"
//     if a Caddy plugin needs W^X memory (rare).
//   - override-dir: drop-in directory (default
//     "/etc/systemd/system/caddy.service.d").
//
// Usage:
//
//	node.Run(caddy.NewHarden())
//	node.Run(caddy.NewHarden().SetProtectHome("false").SetReadWritePaths("/var/lib/caddy /var/log/caddy /srv/uploads"))
//
// Execution Flow:
//  1. Creates the systemd override directory (default
//     /etc/systemd/system/caddy.service.d) with mode 755
//  2. Writes override.conf with the sandboxing directives, overwriting any
//     existing override
//  3. Runs `systemctl daemon-reload` so systemd picks up the override
//  4. Restarts the caddy unit so the sandboxing directives apply to the
//     running process (daemon-reload alone does not apply them)
//
// Prerequisites:
//   - Caddy must be installed (see Install)
//   - Root SSH access required to write under /etc/systemd/system
//
// Related Skills:
//   - caddy.Install: Install Caddy via apt
//   - caddy.Restart: Upload a Caddyfile and reload the service
//   - caddy.Status: Show the Caddy systemd unit status
type Harden struct {
	*types.BaseSkill
}

// Compile-time assertion that Harden implements types.RunnableInterface.
var _ types.RunnableInterface = (*Harden)(nil)

// Check returns true when the override file content differs from the desired
// content, indicating a change is needed. Returns true in dry-run mode.
func (h *Harden) Check() (bool, error) {
	cfg := h.GetNodeConfig()
	if cfg.IsDryRunMode {
		return true, nil
	}
	// Conservative: always apply to guarantee the override matches the
	// configured args. A content-diff check would require reading the remote
	// file; the FileCreate sub-skill already skips the write when content,
	// mode, and owner all match.
	return true, nil
}

// Run applies the systemd unit override for Caddy hardening.
func (h *Harden) Run() types.Result {
	cfg := h.GetNodeConfig()

	overrideDir := h.GetArg(ArgOverrideDir)
	if overrideDir == "" {
		overrideDir = DefaultOverrideDir
	}
	protectHome := h.GetArg(ArgProtectHome)
	if protectHome == "" {
		protectHome = DefaultProtectHome
	}
	protectSystem := h.GetArg(ArgProtectSystem)
	if protectSystem == "" {
		protectSystem = DefaultProtectSystem
	}
	readWritePaths := h.GetArg(ArgReadWritePaths)
	if readWritePaths == "" {
		readWritePaths = DefaultReadWritePaths
	}
	memoryDenyWriteExecute := h.GetArg(ArgMemoryDenyWriteExecute)
	if memoryDenyWriteExecute == "" {
		memoryDenyWriteExecute = DefaultMemoryDenyWriteExecute
	}

	// Step 1: Create the systemd override directory.
	dirResult := runSub(fs.NewDirCreate().
		SetPath(overrideDir).
		SetMode("755"), cfg)
	if dirResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create Caddy systemd override directory",
			Error:   dirResult.Error,
		}
	}

	// Step 2: Write the systemd drop-in override with hardening directives.
	// The official Caddy package already ships with some of these, but we
	// explicitly set them all to ensure they're present regardless of package
	// version changes.
	overrideContent := buildOverrideContent(protectSystem, protectHome, readWritePaths, memoryDenyWriteExecute)

	overridePath := strings.TrimRight(overrideDir, "/") + "/override.conf"
	fileResult := runSub(fs.NewFileCreate().
		SetPath(overridePath).
		SetContent(overrideContent).
		SetMode("644").
		SetOverwrite(true), cfg)
	if fileResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write Caddy systemd override",
			Error:   fileResult.Error,
		}
	}

	// Step 3: Reload systemd to pick up the override.
	daemonResult := runSub(systemctl.NewDaemonReload(), cfg)
	if daemonResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reload systemd",
			Error:   daemonResult.Error,
		}
	}

	// Step 4: Restart Caddy so the sandboxing directives (ProtectSystem,
	// ProtectHome, etc.) take effect. daemon-reload alone does not apply
	// them to the currently running process.
	restartResult := runSub(systemctl.NewRestart().SetService(DefaultCaddyService), cfg)
	if restartResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart Caddy after hardening",
			Error:   restartResult.Error,
		}
	}

	return types.Result{
		Changed: true,
		Message: "Caddy systemd unit hardened and restarted successfully",
	}
}

// buildOverrideContent renders the systemd drop-in override file content from
// the configurable directives. The non-configurable directives are
// hardening best-practices that are safe for the standard apt-installed
// Caddy and should not vary between deployments.
func buildOverrideContent(protectSystem, protectHome, readWritePaths, memoryDenyWriteExecute string) string {
	var b strings.Builder
	b.WriteString("[Service]\n")
	b.WriteString("# Sandbox: make most of the filesystem read-only\n")
	b.WriteString(fmt.Sprintf("ProtectSystem=%s\n", protectSystem))
	b.WriteString("# Sandbox: hide /home, /root, /run/user (set to false if web root is under /home)\n")
	b.WriteString(fmt.Sprintf("ProtectHome=%s\n", protectHome))
	b.WriteString("# Sandbox: private /tmp\n")
	b.WriteString("PrivateTmp=true\n")
	b.WriteString("# Sandbox: no access to physical devices\n")
	b.WriteString("PrivateDevices=true\n")
	b.WriteString("# Sandbox: cannot gain new privileges\n")
	b.WriteString("NoNewPrivileges=true\n")
	b.WriteString("# Capabilities: port binding + network manipulation (matches official caddy.service)\n")
	b.WriteString("AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE\n")
	b.WriteString("CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE\n")
	b.WriteString("# Restrict what the process can do\n")
	b.WriteString("ProtectKernelTunables=true\n")
	b.WriteString("ProtectKernelModules=true\n")
	b.WriteString("ProtectControlGroups=true\n")
	b.WriteString("RestrictSUIDSGID=true\n")
	b.WriteString("RestrictNamespaces=true\n")
	b.WriteString("LockPersonality=true\n")
	b.WriteString(fmt.Sprintf("MemoryDenyWriteExecute=%s\n", memoryDenyWriteExecute))
	b.WriteString("# Allow writing to Caddy's data/cache and log directories\n")
	b.WriteString(fmt.Sprintf("ReadWritePaths=%s\n", readWritePaths))
	return b.String()
}

// SetArgs sets the arguments for the Caddy hardening.
// Returns Harden for fluent method chaining.
func (h *Harden) SetArgs(args map[string]string) types.RunnableInterface {
	h.BaseSkill.SetArgs(args)
	return h
}

// SetArg sets a single argument for the Caddy hardening.
// Returns Harden for fluent method chaining.
func (h *Harden) SetArg(key, value string) types.RunnableInterface {
	h.BaseSkill.SetArg(key, value)
	return h
}

// SetOverrideDir sets the systemd drop-in directory and returns Harden for
// chaining. Default: "/etc/systemd/system/caddy.service.d".
func (h *Harden) SetOverrideDir(dir string) *Harden {
	h.BaseSkill.SetArg(ArgOverrideDir, dir)
	return h
}

// SetProtectHome sets the systemd ProtectHome directive value and returns
// Harden for chaining. Accepted values: "true", "false", "read-only".
// Default: "true". Set to "false" if the web root is under /home.
func (h *Harden) SetProtectHome(value string) *Harden {
	h.BaseSkill.SetArg(ArgProtectHome, value)
	return h
}

// SetReadWritePaths sets the space-separated ReadWritePaths list and returns
// Harden for chaining. Default: "/var/lib/caddy /var/log/caddy".
func (h *Harden) SetReadWritePaths(paths string) *Harden {
	h.BaseSkill.SetArg(ArgReadWritePaths, paths)
	return h
}

// SetMemoryDenyWriteExecute sets the MemoryDenyWriteExecute directive value
// and returns Harden for chaining. Accepted values: "true", "false".
// Default: "true". Set to "false" if a Caddy plugin needs W^X memory.
func (h *Harden) SetMemoryDenyWriteExecute(value string) *Harden {
	h.BaseSkill.SetArg(ArgMemoryDenyWriteExecute, value)
	return h
}

// SetProtectSystem sets the systemd ProtectSystem directive value and
// returns Harden for chaining. Accepted values: "strict", "full", "true".
// Default: "strict". Set to "full" if Caddy manages ACME certificates
// itself and needs to write to /etc/letsencrypt.
func (h *Harden) SetProtectSystem(value string) *Harden {
	h.BaseSkill.SetArg(ArgProtectSystem, value)
	return h
}

// SetID sets the ID for the Caddy hardening.
// Returns Harden for fluent method chaining.
func (h *Harden) SetID(id string) types.RunnableInterface {
	h.BaseSkill.SetID(id)
	return h
}

// SetDescription sets the description for the Caddy hardening.
// Returns Harden for fluent method chaining.
func (h *Harden) SetDescription(description string) types.RunnableInterface {
	h.BaseSkill.SetDescription(description)
	return h
}

// SetTimeout sets the timeout for the Caddy hardening.
// Returns Harden for fluent method chaining.
func (h *Harden) SetTimeout(timeout time.Duration) types.RunnableInterface {
	h.BaseSkill.SetTimeout(timeout)
	return h
}

// NewHarden creates a new caddy-harden skill.
//
// Returns a Harden skill configured with skills.IDCaddyHarden identifier and
// description "Harden Caddy systemd unit with sandboxing directives".
//
// Example:
//
//	node.Run(caddy.NewHarden())
//	node.Run(caddy.NewHarden().SetProtectHome("false"))
func NewHarden() *Harden {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDCaddyHarden)
	pb.SetDescription("Harden Caddy systemd unit with sandboxing directives")
	return &Harden{BaseSkill: pb}
}
