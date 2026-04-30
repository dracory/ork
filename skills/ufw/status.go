package ufw

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UfwStatus displays the current UFW firewall configuration and state.
// This read-only skill checks whether UFW is installed, enabled, and
// displays all active rules. Use this to verify firewall configuration.
//
// Usage:
//
//	go run . --playbook=ufw-status
//
// Execution Flow:
//  1. Runs 'ufw status verbose' to get detailed status
//  2. Displays firewall state (active/inactive)
//  3. Shows default policies
//  4. Lists all configured rules with numbers
//
// Output Information:
//   - Status: active or inactive
//   - Default Policy: incoming/outgoing/routed
//   - Rules: numbered list with action, direction, and target
//   - Logging: current logging level
//
// Understanding the Output:
//   - Status active: Firewall is enforcing rules
//   - To Action: From -> Destination direction
//   - Anywhere: Applies to all IP addresses
//   - Numbers: Use with 'ufw delete <number>' to remove rules
//
// Prerequisites:
//   - UFW must be installed (use ufw-install playbook)
//   - Root SSH access required
//
// Related Playbooks:
//   - ufw-install: Install UFW firewall
//   - ufw-allow: Allow additional ports
type UfwStatus struct {
	*types.BaseSkill
}

// Check always returns false since this is a read-only skill.
func (u *UfwStatus) Check() (bool, error) {
	return false, nil
}

// Run executes the skill and returns detailed result.
func (u *UfwStatus) Run() types.Result {
	cfg := u.GetNodeConfig()
	cmdStatus := types.Command{Command: "ufw status verbose", Description: "Check UFW status"}

	// Check for dry-run mode - display actual command
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStatus.Command)
		return types.Result{
			Changed: false,
			Message: "Would check UFW firewall status",
		}
	}

	cfg.GetLoggerOrDefault().Info("checking UFW status")

	output, err := ssh.Run(cfg, cmdStatus)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check UFW status",
			Error:   fmt.Errorf("failed to check UFW status: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("UFW status retrieved", "output", output)
	return types.Result{
		Changed: false,
		Message: "UFW status retrieved",
		Details: map[string]string{
			"status": output,
		},
	}
}

// SetArgs sets the arguments for UFW status.
// Returns UfwStatus for fluent method chaining.
func (u *UfwStatus) SetArgs(args map[string]string) types.RunnableInterface {
	u.BaseSkill.SetArgs(args)
	return u
}

// SetArg sets a single argument for UFW status.
// Returns UfwStatus for fluent method chaining.
func (u *UfwStatus) SetArg(key, value string) types.RunnableInterface {
	u.BaseSkill.SetArg(key, value)
	return u
}

// SetID sets the ID for UFW status.
// Returns UfwStatus for fluent method chaining.
func (u *UfwStatus) SetID(id string) types.RunnableInterface {
	u.BaseSkill.SetID(id)
	return u
}

// SetDescription sets the description for UFW status.
// Returns UfwStatus for fluent method chaining.
func (u *UfwStatus) SetDescription(description string) types.RunnableInterface {
	u.BaseSkill.SetDescription(description)
	return u
}

// SetTimeout sets the timeout for UFW status.
// Returns UfwStatus for fluent method chaining.
func (u *UfwStatus) SetTimeout(timeout time.Duration) types.RunnableInterface {
	u.BaseSkill.SetTimeout(timeout)
	return u
}

// NewUfwStatus creates a new ufw-status skill.
func NewUfwStatus() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwStatus)
	pb.SetDescription("Display UFW firewall configuration and status (read-only)")
	return &UfwStatus{BaseSkill: pb}
}
