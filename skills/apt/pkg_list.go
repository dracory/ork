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

// PkgList lists installed packages via `apt list --installed`.
// This is a read-only skill that does not modify system state.
//
// Usage:
//
//	node.Run(apt.NewPkgList())
//	// filter to a single package:
//	node.Run(apt.NewPkgList().SetPackage("nginx"))
//
// Execution Flow:
//  1. Runs `apt list --installed`, optionally filtered to a single package
//  2. Strips the "Listing..." header line emitted by apt
//  3. Reports the count and full list of installed packages
//
// Args:
//   - package: Optional single package name to filter on, e.g. "nginx"
//
// Expected Output:
//   - Success: Message indicating the number of installed packages (or "no packages")
//   - Failure: Error with details of the apt command failure
//
// Result Details:
//   - installed_count: Number of installed packages matching the query (as string)
//   - packages: Full output from apt list --installed (when packages exist)
//
// Use Cases:
//   - Inventory installed software for auditing and compliance
//   - Verify a package is present without installing it
//   - Pre-flight check before provisioning or hardening playbooks
type PkgList struct {
	*types.BaseSkill
}

// Compile-time assertion that PkgList implements types.RunnableInterface.
var _ types.RunnableInterface = (*PkgList)(nil)

// Check always returns false since PkgList is read-only.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since apt list only queries
// package information, this always returns false.
func (a *PkgList) Check() (bool, error) {
	return false, nil
}

// Run executes apt list --installed and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - installed_count: Number of installed packages matching the query
//   - packages: Full output from apt list --installed (when packages exist)
func (a *PkgList) Run() types.Result {
	cfg := a.GetNodeConfig()

	// `apt list` prints a "Listing..." header line on stderr/stdout that is
	// not actual data; `tail -n +2` drops it. stderr is discarded so the
	// apt warning ("apt does not have a stable CLI interface") does not
	// leak into the captured output.
	cmdStr := "apt list --installed 2>/dev/null | tail -n +2"

	pkg := strings.TrimSpace(a.GetArg(ArgPackage))
	if pkg != "" {
		cmdStr = fmt.Sprintf("apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^%s/", skills.ShellEscapeGrep(pkg))
	}

	cmdList := types.Command{Command: cmdStr, Description: "List installed packages"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdList.Command)
		return types.Result{
			Changed: false,
			Message: "Would list installed packages",
		}
	}

	cfg.GetLoggerOrDefault().Info("listing installed packages")
	output, err := ssh.Run(cfg, cmdList)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to list installed packages",
			Error:   fmt.Errorf("failed to list installed packages: %w", err),
		}
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		cfg.GetLoggerOrDefault().Info("no installed packages matched")
		return types.Result{
			Changed: false,
			Message: "No installed packages matched",
			Details: map[string]string{
				"installed_count": "0",
			},
		}
	}

	count := 0
	for _, line := range strings.Split(trimmed, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	cfg.GetLoggerOrDefault().Info("installed packages", "count", count)
	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("%d installed packages", count),
		Details: map[string]string{
			"installed_count": fmt.Sprintf("%d", count),
			"packages":        trimmed,
		},
	}
}

// SetArgs sets the arguments for apt list.
// Returns PkgList for fluent method chaining.
func (a *PkgList) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// SetPackage sets a single package name to filter on and returns PkgList for chaining.
// Example: SetPackage("nginx")
func (a *PkgList) SetPackage(pkg string) *PkgList {
	a.BaseSkill.SetArg(ArgPackage, pkg)
	return a
}

// WithNodeConfig sets the node config and returns PkgList for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *PkgList) WithNodeConfig(cfg types.NodeConfig) *PkgList {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// SetArg sets a single argument for apt list.
// Returns PkgList for fluent method chaining.
func (a *PkgList) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt list.
// Returns PkgList for fluent method chaining.
func (a *PkgList) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt list.
// Returns PkgList for fluent method chaining.
func (a *PkgList) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt list.
// Returns PkgList for fluent method chaining.
func (a *PkgList) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// NewPkgList creates a new apt-list skill.
//
// Returns:
//
//	An PkgList skill configured with skills.IDPkgList identifier
//	and description "List installed packages (read-only)".
//
// Example:
//
//	node.Run(apt.NewPkgList())
//	node.Run(apt.NewPkgList().SetPackage("nginx"))
func NewPkgList() *PkgList {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPkgList)
	pb.SetDescription("List installed packages (read-only)")
	return &PkgList{BaseSkill: pb}
}

// Deprecated: Use NewPkgList instead. NewAptList will be removed in a future version.
func NewAptList() *PkgList { return NewPkgList() }
