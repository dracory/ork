package swap

// Package swap documentation is in create.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SwapStatus shows current swap usage.
// This is a read-only skill that displays information about active swap
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
	*types.BaseSkill
}

// Check always returns false since SwapStatus is read-only.
// Per the skill interface convention, the bool return indicates whether
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
func (s *SwapStatus) Run() types.Result {
	cfg := s.GetNodeConfig()
	cmdStatus := types.Command{Command: "swapon --show", Description: "Check swap status"}

	// Check for dry-run mode - display actual command
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStatus.Command)
		return types.Result{
			Changed: false,
			Message: "Would check swap status",
		}
	}

	output, err := ssh.Run(cfg, cmdStatus)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to get swap status",
			Error:   fmt.Errorf("failed to get swap status: %w", err),
		}
	}

	if strings.TrimSpace(output) == "" {
		cfg.GetLoggerOrDefault().Info("no swap active")
		return types.Result{
			Changed: false,
			Message: "No swap is currently active",
			Details: map[string]string{
				"active": "false",
			},
		}
	}

	cfg.GetLoggerOrDefault().Info("swap status", "status", output)
	return types.Result{
		Changed: false, // Read-only operation
		Message: "Swap is active",
		Details: map[string]string{
			"active": "true",
			"status": output,
		},
	}
}

// SetArgs sets the arguments for swap status.
// Returns SwapStatus for fluent method chaining.
func (s *SwapStatus) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for swap status.
// Returns SwapStatus for fluent method chaining.
func (s *SwapStatus) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for swap status.
// Returns SwapStatus for fluent method chaining.
func (s *SwapStatus) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for swap status.
// Returns SwapStatus for fluent method chaining.
func (s *SwapStatus) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for swap status.
// Returns SwapStatus for fluent method chaining.
func (s *SwapStatus) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSwapStatus creates a new swap-status skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapStatus identifier
//	and description "Show swap status and usage".
func NewSwapStatus() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSwapStatus)
	pb.SetDescription("Show swap status and usage")
	return &SwapStatus{BaseSkill: pb}
}
