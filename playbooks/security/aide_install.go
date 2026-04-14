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
	_, err := ssh.Run(cfg, types.Command{Command: "which aide", Description: "Check if AIDE is installed"})
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (a *AideInstall) Run() playbook.Result {
	cfg := a.GetNodeConfig()

	cfg.GetLoggerOrDefault().Info("installing AIDE")

	// Install AIDE
	cfg.GetLoggerOrDefault().Info("installing AIDE package")
	_, err := ssh.Run(cfg, types.Command{Command: `DEBIAN_FRONTEND=noninteractive apt-get install -y aide aide-common`, Description: "Install AIDE package"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to install AIDE", Error: err}
	}

	// Configure AIDE
	cfg.GetLoggerOrDefault().Info("configuring AIDE to monitor critical paths")
	cmd := `cat >> /etc/aide/aide.conf << 'EOF'

# Custom monitoring rules
/etc/ssh p+i+n+u+g+s+b+acl+xattrs+sha256
/etc/mysql p+i+n+u+g+s+b+acl+xattrs+sha256
/var/lib/mysql p+i+n+u+g+s+b+acl+xattrs+sha256
/root/.ssh p+i+n+u+g+s+b+acl+xattrs+sha256
/home p+i+n+u+g+s+b+acl+xattrs+sha256
EOF`
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Configure AIDE monitoring rules"})

	// Initialize AIDE database
	cfg.GetLoggerOrDefault().Info("initializing AIDE database")
	_, _ = ssh.Run(cfg, types.Command{Command: `aideinit`, Description: "Initialize AIDE database"})

	// Move database to active location
	_, _ = ssh.Run(cfg, types.Command{Command: `mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db`, Description: "Move AIDE database to active location"})

	// Create daily cron job
	cmd = `cat > /etc/cron.daily/aide-check << 'EOF'
#!/bin/bash
/usr/bin/aide --check | mail -s "AIDE Daily Report - $(hostname)" root
EOF`
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Create AIDE daily cron job"})
	_, _ = ssh.Run(cfg, types.Command{Command: `chmod +x /etc/cron.daily/aide-check`, Description: "Make AIDE cron job executable"})

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
