// Package apt provides playbooks for managing Debian/Ubuntu packages via apt.
// It includes operations for checking package status, updating the package database,
// and installing available upgrades.
package apt

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// PkgStatus shows available package updates without installing them.
// This is a read-only skill that reports how many packages are available
// for upgrade without modifying the system.
//
// Usage:
//
//	node.Run(apt.NewPkgStatus())
//
// Execution Flow:
//  1. Lists upgradable packages with apt list --upgradable
//  2. Reports count and details of available updates
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
type PkgStatus struct {
	*types.BaseSkill
}

// Compile-time assertion that PkgStatus implements types.RunnableInterface.
var _ types.RunnableInterface = (*PkgStatus)(nil)

// Check always returns false since PkgStatus is read-only.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since apt-status only queries
// package information, this always returns false.
func (a *PkgStatus) Check() (bool, error) {
	return false, nil
}

// Run executes apt status check and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - upgradable_count: Number of packages available for upgrade
//   - packages: Full output from apt list --upgradable (when packages exist)
func (a *PkgStatus) Run() types.Result {
	cfg := a.GetNodeConfig()

	cmdList := types.Command{Command: "apt list --upgradable 2>/dev/null | tail -n +2", Description: "List upgradable packages"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdList.Command)
		return types.Result{
			Changed: false,
			Message: "Would check for available package updates",
		}
	}

	cfg.GetLoggerOrDefault().Info("checking for available updates")
	output, err := ssh.Run(cfg, cmdList)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to list upgradable packages",
			Error:   fmt.Errorf("failed to list upgradable packages: %w", err),
		}
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" || trimmed == "0" {
		cfg.GetLoggerOrDefault().Info("all packages are up to date")
		return types.Result{
			Changed: false,
			Message: "All packages are up to date",
			Details: map[string]string{
				"upgradable_count": "0",
			},
		}
	}

	lineCount := 0
	for _, line := range strings.Split(trimmed, "\n") {
		if strings.TrimSpace(line) != "" {
			lineCount++
		}
	}
	cfg.GetLoggerOrDefault().Info("available upgrades", "packages", trimmed)
	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("%d packages available for upgrade", lineCount),
		Details: map[string]string{
			"upgradable_count": fmt.Sprintf("%d", lineCount),
			"packages":         trimmed,
		},
	}
}

// SetArgs sets the arguments for apt status.
// Returns PkgStatus for fluent method chaining.
func (a *PkgStatus) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// SetArg sets a single argument for apt status.
// Returns PkgStatus for fluent method chaining.
func (a *PkgStatus) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt status.
// Returns PkgStatus for fluent method chaining.
func (a *PkgStatus) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt status.
// Returns PkgStatus for fluent method chaining.
func (a *PkgStatus) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt status.
// Returns PkgStatus for fluent method chaining.
func (a *PkgStatus) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// WithNodeConfig sets the node config and returns PkgStatus for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *PkgStatus) WithNodeConfig(cfg types.NodeConfig) *PkgStatus {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// NewPkgStatus creates a new apt-status skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDPkgStatus identifier
//	and description "Show available package updates (read-only)".
func NewPkgStatus() *PkgStatus {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPkgStatus)
	pb.SetDescription("Show available package updates (read-only)")
	return &PkgStatus{BaseSkill: pb}
}

// Deprecated: Use NewPkgStatus instead. NewAptStatus will be removed in a future version.
func NewAptStatus() *PkgStatus { return NewPkgStatus() }
