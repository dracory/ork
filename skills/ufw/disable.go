package ufw

import (
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Disable disables the UFW firewall.
// This skill deactivates UFW and stops enforcing firewall rules.
//
// Usage:
//
//	go run . --playbook=ufw-disable
//
// Execution Flow:
//  1. Executes `ufw disable`
//  2. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed
//   - Root SSH access required
//
// Warning:
//   - Disabling UFW removes all firewall protection
//   - Server becomes vulnerable to attacks
//
// Related Playbooks:
//   - ufw-enable: Enable UFW firewall
//   - ufw-status: Verify UFW status
type Disable struct {
	*types.BaseSkill
}

// Check determines if UFW needs to be disabled.
func (d *Disable) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and disables UFW.
func (d *Disable) Run() types.Result {
	cfg := d.GetNodeConfig()

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: "Would disable UFW firewall",
		}
	}

	// Disable UFW
	cmd := types.Command{
		Command:     "ufw disable",
		Description: "Disable UFW firewall",
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to disable UFW",
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: "Disabled UFW firewall",
		Details: map[string]string{
			"output": output,
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

// NewDisable creates a new ufw-disable skill.
func NewDisable() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwDisable)
	pb.SetDescription("Disable UFW firewall")
	return &Disable{BaseSkill: pb}
}
