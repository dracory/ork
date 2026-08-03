package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Paths written by this skill. Kept as constants so they are referenced from a
// single place, avoiding drift between the install, write, and validation
// commands.
const (
	pathAutoUpgrades       = "/etc/apt/apt.conf.d/20auto-upgrades"
	pathUnattendedUpgrades = "/etc/apt/apt.conf.d/50unattended-upgrades"
)

// autoUpgradesContent is written to /etc/apt/apt.conf.d/20auto-upgrades to
// enable automatic apt update and unattended upgrade runs.
// Kept as a constant so the content is shell-escaped at runtime instead of
// embedded in a fragile heredoc.
const autoUpgradesContent = `APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
APT::Periodic::Verbose "0";
`

// unattendedUpgradesContent is written to /etc/apt/apt.conf.d/50unattended-upgrades.
// Only security origins are allowed. Automatic reboots are intentionally
// disabled — reboots stay under a separate reboot playbook's control to
// avoid unattended downtime.
//
// Allowed-Origins (rather than Origins-Pattern) is used because it is the
// syntax that works identically on both Debian and Ubuntu via the
// ${distro_id} and ${distro_codename} macros, without requiring OS-specific
// origin patterns.
//
// Automatic-Reboot-Time and Automatic-Reboot-WithUsers are kept despite
// Automatic-Reboot being "false": they are inert today but self-document
// the intended behavior if an operator later flips Automatic-Reboot to "true".
const unattendedUpgradesContent = `Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
    "${distro_id}ESMApps:${distro_codename}-security";
    "${distro_id}ESM:${distro_codename}-infra-security";
};
Unattended-Upgrade::Package-Blacklist {
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::InstallOnShutdown "false";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Automatic-Reboot-Time "04:00";
Unattended-Upgrade::Automatic-Reboot-WithUsers "false";
Unattended-Upgrade::Mail "root";
Unattended-Upgrade::MailOnlyOnError "true";
`

// UnattendedUpgradesInstall installs and configures unattended-upgrades for
// automatic security updates. This keeps the server patched without requiring
// manual apt-upgrade runs for security fixes.
//
// Usage:
//
//	node.Run(security.NewUnattendedUpgradesInstall())
//
// Execution Flow:
//  1. Installs unattended-upgrades and apt-listchanges packages
//     (apt-listchanges surfaces changelogs during upgrades; its Debian
//     default config is reasonable so no custom config is written)
//  2. Writes /etc/apt/apt.conf.d/20auto-upgrades to enable automatic updates
//     (skipped if the file already has the correct content — idempotent)
//  3. Writes /etc/apt/apt.conf.d/50unattended-upgrades to restrict upgrades
//     to security origins only (skipped if content already matches)
//  4. Validates the configuration with `apt-config dump` to catch syntax
//     errors that would silently break security updates
//
// Configuration:
//   - Only security origins are allowed (no full distro upgrades)
//   - Automatic reboots are disabled (controlled by a separate reboot playbook)
//   - Unused kernel packages and dependencies are auto-removed
//   - Mail is sent to root only on errors
//
// Prerequisites:
//   - Root SSH access required
//
// Related Playbooks:
//   - reboot: Controlled server reboots
type UnattendedUpgradesInstall struct {
	*types.BaseSkill
}

// Compile-time assertion that UnattendedUpgradesInstall implements types.RunnableInterface.
var _ types.RunnableInterface = (*UnattendedUpgradesInstall)(nil)

// Check determines if unattended-upgrades needs to be installed.
// Returns true if the package is not yet installed.
func (u *UnattendedUpgradesInstall) Check() (bool, error) {
	cfg := u.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if unattended-upgrades is installed")
		return true, nil
	}

	// dpkg-query exits 0 if the package is recorded in the db, 1 if missing.
	// Required must be true so that ssh.Run propagates the non-zero exit code
	// instead of swallowing it (ssh.Run suppresses errors when Required is false).
	cmdCheck := types.Command{
		Command:     "dpkg-query -W -- unattended-upgrades 2>/dev/null",
		Description: "Check if unattended-upgrades is installed",
		Required:    true,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		// A non-zero exit code means the package is not installed — treat as
		// "needs install". A non-exit error (SSH connection failure, timeout,
		// auth error) is a real failure and is propagated so the caller can
		// distinguish "package missing" from "cannot reach the server".
		if ssh.IsExitError(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// Run executes the skill and returns detailed result.
func (u *UnattendedUpgradesInstall) Run() types.Result {
	cfg := u.GetNodeConfig()

	// Check if unattended-upgrades is already installed (idempotency).
	needsInstall, err := u.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check if unattended-upgrades is installed",
			Error:   err,
		}
	}
	if !needsInstall {
		return types.Result{
			Changed: false,
			Message: "Unattended-upgrades is already installed",
		}
	}

	// Define the install command.
	// apt-listchanges is installed alongside unattended-upgrades because it is
	// the recommended companion that surfaces changelogs during upgrades; its
	// Debian default config is reasonable so no custom config is written for it.
	cmdInstallStr := ""
	cmdInstallStr += skills.DebianNonInteractive                               // prevent interactive prompts
	cmdInstallStr += " apt-get install -y unattended-upgrades apt-listchanges" // install packages, auto-confirm
	cmdInstallStr += skills.DpkgConfOptions                                    // keep local config, use maintainer default if unmodified

	cmdInstall := types.Command{Command: cmdInstallStr, Description: "Install unattended-upgrades package", Required: true}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command, "description", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would write 20auto-upgrades config")
		cfg.GetLoggerOrDefault().Info("dry-run: would write 50unattended-upgrades config")
		cfg.GetLoggerOrDefault().Info("dry-run: would validate config with apt-config dump")
		return types.Result{
			Changed: true,
			Message: "Would install and configure unattended-upgrades",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing unattended-upgrades")

	// Step 1: Install unattended-upgrades package.
	cfg.GetLoggerOrDefault().Info("installing unattended-upgrades package")
	if _, err := ssh.Run(cfg, cmdInstall); err != nil {
		return types.Result{Changed: false, Message: "Failed to install unattended-upgrades", Error: err}
	}

	// Step 2: Write 20auto-upgrades config.
	// fs.FileCreate gives content-based idempotency for free: it reads the
	// existing file, compares content, and skips the write if they match
	// (same behavior as Ansible's template module).
	cfg.GetLoggerOrDefault().Info("writing 20auto-upgrades config")
	autoResult := types.RunSub(fs.NewFileCreate().
		SetPath(pathAutoUpgrades).
		SetContent(autoUpgradesContent).
		SetMode("644").
		SetOverwrite(true), cfg)
	if autoResult.Error != nil {
		return types.Result{Changed: false, Message: "Failed to write 20auto-upgrades config", Error: autoResult.Error}
	}

	// Step 3: Write 50unattended-upgrades config.
	cfg.GetLoggerOrDefault().Info("writing 50unattended-upgrades config")
	unattendedResult := types.RunSub(fs.NewFileCreate().
		SetPath(pathUnattendedUpgrades).
		SetContent(unattendedUpgradesContent).
		SetMode("644").
		SetOverwrite(true), cfg)
	if unattendedResult.Error != nil {
		return types.Result{Changed: false, Message: "Failed to write 50unattended-upgrades config", Error: unattendedResult.Error}
	}

	// Step 4: Validate the configuration with apt-config dump.
	// This catches syntax errors in the apt.conf.d files that would silently
	// break security updates. No Ansible role does this — it's a defensive
	// improvement.
	cfg.GetLoggerOrDefault().Info("validating unattended-upgrades configuration")
	if err := validateConfig(cfg); err != nil {
		return types.Result{Changed: true, Message: "Unattended-upgrades installed but config validation failed", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("unattended-upgrades installation complete")
	return types.Result{
		Changed: true,
		Message: "Unattended-upgrades installed and configured (security updates only, no auto-reboot)",
	}
}

// validateConfig runs `apt-config dump` to verify that apt recognizes the
// unattended-upgrades configuration. This catches syntax errors in the
// apt.conf.d files that would silently break security updates.
func validateConfig(cfg types.NodeConfig) error {
	cmdValidate := types.Command{
		Command:     "apt-config dump APT::Periodic::Unattended-Upgrade",
		Description: "Validate unattended-upgrades config",
		Required:    true,
	}
	output, err := ssh.Run(cfg, cmdValidate)
	if err != nil {
		return fmt.Errorf("apt-config dump failed: %w", err)
	}
	// apt-config dump prints the configured value, e.g.:
	//   APT::Periodic::Unattended-Upgrade "1";
	// If the line is absent or the value is "0", the config is not active.
	if !strings.Contains(output, `APT::Periodic::Unattended-Upgrade "1"`) {
		return fmt.Errorf("apt-config dump did not confirm Unattended-Upgrade is enabled, got: %q", strings.TrimSpace(output))
	}
	return nil
}

// SetArgs sets the arguments for unattended-upgrades installation.
// Returns UnattendedUpgradesInstall for fluent method chaining.
func (u *UnattendedUpgradesInstall) SetArgs(args map[string]string) types.RunnableInterface {
	u.BaseSkill.SetArgs(args)
	return u
}

// SetArg sets a single argument for unattended-upgrades installation.
// Returns UnattendedUpgradesInstall for fluent method chaining.
func (u *UnattendedUpgradesInstall) SetArg(key, value string) types.RunnableInterface {
	u.BaseSkill.SetArg(key, value)
	return u
}

// SetID sets the ID for unattended-upgrades installation.
// Returns UnattendedUpgradesInstall for fluent method chaining.
func (u *UnattendedUpgradesInstall) SetID(id string) types.RunnableInterface {
	u.BaseSkill.SetID(id)
	return u
}

// SetDescription sets the description for unattended-upgrades installation.
// Returns UnattendedUpgradesInstall for fluent method chaining.
func (u *UnattendedUpgradesInstall) SetDescription(description string) types.RunnableInterface {
	u.BaseSkill.SetDescription(description)
	return u
}

// SetTimeout sets the timeout for unattended-upgrades installation.
// Returns UnattendedUpgradesInstall for fluent method chaining.
func (u *UnattendedUpgradesInstall) SetTimeout(timeout time.Duration) types.RunnableInterface {
	u.BaseSkill.SetTimeout(timeout)
	return u
}

// NewUnattendedUpgradesInstall creates a new unattended-upgrades-install skill.
func NewUnattendedUpgradesInstall() *UnattendedUpgradesInstall {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUnattendedUpgradesInstall)
	pb.SetDescription("Install and configure unattended-upgrades for security updates")
	return &UnattendedUpgradesInstall{BaseSkill: pb}
}
