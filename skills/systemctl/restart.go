package systemctl

// Package systemctl documentation is in status.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Restart restarts a systemd unit via `systemctl restart <service>`.
// Restart is an explicit action: like the reboot skill, it is always
// "needed" when requested, so Check always returns true. Restart stops and
// starts the unit, applying any configuration changes (e.g. sandboxing
// directives) that daemon-reload alone does not apply to a running process.
//
// Usage:
//
//	node.Run(systemctl.NewRestart().SetService("caddy"))
//
// Execution Flow:
//  1. Connects to the remote server via SSH
//  2. Runs `systemctl restart <service>`
//  3. Reports success or failure
//
// Result Details:
//   - output: Full output from systemctl restart
//   - service: The unit name that was restarted
type Restart struct {
	*types.BaseSkill
}

// Compile-time assertion that Restart implements types.RunnableInterface.
var _ types.RunnableInterface = (*Restart)(nil)

// Check always returns true for restart.
// Per the skill interface convention, the bool return indicates whether the
// operation would modify system state. Restart is an explicit action that
// always modifies the unit's process state, so this always returns true.
// Like the reboot skill, restart is always "needed" when requested by the user.
func (r *Restart) Check() (bool, error) {
	return true, nil
}

// Run executes systemctl restart and returns the result.
// Changed is true when the restart succeeds, false on failure.
//
// Result.Details contains:
//   - output: Full output from systemctl restart
//   - service: The unit name that was restarted
func (r *Restart) Run() types.Result {
	service := r.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := r.GetNodeConfig()
	cmdStr := fmt.Sprintf("systemctl restart %s", skills.ShellEscapeArg(service))
	cmd := types.Command{
		Command:     cmdStr,
		Description: "Restart unit: " + service,
		Required:    true,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: "Would restart unit: " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("restarting unit", "service", service, "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart unit: " + service,
			Error:   fmt.Errorf("systemctl restart %s failed: %w\nOutput: %s", service, err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("unit restarted", "service", service, "host", cfg.SSHHost)
	return types.Result{
		Changed: true,
		Message: "Unit restarted: " + service,
		Details: map[string]string{
			"output":  output,
			"service": service,
		},
	}
}

// SetArgs sets the arguments for restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetService sets the systemd unit name and returns Restart for chaining.
func (r *Restart) SetService(service string) *Restart {
	r.BaseSkill.SetArg(ArgService, service)
	return r
}

// WithNodeConfig sets the node config and returns Restart for chaining.
func (r *Restart) WithNodeConfig(cfg types.NodeConfig) *Restart {
	r.BaseSkill.SetNodeConfig(cfg)
	return r
}

// SetID sets the ID for restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewRestart creates a new systemctl-restart skill.
//
// Returns a Restart skill configured with IDSystemctlRestart identifier and
// description "Restart a systemd unit".
//
// Example:
//
//	node.Run(systemctl.NewRestart().SetService("caddy"))
func NewRestart() *Restart {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlRestart)
	pb.SetDescription("Restart a systemd unit")
	return &Restart{BaseSkill: pb}
}
