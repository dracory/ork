package ufw

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Delete removes UFW firewall rules by rule number.
// This skill deletes rules using their number from 'ufw status numbered'.
//
// Usage:
//
//	go run . --playbook=ufw-delete --arg=number=<rule_number>
//
// Args:
//   - number: Rule number to delete (required)
//
// Execution Flow:
//   1. Validates rule number parameter
//   2. Executes `ufw delete <number>`
//   3. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed and enabled
//   - Root SSH access required
//   - Use 'ufw-status' to get rule numbers
//
// Related Playbooks:
//   - ufw-status: Get rule numbers
//   - ufw-install: Install UFW firewall
type Delete struct {
	*types.BaseSkill
}

// Check determines if the rule needs to be deleted.
func (d *Delete) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and deletes the UFW rule.
func (d *Delete) Run() types.Result {
	cfg := d.GetNodeConfig()
	number := cfg.GetArgOr(ArgNumber, "")

	if number == "" {
		return types.Result{
			Changed: false,
			Message: "Rule number parameter is required",
			Error:   fmt.Errorf("use --arg=number=<rule_number>"),
		}
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would delete UFW rule number %s", number),
		}
	}

	// Delete the rule
	cmd := types.Command{
		Command:     fmt.Sprintf("ufw --force delete %s", number),
		Description: fmt.Sprintf("Delete UFW rule number %s", number),
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to delete UFW rule number %s", number),
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Deleted UFW rule number %s", number),
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for the delete operation.
// Returns Delete for fluent method chaining.
func (d *Delete) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument for the delete operation.
// Returns Delete for fluent method chaining.
func (d *Delete) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID for the delete operation.
// Returns Delete for fluent method chaining.
func (d *Delete) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for the delete operation.
// Returns Delete for fluent method chaining.
func (d *Delete) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for the delete operation.
// Returns Delete for fluent method chaining.
func (d *Delete) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDelete creates a new ufw-delete skill.
func NewDelete() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwDelete)
	pb.SetDescription("Delete UFW firewall rule by number")
	return &Delete{BaseSkill: pb}
}
