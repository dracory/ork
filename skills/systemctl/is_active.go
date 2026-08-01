package systemctl

// Package systemctl documentation is in status.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// IsActive checks whether a systemd unit is currently active by running
// `systemctl is-active <service>`. This is a read-only skill: it never
// modifies system state, so Changed is always false and Check always
// returns false.
//
// Unlike Status, IsActive is designed for programmatic checks: it captures
// the single-word state ("active", "inactive", "failed", etc.) in
// Result.Details["state"] and does not treat a non-active unit as an error.
//
// Usage:
//
//	result := node.Run(systemctl.NewIsActive().SetService("caddy"))
//	state := result.Details["state"] // "active", "inactive", "failed", ...
//
// Result Details:
//   - state: The unit's active state word from systemctl is-active
//   - service: The unit name that was queried
type IsActive struct {
	*types.BaseSkill
}

// Compile-time assertion that IsActive implements types.RunnableInterface.
var _ types.RunnableInterface = (*IsActive)(nil)

// Check always returns false since IsActive is read-only.
func (s *IsActive) Check() (bool, error) {
	return false, nil
}

// Run executes the is-active query and returns the result.
// Changed is always false since this is a read-only operation.
// `systemctl is-active` exits non-zero for inactive/failed units, so the
// error from ssh.Run is intentionally ignored — the state word is captured
// from stdout regardless of exit code.
func (s *IsActive) Run() types.Result {
	service := s.GetArg(ArgService)
	if service == "" {
		return types.Result{
			Changed: false,
			Message: "No service specified",
			Error:   fmt.Errorf("no service specified: set the %q argument", ArgService),
		}
	}

	cfg := s.GetNodeConfig()
	cmdStr := fmt.Sprintf("systemctl is-active %s", skills.ShellEscapeArg(service))
	cmd := types.Command{
		Command:     cmdStr,
		Description: "Check if unit is active: " + service,
		// Required is false: is-active exits non-zero for inactive units,
		// which is a valid result, not a hard failure.
		Required: false,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: false,
			Message: "Would check if unit is active: " + service,
		}
	}

	cfg.GetLoggerOrDefault().Info("checking if unit is active", "service", service, "host", cfg.SSHHost)
	output, err := ssh.Run(cfg, cmd)
	if err != nil && !ssh.IsExitError(err) {
		return types.Result{
			Changed: false,
			Message: "Failed to check if unit is active: " + service,
			Error:   fmt.Errorf("failed to check if unit %s is active: %w", service, err),
		}
	}
	state := strings.TrimSpace(output)

	return types.Result{
		Changed: false,
		Message: "Unit " + service + " is " + state,
		Details: map[string]string{
			"state":   state,
			"service": service,
		},
	}
}

// SetArgs sets the arguments for is-active.
// Returns IsActive for fluent method chaining.
func (s *IsActive) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for is-active.
// Returns IsActive for fluent method chaining.
func (s *IsActive) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetService sets the systemd unit name and returns IsActive for chaining.
func (s *IsActive) SetService(service string) *IsActive {
	s.BaseSkill.SetArg(ArgService, service)
	return s
}

// WithNodeConfig sets the node config and returns IsActive for chaining.
func (s *IsActive) WithNodeConfig(cfg types.NodeConfig) *IsActive {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetID sets the ID for is-active.
// Returns IsActive for fluent method chaining.
func (s *IsActive) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for is-active.
// Returns IsActive for fluent method chaining.
func (s *IsActive) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for is-active.
// Returns IsActive for fluent method chaining.
func (s *IsActive) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewIsActive creates a new systemctl-is-active skill.
//
// Returns an IsActive skill configured with IDSystemctlIsActive identifier
// and description "Check if a systemd unit is active (read-only)".
func NewIsActive() *IsActive {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSystemctlIsActive)
	pb.SetDescription("Check if a systemd unit is active (read-only)")
	return &IsActive{BaseSkill: pb}
}
