package dpkg

// Package dpkg documentation is in constants.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// IsPkgInstalled checks whether a specific package is installed via
// `dpkg-query`. This is a read-only skill that does not modify system state.
//
// It is lighter than apt.IsPkgInstalled (which uses `apt list --installed | grep`)
// because dpkg-query is a direct database lookup with no pipeline. Use this
// skill when you only need a boolean answer; use apt.IsPkgInstalled when you need
// version, architecture, or suite metadata.
//
// Usage:
//
//	node.Run(dpkg.NewIsPkgInstalled().SetPackage("nginx"))
//	// or equivalently:
//	node.Run(dpkg.NewIsPkgInstalled().SetArg(dpkg.ArgPackage, "nginx"))
//
// Execution Flow:
//  1. Runs `dpkg-query -W -f='${Status}' <package>`
//  2. Checks if the output contains "install ok installed"
//  3. Reports boolean installed status
//
// Args:
//   - package: Single package name to check (required), e.g. "nginx"
//
// Expected Output:
//   - Installed:   "package 'nginx' is installed"
//   - Not found:   "package 'nginx' is not installed"
//   - Missing arg: Error indicating the package argument was not set
//
// Result Details:
//   - installed: "true" or "false"
//
// Use Cases:
//   - Pre-flight check before installing a package
//   - Conditional branching in playbooks
//   - Idempotency check for skills that wrap apt-get install
type IsPkgInstalled struct {
	*types.BaseSkill
}

// Compile-time assertion that IsPkgInstalled implements types.RunnableInterface.
var _ types.RunnableInterface = (*IsPkgInstalled)(nil)

// Check returns true if the specified package is currently installed.
// Returns an error if no package argument is set.
func (d *IsPkgInstalled) Check() (bool, error) {
	pkg := strings.TrimSpace(d.GetArg(ArgPackage))
	if pkg == "" {
		return false, fmt.Errorf("no package specified: set the %q argument", ArgPackage)
	}

	cfg := d.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if package is installed", "package", pkg)
		return false, nil
	}

	return d.queryInstalled(pkg)
}

// Run executes the installed-package check and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - installed: "true" or "false"
func (d *IsPkgInstalled) Run() types.Result {
	pkg := strings.TrimSpace(d.GetArg(ArgPackage))
	if pkg == "" {
		return types.Result{
			Changed: false,
			Message: "No package specified",
			Error:   fmt.Errorf("no package specified: set the %q argument", ArgPackage),
		}
	}

	cfg := d.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if package is installed", "package", pkg)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Would check if package '%s' is installed", pkg),
		}
	}

	cfg.GetLoggerOrDefault().Info("checking if package is installed", "package", pkg)
	isInstalled, err := d.queryInstalled(pkg)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to check if package '%s' is installed", pkg),
			Error:   fmt.Errorf("failed to check if package '%s' is installed: %w", pkg, err),
		}
	}

	if isInstalled {
		cfg.GetLoggerOrDefault().Info("package is installed", "package", pkg)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("package '%s' is installed", pkg),
			Details: map[string]string{
				"installed": "true",
			},
		}
	}

	cfg.GetLoggerOrDefault().Info("package is not installed", "package", pkg)
	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("package '%s' is not installed", pkg),
		Details: map[string]string{
			"installed": "false",
		},
	}
}

// queryInstalled runs `dpkg-query -W -f='${Status}' <package>` and returns
// true if the package is recorded as "install ok installed". The package
// name is shell-escaped to prevent injection.
func (d *IsPkgInstalled) queryInstalled(pkg string) (bool, error) {
	cfg := d.GetNodeConfig()

	cmdStr := fmt.Sprintf("dpkg-query -W -f='${Status}' -- %s 2>/dev/null", skills.ShellEscapeArg(pkg))

	cmd := types.Command{
		Command:     cmdStr,
		Description: fmt.Sprintf("Check if package is installed: %s", pkg),
		Required:    true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		// dpkg-query exits non-zero when the package is not in the database.
		// Treat that as "not installed" rather than an error.
		return false, nil
	}

	return strings.Contains(strings.TrimSpace(output), "install ok installed"), nil
}

// SetArgs sets the arguments for dpkg-is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (d *IsPkgInstalled) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetPackage sets the package name to check and returns IsPkgInstalled for chaining.
// Example: SetPackage("nginx")
func (d *IsPkgInstalled) SetPackage(pkg string) *IsPkgInstalled {
	d.BaseSkill.SetArg(ArgPackage, pkg)
	return d
}

// WithNodeConfig sets the node config and returns IsPkgInstalled for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (d *IsPkgInstalled) WithNodeConfig(cfg types.NodeConfig) *IsPkgInstalled {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetArg sets a single argument for dpkg-is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (d *IsPkgInstalled) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID for dpkg-is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (d *IsPkgInstalled) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description for dpkg-is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (d *IsPkgInstalled) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout for dpkg-is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (d *IsPkgInstalled) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewIsPkgInstalled creates a new dpkg-is-installed skill.
// Set the package to check via SetPackage("nginx").
//
// Returns:
//
//	An IsPkgInstalled skill configured with skills.IDDpkgIsPkgInstalled identifier
//	and description "Check if a package is installed via dpkg-query (read-only)".
//
// Example:
//
//	node.Run(dpkg.NewIsPkgInstalled().SetPackage("nginx"))
func NewIsPkgInstalled() *IsPkgInstalled {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDDpkgIsPkgInstalled)
	pb.SetDescription("Check if a package is installed via dpkg-query (read-only)")
	return &IsPkgInstalled{BaseSkill: pb}
}
