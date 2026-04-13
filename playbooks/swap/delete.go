package swap

// Package swap documentation is in create.go

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapDelete removes the swap file.
// This playbook disables swap, removes it from /etc/fstab, deletes the swap
// file at /swapfile, and cleans up system configuration.
//
// Usage:
//
//	go run . --playbook=swap-delete
//
// Execution Flow:
//  1. Checks if swap exists using swapon --show
//  2. If swap exists:
//     a. Disables swap with swapoff
//     b. Removes /swapfile entry from /etc/fstab
//     c. Deletes /swapfile
//  3. Reports success or no-op if no swap existed
//
// Expected Output:
//   - Success (swap removed): "Swap file removed" with file path detail
//   - Success (no swap): "No swap file to remove"
//   - Failure: Error with specific failure reason
//
// Result Details:
//   - file: Path to the removed swap file ("/swapfile")
//
// Use Cases:
//   - Remove swap before resizing partitions
//   - Reclaim disk space from oversized swap
//   - Clean up temporary swap configuration
//
// Idempotency:
//   - Reports Changed=true when swap is removed
//   - Reports Changed=false when no swap exists
//
// Safety:
//   - Uses swapoff with error suppression (|| true) to handle edge cases
//   - Safely removes fstab entry with sed pattern matching
//   - Idempotent - safe to run multiple times
type SwapDelete struct {
	*playbook.BasePlaybook
}

// Check determines if swap needs to be removed.
// Per the playbook interface convention, returns true if swap exists
// (meaning Run would modify the system by removing it), false if no swap exists.
//
// This method runs swapon --show to detect active swap devices.
// Non-empty output indicates swap is active and can be removed.
func (s *SwapDelete) Check() (bool, error) {
	cfg := s.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is not empty, swap exists - need to remove
	return strings.TrimSpace(output) != "", nil
}

// Run removes the swap file and returns detailed result.
// Changed is true when swap is removed, false if no swap existed.
//
// Result.Details contains:
//   - file: Path to the removed swap file ("/swapfile")
func (s *SwapDelete) Run() playbook.Result {
	cfg := s.GetConfig()
	// Check if swap exists
	needsDelete, err := s.Check()
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to check swap status",
			Error:   err,
		}
	}

	if !needsDelete {
		return playbook.Result{
			Changed: false,
			Message: "No swap file to remove",
		}
	}

	log.Println("Removing swap file...")

	// Turn off swap
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapoff /swapfile 2>/dev/null || true")

	// Remove from fstab
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `sed -i '/\/swapfile/d' /etc/fstab`)

	// Delete file
	_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "rm -f /swapfile")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to remove swap file",
			Error:   fmt.Errorf("failed to remove swap file: %w", err),
		}
	}

	log.Println("Swap file removed successfully")
	return playbook.Result{
		Changed: true,
		Message: "Swap file removed",
		Details: map[string]string{
			"file": "/swapfile",
		},
	}
}

// NewSwapDelete creates a new swap-delete playbook.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapDelete identifier
//	and description "Remove the swap file".
func NewSwapDelete() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapDelete)
	pb.SetDescription("Remove the swap file")
	return &SwapDelete{BasePlaybook: pb}
}
