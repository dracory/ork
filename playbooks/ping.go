// Package playbooks provides reusable playbook implementations for common
// server automation tasks.
package playbooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// Ping checks SSH connectivity and shows basic server info.
// It runs 'uptime' to verify the connection works and display load.
type Ping struct{}

// Name returns the playbook identifier.
func (p *Ping) Name() string {
	return playbook.NamePing
}

// Description returns what this playbook does.
func (p *Ping) Description() string {
	return "Check SSH connectivity and show server uptime/load"
}

// Check always returns false for ping since it doesn't modify the system.
// It verifies connectivity by attempting to run a command.
func (p *Ping) Check(cfg config.Config) (bool, error) {
	// Ping never changes the system, so we always return false
	// The error indicates if the check itself failed (connection issue)
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
	if err != nil {
		return false, err
	}
	return false, nil
}

// Run executes the ping playbook and returns detailed result.
// Changed is always false since ping doesn't modify the system.
func (p *Ping) Run(cfg config.Config) playbook.Result {
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

// NewPing creates a new ping playbook instance.
func NewPing() *Ping {
	return &Ping{}
}
