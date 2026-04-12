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
type Ping struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (p *Ping) GetID() string {
	return playbook.IDPing
}

// SetID sets the playbook identifier.
func (p *Ping) SetID(id string) playbook.Playbook {
	return p
}

// GetDescription returns what this playbook does.
func (p *Ping) GetDescription() string {
	return "Check SSH connectivity and show server uptime/load"
}

// SetDescription sets the playbook description.
func (p *Ping) SetDescription(description string) playbook.Playbook {
	return p
}

// GetConfig returns the current node configuration.
func (p *Ping) GetConfig() config.Config {
	return p.cfg
}

// GetOptions returns the current playbook options.
func (p *Ping) GetOptions() *playbook.PlaybookOptions {
	return p.opts
}

// SetConfig sets the node configuration for this playbook.
func (p *Ping) SetConfig(cfg config.Config) playbook.Playbook {
	p.cfg = cfg
	return p
}

// SetOptions sets the playbook-specific options.
func (p *Ping) SetOptions(opts *playbook.PlaybookOptions) playbook.Playbook {
	p.opts = opts
	return p
}

// Check always returns false for ping since it doesn't modify the system.
// It verifies connectivity by attempting to run a command.
func (p *Ping) Check() (bool, error) {
	// Ping never changes the system, so we always return false
	// The error indicates if the check itself failed (connection issue)
	_, err := ssh.RunOnce(p.cfg.SSHHost, p.cfg.SSHPort, p.cfg.RootUser, p.cfg.SSHKey, "uptime")
	if err != nil {
		return false, err
	}
	return false, nil
}

// Run executes the ping playbook and returns detailed result.
// Changed is always false since ping doesn't modify the system.
func (p *Ping) Run() playbook.Result {
	output, err := ssh.RunOnce(p.cfg.SSHHost, p.cfg.SSHPort, p.cfg.RootUser, p.cfg.SSHKey, "uptime")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to ping %s", p.cfg.SSHHost),
			Error:   fmt.Errorf("failed to ping %s: %w", p.cfg.SSHHost, err),
		}
	}

	log.Printf("%s is alive: %s", p.cfg.SSHHost, strings.TrimSpace(output))

	return playbook.Result{
		Changed: false, // Ping never changes the system
		Message: fmt.Sprintf("%s is alive", p.cfg.SSHHost),
		Details: map[string]string{
			"uptime": strings.TrimSpace(output),
		},
	}
}

// NewPing creates a new ping playbook instance.
func NewPing() *Ping {
	return &Ping{}
}
