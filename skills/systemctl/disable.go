package systemctl

// Package systemctl documentation is in status.go

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Disable disables a systemd unit so it no longer starts at boot, via
// `systemctl disable <service>`. When the ArgStop argument is set to "true",
// it also stops the unit immediately (`systemctl stop <service>`).
//
// Usage:
//
//	// Disable only (stop on next boot)
//	node.Run(systemctl.NewDisable().SetService("old-service"))
//
//	// Disable and stop immediately
//	node.Run(systemctl.NewDisable().SetService("old-service").SetArg(systemctl.ArgStop, "true"))
//
// Idempotency:
//   - Check runs `systemctl is-enabled <service>` and returns true only if
//     the unit is currently enabled (or the check fails in a way that
//     suggests the unit exists but is enabled)
//   - When ArgStop is "true", Check also verifies the unit is active, so an
//     already-stopped disabled unit is left untouched
//
// Result Details:
//   - output: Full output from the disable (and stop) command(s)
//   - service: The unit name that was disabled
//   - stopped: "true" if the unit was also stopped, "false" otherwise
type Disable struct {
	*types.BaseSkill
}

// Compile-time assertion that Disable implements types.RunnableInterface.
var _ types.RunnableInterface = (*Disable)(nil)

// shouldStop returns true if the ArgStop argument is set to "true".
func (d *Disable) shouldStop() bool {
	return d.GetArg(ArgStop) == "true"
}

// Check determines if the unit needs to be disabled (and optionally stopped).
// Returns true if the unit is currently enabled, or if ArgStop is "true" and
// the unit is currently active. Returns false if the unit is already in the
// desired state (disabled, and stopped if requested).
func (d *Disable) Check() (bool, error) {
	service := d.GetArg(ArgService)
	if service == "" {
		return false, fmt.Errorf("no service specified: set the %q argument", ArgService)
	}

	cfg := d.GetNodeConfig()

	if cfg.IsDryRunMode {
		return true, nil
	}

	// is-enabled exits 0 if enabled, non-zero if disabled or not found.
	cmdEnabled := types.Command{
		Command:     fmt.Sprintf("systemctl is-enabled %s", skills.ShellEscapeArg(service)),
		Description: "Check if unit is enabled: " + service,
		Required:    true,
	}
	_, err := ssh.Run(cfg, cmdEnabled)
	if err != nil {
		if !ssh.IsExitError(err) {
			return false, err
		}
		// Already disabled (or unit file missing). If stop is requested,
		// check if it's still active.
		if d.shouldStop() {
			cmdActive := types.Command{
				Command:     fmt.Sprintf("systemctl is-active %s", skills.ShellEscapeArg(service)),
				Description: "Check if unit is active: " + service,
				Required:    true,
			}
			_, actErr := ssh.Run(cfg, cmdActive)
			if actErr != nil {
				if !ssh.IsExitError(actErr) {
					return false, actErr
				}
				// Disabled and not active — nothing to do.
				return false, nil
			}
			// Disabled but still running — needs stopping.
			return true, nil
		}
		return false, nil // already disabled, no stop requested
	}

	// Currently enabled — needs disabling.
	return true, nil
}

// Run disables the unit and optionally stops it.
// Changed is true when the unit was disabled (or stopped), false if it was
// already in the desired state.
//
// Idempotency:
//   - Run calls Check() first and returns early with Changed=false if the
//     unit is already disabled (and stopped when ArgStop is set), matching the
//     apt-install / apt-upgrade Check-Run pattern.
//
// Result.Details contains:
//   - output: Full output from the disable (and stop) command(s)
//   - service: The unit name that was disabled
//   - stopped: "true" if the unit was also stopped, "false" otherwise
func (d *Disable) Run() types.Result {
	service := d.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := d.GetNodeConfig()
	stop := d.shouldStop()

	// Idempotency: skip if the unit is already in the desired state.
	// In dry-run mode Check() returns true so the dry-run guard below is reached.
	if !cfg.IsDryRunMode {
		needsDisable, err := d.Check()
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to check if unit needs disabling",
				Error:   err,
			}
		}
		if !needsDisable {
			return types.Result{
				Changed: false,
				Message: "Unit already disabled" + ternary(stop, " and stopped", "") + ": " + service,
			}
		}
	}

	escaped := skills.ShellEscapeArg(service)

	cmdStr := fmt.Sprintf("systemctl disable %s", escaped)
	if stop {
		cmdStr += " && systemctl stop " + escaped
	}

	cmd := types.Command{
		Command:     cmdStr,
		Description: "Disable unit" + ternary(stop, " and stop", "") + ": " + service,
		Required:    true,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Would disable unit" + ternary(stop, " and stop", "") + ": " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("disabling unit", "service", service, "stop", stop, "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to disable unit: " + service,
			Error:   fmt.Errorf("systemctl disable %s failed: %w\nOutput: %s", service, err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("unit disabled", "service", service, "stop", stop, "host", cfg.SSHHost)
	return types.Result{
		Changed: true,
		Message: "Unit disabled" + ternary(stop, " and stopped", "") + ": " + service,
		Details: map[string]string{
			"output":  output,
			"service": service,
			"stopped": strconv.FormatBool(stop),
		},
	}
}

// SetArgs sets the arguments for disable.
// Returns Disable for fluent method chaining.
func (d *Disable) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument for disable.
// Returns Disable for fluent method chaining.
func (d *Disable) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetService sets the systemd unit name and returns Disable for chaining.
func (d *Disable) SetService(service string) *Disable {
	d.BaseSkill.SetArg(ArgService, service)
	return d
}

// SetStop sets whether to also stop the unit after disabling and returns
// Disable for chaining. Pass true to also run `systemctl stop`.
func (d *Disable) SetStop(stop bool) *Disable {
	d.BaseSkill.SetArg(ArgStop, strconv.FormatBool(stop))
	return d
}

// WithNodeConfig sets the node config and returns Disable for chaining.
func (d *Disable) WithNodeConfig(cfg types.NodeConfig) *Disable {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetID sets the ID for disable.
// Returns Disable for fluent method chaining.
func (d *Disable) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for disable.
// Returns Disable for fluent method chaining.
func (d *Disable) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for disable.
// Returns Disable for fluent method chaining.
func (d *Disable) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDisable creates a new systemctl-disable skill.
//
// Returns a Disable skill configured with IDSystemctlDisable identifier and
// description "Disable a systemd unit (optionally stop it)".
func NewDisable() *Disable {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlDisable)
	pb.SetDescription("Disable a systemd unit (optionally stop it)")
	return &Disable{BaseSkill: pb}
}
