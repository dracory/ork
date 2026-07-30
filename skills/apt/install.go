package apt

// Package apt documentation is in status.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// AptInstall installs one or more packages using apt-get.
// Packages are specified as a space-separated list via the ArgPackages argument.
//
// Usage:
//
//	node.Run(apt.NewAptInstall().SetArg(apt.ArgPackages, "nodejs npm"))
//
// Execution Flow:
//  1. Checks if the packages are already installed via dpkg-query
//  2. If any package is missing, runs apt-get install with DEBIAN_FRONTEND=noninteractive
//  3. Reports success or failure
//
// Args:
//   - packages: Space-separated list of package names (required), e.g. "nodejs npm"
//
// Expected Output:
//   - Success: "Packages installed: <packages>" message
//   - Failure: Error with apt output details
//
// Result Details:
//   - output: Full output from apt-get install command
//   - packages: The package list that was installed
//
// Use Cases:
//   - Install any set of system packages
//   - Part of a larger provisioning playbook
//
// Idempotency:
//   - Check() uses dpkg-query to skip installation if all packages are present
//   - apt-get install is itself idempotent; already-installed packages are left untouched
type AptInstall struct {
	*types.BaseSkill
}

// shellEscapePackages splits a space-separated package list, escapes each
// name with skills.ShellEscapeArg, and rejoins them. This prevents shell
// injection when interpolating user-supplied package names into commands.
func shellEscapePackages(packages string) string {
	parts := strings.Fields(packages)
	for i, p := range parts {
		parts[i] = skills.ShellEscapeArg(p)
	}
	return strings.Join(parts, " ")
}

// Check determines if any of the specified packages need to be installed.
// Returns true if at least one package is not currently installed, false if all are present.
func (a *AptInstall) Check() (bool, error) {
	packages := a.GetArg(ArgPackages)
	if packages == "" {
		return false, fmt.Errorf("no packages specified: set the %q argument", ArgPackages)
	}

	cfg := a.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if packages are installed")
		return true, nil
	}

	// dpkg-query exits 0 if all packages are recorded in the db, 1 if any is missing.
	// Required must be true so that ssh.Run propagates the non-zero exit code
	// instead of swallowing it (ssh.Run suppresses errors when Required is false).
	cmdCheck := types.Command{
		Command:     fmt.Sprintf("dpkg-query -W -- %s 2>/dev/null", shellEscapePackages(packages)),
		Description: "Check if packages are installed: " + packages,
		Required:    true,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		// dpkg-query exits 1 when a package is not found in the database.
		// At this point SSH is already established (Run calls Check internally),
		// so an error here is far more likely to be "package missing" than a
		// connection failure. Treat it as "needs install".
		return true, nil
	}
	return false, nil // all packages installed
}

// Run installs the packages specified in the ArgPackages argument.
// Changed is true when apt-get install actually installs packages,
// false when all packages are already installed or validation fails.
//
// Result.Details contains:
//   - output: Full output from apt-get install command
//   - packages: The packages that were installed
func (a *AptInstall) Run() types.Result {
	packages := a.GetArg(ArgPackages)
	if packages == "" {
		return types.Result{
			Changed: false,
			Message: "No packages specified",
			Error:   fmt.Errorf("no packages specified: set the %q argument", ArgPackages),
		}
	}

	cfg := a.GetNodeConfig()

	// Check if all packages are already installed (idempotency)
	needsInstall, err := a.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check installed packages",
			Error:   err,
		}
	}

	if !needsInstall {
		return types.Result{
			Changed: false,
			Message: "All packages already installed: " + packages,
		}
	}

	// See skills.DebianNonInteractive and skills.DpkgConfOptions for details
	cmdInstallStr := ""
	cmdInstallStr += skills.DebianNonInteractive   // prevent interactive prompts
	cmdInstallStr += " apt-get install -y -- "     // install packages, auto-confirm, -- prevents option injection
	cmdInstallStr += shellEscapePackages(packages) // escape each package name
	cmdInstallStr += skills.DpkgConfOptions        // keep local config, use maintainer default if unmodified

	cmdInstall := types.Command{
		Command:     cmdInstallStr,
		Description: "Install packages: " + packages,
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command)
		return types.Result{
			Changed: true,
			Message: "Would install packages: " + packages,
		}
	}

	cfg.GetLoggerOrDefault().Info("installing packages", "packages", packages)
	output, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Package installation failed",
			Error:   fmt.Errorf("apt-get install failed for %s: %w\nOutput: %s", packages, err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("packages installed", "packages", packages)
	return types.Result{
		Changed: true,
		Message: "Packages installed: " + packages,
		Details: map[string]string{
			"output":   output,
			"packages": packages,
		},
	}
}

// SetArgs sets the arguments for apt install.
// Returns AptInstall for fluent method chaining.
func (a *AptInstall) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// WithNodeConfig sets the node config and returns AptInstall for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *AptInstall) WithNodeConfig(cfg types.NodeConfig) *AptInstall {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// SetArg sets a single argument for apt install.
// Returns AptInstall for fluent method chaining.
func (a *AptInstall) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt install.
// Returns AptInstall for fluent method chaining.
func (a *AptInstall) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt install.
// Returns AptInstall for fluent method chaining.
func (a *AptInstall) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt install.
// Returns AptInstall for fluent method chaining.
func (a *AptInstall) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// NewAptInstall creates a new apt-install skill.
// Set the packages to install via SetArg(apt.ArgPackages, "pkg1 pkg2").
//
// Returns:
//
//	A AptInstall skill configured with IDAptInstall identifier
//	and description "Install packages (apt-get install)".
//
// Example:
//
//	node.Run(apt.NewAptInstall().SetArg(apt.ArgPackages, "nodejs npm"))
func NewAptInstall() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDAptInstall)
	pb.SetDescription("Install packages (apt-get install)")
	return &AptInstall{BaseSkill: pb}
}
