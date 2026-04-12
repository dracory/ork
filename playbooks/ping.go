// Package playbooks provides reusable playbook implementations for common
// server automation tasks.
package playbooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// NewPing creates a new ping playbook instance.
func NewPing() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDPing)
	pb.SetDescription("Check SSH connectivity and show server uptime/load")
	return &ping{BasePlaybook: pb}
}

// Ping checks SSH connectivity and shows basic server info.
// It runs 'uptime' to verify the connection works and display load.
type ping struct {
	*playbook.BasePlaybook
}

// Check always returns false for ping since it doesn't modify the system.
// It verifies connectivity by attempting to run a command.
func (p *ping) Check() (bool, error) {
	// Ping never changes the system, so we always return false
	// The error indicates if the check itself failed (connection issue)
	cfg := p.GetConfig()
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
	if err != nil {
		return false, err
	}
	return false, nil
}

// Run executes the ping playbook and returns detailed result.
// Changed is always false since ping doesn't modify the system.
func (p *ping) Run() playbook.Result {
	cfg := p.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to ping %s", cfg.SSHHost),
			Error:   fmt.Errorf("failed to ping %s: %w", cfg.SSHHost, err),
		}
	}

	log.Printf("%s is alive: %s", cfg.SSHHost, strings.TrimSpace(output))

	return playbook.Result{
		Changed: false, // Ping never changes the system
		Message: fmt.Sprintf("%s is alive", cfg.SSHHost),
		Details: map[string]string{
			"uptime": strings.TrimSpace(output),
		},
	}
}
