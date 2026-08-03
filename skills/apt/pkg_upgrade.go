package apt

// Package apt documentation is in pkg_status.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// PkgUpgrade installs available package updates.
// This skill runs apt-get upgrade to install all available package updates.
// It first checks if updates are available by querying the package database,
// then installs them only if needed.
//
// Usage:
//
//	node.Run(apt.NewPkgUpgrade())
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
type PkgUpgrade struct {
	*types.BaseSkill
}

// Compile-time assertion that PkgUpgrade implements types.RunnableInterface.
var _ types.RunnableInterface = (*PkgUpgrade)(nil)

// Check determines if there are packages that need upgrading.
// Per the skill interface convention, returns true if upgrades are available
// (meaning Run would modify the system), false if system is already up to date.
//
// This method counts upgradable packages using apt list --upgradable.
// It does not run apt-get update; callers who need fresh package lists should
// run PkgUpdate first or use Run() which updates before upgrading.
func (a *PkgUpgrade) Check() (bool, error) {
	cfg := a.GetNodeConfig()

	// In dry-run mode, assume upgrades are needed so Run() reaches its dry-run guard
	if cfg.IsDryRunMode {
		return true, nil
	}

	// Check for upgradable packages
	cmdCheck := types.Command{Command: "apt list --upgradable 2>/dev/null | grep -c '\\[upgradable from:' || echo 0", Description: "Check for upgradable packages", Required: true}
	output, err := ssh.Run(cfg, cmdCheck)
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
func (a *PkgUpgrade) Run() types.Result {
	cfg := a.GetNodeConfig()

	// Update package lists before checking for upgrades
	if !cfg.IsDryRunMode {
		cmdUpdate := types.Command{Command: "apt-get update -qq", Description: "Update package lists", Required: true}
		_, err := ssh.Run(cfg, cmdUpdate)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to update package lists",
				Error:   fmt.Errorf("failed to update package lists: %w", err),
			}
		}
	}

	// Check if upgrades are needed
	needsUpgrade, err := a.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check for upgrades",
			Error:   err,
		}
	}

	if !needsUpgrade {
		return types.Result{
			Changed: false,
			Message: "All packages are up to date",
		}
	}

	// See skills.DebianNonInteractive and skills.DpkgConfOptions for details
	cmdUpgradeStr := ""
	cmdUpgradeStr += skills.DebianNonInteractive // prevent interactive prompts
	cmdUpgradeStr += " apt-get upgrade -y"       // upgrade all packages, auto-confirm
	cmdUpgradeStr += skills.DpkgConfOptions      // keep local config, use maintainer default if unmodified

	cmdUpgrade := types.Command{
		Command:     cmdUpgradeStr,
		Description: "Upgrade packages (keep local config files)",
		Required:    true,
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdUpgrade.Command)
		return types.Result{
			Changed: true,
			Message: "Would upgrade packages: " + cmdUpgrade.Command,
		}
	}

	cfg.GetLoggerOrDefault().Info("running apt upgrade")
	output, err := ssh.Run(cfg, cmdUpgrade)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Apt upgrade failed",
			Error:   fmt.Errorf("apt upgrade failed: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("apt upgrade completed")
	return types.Result{
		Changed: true,
		Message: "Packages upgraded successfully",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for apt upgrade.
// Returns PkgUpgrade for fluent method chaining.
func (a *PkgUpgrade) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// SetArg sets a single argument for apt upgrade.
// Returns PkgUpgrade for fluent method chaining.
func (a *PkgUpgrade) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt upgrade.
// Returns PkgUpgrade for fluent method chaining.
func (a *PkgUpgrade) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt upgrade.
// Returns PkgUpgrade for fluent method chaining.
func (a *PkgUpgrade) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt upgrade.
// Returns PkgUpgrade for fluent method chaining.
func (a *PkgUpgrade) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// WithNodeConfig sets the node config and returns PkgUpgrade for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *PkgUpgrade) WithNodeConfig(cfg types.NodeConfig) *PkgUpgrade {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// NewPkgUpgrade creates a new apt-upgrade skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDPkgUpgrade identifier
//	and description "Install available package updates (apt-get upgrade)".
func NewPkgUpgrade() *PkgUpgrade {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPkgUpgrade)
	pb.SetDescription("Install available package updates (apt-get upgrade)")
	return &PkgUpgrade{BaseSkill: pb}
}

// Deprecated: Use NewPkgUpgrade instead. NewAptUpgrade will be removed in a future version.
func NewAptUpgrade() *PkgUpgrade { return NewPkgUpgrade() }
