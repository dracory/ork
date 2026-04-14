// Package ping provides a playbook for testing SSH connectivity to remote servers.
// It is the simplest health check playbook, used to verify SSH configuration
// and server accessibility before running more complex operations.
package ping

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// PlaybookPing tests SSH connectivity to the remote server.
// This is the simplest health check playbook that verifies the SSH
// connection is working before running more complex operations.
//
// Usage:
//
//	go run . --playbook=ping
//
// Execution Flow:
//  1. Establishes SSH connection using configured SSH key
//  2. Runs the 'uptime' command on the remote server
//  3. Reports success or failure
//
// Expected Output:
//   - Success: "SSH connection successful" message from remote server
//   - Failure: Fatal error with connection details
//
// Use Cases:
//   - Verify SSH configuration is correct
//   - Test server accessibility
//   - Initial connectivity validation before running other playbooks
//
// Prerequisites:
//   - SSH key must be accessible at ~/.ssh/ with correct permissions
//   - Root user must have SSH key authentication enabled
type Ping struct {
	*playbook.BasePlaybook
}

// Check verifies SSH connectivity to the remote server.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since ping is read-only,
// this always returns false. The error indicates connection failures.
func (p *Ping) Check() (bool, error) {
	// Ping never changes the system, so we always return false
	// The error indicates if the check itself failed (connection issue)
	cfg := p.GetNodeConfig()
	cmdCheck := types.Command{Command: "uptime", Description: "Check server uptime"}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		return false, err
	}
	return false, nil
}

// Run executes the ping playbook and returns detailed result.
// Changed is always false since ping doesn't modify the system.
// On success, Result.Details contains an 'uptime' key with the server's
// uptime/load string from the remote command execution.
func (p *Ping) Run() playbook.Result {
	cfg := p.GetNodeConfig()
	cmdUptime := types.Command{Command: "uptime", Description: "Check server uptime"}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd: ", cmdUptime.Command)
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("Would ping: %s", cfg.SSHHost),
		}
	}

	cfg.GetLoggerOrDefault().Info("running command", "cmd", cmdUptime.Command, "description", cmdUptime.Description)
	output, err := ssh.Run(cfg, cmdUptime)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to ping %s", cfg.SSHHost),
			Error:   fmt.Errorf("failed to ping %s: %w", cfg.SSHHost, err),
		}
	}

	cfg.GetLoggerOrDefault().Info("host is alive", "host", cfg.SSHHost, "uptime", strings.TrimSpace(output))

	return playbook.Result{
		Changed: false, // Ping never changes the system
		Message: fmt.Sprintf("%s is alive", cfg.SSHHost),
		Details: map[string]string{
			"uptime": strings.TrimSpace(output),
		},
	}
}

// NewPing creates a new ping playbook instance.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDPing identifier
//	and description "Check SSH connectivity and show server uptime/load".
//
// The returned playbook can be registered with the playbook registry
// and executed via the CLI or programmatically.
func NewPing() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDPing)
	pb.SetDescription("Check SSH connectivity and show server uptime/load")
	return &Ping{BasePlaybook: pb}
}
