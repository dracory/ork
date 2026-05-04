package ufw

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Deny configures UFW firewall rules to deny traffic on a port.
// This skill blocks traffic on specified TCP/UDP ports with an optional comment.
//
// Usage:
//
//	go run . --playbook=ufw-deny --arg=port=<port> [--arg=protocol=<tcp|udp>] [--arg=comment=<comment>]
//
// Args:
//   - port: Port number to deny (required)
//   - protocol: Protocol to use - "tcp" or "udp" (default: "tcp")
//   - comment: Optional comment for the rule
//
// Execution Flow:
//   1. Validates port parameter
//   2. Executes `ufw deny <port>/<protocol> [comment]`
//   3. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed and enabled
//   - Root SSH access required
//
// Related Playbooks:
//   - ufw-install: Install UFW firewall
//   - ufw-status: Verify UFW status
//   - ufw-allow: Allow ports (inverse of deny)
type Deny struct {
	*types.BaseSkill
}

// Check determines if the rule needs to be added.
func (d *Deny) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and adds the UFW deny rule.
func (d *Deny) Run() types.Result {
	cfg := d.GetNodeConfig()
	port := cfg.GetArgOr(ArgPort, "")
	protocol := cfg.GetArgOr(ArgProtocol, "tcp")
	comment := cfg.GetArgOr(ArgComment, "")

	if port == "" {
		return types.Result{
			Changed: false,
			Message: "Port parameter is required",
			Error:   fmt.Errorf("use --arg=port=<port_number>"),
		}
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would deny port %s/%s in UFW", port, protocol),
		}
	}

	// Build the command
	cmdStr := fmt.Sprintf("ufw deny %s/%s", port, protocol)
	if comment != "" {
		cmdStr += fmt.Sprintf(" comment '%s'", comment)
	}

	// Add the deny rule
	cmd := types.Command{
		Command:     cmdStr,
		Description: fmt.Sprintf("Deny port %s/%s in UFW", port, protocol),
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to deny port %s/%s in UFW", port, protocol),
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Denied port %s/%s in UFW", port, protocol),
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for the deny rule.
// Returns Deny for fluent method chaining.
func (d *Deny) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument for the deny rule.
// Returns Deny for fluent method chaining.
func (d *Deny) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID for the deny rule.
// Returns Deny for fluent method chaining.
func (d *Deny) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for the deny rule.
// Returns Deny for fluent method chaining.
func (d *Deny) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for the deny rule.
// Returns Deny for fluent method chaining.
func (d *Deny) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDeny creates a new ufw-deny skill.
func NewDeny() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwDeny)
	pb.SetDescription("Deny port in UFW firewall")
	return &Deny{BaseSkill: pb}
}
