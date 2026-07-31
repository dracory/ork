package php

// Package php documentation is in constants.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Uninstall removes all PHP packages for a version and cleans up FPM configuration.
//
// It stops and disables the php<version>-fpm service, purges all php<version>-*
// packages via apt, and removes the /etc/php/<version> directory.
//
// Args:
//   - version (required): PHP version to remove, e.g. "8.3"
//
// Execution Flow:
//  1. Validates version arg
//  2. Stops and disables php<version>-fpm service
//  3. Purges all php<version>-* packages via apt
//  4. Removes /etc/php/<version> directory
//
// Usage:
//
//	node.Run(php.NewUninstall().SetVersion("8.3"))
//
// Idempotency:
//   - Check() verifies whether any php<version>-* packages are installed.
//     When none are installed, Run() is a no-op.
type Uninstall struct {
	*types.BaseSkill
}

// Compile-time assertion that Uninstall implements types.RunnableInterface.
var _ types.RunnableInterface = (*Uninstall)(nil)

// Check determines if PHP needs to be uninstalled.
// Returns true if any php<version>-* packages are installed, false otherwise.
func (s *Uninstall) Check() (bool, error) {
	version := s.GetArg(ArgVersion)
	if version == "" {
		return false, fmt.Errorf("no version specified: set the %q argument", ArgVersion)
	}

	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if PHP is installed")
		return true, nil
	}

	// Check if any php<version>-* packages are installed.
	// dpkg-query exits 0 if the package is recorded, 1 if missing.
	cmdCheck := types.Command{
		Command:     fmt.Sprintf("dpkg-query -W 'php%s-*' 2>/dev/null", skills.ShellEscapeArg(version)),
		Description: "Check if php" + version + " packages are installed",
		Required:    true,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		// No packages found — nothing to uninstall.
		return false, nil
	}
	return true, nil // packages exist, need to uninstall
}

// Run uninstalls PHP packages and cleans up FPM configuration.
// Changed is true when packages were purged, false when nothing was installed
// or validation fails.
//
// Result.Details contains:
//   - version: The PHP version that was removed
//   - output: Full output from the apt-get purge command
func (s *Uninstall) Run() types.Result {
	version := s.GetArg(ArgVersion)
	if version == "" {
		return types.Result{
			Changed: false,
			Message: "No version specified",
			Error:   fmt.Errorf("no version specified: set the %q argument", ArgVersion),
		}
	}

	cfg := s.GetNodeConfig()

	// Check if anything needs to be uninstalled (idempotency).
	needsChange, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check PHP installation",
			Error:   err,
		}
	}

	if !needsChange {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("PHP %s is not installed", version),
		}
	}

	// Check for dry-run mode.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would uninstall PHP", "version", version)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would uninstall PHP %s", version),
		}
	}

	// Stop and disable php<version>-fpm service.
	cmdStop := types.Command{
		Command:     fmt.Sprintf("systemctl stop php%s-fpm 2>/dev/null; systemctl disable php%s-fpm 2>/dev/null", skills.ShellEscapeArg(version), skills.ShellEscapeArg(version)),
		Description: "Stop and disable php" + version + "-fpm",
	}
	ssh.Run(cfg, cmdStop) // best effort — ignore errors if service not running

	// Purge php<version>-* packages.
	// See skills.DebianNonInteractive and skills.DpkgConfOptions for details.
	cmdPurgeStr := ""
	cmdPurgeStr += skills.DebianNonInteractive   // prevent interactive prompts
	cmdPurgeStr += " apt-get purge -y -- "       // purge packages, auto-confirm, -- prevents option injection
	cmdPurgeStr += skills.ShellEscapeArg("php" + version + "-*")
	cmdPurgeStr += skills.DpkgConfOptions        // keep local config, use maintainer default if unmodified

	cmdPurge := types.Command{
		Command:     cmdPurgeStr,
		Description: "Purge php" + version + "-* packages",
	}
	cfg.GetLoggerOrDefault().Info("purging PHP packages", "version", version)
	purgeOutput, err := ssh.Run(cfg, cmdPurge)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "PHP package purge failed",
			Error:   fmt.Errorf("apt-get purge failed for php%s-*: %w\nOutput: %s", version, err, purgeOutput),
		}
	}

	// Remove /etc/php/<version> directory.
	cmdRemoveDir := types.Command{
		Command:     fmt.Sprintf("rm -rf %s", skills.ShellEscapeArg("/etc/php/"+version)),
		Description: "Remove /etc/php/" + version + " directory",
	}
	ssh.Run(cfg, cmdRemoveDir) // best effort — may not exist

	cfg.GetLoggerOrDefault().Info("PHP uninstalled", "version", version)
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("PHP %s uninstalled", version),
		Details: map[string]string{
			"version": version,
			"output":  purgeOutput,
		},
	}
}

// SetArgs sets the arguments for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Uninstall) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetVersion sets the PHP version to uninstall (e.g. "8.3") and returns Uninstall for chaining.
func (s *Uninstall) SetVersion(version string) *Uninstall {
	s.BaseSkill.SetArg(ArgVersion, version)
	return s
}

// WithNodeConfig sets the node config and returns Uninstall for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *Uninstall) WithNodeConfig(cfg types.NodeConfig) *Uninstall {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArg sets a single argument for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Uninstall) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Uninstall) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Uninstall) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Uninstall) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewUninstall creates a new php-uninstall skill.
// Set the PHP version via SetVersion("8.3").
//
// Returns:
//
//	An Uninstall skill configured with skills.IDPhpUninstall identifier
//	and description "Uninstall PHP and remove FPM configuration".
//
// Example:
//
//	node.Run(php.NewUninstall().SetVersion("8.3"))
func NewUninstall() *Uninstall {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpUninstall)
	pb.SetDescription("Uninstall PHP and remove FPM configuration")
	return &Uninstall{BaseSkill: pb}
}
