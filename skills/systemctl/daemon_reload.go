package systemctl

// Package systemctl documentation is in status.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// DaemonReload reloads the systemd manager configuration.
// This is required after creating or modifying unit files (e.g. drop-in
// overrides in /etc/systemd/system/<unit>.service.d/) so that systemd picks
// up the changes. It does not restart any units; use Restart or Reload for
// that.
//
// Usage:
//
//	node.Run(systemctl.NewDaemonReload())
//
// Execution Flow:
//  1. Connects to the remote server via SSH
//  2. Runs `systemctl daemon-reload`
//  3. Reports success or failure
//
// Result Details:
//   - output: Full output from systemctl daemon-reload
type DaemonReload struct {
	*types.BaseSkill
}

// Compile-time assertion that DaemonReload implements types.RunnableInterface.
var _ types.RunnableInterface = (*DaemonReload)(nil)

// Check always returns true for daemon-reload.
// Per the skill interface convention, the bool return indicates whether the
// operation would modify system state. daemon-reload always reloads the
// manager configuration (a state change), so this always returns true.
// Like the reboot skill, daemon-reload is an explicit action that is always
// "needed" when requested.
func (d *DaemonReload) Check() (bool, error) {
	return true, nil
}

// Run executes systemctl daemon-reload and returns the result.
// Changed is true when the reload succeeds, false on failure.
//
// Result.Details contains:
//   - output: Full output from systemctl daemon-reload
func (d *DaemonReload) Run() types.Result {
	cfg := d.GetNodeConfig()
	cmd := types.Command{
		Command:     "systemctl daemon-reload",
		Description: "Reload systemd manager configuration",
		Required:    true,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Would reload systemd manager configuration",
		}
	}

	cfg.GetLoggerOrDefault().Info("reloading systemd manager", "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reload systemd manager",
			Error:   fmt.Errorf("systemctl daemon-reload failed: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("systemd manager reloaded", "host", cfg.SSHHost)
	return types.Result{
		Changed: true,
		Message: "Systemd manager configuration reloaded",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for daemon-reload.
// Returns DaemonReload for fluent method chaining.
func (d *DaemonReload) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument for daemon-reload.
// Returns DaemonReload for fluent method chaining.
func (d *DaemonReload) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// WithNodeConfig sets the node config and returns DaemonReload for chaining.
func (d *DaemonReload) WithNodeConfig(cfg types.NodeConfig) *DaemonReload {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetID sets the ID for daemon-reload.
// Returns DaemonReload for fluent method chaining.
func (d *DaemonReload) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for daemon-reload.
// Returns DaemonReload for fluent method chaining.
func (d *DaemonReload) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for daemon-reload.
// Returns DaemonReload for fluent method chaining.
func (d *DaemonReload) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDaemonReload creates a new systemctl-daemon-reload skill.
//
// Returns a DaemonReload skill configured with IDSystemctlDaemonReload
// identifier and description "Reload systemd manager configuration".
func NewDaemonReload() *DaemonReload {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlDaemonReload)
	pb.SetDescription("Reload systemd manager configuration")
	return &DaemonReload{BaseSkill: pb}
}
