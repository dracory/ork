package playbooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
)

// AptUpdate refreshes the package database.
type AptUpdate struct{}

// Name returns the playbook identifier.
func (a *AptUpdate) Name() string {
	return "apt-update"
}

// Description returns what this playbook does.
func (a *AptUpdate) Description() string {
	return "Refresh package database (apt-get update)"
}

// Run executes apt-get update.
func (a *AptUpdate) Run(cfg config.Config) error {
	log.Println("Running apt update...")

	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get update -y")
	if err != nil {
		return fmt.Errorf("apt update failed: %w\nOutput: %s", err, output)
	}

	log.Println("Apt update completed successfully")
	return nil
}

// NewAptUpdate creates a new apt-update playbook.
func NewAptUpdate() *AptUpdate {
	return &AptUpdate{}
}

// AptUpgrade installs available package updates.
type AptUpgrade struct{}

// Name returns the playbook identifier.
func (a *AptUpgrade) Name() string {
	return "apt-upgrade"
}

// Description returns what this playbook does.
func (a *AptUpgrade) Description() string {
	return "Install available package updates (apt-get upgrade)"
}

// Run executes apt-get upgrade.
func (a *AptUpgrade) Run(cfg config.Config) error {
	log.Println("Running apt upgrade...")

	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get upgrade -y")
	if err != nil {
		return fmt.Errorf("apt upgrade failed: %w\nOutput: %s", err, output)
	}

	log.Println("Apt upgrade completed successfully")
	return nil
}

// NewAptUpgrade creates a new apt-upgrade playbook.
func NewAptUpgrade() *AptUpgrade {
	return &AptUpgrade{}
}

// AptStatus shows available package updates without installing them.
type AptStatus struct{}

// Name returns the playbook identifier.
func (a *AptStatus) Name() string {
	return "apt-status"
}

// Description returns what this playbook does.
func (a *AptStatus) Description() string {
	return "Show available package updates (read-only)"
}

// Run checks for available updates.
func (a *AptStatus) Run(cfg config.Config) error {
	log.Println("Checking for available updates...")

	// First update package lists
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get update -qq")
	if err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Then list upgradable packages
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt list --upgradable 2>/dev/null | tail -n +2")
	if err != nil {
		return fmt.Errorf("failed to list upgradable packages: %w", err)
	}

	count := strings.TrimSpace(output)
	if count == "" || count == "0" {
		log.Println("All packages are up to date")
	} else {
		log.Printf("Available upgrades:\n%s", output)
	}

	return nil
}

// NewAptStatus creates a new apt-status playbook.
func NewAptStatus() *AptStatus {
	return &AptStatus{}
}
