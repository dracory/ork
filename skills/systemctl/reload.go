package systemctl

// Package systemctl documentation is in status.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Reload reloads a systemd unit's configuration via `systemctl reload`.
// Reload is a graceful operation that asks the service to re-read its config
// without stopping (zero-downtime). Not all services support reload; when
// reload fails, this skill automatically falls back to `systemctl restart`
// to ensure the new configuration takes effect. This matches the common
// `systemctl reload <unit> || systemctl restart <unit>` idiom.
//
// Usage:
//
//	node.Run(systemctl.NewReload().SetService("caddy"))
//
// Execution Flow:
//  1. Connects to the remote server via SSH
//  2. Runs `systemctl reload <service>`
//  3. If reload fails (non-zero exit), runs `systemctl restart <service>`
//  4. Reports which operation succeeded (or both failures)
//
// Result Details:
//   - output: Full output from the successful (or final) command
//   - service: The unit name that was reloaded
//   - method: "reload" or "restart" (which command actually applied the change)
type Reload struct {
	*types.BaseSkill
}

// Compile-time assertion that Reload implements types.RunnableInterface.
var _ types.RunnableInterface = (*Reload)(nil)

// Check always returns true for reload.
// Per the skill interface convention, the bool return indicates whether the
// operation would modify system state. Reload (or its restart fallback)
// always modifies the unit's process state, so this always returns true.
func (r *Reload) Check() (bool, error) {
	return true, nil
}

// Run executes systemctl reload (with restart fallback) and returns the result.
// Changed is true when either reload or restart succeeds, false if both fail.
//
// Result.Details contains:
//   - output: Full output from the successful (or final) command
//   - service: The unit name that was reloaded
//   - method: "reload" or "restart" indicating which command applied the change
func (r *Reload) Run() types.Result {
	service := r.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := r.GetNodeConfig()
	escaped := skills.ShellEscapeArg(service)

	cmdReload := types.Command{
		Command:     fmt.Sprintf("systemctl reload %s", escaped),
		Description: "Reload unit: " + service,
		Required:    true, // propagate error so we can fall back to restart
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdReload.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Would reload unit (with restart fallback): " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("reloading unit", "service", service, "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmdReload)
	if err == nil {
		cfg.GetLoggerOrDefault().Info("unit reloaded", "service", service, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Unit reloaded: " + service,
			Details: map[string]string{
				"output":  output,
				"service": service,
				"method":  "reload",
			},
		}
	}

	// Reload failed. If it's an SSH connection error (not a command exit
	// error), propagate it directly — attempting restart would also fail
	// and produce a misleading "both failed" error message.
	if !ssh.IsExitError(err) {
		return types.Result{
			Changed: false,
			Message: "Failed to reload unit: " + service,
			Error:   fmt.Errorf("systemctl reload %s failed: %w\nOutput: %s", service, err, output),
		}
	}

	// Exit error — service may not support reload, or unit not active.
	// Fall back to restart so the configuration change still takes effect.
	cfg.GetLoggerOrDefault().Info("reload failed, falling back to restart", "service", service, "reload_error", err)

	cmdRestart := types.Command{
		Command:     fmt.Sprintf("systemctl restart %s", escaped),
		Description: "Restart unit (fallback after reload failed): " + service,
		Required:    true,
	}

	restartOutput, restartErr := ssh.Run(cfg, cmdRestart)
	if restartErr != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reload and restart unit: " + service,
			Error: fmt.Errorf("systemctl reload %s failed: %w (reload output: %s); "+
				"systemctl restart %s also failed: %v (restart output: %s)",
				service, err, output, service, restartErr, restartOutput),
		}
	}

	cfg.GetLoggerOrDefault().Info("unit restarted after reload fallback", "service", service, "host", cfg.SSHHost)
	return types.Result{
		Changed: true,
		Message: "Unit reloaded via restart fallback: " + service,
		Details: map[string]string{
			"output":  restartOutput,
			"service": service,
			"method":  "restart",
		},
	}
}

// SetArgs sets the arguments for reload.
// Returns Reload for fluent method chaining.
func (r *Reload) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for reload.
// Returns Reload for fluent method chaining.
func (r *Reload) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetService sets the systemd unit name and returns Reload for chaining.
func (r *Reload) SetService(service string) *Reload {
	r.BaseSkill.SetArg(ArgService, service)
	return r
}

// WithNodeConfig sets the node config and returns Reload for chaining.
func (r *Reload) WithNodeConfig(cfg types.NodeConfig) *Reload {
	r.BaseSkill.SetNodeConfig(cfg)
	return r
}

// SetID sets the ID for reload.
// Returns Reload for fluent method chaining.
func (r *Reload) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for reload.
// Returns Reload for fluent method chaining.
func (r *Reload) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for reload.
// Returns Reload for fluent method chaining.
func (r *Reload) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewReload creates a new systemctl-reload skill.
//
// Returns a Reload skill configured with IDSystemctlReload identifier and
// description "Reload a systemd unit (with restart fallback)".
//
// Example:
//
//	node.Run(systemctl.NewReload().SetService("caddy"))
func NewReload() *Reload {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlReload)
	pb.SetDescription("Reload a systemd unit (with restart fallback)")
	return &Reload{BaseSkill: pb}
}
