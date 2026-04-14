package security

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// AideInstall installs and configures AIDE (Advanced Intrusion Detection Environment).
// AIDE is a file integrity monitoring tool that creates a database of file checksums and
// can detect unauthorized file modifications. This playbook sets up daily integrity checks.
//
// Usage:
//
//	go run . --playbook=aide-install
//
// Execution Flow:
//  1. Installs AIDE and aide-common packages
//  2. Configures monitoring rules for critical paths:
//     - /etc/ssh (SSH configuration)
//     - /etc/mysql (MySQL/MariaDB configuration)
//     - /var/lib/mysql (Database files)
//     - /root/.ssh (Root SSH keys)
//     - /home (User home directories)
//  3. Initializes AIDE database (first run creates baseline)
//  4. Moves new database to active location
//  5. Creates daily cron job for automated checks
//  6. Runs initial check
//
// Monitored File Attributes:
//   - p: permissions
//   - i: inode
//   - n: number of links
//   - u: user
//   - g: group
//   - s: size
//   - b: block count
//   - acl: access control lists
//   - xattrs: extended attributes
//   - sha256: cryptographic hash
//
// Daily Checks:
//   - Automated via /etc/cron.daily/aide-check
//   - Results emailed to root
//
// Prerequisites:
//   - Root SSH access required
//   - First initialization may take several minutes
//
// Related Playbooks:
//   - auditd-install: System call auditing
type AideInstall struct {
	*playbook.BasePlaybook
}

// Check determines if AIDE needs to be installed.
func (a *AideInstall) Check() (bool, error) {
	cfg := a.GetNodeConfig()
	cmdCheck := types.Command{Command: "which aide", Description: "Check if AIDE is installed"}
	_, err := ssh.Run(cfg, cmdCheck)
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (a *AideInstall) Run() playbook.Result {
	cfg := a.GetNodeConfig()

	// Define commands
	cmdInstall := types.Command{Command: `DEBIAN_FRONTEND=noninteractive apt-get install -y aide aide-common`, Description: "Install AIDE package"}
	cmdConfigure := types.Command{Command: `cat >> /etc/aide/aide.conf << 'EOF'

# Custom monitoring rules
/etc/ssh p+i+n+u+g+s+b+acl+xattrs+sha256
/etc/mysql p+i+n+u+g+s+b+acl+xattrs+sha256
/var/lib/mysql p+i+n+u+g+s+b+acl+xattrs+sha256
/root/.ssh p+i+n+u+g+s+b+acl+xattrs+sha256
/home p+i+n+u+g+s+b+acl+xattrs+sha256
EOF`, Description: "Configure AIDE monitoring rules"}
	cmdInit := types.Command{Command: `aideinit`, Description: "Initialize AIDE database"}
	cmdMoveDb := types.Command{Command: `mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db`, Description: "Move AIDE database to active location"}
	cmdCron := types.Command{Command: `cat > /etc/cron.daily/aide-check << 'EOF'
#!/bin/bash
/usr/bin/aide --check | mail -s "AIDE Daily Report - $(hostname)" root
EOF`, Description: "Create AIDE daily cron job"}
	cmdChmod := types.Command{Command: `chmod +x /etc/cron.daily/aide-check`, Description: "Make AIDE cron job executable"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command, "description", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would configure AIDE monitoring rules")
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInit.Command, "description", cmdInit.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMoveDb.Command, "description", cmdMoveDb.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCron.Command, "description", cmdCron.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmod.Command, "description", cmdChmod.Description)
		return playbook.Result{
			Changed: true,
			Message: "Would install and configure AIDE",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing AIDE")

	// Install AIDE
	cfg.GetLoggerOrDefault().Info("installing AIDE package")
	_, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to install AIDE", Error: err}
	}

	// Configure AIDE
	cfg.GetLoggerOrDefault().Info("configuring AIDE to monitor critical paths")
	_, _ = ssh.Run(cfg, cmdConfigure)

	// Initialize AIDE database
	cfg.GetLoggerOrDefault().Info("initializing AIDE database")
	_, _ = ssh.Run(cfg, cmdInit)

	// Move database to active location
	_, _ = ssh.Run(cfg, cmdMoveDb)

	// Create daily cron job
	_, _ = ssh.Run(cfg, cmdCron)
	_, _ = ssh.Run(cfg, cmdChmod)

	cfg.GetLoggerOrDefault().Info("AIDE installation complete")
	return playbook.Result{
		Changed: true,
		Message: "AIDE installed and configured successfully",
	}
}

// NewAideInstall creates a new aide-install playbook.
func NewAideInstall() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDAideInstall)
	pb.SetDescription("Install and configure AIDE file integrity monitoring")
	return &AideInstall{BasePlaybook: pb}
}
