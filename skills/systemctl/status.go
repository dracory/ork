// Package systemctl provides skills for managing systemd units on remote
// servers via SSH. It covers the common systemctl verbs used in provisioning
// and maintenance playbooks: daemon-reload, restart, reload (with restart
// fallback), status, is-active, enable, and disable.
//
// Each skill shell-escapes the unit name via skills.ShellEscapeArg to prevent
// injection, honors dry-run mode, and reports a structured types.Result with
// the command output in Details.
//
// Usage:
//
//	node.Run(systemctl.NewRestart().SetService("caddy"))
//	node.Run(systemctl.NewReload().SetService("caddy"))
//	node.Run(systemctl.NewStatus().SetService("caddy"))
//	node.Run(systemctl.NewDaemonReload())
//	node.Run(systemctl.NewEnable().SetService("mariadb-backup.timer").SetArg(systemctl.ArgStart, "true"))
package systemctl

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Status shows the status of a systemd unit via `systemctl status`.
// This is a read-only skill: it never modifies system state, so Changed is
// always false and Check always returns false.
//
// Usage:
//
//	node.Run(systemctl.NewStatus().SetService("caddy"))
//
// Execution Flow:
//  1. Connects to the remote server via SSH
//  2. Runs `systemctl status <service> --no-pager -l`
//  3. Returns the full status output in Result.Details["output"]
//
// The `--no-pager -l` flags disable the pager and expand ellipsized lines so
// the output is safe to capture over SSH.
//
// Result Details:
//   - output: Full output from systemctl status
//   - service: The unit name that was queried
type Status struct {
	*types.BaseSkill
}

// Compile-time assertion that Status implements types.RunnableInterface.
var _ types.RunnableInterface = (*Status)(nil)

// Check always returns false since Status is read-only.
// Per the skill interface convention, the bool return indicates whether the
// operation would modify system state. Since status only queries unit
// information, this always returns false.
func (s *Status) Check() (bool, error) {
	return false, nil
}

// Run executes the status query and returns the result.
// Changed is always false since this is a read-only operation.
// Error is non-nil only if no service is configured; a failing systemctl
// status (e.g. unit not found) is reported in the output, not as an error,
// because `systemctl status` exits non-zero for inactive units.
func (s *Status) Run() types.Result {
	service := s.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := s.GetNodeConfig()
	cmdStr := fmt.Sprintf("systemctl status %s --no-pager -l", skills.ShellEscapeArg(service))
	cmd := types.Command{
		Command:     cmdStr,
		Description: "Show status of unit: " + service,
		// Required is false: systemctl status exits non-zero for inactive or
		// failed units, which is useful information rather than a hard failure.
		Required: false,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: false,
			Message: "Would show status of unit: " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("checking unit status", "service", service, "host", cfg.SSHHost)
	output, _ := ssh.Run(cfg, cmd)

	return types.Result{
		Changed: false,
		Message: "Status of unit " + service + ":\n" + output,
		Details: map[string]string{
			"output":  output,
			"service": service,
		},
	}
}

// SetArgs sets the arguments for status.
// Returns Status for fluent method chaining.
func (s *Status) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for status.
// Returns Status for fluent method chaining.
func (s *Status) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetService sets the systemd unit name and returns Status for chaining.
// Example: SetService("caddy")
func (s *Status) SetService(service string) *Status {
	s.BaseSkill.SetArg(ArgService, service)
	return s
}

// WithNodeConfig sets the node config and returns Status for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *Status) WithNodeConfig(cfg types.NodeConfig) *Status {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetID sets the ID for status.
// Returns Status for fluent method chaining.
func (s *Status) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for status.
// Returns Status for fluent method chaining.
func (s *Status) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for status.
// Returns Status for fluent method chaining.
func (s *Status) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewStatus creates a new systemctl-status skill.
//
// Returns a Status skill configured with IDSystemctlStatus identifier and
// description "Show systemd unit status (read-only)".
//
// Example:
//
//	node.Run(systemctl.NewStatus().SetService("caddy"))
func NewStatus() *Status {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlStatus)
	pb.SetDescription("Show systemd unit status (read-only)")
	return &Status{BaseSkill: pb}
}
