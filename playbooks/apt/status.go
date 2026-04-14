// Package apt provides playbooks for managing Debian/Ubuntu packages via apt.
// It includes operations for checking package status, updating the package database,
// and installing available upgrades.
package apt

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// AptStatus shows available package updates without installing them.
// This is a read-only playbook that refreshes the package database and reports
// how many packages are available for upgrade without modifying the system.
//
// Usage:
//
//	go run . --playbook=apt-status
//
// Execution Flow:
//  1. Runs apt-get update to refresh package lists
//  2. Lists upgradable packages with apt list --upgradable
//  3. Reports count and details of available updates
//
// Expected Output:
//   - Success: Message indicating number of packages available for upgrade (or "up to date")
//   - Failure: Error with details of the apt command failure
//
// Result Details:
//   - upgradable_count: Number of packages available for upgrade (as string)
//   - packages: Full list of upgradable packages (when packages are available)
//
// Use Cases:
//   - Monitor available security updates without installing them
//   - Pre-flight check before maintenance windows
//   - Reporting and compliance auditing
type AptStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since AptStatus is read-only.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since apt-status only queries
// package information, this always returns false.
func (a *AptStatus) Check() (bool, error) {
	return false, nil
}

// Run executes apt status check and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - upgradable_count: Number of packages available for upgrade
//   - packages: Full output from apt list --upgradable (when packages exist)
func (a *AptStatus) Run() playbook.Result {
	cfg := a.GetNodeConfig()

	cmdUpdate := types.Command{Command: "apt-get update -qq", Description: "Update package lists"}
	cmdList := types.Command{Command: "apt list --upgradable 2>/dev/null | tail -n +2", Description: "List upgradable packages"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdUpdate.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdList.Command)
		return playbook.Result{
			Changed: false,
			Message: "Would check for available package updates",
		}
	}

	cfg.GetLoggerOrDefault().Info("checking for available updates")
	_, err := ssh.Run(cfg, cmdUpdate)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to update package lists",
			Error:   fmt.Errorf("failed to update package lists: %w", err),
		}
	}

	output, err := ssh.Run(cfg, cmdList)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list upgradable packages",
			Error:   fmt.Errorf("failed to list upgradable packages: %w", err),
		}
	}

	count := strings.TrimSpace(output)
	if count == "" || count == "0" {
		cfg.GetLoggerOrDefault().Info("all packages are up to date")
		return playbook.Result{
			Changed: false,
			Message: "All packages are up to date",
			Details: map[string]string{
				"upgradable_count": "0",
			},
		}
	}

	cfg.GetLoggerOrDefault().Info("available upgrades", "packages", output)
	return playbook.Result{
		Changed: false,
		Message: fmt.Sprintf("%d packages available for upgrade", strings.Count(output, "\n")+1),
		Details: map[string]string{
			"upgradable_count": fmt.Sprintf("%d", strings.Count(output, "\n")+1),
			"packages":         output,
		},
	}
}

// NewAptStatus creates a new apt-status playbook.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDAptStatus identifier
//	and description "Show available package updates (read-only)".
func NewAptStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDAptStatus)
	pb.SetDescription("Show available package updates (read-only)")
	return &AptStatus{BasePlaybook: pb}
}
