package ufw

import (
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Enable enables the UFW firewall.
// This skill activates UFW to start enforcing firewall rules.
//
// Usage:
//
//	go run . --playbook=ufw-enable
//
// Execution Flow:
//  1. Executes `ufw --force enable`
//  2. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed
//   - Root SSH access required
//   - Ensure SSH port is allowed before enabling
//
// Related Playbooks:
//   - ufw-disable: Disable UFW firewall
//   - ufw-status: Verify UFW status
//   - ufw-install: Install and enable UFW
type Enable struct {
	*types.BaseSkill
}

// Check determines if UFW needs to be enabled.
func (e *Enable) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and enables UFW.
func (e *Enable) Run() types.Result {
	cfg := e.GetNodeConfig()

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: "Would enable UFW firewall",
		}
	}

	// Enable UFW
	cmd := types.Command{
		Command:     "ufw --force enable",
		Description: "Enable UFW firewall",
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable UFW",
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: "Enabled UFW firewall",
		Details: map[string]string{
			"output": output,
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

// NewEnable creates a new ufw-enable skill.
func NewEnable() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwEnable)
	pb.SetDescription("Enable UFW firewall")
	return &Enable{BaseSkill: pb}
}
