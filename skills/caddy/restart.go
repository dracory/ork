package caddy

import (
	"fmt"
	"os"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Restart uploads a local Caddyfile and restarts Caddy via systemd.
// It reads the Caddyfile from the local filesystem (ArgCaddyfilePath), uploads
// it to the remote server (ArgCaddyfileRemotePath), validates the syntax with
// `caddy validate`, then reloads the caddy systemd unit (with a restart
// fallback), and finally verifies the service is active.
//
// Usage:
//
//	node.Run(caddy.NewRestart().SetCaddyfilePath("webserver/Caddyfile"))
//	node.Run(caddy.NewRestart()) // uses DefaultCaddyfilePath ("Caddyfile")
//
// Execution Flow:
//  1. Reads the Caddyfile from the local path (ArgCaddyfilePath, default "Caddyfile")
//  2. Ensures the Caddy log directory exists and fixes ownership of any existing
//     access.log that may have been created by a pre-hardening Caddy run as root
//  3. Uploads it to the remote path (ArgCaddyfileRemotePath, default
//     "/etc/caddy/Caddyfile") owned by root:caddy with mode 644
//  4. Validates the Caddyfile syntax with `caddy validate`
//  5. Reloads the caddy systemd unit (graceful, zero-downtime, with restart fallback)
//  6. Verifies the caddy service is active via `systemctl is-active`
//
// Args:
//   - caddyfile-path: Local path to the Caddyfile to upload (default: "Caddyfile")
//   - caddyfile-remote-path: Remote path to write the Caddyfile (default:
//     "/etc/caddy/Caddyfile")
//
// Prerequisites:
//   - Caddy must be installed (see Install)
//   - The local Caddyfile must exist at the configured path
//
// Related Skills:
//   - caddy.Install: Install Caddy via apt
//   - caddy.Status: Show the Caddy systemd unit status
type Restart struct {
	*types.BaseSkill
}

// Compile-time assertion that Restart implements types.RunnableInterface.
var _ types.RunnableInterface = (*Restart)(nil)

// Check always returns true since Restart is intentionally non-idempotent —
// it re-uploads the Caddyfile and reloads on every run to apply the current
// local configuration.
func (r *Restart) Check() (bool, error) {
	return true, nil
}

// Run uploads the Caddyfile and restarts Caddy via systemd.
func (r *Restart) Run() types.Result {
	cfg := r.GetNodeConfig()

	localPath := r.GetArg(ArgCaddyfilePath)
	if localPath == "" {
		localPath = DefaultCaddyfilePath
	}
	remotePath := r.GetArg(ArgCaddyfileRemotePath)
	if remotePath == "" {
		remotePath = DefaultCaddyfileRemotePath
	}

	// Dry-run mode: log intent and return without reading local files or executing SSH.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would upload Caddyfile and restart Caddy",
			"localPath", localPath, "remotePath", remotePath)
		return types.Result{
			Changed: true,
			Message: "Would upload Caddyfile and restart Caddy: " + localPath + " -> " + remotePath,
		}
	}

	// Step 1: Read Caddyfile from local file.
	caddyfileContent, err := os.ReadFile(localPath)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to read Caddyfile: " + localPath,
			Error:   err,
		}
	}

	// Step 2: Ensure the Caddy log directory exists. If caddy-harden sets
	// ProtectSystem=strict, Caddy cannot create /var/log/caddy inside the
	// sandbox — it must already exist with the right ownership.
	// Also fix ownership of an existing access.log that may have been created
	// by a pre-hardening Caddy run as root (mode 600, owned by root).
	logDirResult := types.RunSub(fs.NewDirCreate().
		SetPath(DefaultCaddyLogDir).
		SetOwner(DefaultCaddyUser+":"+DefaultCaddyUser).
		SetMode("755"), cfg)
	if logDirResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create Caddy log directory",
			Error:   logDirResult.Error,
		}
	}

	logOwnerResult := types.RunSub(fs.NewChangeOwner().
		SetPath(DefaultCaddyLogDir+"/access.log").
		SetOwner(DefaultCaddyUser+":"+DefaultCaddyUser), cfg)
	if logOwnerResult.Error != nil {
		cfg.GetLoggerOrDefault().Warn("Failed to fix Caddy access.log ownership (non-fatal)",
			"error", logOwnerResult.Error)
	}

	// Step 3: Upload Caddyfile to the remote path (owned by root:caddy,
	// group-readable so the caddy process can read it).
	fileResult := types.RunSub(fs.NewFileCreate().
		SetPath(remotePath).
		SetContent(string(caddyfileContent)).
		SetOwner("root:"+DefaultCaddyUser).
		SetMode("644").
		SetOverwrite(true), cfg)
	if fileResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to upload Caddyfile",
			Error:   fileResult.Error,
		}
	}

	// Step 4: Validate Caddyfile before applying.
	cmdValidate := types.Command{
		Command:     "caddy validate --config " + skills.ShellEscapeArg(remotePath),
		Description: "Validate Caddyfile syntax",
		Required:    true,
	}
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdValidate.Command)
	} else {
		output, err := ssh.Run(cfg, cmdValidate)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Caddyfile validation failed",
				Error:   fmt.Errorf("%w: %s", err, output),
			}
		}
	}

	// Step 5: Reload Caddy via systemd (graceful, zero-downtime, with restart
	// fallback handled by the systemctl.Reload skill).
	reloadResult := types.RunSub(systemctl.NewReload().SetService(DefaultCaddyService), cfg)
	if reloadResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reload/restart Caddy",
			Error:   reloadResult.Error,
		}
	}

	// Step 6: Verify Caddy is running.
	activeResult := types.RunSub(systemctl.NewIsActive().SetService(DefaultCaddyService), cfg)
	if activeResult.Error != nil {
		return types.Result{
			Changed: true,
			Message: "Caddy restarted but failed to verify active state",
			Error:   activeResult.Error,
		}
	}
	state := activeResult.Details["state"]
	if state != "active" {
		return types.Result{
			Changed: true,
			Message: "Caddy restart failed — service state is: " + state,
			Error:   fmt.Errorf("caddy service is not active after restart (state: %s)", state),
		}
	}

	return types.Result{
		Changed: true,
		Message: "Caddy restarted successfully via systemd: " + state,
	}
}

// SetArgs sets the arguments for the Caddy restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for the Caddy restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetCaddyfilePath sets the local path to the Caddyfile to upload and returns
// Restart for chaining. Example: SetCaddyfilePath("webserver/Caddyfile")
func (r *Restart) SetCaddyfilePath(path string) *Restart {
	r.BaseSkill.SetArg(ArgCaddyfilePath, path)
	return r
}

// SetCaddyfileRemotePath sets the remote path where the Caddyfile is written
// and returns Restart for chaining. Defaults to "/etc/caddy/Caddyfile".
func (r *Restart) SetCaddyfileRemotePath(path string) *Restart {
	r.BaseSkill.SetArg(ArgCaddyfileRemotePath, path)
	return r
}

// SetID sets the ID for the Caddy restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for the Caddy restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for the Caddy restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewRestart creates a new caddy-restart skill.
//
// Returns a Restart skill configured with skills.IDCaddyRestart identifier and
// description "Upload Caddyfile and restart Caddy via systemd".
//
// Example:
//
//	node.Run(caddy.NewRestart().SetCaddyfilePath("webserver/Caddyfile"))
func NewRestart() *Restart {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDCaddyRestart)
	pb.SetDescription("Upload Caddyfile and restart Caddy via systemd")
	return &Restart{BaseSkill: pb}
}
