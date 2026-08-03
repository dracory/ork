package apt

// Package apt documentation is in pkg_status.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// PkgUpdate refreshes the package database.
// This skill runs apt-get update to download the latest package lists
// from configured repositories. This is a mutating operation that changes
// the local package cache.
//
// Usage:
//
//	node.Run(apt.NewPkgUpdate())
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
type PkgUpdate struct {
	*types.BaseSkill
}

// Compile-time assertion that PkgUpdate implements types.RunnableInterface.
var _ types.RunnableInterface = (*PkgUpdate)(nil)

// Check always returns true for apt-update since cache refresh is always beneficial.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since apt-update always modifies
// the package cache timestamp, this always returns true.
//
// Note: The cost of checking if update is needed is similar to just running it,
// so we skip the check and always execute.
func (a *PkgUpdate) Check() (bool, error) {
	cfg := a.GetNodeConfig()

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if apt update is needed")
		return true, nil
	}

	return true, nil // Always run apt update
}

// Run executes apt-get update and returns the result.
// Changed is always true because the package cache is refreshed.
//
// Result.Details contains:
//   - output: Full output from apt-get update command
func (a *PkgUpdate) Run() types.Result {
	cfg := a.GetNodeConfig()
	cmdUpdate := types.Command{Command: "apt-get update -y", Description: "Update package database", Required: true}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdUpdate.Command)
		return types.Result{
			Changed: true,
			Message: "Would update package database: " + cmdUpdate.Command,
		}
	}

	cfg.GetLoggerOrDefault().Info("running apt update")
	output, err := ssh.Run(cfg, cmdUpdate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Apt update failed",
			Error:   fmt.Errorf("apt update failed: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("apt update completed")
	return types.Result{
		Changed: true, // Cache was refreshed
		Message: "Package database updated",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for apt update.
// Returns PkgUpdate for fluent method chaining.
func (a *PkgUpdate) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// WithNodeConfig sets the node config and returns PkgUpdate for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *PkgUpdate) WithNodeConfig(cfg types.NodeConfig) *PkgUpdate {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// SetArg sets a single argument for apt update.
// Returns PkgUpdate for fluent method chaining.
func (a *PkgUpdate) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt update.
// Returns PkgUpdate for fluent method chaining.
func (a *PkgUpdate) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt update.
// Returns PkgUpdate for fluent method chaining.
func (a *PkgUpdate) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt update.
// Returns PkgUpdate for fluent method chaining.
func (a *PkgUpdate) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// NewPkgUpdate creates a new apt-update skill.
//
// Returns:
//
//	A PkgUpdate skill configured with IDPkgUpdate identifier
//	and description "Refresh package database (apt-get update)".
func NewPkgUpdate() *PkgUpdate {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPkgUpdate)
	pb.SetDescription("Refresh package database (apt-get update)")
	return &PkgUpdate{BaseSkill: pb}
}

// Deprecated: Use NewPkgUpdate instead. NewAptUpdate will be removed in a future version.
func NewAptUpdate() *PkgUpdate { return NewPkgUpdate() }
