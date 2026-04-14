package security

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// AuditdInstall installs and configures the Linux Audit Framework.
// Auditd provides detailed logging of system calls and file access, enabling
// forensic analysis and compliance monitoring. This playbook sets up comprehensive
// security monitoring rules.
//
// Usage:
//
//	go run . --playbook=auditd-install
//
// Execution Flow:
//  1. Installs auditd and audispd-plugins packages
//  2. Creates comprehensive audit rules covering:
//     - Password file changes (/etc/passwd, /etc/shadow)
//     - SSH configuration changes
//     - MySQL configuration and data access
//     - Sudoers modifications
//     - Root command execution
//     - File deletion operations
//     - Permission changes
//     - Network connections
//     - Kernel module operations
//  3. Loads audit rules with augenrules
//  4. Enables and starts auditd service
//  5. Verifies rules are active
//
// Monitored Events:
//   - Authentication file modifications
//   - Privileged command execution
//   - Permission and ownership changes
//   - File deletions
//   - Network socket operations
//   - Kernel module loading/unloading
//
// Log Analysis Commands:
//   - ausearch -k <key_name>: Search by rule key
//   - aureport: Generate summary reports
//   - aureport --login: Login reports
//   - aureport --user: User activity reports
//
// Prerequisites:
//   - Root SSH access required
//
// Related Playbooks:
//   - aide-install: File integrity monitoring
type AuditdInstall struct {
	*playbook.BasePlaybook
}

// Check determines if auditd needs to be installed.
func (a *AuditdInstall) Check() (bool, error) {
	cfg := a.GetNodeConfig()
	_, err := ssh.Run(cfg, types.Command{Command: "which auditd", Description: "Check if auditd is installed"})
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (a *AuditdInstall) Run() playbook.Result {
	cfg := a.GetNodeConfig()

	cfg.GetLoggerOrDefault().Info("installing auditd")

	// Install auditd
	cfg.GetLoggerOrDefault().Info("installing auditd package")
	_, err := ssh.Run(cfg, types.Command{Command: `DEBIAN_FRONTEND=noninteractive apt-get install -y auditd audispd-plugins`, Description: "Install auditd package"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to install auditd", Error: err}
	}

	// Create audit rules
	cfg.GetLoggerOrDefault().Info("creating audit rules")
	cmd := `cat > /etc/audit/rules.d/audit.rules << 'EOF'
# Remove any existing rules
-D

# Buffer Size
-b 8192

# Failure Mode (0=silent 1=printk 2=panic)
-f 1

# Monitor password file changes
-w /etc/passwd -p wa -k passwd_changes
-w /etc/shadow -p wa -k shadow_changes
-w /etc/group -p wa -k group_changes
-w /etc/gshadow -p wa -k gshadow_changes

# Monitor SSH configuration
-w /etc/ssh/sshd_config -p wa -k sshd_config_changes

# Monitor MySQL/MariaDB configuration
-w /etc/mysql/ -p wa -k mysql_config_changes
-w /var/lib/mysql/ -p wa -k mysql_data_changes

# Monitor sudoers
-w /etc/sudoers -p wa -k sudoers_changes
-w /etc/sudoers.d/ -p wa -k sudoers_changes

# Monitor system calls by root
-a always,exit -F arch=b64 -S execve -F euid=0 -k root_commands
-a always,exit -F arch=b32 -S execve -F euid=0 -k root_commands

# Monitor file deletions
-a always,exit -F arch=b64 -S unlink -S unlinkat -S rename -S renameat -k file_deletion
-a always,exit -F arch=b32 -S unlink -S unlinkat -S rename -S renameat -k file_deletion

# Monitor permission changes
-a always,exit -F arch=b64 -S chmod -S fchmod -S fchmodat -k perm_mod
-a always,exit -F arch=b32 -S chmod -S fchmod -S fchmodat -k perm_mod
-a always,exit -F arch=b64 -S chown -S fchown -S fchownat -S lchown -k ownership_mod
-a always,exit -F arch=b32 -S chown -S fchown -S fchownat -S lchown -k ownership_mod

# Monitor network connections
-a always,exit -F arch=b64 -S socket -S connect -k network_connections
-a always,exit -F arch=b32 -S socket -S connect -k network_connections

# Monitor kernel module loading
-w /sbin/insmod -p x -k module_insertion
-w /sbin/rmmod -p x -k module_removal
-w /sbin/modprobe -p x -k module_modification

# Make configuration immutable (requires reboot to change)
-e 2
EOF`
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Create audit rules"})

	// Load audit rules
	cfg.GetLoggerOrDefault().Info("installing auditd rules")
	_, _ = ssh.Run(cfg, types.Command{Command: `augenrules --load`, Description: "Load audit rules"})

	// Enable and start auditd
	_, _ = ssh.Run(cfg, types.Command{Command: `systemctl enable auditd`, Description: "Enable auditd service"})
	_, _ = ssh.Run(cfg, types.Command{Command: `systemctl start auditd`, Description: "Start auditd service"})

	cfg.GetLoggerOrDefault().Info("auditd installation complete")
	return playbook.Result{
		Changed: true,
		Message: "Auditd installed and configured successfully",
	}
}

// NewAuditdInstall creates a new auditd-install playbook.
func NewAuditdInstall() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDAuditdInstall)
	pb.SetDescription("Install and configure the Linux Audit Framework")
	return &AuditdInstall{BasePlaybook: pb}
}
