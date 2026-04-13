package apt

// Package apt documentation is in status.go

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// AptUpgrade installs available package updates.
// This playbook runs apt-get upgrade to install all available package updates.
// It first checks if updates are available by querying the package database,
// then installs them only if needed.
//
// Usage:
//
//	go run . --playbook=apt-upgrade
//
// Execution Flow:
//  1. Runs apt-get update to refresh package lists
//  2. Checks for available upgrades with apt list --upgradable
//  3. If packages need upgrading, runs apt-get upgrade -y
//  4. Reports success with details of what was upgraded
//
// Expected Output:
//   - Success (packages upgraded): "Packages upgraded successfully" with output details
//   - Success (no upgrades): "All packages are up to date"
//   - Failure: Error with apt command output details
//
// Result Details:
//   - output: Full output from apt-get upgrade command (when upgrades occur)
//
// Use Cases:
//   - Apply security updates to production servers
//   - Regular maintenance and patch management
//   - Pre-deployment system updates
//
// Idempotency:
//   - Reports Changed=false when no packages need upgrading
//   - Reports Changed=true when packages are actually upgraded
type AptUpgrade struct {
	*playbook.BasePlaybook
}

// Check determines if there are packages that need upgrading.
// Per the playbook interface convention, returns true if upgrades are available
// (meaning Run would modify the system), false if system is already up to date.
//
// This method first runs apt-get update to ensure package lists are current,
// then counts upgradable packages using apt list --upgradable.
func (a *AptUpgrade) Check() (bool, error) {
	cfg := a.GetConfig()
	// First ensure package lists are updated
	_, err := ssh.Run(cfg, "apt-get update -qq")
	if err != nil {
		return false, fmt.Errorf("failed to update package lists: %w", err)
	}

	// Check for upgradable packages
	output, err := ssh.Run(cfg, "apt list --upgradable 2>/dev/null | grep -c '\\[upgradable from:' || echo 0")
	if err != nil {
		return false, fmt.Errorf("failed to check for upgrades: %w", err)
	}

	count := strings.TrimSpace(output)
	return count != "0" && count != "", nil
}

// Run executes apt-get upgrade and returns detailed result.
// Changed is true when packages are actually upgraded, false when system is up to date.
//
// Result.Details contains:
//   - output: Full output from apt-get upgrade command (when upgrades occur)
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

	cfg := a.GetConfig()
	output, err := ssh.Run(cfg, "apt-get upgrade -y")
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
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDAptUpgrade identifier
//	and description "Install available package updates (apt-get upgrade)".
func NewAptUpgrade() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDAptUpgrade)
	pb.SetDescription("Install available package updates (apt-get upgrade)")
	return &AptUpgrade{BasePlaybook: pb}
}
