package playbooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// AptUpdate refreshes the package database.
type AptUpdate struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (a *AptUpdate) GetID() string {
	return playbook.IDAptUpdate
}

// SetID sets the playbook identifier.
func (a *AptUpdate) SetID(id string) playbook.PlaybookInterface {
	return a
}

// GetDescription returns what this playbook does.
func (a *AptUpdate) GetDescription() string {
	return "Refresh package database (apt-get update)"
}

// SetDescription sets the playbook description.
func (a *AptUpdate) SetDescription(description string) playbook.PlaybookInterface {
	return a
}

// GetConfig returns the current node configuration.
func (a *AptUpdate) GetConfig() config.Config {
	return a.cfg
}

// GetOptions returns the current playbook options.
func (a *AptUpdate) GetOptions() *playbook.PlaybookOptions {
	return a.opts
}

// SetConfig sets the node configuration for this playbook.
func (a *AptUpdate) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	a.cfg = cfg
	return a
}

// SetOptions sets the playbook-specific options.
func (a *AptUpdate) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	a.opts = opts
	return a
}

// Check always returns true for apt-update since cache refresh is always beneficial.
// The cost of checking if update is needed is similar to just running it.
func (a *AptUpdate) Check() (bool, error) {
	return true, nil // Always run apt update
}

// Run executes apt-get update and returns the result.
func (a *AptUpdate) Run() playbook.Result {
	log.Println("Running apt update...")

	output, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey, "apt-get update -y")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Apt update failed",
			Error:   fmt.Errorf("apt update failed: %w\nOutput: %s", err, output),
		}
	}

	log.Println("Apt update completed successfully")
	return playbook.Result{
		Changed: true, // Cache was refreshed
		Message: "Package database updated",
		Details: map[string]string{
			"output": output,
		},
	}
}

// NewAptUpdate creates a new apt-update playbook.
func NewAptUpdate() *AptUpdate {
	return &AptUpdate{}
}

// AptUpgrade installs available package updates.
type AptUpgrade struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (a *AptUpgrade) GetID() string {
	return playbook.IDAptUpgrade
}

// SetID sets the playbook identifier.
func (a *AptUpgrade) SetID(id string) playbook.PlaybookInterface {
	return a
}

// GetDescription returns what this playbook does.
func (a *AptUpgrade) GetDescription() string {
	return "Install available package updates (apt-get upgrade)"
}

// SetDescription sets the playbook description.
func (a *AptUpgrade) SetDescription(description string) playbook.PlaybookInterface {
	return a
}

// GetConfig returns the current node configuration.
func (a *AptUpgrade) GetConfig() config.Config {
	return a.cfg
}

// GetOptions returns the current playbook options.
func (a *AptUpgrade) GetOptions() *playbook.PlaybookOptions {
	return a.opts
}

// SetConfig sets the node configuration for this playbook.
func (a *AptUpgrade) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	a.cfg = cfg
	return a
}

// SetOptions sets the playbook-specific options.
func (a *AptUpgrade) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	a.opts = opts
	return a
}

// Check determines if there are packages that need upgrading.
// Returns true if upgrades are available, false if system is up to date.
func (a *AptUpgrade) Check() (bool, error) {
	// First ensure package lists are updated
	_, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey, "apt-get update -qq")
	if err != nil {
		return false, fmt.Errorf("failed to update package lists: %w", err)
	}

	// Check for upgradable packages
	output, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey,
		"apt list --upgradable 2>/dev/null | grep -c '\\[upgradable from:' || echo 0")
	if err != nil {
		return false, fmt.Errorf("failed to check for upgrades: %w", err)
	}

	count := strings.TrimSpace(output)
	return count != "0" && count != "", nil
}

// Run executes apt-get upgrade and returns detailed result.
func (a *AptUpgrade) Run() playbook.Result {
	// Check if upgrades are needed
	needsUpgrade, err := a.Check()
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to check for upgrades",
			Error:   err,
		}
	}

	if !needsUpgrade {
		return playbook.Result{
			Changed: false,
			Message: "All packages are up to date",
		}
	}

	log.Println("Running apt upgrade...")

	output, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey, "apt-get upgrade -y")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Apt upgrade failed",
			Error:   fmt.Errorf("apt upgrade failed: %w\nOutput: %s", err, output),
		}
	}

	log.Println("Apt upgrade completed successfully")
	return playbook.Result{
		Changed: true,
		Message: "Packages upgraded successfully",
		Details: map[string]string{
			"output": output,
		},
	}
}

// NewAptUpgrade creates a new apt-upgrade playbook.
func NewAptUpgrade() *AptUpgrade {
	return &AptUpgrade{}
}

// AptStatus shows available package updates without installing them.
type AptStatus struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (a *AptStatus) GetID() string {
	return playbook.IDAptStatus
}

// SetID sets the playbook identifier.
func (a *AptStatus) SetID(id string) playbook.PlaybookInterface {
	return a
}

// GetDescription returns what this playbook does.
func (a *AptStatus) GetDescription() string {
	return "Show available package updates (read-only)"
}

// SetDescription sets the playbook description.
func (a *AptStatus) SetDescription(description string) playbook.PlaybookInterface {
	return a
}

// GetConfig returns the current node configuration.
func (a *AptStatus) GetConfig() config.Config {
	return a.cfg
}

// GetOptions returns the current playbook options.
func (a *AptStatus) GetOptions() *playbook.PlaybookOptions {
	return a.opts
}

// SetConfig sets the node configuration for this playbook.
func (a *AptStatus) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	a.cfg = cfg
	return a
}

// SetOptions sets the playbook-specific options.
func (a *AptStatus) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	a.opts = opts
	return a
}

// Check always returns false since AptStatus is read-only.
func (a *AptStatus) Check() (bool, error) {
	return false, nil
}

// Run executes apt status check and returns detailed result.
func (a *AptStatus) Run() playbook.Result {
	log.Println("Checking for available updates...")

	// First update package lists
	_, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey, "apt-get update -qq")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to update package lists",
			Error:   fmt.Errorf("failed to update package lists: %w", err),
		}
	}

	// Then list upgradable packages
	output, err := ssh.RunOnce(a.cfg.SSHHost, a.cfg.SSHPort, a.cfg.RootUser, a.cfg.SSHKey, "apt list --upgradable 2>/dev/null | tail -n +2")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list upgradable packages",
			Error:   fmt.Errorf("failed to list upgradable packages: %w", err),
		}
	}

	count := strings.TrimSpace(output)
	if count == "" || count == "0" {
		log.Println("All packages are up to date")
		return playbook.Result{
			Changed: false,
			Message: "All packages are up to date",
			Details: map[string]string{
				"upgradable_count": "0",
			},
		}
	}

	log.Printf("Available upgrades:\n%s", output)
	return playbook.Result{
		Changed: false, // Read-only operation
		Message: fmt.Sprintf("%d packages available for upgrade", strings.Count(output, "\n")+1),
		Details: map[string]string{
			"upgradable_count": fmt.Sprintf("%d", strings.Count(output, "\n")+1),
			"packages":         output,
		},
	}
}

// NewAptStatus creates a new apt-status playbook.
func NewAptStatus() *AptStatus {
	return &AptStatus{}
}
