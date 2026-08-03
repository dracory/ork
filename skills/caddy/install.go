package caddy

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/apt"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Install installs Caddy via apt from the official Caddy repository.
// This replaces manual-download approaches with a package-managed installation
// that receives security updates via apt upgrade.
//
// Usage:
//
//	node.Run(caddy.NewInstall())
//
// Execution Flow:
//  1. Installs prerequisite packages for adding apt repositories
//     (debian-keyring, debian-archive-keyring, apt-transport-https, curl, gnupg)
//  2. Adds the Caddy GPG key to the keyring (official method, no deprecated apt-key)
//  3. Adds the Caddy apt repository to sources.list.d
//  4. Updates the apt cache to pick up the new repository
//  5. Installs the caddy package via apt
//  6. Creates the Caddy log directory (/var/log/caddy) owned by the caddy user
//  7. Verifies the caddy user exists (created by the apt package)
//
// The GPG key import uses --batch --yes to prevent gpg from trying to open
// /dev/tty, which fails over non-interactive SSH sessions with
// "gpg: cannot open '/dev/tty': No such device or address".
//
// Prerequisites:
//   - Root SSH access required
//   - Internet connectivity for package installation
//
// Related Skills:
//   - caddy.Restart: Upload a Caddyfile and reload the service
//   - caddy.Status: Show the Caddy systemd unit status
type Install struct {
	*types.BaseSkill
}

// Compile-time assertion that Install implements types.RunnableInterface.
var _ types.RunnableInterface = (*Install)(nil)

// Check determines if Caddy needs to be installed.
// Returns true if the caddy package is not yet installed.
func (i *Install) Check() (bool, error) {
	cfg := i.GetNodeConfig()
	cmdCheck := types.Command{
		Command:     "dpkg-query -W -- caddy 2>/dev/null",
		Description: "Check if caddy is installed",
		Required:    true,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if caddy is installed")
		return true, nil
	}

	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		return true, nil // package not installed
	}
	return false, nil
}

// Run executes the Caddy installation via apt.
// Changed is true when the installation completes successfully.
func (i *Install) Run() types.Result {
	cfg := i.GetNodeConfig()

	// Check if Caddy is already installed (idempotency)
	needsInstall, err := i.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check if Caddy is installed",
			Error:   err,
		}
	}
	if !needsInstall {
		return types.Result{
			Changed: false,
			Message: "Caddy is already installed",
		}
	}

	// Step 1: Install prerequisite packages for adding apt repositories.
	prereqResult := types.RunSub(apt.NewPkgInstall().SetPackages(
		"debian-keyring",
		"debian-archive-keyring",
		"apt-transport-https",
		"curl",
		"gnupg",
	), cfg)
	if prereqResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to install prerequisite packages",
			Error:   prereqResult.Error,
		}
	}

	// Step 2: Add Caddy GPG key to keyring (official method, no deprecated apt-key).
	// --batch --yes prevents gpg from trying to open /dev/tty for interactive
	// prompts, which fails over non-interactive SSH sessions.
	cmdAddKey := types.Command{
		Command:     "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg",
		Description: "Add Caddy GPG key to keyring",
		Required:    true,
	}
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAddKey.Command)
	} else {
		output, err := ssh.Run(cfg, cmdAddKey)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to add Caddy GPG key",
				Error:   fmt.Errorf("%w: %s", err, output),
			}
		}
	}

	// Step 3: Add Caddy apt repository.
	cmdAddRepo := types.Command{
		Command:     "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list",
		Description: "Add Caddy apt repository",
		Required:    true,
	}
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAddRepo.Command)
	} else {
		output, err := ssh.Run(cfg, cmdAddRepo)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to add Caddy apt repository",
				Error:   fmt.Errorf("%w: %s", err, output),
			}
		}
	}

	// Step 3b: Ensure keyring and sources list are world-readable.
	// apt runs repository checks as the _apt user, not root — a restrictive
	// umask (e.g. 027) can create these files without o+r, causing GPG
	// signature verification failures during apt update.
	cmdChmodKeyring := types.Command{
		Command:     "chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg /etc/apt/sources.list.d/caddy-stable.list",
		Description: "Ensure Caddy keyring and sources list are world-readable",
		Required:    true,
	}
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmodKeyring.Command)
	} else {
		output, err := ssh.Run(cfg, cmdChmodKeyring)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to set permissions on Caddy keyring/sources list",
				Error:   fmt.Errorf("%w: %s", err, output),
			}
		}
	}

	// Step 4: Update apt cache to pick up the new repository.
	updateResult := types.RunSub(apt.NewPkgUpdate(), cfg)
	if updateResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to update apt cache after adding Caddy repo",
			Error:   updateResult.Error,
		}
	}

	// Step 5: Install Caddy via apt.
	installResult := types.RunSub(apt.NewPkgInstall().SetPackages("caddy"), cfg)
	if installResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to install Caddy",
			Error:   installResult.Error,
		}
	}

	// Step 6: Create Caddy log directory for structured JSON access logs.
	logDirResult := types.RunSub(fs.NewDirCreate().
		SetPath(DefaultCaddyLogDir).
		SetOwner(DefaultCaddyUser+":"+DefaultCaddyUser).
		SetMode("755"), cfg)
	if logDirResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create Caddy log directory",
			Error:   logDirResult.Error,
		}
	}

	// Step 7: Ensure Caddy service is enabled and started.
	// The apt package postinst should do this, but we make it explicit to
	// match Ansible community roles (paultibbetts.caddy, maxhoesel.caddy).
	enableResult := types.RunSub(systemctl.NewEnable().
		SetService(DefaultCaddyService).
		SetStart(true), cfg)
	if enableResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable and start Caddy service",
			Error:   enableResult.Error,
		}
	}

	// Step 8: Verify the caddy user exists (created by the apt package).
	// Non-required: a failure here is informational, not fatal.
	cmdCheckUser := types.Command{
		Command:     "id " + skills.ShellEscapeArg(DefaultCaddyUser),
		Description: "Verify caddy user exists",
		Required:    false,
	}
	if !cfg.IsDryRunMode {
		ssh.Run(cfg, cmdCheckUser)
	}

	return types.Result{
		Changed: true,
		Message: "Caddy installed successfully via apt",
	}
}

// SetArgs sets the arguments for the Caddy installation.
// Returns Install for fluent method chaining.
func (i *Install) SetArgs(args map[string]string) types.RunnableInterface {
	i.BaseSkill.SetArgs(args)
	return i
}

// SetArg sets a single argument for the Caddy installation.
// Returns Install for fluent method chaining.
func (i *Install) SetArg(key, value string) types.RunnableInterface {
	i.BaseSkill.SetArg(key, value)
	return i
}

// SetID sets the ID for the Caddy installation.
// Returns Install for fluent method chaining.
func (i *Install) SetID(id string) types.RunnableInterface {
	i.BaseSkill.SetID(id)
	return i
}

// SetDescription sets the description for the Caddy installation.
// Returns Install for fluent method chaining.
func (i *Install) SetDescription(description string) types.RunnableInterface {
	i.BaseSkill.SetDescription(description)
	return i
}

// SetTimeout sets the timeout for the Caddy installation.
// Returns Install for fluent method chaining.
func (i *Install) SetTimeout(timeout time.Duration) types.RunnableInterface {
	i.BaseSkill.SetTimeout(timeout)
	return i
}

// NewInstall creates a new caddy-install skill.
//
// Returns an Install skill configured with skills.IDCaddyInstall identifier and
// description "Install Caddy web server via apt from official repo".
//
// Example:
//
//	node.Run(caddy.NewInstall())
func NewInstall() *Install {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDCaddyInstall)
	pb.SetDescription("Install Caddy web server via apt from official repo")
	return &Install{BaseSkill: pb}
}
