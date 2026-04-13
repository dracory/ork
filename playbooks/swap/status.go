package swap

// Package swap documentation is in create.go

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapStatus shows current swap usage.
// This is a read-only playbook that displays information about active swap
// devices, including size, usage, and priority. It reports whether swap is
// enabled and provides detailed status output from swapon.
//
// Usage:
//
//	go run . --playbook=swap-status
//
// Execution Flow:
//  1. Connects to remote server via SSH
//  2. Runs swapon --show to get swap status
//  3. Reports swap information or indicates no swap is active
//
// Expected Output:
//   - Success (swap active): "Swap is active" with status details
//   - Success (no swap): "No swap is currently active"
//   - Failure: Error with swapon command failure details
//
// Result Details:
//   - active: "true" when swap exists, "false" when no swap
//   - status: Full output from swapon --show (when swap exists)
//
// Use Cases:
//   - Monitor swap usage during memory pressure
//   - Verify swap configuration after creation/deletion
//   - Inventory current system configuration
//
// Idempotency:
//   - Always reports Changed=false since this is read-only
type SwapStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since SwapStatus is read-only.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since swap-status only queries
// swap information, this always returns false.
func (s *SwapStatus) Check() (bool, error) {
	return false, nil
}

// Run displays swap status and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - active: "true" when swap exists, "false" when no swap active
//   - status: Full output from swapon --show command (when swap active)
func (s *SwapStatus) Run() playbook.Result {
	cfg := s.GetConfig()
	output, err := ssh.Run(cfg, "swapon --show")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to get swap status",
			Error:   fmt.Errorf("failed to get swap status: %w", err),
		}
	}

	if strings.TrimSpace(output) == "" {
		cfg.GetLoggerOrDefault().Info("no swap active")
		return playbook.Result{
			Changed: false,
			Message: "No swap is currently active",
			Details: map[string]string{
				"active": "false",
			},
		}
	}

	cfg.GetLoggerOrDefault().Info("swap status", "status", output)
	return playbook.Result{
		Changed: false, // Read-only operation
		Message: "Swap is active",
		Details: map[string]string{
			"active": "true",
			"status": output,
		},
	}
}

// NewSwapStatus creates a new swap-status playbook.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapStatus identifier
//	and description "Show swap status and usage".
func NewSwapStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapStatus)
	pb.SetDescription("Show swap status and usage")
	return &SwapStatus{BasePlaybook: pb}
}
