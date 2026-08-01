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

// Enable enables a systemd unit so it starts automatically at boot, via
// `systemctl enable <service>`. When the ArgStart argument is set to "true",
// it also starts the unit immediately (`systemctl start <service>`), matching
// the common "enable && start" idiom used for timers and one-shot services.
//
// Usage:
//
//	// Enable only (start on next boot)
//	node.Run(systemctl.NewEnable().SetService("caddy"))
//
//	// Enable and start immediately
//	node.Run(systemctl.NewEnable().SetService("mariadb-backup.timer").SetArg(systemctl.ArgStart, "true"))
//
// Idempotency:
//   - Check runs `systemctl is-enabled <service>` and returns true only if
//     the unit is not yet enabled (or the check itself fails, which usually
//     means the unit file does not exist yet)
//   - When ArgStart is "true", Check also verifies the unit is not active, so
//     an already-running enabled unit is left untouched
//
// Result Details:
//   - output: Full output from the enable (and start) command(s)
//   - service: The unit name that was enabled
//   - started: "true" if the unit was also started, "false" otherwise
type Enable struct {
	*types.BaseSkill
}

// Compile-time assertion that Enable implements types.RunnableInterface.
var _ types.RunnableInterface = (*Enable)(nil)

// shouldStart returns true if the ArgStart argument is set to "true".
func (e *Enable) shouldStart() bool {
	return e.GetArg(ArgStart) == "true"
}

// Check determines if the unit needs to be enabled (and optionally started).
// Returns true if the unit is not currently enabled, or if ArgStart is "true"
// and the unit is not currently active. Returns false if the unit is already
// in the desired state.
func (e *Enable) Check() (bool, error) {
	service := e.GetArg(ArgService)
	if service == "" {
		return false, fmt.Errorf("no service specified: set the %q argument", ArgService)
	}

	cfg := e.GetNodeConfig()

	if cfg.IsDryRunMode {
		return true, nil
	}

	// is-enabled exits 0 if enabled, non-zero otherwise.
	cmdEnabled := types.Command{
		Command:     fmt.Sprintf("systemctl is-enabled %s", skills.ShellEscapeArg(service)),
		Description: "Check if unit is enabled: " + service,
		Required:    true, // propagate non-zero exit so we can detect "not enabled"
	}
	_, err := ssh.Run(cfg, cmdEnabled)
	if err != nil {
		if !ssh.IsExitError(err) {
			return false, err
		}
		// Not enabled (or unit file missing) — needs enabling.
		return true, nil
	}

	// Already enabled. If start is requested, check if it's also active.
	if e.shouldStart() {
		cmdActive := types.Command{
			Command:     fmt.Sprintf("systemctl is-active %s", skills.ShellEscapeArg(service)),
			Description: "Check if unit is active: " + service,
			Required:    true,
		}
		_, err := ssh.Run(cfg, cmdActive)
		if err != nil {
			if !ssh.IsExitError(err) {
				return false, err
			}
			// Enabled but not active — needs starting.
			return true, nil
		}
	}

	return false, nil // already enabled (and active if start was requested)
}

// Run enables the unit and optionally starts it.
// Changed is true when the unit was enabled (or started), false if it was
// already in the desired state.
//
// Idempotency:
//   - Run calls Check() first and returns early with Changed=false if the
//     unit is already enabled (and active when ArgStart is set), matching the
//     apt-install / apt-upgrade Check-Run pattern.
//
// Result.Details contains:
//   - output: Full output from the enable (and start) command(s)
//   - service: The unit name that was enabled
//   - started: "true" if the unit was also started, "false" otherwise
func (e *Enable) Run() types.Result {
	service := e.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := e.GetNodeConfig()
	start := e.shouldStart()

	// Idempotency: skip if the unit is already in the desired state.
	// In dry-run mode Check() returns true so the dry-run guard below is reached.
	if !cfg.IsDryRunMode {
		needsEnable, err := e.Check()
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to check if unit needs enabling",
				Error:   err,
			}
		}
		if !needsEnable {
			return types.Result{
				Changed: false,
				Message: "Unit already enabled" + ternary(start, " and active", "") + ": " + service,
			}
		}
	}

	escaped := skills.ShellEscapeArg(service)

	// Build the command. When start is requested, chain enable && start so
	// that a failed enable aborts the start (matching the playbook idiom).
	cmdStr := fmt.Sprintf("systemctl enable %s", escaped)
	if start {
		cmdStr += " && systemctl start " + escaped
	}

	cmd := types.Command{
		Command:     cmdStr,
		Description: "Enable unit" + ternary(start, " and start", "") + ": " + service,
		Required:    true,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Would enable unit" + ternary(start, " and start", "") + ": " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("enabling unit", "service", service, "start", start, "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable unit: " + service,
			Error:   fmt.Errorf("systemctl enable %s failed: %w\nOutput: %s", service, err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("unit enabled", "service", service, "start", start, "host", cfg.SSHHost)
	return types.Result{
		Changed: true,
		Message: "Unit enabled" + ternary(start, " and started", "") + ": " + service,
		Details: map[string]string{
			"output":  output,
			"service": service,
			"started": strconv.FormatBool(start),
		},
	}
}

// SetArgs sets the arguments for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetArgs(args map[string]string) types.RunnableInterface {
	e.BaseSkill.SetArgs(args)
	return e
}

// SetArg sets a single argument for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetArg(key, value string) types.RunnableInterface {
	e.BaseSkill.SetArg(key, value)
	return e
}

// SetService sets the systemd unit name and returns Enable for chaining.
func (e *Enable) SetService(service string) *Enable {
	e.BaseSkill.SetArg(ArgService, service)
	return e
}

// SetStart sets whether to also start the unit after enabling and returns
// Enable for chaining. Pass true to also run `systemctl start`.
func (e *Enable) SetStart(start bool) *Enable {
	e.BaseSkill.SetArg(ArgStart, strconv.FormatBool(start))
	return e
}

// WithNodeConfig sets the node config and returns Enable for chaining.
func (e *Enable) WithNodeConfig(cfg types.NodeConfig) *Enable {
	e.BaseSkill.SetNodeConfig(cfg)
	return e
}

// SetID sets the ID for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetID(id string) types.RunnableInterface {
	e.BaseSkill.SetID(id)
	return e
}

// SetDescription sets the description for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetDescription(description string) types.RunnableInterface {
	e.BaseSkill.SetDescription(description)
	return e
}

// SetTimeout sets the timeout for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetTimeout(timeout time.Duration) types.RunnableInterface {
	e.BaseSkill.SetTimeout(timeout)
	return e
}

// NewEnable creates a new systemctl-enable skill.
//
// Returns an Enable skill configured with IDSystemctlEnable identifier and
// description "Enable a systemd unit (optionally start it)".
func NewEnable() *Enable {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlEnable)
	pb.SetDescription("Enable a systemd unit (optionally start it)")
	return &Enable{BaseSkill: pb}
}
