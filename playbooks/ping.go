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

// Run executes the ping playbook.
func (p *Ping) Run(cfg config.Config) error {
	log.Printf("Pinging %s...", cfg.SSHHost)

	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
	if err != nil {
		return fmt.Errorf("failed to ping %s: %w", cfg.SSHHost, err)
	}

	log.Printf("%s is alive: %s", cfg.SSHHost, strings.TrimSpace(output))
	return nil
}

// NewPing creates a new ping playbook instance.
func NewPing() *Ping {
	return &Ping{}
}
