package php

// Package php documentation is in constants.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Install installs PHP with extensions and configures FPM pool to run as a non-root user.
//
// It installs the specified PHP version along with a default or custom set of
// extensions, then configures the PHP-FPM pool so that the FPM worker processes
// run as the given user. Finally it restarts and enables the php<version>-fpm
// service so the configuration takes effect immediately and survives reboots.
//
// Args:
//   - version (required): PHP version, e.g. "8.3"
//   - user (required): User to run PHP-FPM as
//   - extensions (optional): Space-separated additional extensions, defaults to DefaultExtensions
//
// Execution Flow:
//  1. Validates version and user args
//  2. Checks if php<version> is already installed and FPM is configured for user
//  3. Runs apt-get update
//  4. Installs php<version>-<ext> for each extension
//  5. Configures FPM pool: sets user, group, listen.owner, listen.group, listen.mode
//  6. Restarts and enables php<version>-fpm service
//
// Usage:
//
//	node.Run(php.NewInstall().SetVersion("8.3").SetUser("deploy"))
//	// or with custom extensions:
//	node.Run(php.NewInstall().SetVersion("8.3").SetUser("deploy").SetExtensions("cli fpm mysql"))
//
// Idempotency:
//   - Check() verifies that the php<version> binary exists AND the FPM pool is
//     configured for the requested user. When both are already correct, Run() is
//     a no-op.
type Install struct {
	*types.BaseSkill
}

// Compile-time assertion that Install implements types.RunnableInterface.
var _ types.RunnableInterface = (*Install)(nil)

// Check determines if PHP needs to be installed or reconfigured.
// Returns true if the php<version> binary is missing or the FPM pool is not
// configured for the requested user, false if everything is already in place.
func (s *Install) Check() (bool, error) {
	version := s.GetArg(ArgVersion)
	if version == "" {
		return false, fmt.Errorf("no version specified: set the %q argument", ArgVersion)
	}

	user := s.GetArg(ArgUser)
	if user == "" {
		return false, fmt.Errorf("no user specified: set the %q argument", ArgUser)
	}

	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if PHP is installed and configured")
		return true, nil
	}

	// Check if the php<version> binary exists.
	cmdCheckBinary := types.Command{
		Command:     fmt.Sprintf("command -v php%s", skills.ShellEscapeArg(version)),
		Description: "Check if php" + version + " binary exists",
		Required:    false,
	}
	_, err := ssh.Run(cfg, cmdCheckBinary)
	if err != nil {
		// Binary missing — needs install.
		return true, nil
	}

	// Check if the FPM pool is configured for the requested user.
	poolPath := fmt.Sprintf(DefaultFpmPoolPath, version)
	cmdCheckUser := types.Command{
		Command:     fmt.Sprintf("grep -q '^user = %s' %s", skills.ShellEscapeArg(user), skills.ShellEscapeArg(poolPath)),
		Description: "Check if FPM pool is configured for user " + user,
		Required:    false,
	}
	_, err = ssh.Run(cfg, cmdCheckUser)
	if err != nil {
		// Pool not configured for this user — needs reconfiguration.
		return true, nil
	}

	return false, nil // everything is already in place
}

// Run installs PHP with extensions and configures FPM.
// Changed is true when PHP was installed or reconfigured, false when everything
// was already in the desired state or validation fails.
//
// Result.Details contains:
//   - version: The PHP version that was installed
//   - user: The user FPM was configured for
//   - extensions: The extensions that were installed
//   - output: Full output from the apt-get install command
func (s *Install) Run() types.Result {
	version := s.GetArg(ArgVersion)
	if version == "" {
		return types.Result{
			Changed: false,
			Message: "No version specified",
			Error:   fmt.Errorf("no version specified: set the %q argument", ArgVersion),
		}
	}

	user := s.GetArg(ArgUser)
	if user == "" {
		return types.Result{
			Changed: false,
			Message: "No user specified",
			Error:   fmt.Errorf("no user specified: set the %q argument", ArgUser),
		}
	}

	extensions := s.GetArg(ArgExtensions)
	if extensions == "" {
		extensions = DefaultExtensions
	}

	cfg := s.GetNodeConfig()

	// Check if everything is already in place (idempotency).
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
			Message: fmt.Sprintf("PHP %s already installed and configured for user %s", version, user),
		}
	}

	// Check for dry-run mode.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would install PHP", "version", version, "user", user, "extensions", extensions)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would install PHP %s", version),
		}
	}

	// Run apt-get update first.
	cmdUpdate := types.Command{
		Command:     "apt-get update -y",
		Description: "Update package database",
	}
	cfg.GetLoggerOrDefault().Info("running apt update before PHP install")
	updateOutput, err := ssh.Run(cfg, cmdUpdate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "apt-get update failed",
			Error:   fmt.Errorf("apt-get update failed: %w\nOutput: %s", err, updateOutput),
		}
	}

	// Build the package list: php<version>-<ext> for each extension.
	parts := strings.Fields(extensions)
	for i, p := range parts {
		parts[i] = "php" + version + "-" + p
	}
	packageList := strings.Join(parts, "")

	// Escape each package name for shell safety.
	escapedParts := make([]string, len(parts))
	for i, p := range parts {
		escapedParts[i] = skills.ShellEscapeArg(p)
	}
	escapedPackages := strings.Join(escapedParts, " ")

	// See skills.DebianNonInteractive and skills.DpkgConfOptions for details.
	cmdInstallStr := ""
	cmdInstallStr += skills.DebianNonInteractive   // prevent interactive prompts
	cmdInstallStr += " apt-get install -y -- "     // install packages, auto-confirm, -- prevents option injection
	cmdInstallStr += escapedPackages               // escape each package name
	cmdInstallStr += skills.DpkgConfOptions        // keep local config, use maintainer default if unmodified

	cmdInstall := types.Command{
		Command:     cmdInstallStr,
		Description: "Install PHP packages: " + packageList,
	}

	cfg.GetLoggerOrDefault().Info("installing PHP packages", "packages", packageList)
	installOutput, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "PHP package installation failed",
			Error:   fmt.Errorf("apt-get install failed for %s: %w\nOutput: %s", packageList, err, installOutput),
		}
	}

	// Configure FPM pool: set user, group, listen.owner, listen.group, listen.mode.
	poolPath := fmt.Sprintf(DefaultFpmPoolPath, version)
	escapedPoolPath := skills.ShellEscapeArg(poolPath)
	escapedUser := skills.ShellEscapeArg(user)

	sedCommands := []string{
		fmt.Sprintf("sed -i 's/^user = .*/user = %s/' %s", escapedUser, escapedPoolPath),
		fmt.Sprintf("sed -i 's/^group = .*/group = %s/' %s", escapedUser, escapedPoolPath),
		fmt.Sprintf("sed -i 's/^listen.owner = .*/listen.owner = %s/' %s", escapedUser, escapedPoolPath),
		fmt.Sprintf("sed -i 's/^listen.group = .*/listen.group = %s/' %s", escapedUser, escapedPoolPath),
		fmt.Sprintf("sed -i 's/^listen.mode = .*/listen.mode = 0660/' %s", escapedPoolPath),
	}

	for _, sedCmd := range sedCommands {
		cmd := types.Command{
			Command:     sedCmd,
			Description: "Configure FPM pool: " + poolPath,
		}
		if _, err := ssh.Run(cfg, cmd); err != nil {
			return types.Result{
				Changed: false,
				Message: "FPM pool configuration failed",
				Error:   fmt.Errorf("failed to configure FPM pool %s: %w", poolPath, err),
			}
		}
	}

	// Restart and enable php<version>-fpm service.
	cmdRestart := types.Command{
		Command:     fmt.Sprintf("systemctl restart php%s-fpm && systemctl enable php%s-fpm", skills.ShellEscapeArg(version), skills.ShellEscapeArg(version)),
		Description: "Restart and enable php" + version + "-fpm",
	}
	if _, err := ssh.Run(cfg, cmdRestart); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart php" + version + "-fpm",
			Error:   fmt.Errorf("failed to restart php%s-fpm: %w", version, err),
		}
	}

	cfg.GetLoggerOrDefault().Info("PHP installed and configured", "version", version, "user", user)
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("PHP %s installed and configured for user %s", version, user),
		Details: map[string]string{
			"version":    version,
			"user":       user,
			"extensions": extensions,
			"output":     installOutput,
		},
	}
}

// SetArgs sets the arguments for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Install) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetVersion sets the PHP version (e.g. "8.3") and returns Install for chaining.
func (s *Install) SetVersion(version string) *Install {
	s.BaseSkill.SetArg(ArgVersion, version)
	return s
}

// SetUser sets the user to run PHP-FPM as and returns Install for chaining.
func (s *Install) SetUser(user string) *Install {
	s.BaseSkill.SetArg(ArgUser, user)
	return s
}

// SetExtensions sets the space-separated list of PHP extensions and returns Install for chaining.
func (s *Install) SetExtensions(extensions string) *Install {
	s.BaseSkill.SetArg(ArgExtensions, extensions)
	return s
}

// WithNodeConfig sets the node config and returns Install for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *Install) WithNodeConfig(cfg types.NodeConfig) *Install {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArg sets a single argument for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Install) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Install) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Install) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *Install) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewInstall creates a new php-install skill.
// Set the PHP version via SetVersion("8.3") and the FPM user via SetUser("deploy").
//
// Returns:
//
//	An Install skill configured with skills.IDPhpInstall identifier
//	and description "Install PHP with extensions and configure FPM".
//
// Example:
//
//	node.Run(php.NewInstall().SetVersion("8.3").SetUser("deploy"))
func NewInstall() *Install {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpInstall)
	pb.SetDescription("Install PHP with extensions and configure FPM")
	return &Install{BaseSkill: pb}
}
