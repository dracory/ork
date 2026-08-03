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

// IsPkgInstalled checks whether a specific package is installed via
// `apt list --installed`. This is a read-only skill that does not modify
// system state.
//
// The package name is specified via the SetPackage method or ArgPackage
// argument. The skill returns whether the package is installed along with
// its version, architecture, and suite (distribution channel) parsed from
// the apt list output line.
//
// Usage:
//
//	node.Run(apt.NewIsPkgInstalled().SetPackage("nginx"))
//	// or equivalently:
//	node.Run(apt.NewIsPkgInstalled().SetArg(apt.ArgPackage, "nginx"))
//
// Execution Flow:
//  1. Runs `apt list --installed` filtered to the given package name
//  2. Parses the matching line (format: name/suite version arch [installed])
//  3. Reports installed status and metadata
//
// Args:
//   - package: Single package name to check (required), e.g. "nginx"
//
// Expected Output:
//   - Installed:   "package 'nginx' is installed (1.18.0-0ubuntu1)"
//   - Not found:   "package 'nginx' is not installed"
//   - Missing arg: Error indicating the package argument was not set
//
// Result Details:
//   - installed:     "true" or "false"
//   - version:       Installed version (e.g. "1.18.0-0ubuntu1"), empty if not installed
//   - architecture:  Package architecture (e.g. "amd64"), empty if not installed
//   - suite:         Distribution suite (e.g. "jammy,now"), empty if not installed
//   - line:          Raw matching apt list line, empty if not installed
//
// Use Cases:
//   - Pre-flight check before configuring a service that depends on a package
//   - Conditional branching in playbooks (install only if missing)
//   - Compliance auditing of installed software versions
type IsPkgInstalled struct {
	*types.BaseSkill
}

// Compile-time assertion that IsPkgInstalled implements types.RunnableInterface.
var _ types.RunnableInterface = (*IsPkgInstalled)(nil)

// Check returns true if the specified package is currently installed.
// This mirrors the boolean semantics of systemctl-is-active: the bool
// indicates whether the desired state ("installed") is met, not whether
// a change is needed. Since this skill is read-only, Run() never reports
// Changed=true regardless of the Check() result.
//
// Returns an error if no package argument is set.
func (a *IsPkgInstalled) Check() (bool, error) {
	pkg := strings.TrimSpace(a.GetArg(ArgPackage))
	if pkg == "" {
		return false, fmt.Errorf("no package specified: set the %q argument", ArgPackage)
	}

	cfg := a.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if package is installed", "package", pkg)
		return false, nil
	}

	line, err := a.queryInstalled(pkg)
	if err != nil {
		return false, err
	}

	return line != "", nil
}

// Run executes the installed-package check and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - installed:     "true" or "false"
//   - version:       Installed version, empty if not installed
//   - architecture:  Package architecture, empty if not installed
//   - suite:         Distribution suite, empty if not installed
//   - line:          Raw matching apt list line, empty if not installed
func (a *IsPkgInstalled) Run() types.Result {
	pkg := strings.TrimSpace(a.GetArg(ArgPackage))
	if pkg == "" {
		return types.Result{
			Changed: false,
			Message: "No package specified",
			Error:   fmt.Errorf("no package specified: set the %q argument", ArgPackage),
		}
	}

	cfg := a.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if package is installed", "package", pkg)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Would check if package '%s' is installed", pkg),
		}
	}

	cfg.GetLoggerOrDefault().Info("checking if package is installed", "package", pkg)
	line, err := a.queryInstalled(pkg)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to check if package '%s' is installed", pkg),
			Error:   fmt.Errorf("failed to check if package '%s' is installed: %w", pkg, err),
		}
	}

	line = strings.TrimSpace(line)
	if line == "" {
		cfg.GetLoggerOrDefault().Info("package is not installed", "package", pkg)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("package '%s' is not installed", pkg),
			Details: map[string]string{
				"installed": "false",
			},
		}
	}

	details := parseAptListLine(line)
	details["installed"] = "true"
	details["line"] = line

	cfg.GetLoggerOrDefault().Info("package is installed", "package", pkg, "version", details["version"])
	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("package '%s' is installed (%s)", pkg, details["version"]),
		Details: details,
	}
}

// queryInstalled runs `apt list --installed` filtered to the package name
// and returns the first matching line (empty string if no match).
// The package name is shell-escaped to prevent injection.
func (a *IsPkgInstalled) queryInstalled(pkg string) (string, error) {
	cfg := a.GetNodeConfig()

	// `apt list --installed` prints a "Listing..." header that we strip with
	// `tail -n +2`. stderr is discarded to suppress apt's "no stable CLI
	// interface" warning. The package name is shell-escaped before being
	// passed to grep so user input cannot inject shell metacharacters.
	cmdStr := fmt.Sprintf(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^%s/",
		skills.ShellEscapeGrep(pkg),
	)

	cmd := types.Command{
		Command:     cmdStr,
		Description: fmt.Sprintf("Check if package is installed: %s", pkg),
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return "", err
	}

	// grep is anchored to ^<pkg>/ so only lines with the exact package name
	// prefix should appear. We still verify an exact name match in Go to
	// handle any edge cases.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		// Return the line whose package name (before the first "/") matches exactly.
		nameEnd := strings.Index(l, "/")
		if nameEnd > 0 && l[:nameEnd] == pkg {
			return l, nil
		}
	}
	return "", nil
}

// parseAptListLine parses a single `apt list --installed` output line into
// its components. Expected format:
//
//	<name>/<suite> <version> <architecture> [installed]
//
// Example:
//
//	nginx/jammy,now 1.18.0-0ubuntu1 amd64 [installed]
//
// Returns a map with keys: version, architecture, suite. Missing fields
// are returned as empty strings rather than errors, since apt output can
// vary across distributions.
func parseAptListLine(line string) map[string]string {
	details := map[string]string{
		"version":      "",
		"architecture": "",
		"suite":        "",
	}

	// Split into "name/suite" and the rest on the first space.
	firstSpace := strings.IndexAny(line, " \t")
	if firstSpace <= 0 {
		return details
	}

	nameSuite := line[:firstSpace]
	rest := strings.TrimSpace(line[firstSpace:])

	// name/suite
	slash := strings.Index(nameSuite, "/")
	if slash > 0 {
		details["suite"] = nameSuite[slash+1:]
	}

	// rest = "version architecture [installed]"
	fields := strings.Fields(rest)
	if len(fields) >= 1 {
		details["version"] = fields[0]
	}
	if len(fields) >= 2 {
		details["architecture"] = fields[1]
	}

	return details
}

// SetArgs sets the arguments for apt is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (a *IsPkgInstalled) SetArgs(args map[string]string) types.RunnableInterface {
	a.BaseSkill.SetArgs(args)
	return a
}

// SetPackage sets the package name to check and returns IsPkgInstalled for chaining.
// Example: SetPackage("nginx")
func (a *IsPkgInstalled) SetPackage(pkg string) *IsPkgInstalled {
	a.BaseSkill.SetArg(ArgPackage, pkg)
	return a
}

// WithNodeConfig sets the node config and returns IsPkgInstalled for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (a *IsPkgInstalled) WithNodeConfig(cfg types.NodeConfig) *IsPkgInstalled {
	a.BaseSkill.SetNodeConfig(cfg)
	return a
}

// SetArg sets a single argument for apt is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (a *IsPkgInstalled) SetArg(key, value string) types.RunnableInterface {
	a.BaseSkill.SetArg(key, value)
	return a
}

// SetID sets the ID for apt is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (a *IsPkgInstalled) SetID(id string) types.RunnableInterface {
	a.BaseSkill.SetID(id)
	return a
}

// SetDescription sets the description for apt is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (a *IsPkgInstalled) SetDescription(description string) types.RunnableInterface {
	a.BaseSkill.SetDescription(description)
	return a
}

// SetTimeout sets the timeout for apt is-installed.
// Returns IsPkgInstalled for fluent method chaining.
func (a *IsPkgInstalled) SetTimeout(timeout time.Duration) types.RunnableInterface {
	a.BaseSkill.SetTimeout(timeout)
	return a
}

// NewIsPkgInstalled creates a new apt-is-installed skill.
// Set the package to check via SetPackage("nginx").
//
// Returns:
//
//	An IsPkgInstalled skill configured with skills.IDIsPkgInstalled identifier
//	and description "Check if a package is installed (read-only)".
//
// Example:
//
//	node.Run(apt.NewIsPkgInstalled().SetPackage("nginx"))
func NewIsPkgInstalled() *IsPkgInstalled {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDIsPkgInstalled)
	pb.SetDescription("Check if a package is installed (read-only)")
	return &IsPkgInstalled{BaseSkill: pb}
}
