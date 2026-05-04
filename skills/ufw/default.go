package ufw

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Default sets UFW default policies for incoming and outgoing traffic.
// This skill configures the default action for traffic not matching specific rules.
//
// Usage:
//
//	go run . --playbook=ufw-default --arg=incoming=<deny|allow|reject> --arg=outgoing=<deny|allow|reject>
//
// Args:
//   - incoming: Policy for incoming traffic - "deny", "allow", or "reject" (default: "deny")
//   - outgoing: Policy for outgoing traffic - "deny", "allow", or "reject" (default: "allow")
//
// Execution Flow:
//   1. Validates policy parameters
//   2. Executes `ufw default <incoming> incoming` and `ufw default <outgoing> outgoing`
//   3. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed
//   - Root SSH access required
//
// Warning:
//   - Changing default policies can lock you out
//   - Ensure SSH port is explicitly allowed before setting incoming to "deny"
//
// Related Playbooks:
//   - ufw-install: Install and configure UFW
//   - ufw-status: Verify UFW status
type Default struct {
	*types.BaseSkill
}

// Check determines if default policies need to be set.
func (d *Default) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and sets default policies.
func (d *Default) Run() types.Result {
	cfg := d.GetNodeConfig()
	incoming := cfg.GetArgOr(ArgIncoming, "deny")
	outgoing := cfg.GetArgOr(ArgOutgoing, "allow")

	// Validate policies
	validPolicies := map[string]bool{"deny": true, "allow": true, "reject": true}
	if !validPolicies[incoming] {
		return types.Result{
			Changed: false,
			Message: "Invalid incoming policy",
			Error:   fmt.Errorf("incoming must be 'deny', 'allow', or 'reject'"),
		}
	}
	if !validPolicies[outgoing] {
		return types.Result{
			Changed: false,
			Message: "Invalid outgoing policy",
			Error:   fmt.Errorf("outgoing must be 'deny', 'allow', or 'reject'"),
		}
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would set UFW defaults: incoming=%s, outgoing=%s", incoming, outgoing),
		}
	}

	// Set default policies
	cmdIncoming := types.Command{
		Command:     fmt.Sprintf("ufw default %s incoming", incoming),
		Description: fmt.Sprintf("Set default incoming policy to %s", incoming),
		Required:    true,
	}
	cmdOutgoing := types.Command{
		Command:     fmt.Sprintf("ufw default %s outgoing", outgoing),
		Description: fmt.Sprintf("Set default outgoing policy to %s", outgoing),
		Required:    true,
	}

	_, err := ssh.Run(cfg, cmdIncoming)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to set default incoming policy to %s", incoming),
			Error:   err,
		}
	}

	_, err = ssh.Run(cfg, cmdOutgoing)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to set default outgoing policy to %s", outgoing),
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Set UFW defaults: incoming=%s, outgoing=%s", incoming, outgoing),
	}
}

// SetArgs sets the arguments for default policies.
// Returns Default for fluent method chaining.
func (d *Default) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument for default policies.
// Returns Default for fluent method chaining.
func (d *Default) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID for default policies.
// Returns Default for fluent method chaining.
func (d *Default) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for default policies.
// Returns Default for fluent method chaining.
func (d *Default) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for default policies.
// Returns Default for fluent method chaining.
func (d *Default) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDefault creates a new ufw-default skill.
func NewDefault() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwDefault)
	pb.SetDescription("Set UFW default policies")
	return &Default{BaseSkill: pb}
}
