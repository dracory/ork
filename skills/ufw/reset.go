package ufw

import (
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Reset resets UFW to factory defaults.
// This skill removes all rules and resets UFW to its initial state.
//
// Usage:
//
//	go run . --playbook=ufw-reset
//
// Execution Flow:
//  1. Executes `ufw --force reset`
//  2. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed
//   - Root SSH access required
//
// Warning:
//   - This removes ALL firewall rules
//   - Server becomes unprotected until new rules are added
//   - SSH access may be lost if port 22 isn't re-allowed
//
// Related Playbooks:
//   - ufw-install: Install and configure UFW
//   - ufw-status: Verify UFW status
//   - ufw-allow: Add rules after reset
type Reset struct {
	*types.BaseSkill
}

// Check determines if UFW needs to be reset.
func (r *Reset) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and resets UFW.
func (r *Reset) Run() types.Result {
	cfg := r.GetNodeConfig()

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: "Would reset UFW to factory defaults",
		}
	}

	// Reset UFW
	cmd := types.Command{
		Command:     "ufw --force reset",
		Description: "Reset UFW to factory defaults",
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reset UFW",
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: "Reset UFW to factory defaults",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for reset.
// Returns Reset for fluent method chaining.
func (r *Reset) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for reset.
// Returns Reset for fluent method chaining.
func (r *Reset) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetID sets the ID for reset.
// Returns Reset for fluent method chaining.
func (r *Reset) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for reset.
// Returns Reset for fluent method chaining.
func (r *Reset) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for reset.
// Returns Reset for fluent method chaining.
func (r *Reset) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewReset creates a new ufw-reset skill.
func NewReset() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwReset)
	pb.SetDescription("Reset UFW to factory defaults")
	return &Reset{BaseSkill: pb}
}
