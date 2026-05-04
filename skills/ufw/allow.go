package ufw

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Allow configures UFW firewall rules to allow traffic on a port.
// This skill allows opening any TCP/UDP port with an optional comment.
//
// Usage:
//
//	go run . --playbook=ufw-allow --arg=port=<port> [--arg=protocol=<tcp|udp>] [--arg=comment=<comment>]
//
// Args:
//   - port: Port number to allow (required)
//   - protocol: Protocol to use - "tcp" or "udp" (default: "tcp")
//   - comment: Optional comment for the rule
//
// Execution Flow:
//  1. Validates port parameter
//  2. Executes `ufw allow <port>/<protocol> [comment]`
//  3. Returns success/failure result
//
// Prerequisites:
//   - UFW must be installed and enabled
//   - Root SSH access required
//
// Related Playbooks:
//   - ufw-install: Install UFW firewall
//   - ufw-status: Verify UFW status
//   - ufw-deny: Deny ports (inverse of allow)
type Allow struct {
	*types.BaseSkill
}

// Check determines if the rule needs to be added.
func (a *Allow) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and adds the UFW rule.
func (a *Allow) Run() types.Result {
	cfg := a.GetNodeConfig()
	port := a.GetArg(ArgPort)
	if port == "" {
		port = cfg.GetArgOr(ArgPort, "")
	}
	protocol := a.GetArg(ArgProtocol)
	if protocol == "" {
		protocol = cfg.GetArgOr(ArgProtocol, "tcp")
	}
	comment := a.GetArg(ArgComment)
	if comment == "" {
		comment = cfg.GetArgOr(ArgComment, "")
	}

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
			Message: fmt.Sprintf("Would allow port %s/%s in UFW", port, protocol),
		}
	}

	// Build the command
	cmdStr := fmt.Sprintf("ufw allow %s/%s", port, protocol)
	if comment != "" {
		cmdStr += fmt.Sprintf(" comment '%s'", comment)
	}

	// Add the port rule
	cmd := types.Command{
		Command:     cmdStr,
		Description: fmt.Sprintf("Allow port %s/%s in UFW", port, protocol),
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to allow port %s/%s in UFW", port, protocol),
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Allowed port %s/%s in UFW", port, protocol),
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for the allow rule.
// Returns Allow for fluent method chaining.
func (a *Allow) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// SetArg sets a single argument for the allow rule.
// Returns Allow for fluent method chaining.
func (a *Allow) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for the allow rule.
// Returns Allow for fluent method chaining.
func (a *Allow) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for the allow rule.
// Returns Allow for fluent method chaining.
func (a *Allow) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for the allow rule.
// Returns Allow for fluent method chaining.
func (a *Allow) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// NewAllow creates a new ufw-allow skill.
func NewAllow() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwAllow)
	pb.SetDescription("Allow port in UFW firewall")
	return &Allow{BaseSkill: pb}
}
