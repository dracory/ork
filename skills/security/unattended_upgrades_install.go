package security

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Paths written by this skill. Kept as constants so the printf and chmod
// commands reference the same value, avoiding drift between the two.
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
//  3. Writes /etc/apt/apt.conf.d/50unattended-upgrades to restrict upgrades
//     to security origins only
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

	// Define commands.
	// apt-listchanges is installed alongside unattended-upgrades because it is
	// the recommended companion that surfaces changelogs during upgrades; its
	// Debian default config is reasonable so no custom config is written for it.
	cmdInstallStr := ""
	cmdInstallStr += skills.DebianNonInteractive                               // prevent interactive prompts
	cmdInstallStr += " apt-get install -y unattended-upgrades apt-listchanges" // install packages, auto-confirm
	cmdInstallStr += skills.DpkgConfOptions                                    // keep local config, use maintainer default if unmodified

	cmdInstall := types.Command{Command: cmdInstallStr, Description: "Install unattended-upgrades package", Required: true}
	cmdAutoUpgrades := types.Command{
		Command:     fmt.Sprintf("printf '%%s' %s > %s", skills.ShellEscapeContent(autoUpgradesContent), skills.ShellEscapeArg(pathAutoUpgrades)),
		Description: "Write 20auto-upgrades config",
		Required:    true,
	}
	cmdUnattended := types.Command{
		Command:     fmt.Sprintf("printf '%%s' %s > %s", skills.ShellEscapeContent(unattendedUpgradesContent), skills.ShellEscapeArg(pathUnattendedUpgrades)),
		Description: "Write 50unattended-upgrades config",
		Required:    true,
	}
	cmdChmodAuto := types.Command{Command: "chmod 644 " + skills.ShellEscapeArg(pathAutoUpgrades), Description: "Set 20auto-upgrades permissions"}
	cmdChmodUnattended := types.Command{Command: "chmod 644 " + skills.ShellEscapeArg(pathUnattendedUpgrades), Description: "Set 50unattended-upgrades permissions"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command, "description", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would write 20auto-upgrades config")
		cfg.GetLoggerOrDefault().Info("dry-run: would write 50unattended-upgrades config")
		return types.Result{
			Changed: true,
			Message: "Would install and configure unattended-upgrades",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing unattended-upgrades")

	// Install unattended-upgrades
	cfg.GetLoggerOrDefault().Info("installing unattended-upgrades package")
	if _, err := ssh.Run(cfg, cmdInstall); err != nil {
		return types.Result{Changed: false, Message: "Failed to install unattended-upgrades", Error: err}
	}

	// Write 20auto-upgrades config
	cfg.GetLoggerOrDefault().Info("writing 20auto-upgrades config")
	if _, err := ssh.Run(cfg, cmdAutoUpgrades); err != nil {
		return types.Result{Changed: false, Message: "Failed to write 20auto-upgrades config", Error: err}
	}
	if _, err := ssh.Run(cfg, cmdChmodAuto); err != nil {
		cfg.GetLoggerOrDefault().Warn("failed to set 20auto-upgrades permissions", "error", err)
	}

	// Write 50unattended-upgrades config
	cfg.GetLoggerOrDefault().Info("writing 50unattended-upgrades config")
	if _, err := ssh.Run(cfg, cmdUnattended); err != nil {
		return types.Result{Changed: false, Message: "Failed to write 50unattended-upgrades config", Error: err}
	}
	if _, err := ssh.Run(cfg, cmdChmodUnattended); err != nil {
		cfg.GetLoggerOrDefault().Warn("failed to set 50unattended-upgrades permissions", "error", err)
	}

	cfg.GetLoggerOrDefault().Info("unattended-upgrades installation complete")
	return types.Result{
		Changed: true,
		Message: "Unattended-upgrades installed and configured (security updates only, no auto-reboot)",
	}
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
