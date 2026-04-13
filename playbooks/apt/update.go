package apt

// Package apt documentation is in status.go

import (
	"fmt"
	"log"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// AptUpdate refreshes the package database.
// This playbook runs apt-get update to download the latest package lists
// from configured repositories. This is a mutating operation that changes
// the local package cache.
//
// Usage:
//
//	go run . --playbook=apt-update
//
// Execution Flow:
//  1. Connects to remote server via SSH
//  2. Runs apt-get update -y to refresh package lists
//  3. Reports success or failure
//
// Expected Output:
//   - Success: "Package database updated" message
//   - Failure: Error with apt output details
//
// Result Details:
//   - output: Full output from apt-get update command
//
// Use Cases:
//   - Prepare system for package installations
//   - Ensure package lists are current before upgrades
//   - Initial server setup
//
// Idempotency:
//   - Always reports Changed=true because the cache modification time is updated
//   - The cost of checking if update is needed is similar to running it
type AptUpdate struct {
	*playbook.BasePlaybook
}

// Check always returns true for apt-update since cache refresh is always beneficial.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since apt-update always modifies
// the package cache timestamp, this always returns true.
//
// Note: The cost of checking if update is needed is similar to just running it,
// so we skip the check and always execute.
func (a *AptUpdate) Check() (bool, error) {
	return true, nil // Always run apt update
}

// Run executes apt-get update and returns the result.
// Changed is always true because the package cache is refreshed.
//
// Result.Details contains:
//   - output: Full output from apt-get update command
func (a *AptUpdate) Run() playbook.Result {
	log.Println("Running apt update...")

	cfg := a.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "apt-get update -y")
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
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDAptUpdate identifier
//	and description "Refresh package database (apt-get update)".
func NewAptUpdate() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDAptUpdate)
	pb.SetDescription("Refresh package database (apt-get update)")
	return &AptUpdate{BasePlaybook: pb}
}
