package swap

// Package swap documentation is in create.go

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SwapDelete removes the swap file.
// This skill disables swap, removes it from /etc/fstab, deletes the swap
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
	*skills.BaseSkill
}

// Check determines if swap needs to be removed.
// Per the skill interface convention, returns true if swap exists
// (meaning Run would modify the system by removing it), false if no swap exists.
//
// This method runs swapon --show to detect active swap devices.
// Non-empty output indicates swap is active and can be removed.
func (s *SwapDelete) Check() (bool, error) {
	cfg := s.GetNodeConfig()
	cmdCheck := types.Command{Command: "swapon --show=NAME --noheadings", Description: "Check if swap exists"}
	output, err := ssh.Run(cfg, cmdCheck)
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
//   - file: Path to the removed swap file
func (s *SwapDelete) Run() types.Result {
	cfg := s.GetNodeConfig()
	swapFilePath := s.GetArg(ArgSwapFilePath)
	if swapFilePath == "" {
		swapFilePath = DefaultSwapFilePath
	}

	// Check if swap exists
	needsDelete, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check swap status",
			Error:   err,
		}
	}

	if !needsDelete {
		return types.Result{
			Changed: false,
			Message: "No swap file to remove",
		}
	}

	cmdSwapoff := types.Command{Command: fmt.Sprintf("swapoff %s 2>/dev/null || true", swapFilePath), Description: "Disable swap"}
	cmdFstab := types.Command{Command: fmt.Sprintf(`sed -i '/%s/d' /etc/fstab`, swapFilePath), Description: "Remove swap from fstab"}
	cmdRm := types.Command{Command: fmt.Sprintf("rm -f %s", swapFilePath), Description: "Delete swap file"}

	cfg.GetLoggerOrDefault().Info("removing swap file", "path", swapFilePath)

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdSwapoff.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdFstab.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRm.Command)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would remove swap file at %s", swapFilePath),
		}
	}

	// Turn off swap
	_, _ = ssh.Run(cfg, cmdSwapoff)

	// Remove from fstab
	_, _ = ssh.Run(cfg, cmdFstab)

	// Delete file
	_, err = ssh.Run(cfg, cmdRm)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to remove swap file",
			Error:   fmt.Errorf("failed to remove swap file: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("swap file removed", "path", swapFilePath)
	return types.Result{
		Changed: true,
		Message: "Swap file removed",
		Details: map[string]string{
			"file": swapFilePath,
		},
	}
}

// NewSwapDelete creates a new swap-delete skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapDelete identifier
//	and description "Remove the swap file".
func NewSwapDelete() types.SkillInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDSwapDelete)
	pb.SetDescription("Remove the swap file")
	return &SwapDelete{BaseSkill: pb}
}
