package security

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SshHarden applies security hardening to SSH server configuration.
// This playbook modifies /etc/ssh/sshd_config to disable insecure authentication
// methods and enforce secure defaults. It backs up the original configuration before
// making changes and validates the new configuration before applying.
//
// Usage:
//
//	go run . --playbook=ssh-harden [--arg=non-root-user=<username>]
//
// Security Changes Applied:
//   - Disable root login (PermitRootLogin no)
//   - Disable password authentication (PasswordAuthentication no)
//   - Enable public key authentication (PubkeyAuthentication yes)
//   - Disable empty passwords (PermitEmptyPasswords no)
//   - Set max authentication attempts to 3 (MaxAuthTries 3)
//   - Disable X11 forwarding (X11Forwarding no)
//   - Set client alive interval to 300 seconds
//   - Set client alive count max to 2
//
// Args:
//   - non-root-user: Username to verify exists before disabling root login (default: "deploy")
//
// Execution Flow:
//  1. Backs up current SSH configuration with timestamp
//  2. Verifies non-root user exists with sudo privileges
//  3. Applies security settings using sed commands
//  4. Validates configuration with sshd -t
//  5. Restarts SSH service if validation passes
//  6. Restores backup if validation fails
//
// WARNING:
//   - After running this, you MUST use SSH key authentication
//   - Root login will be disabled - ensure non-root user exists
//   - Create a non-root user first with user-create playbook
//
// Prerequisites:
//   - Root SSH access required
//   - Ensure SSH key authentication is working before running
//   - Create non-root user with sudo access first
//
// Related Playbooks:
//   - user-create: Create non-root user before disabling root login
type SshHarden struct {
	*playbook.BasePlaybook
}

// Check always returns true since we want to verify and apply security settings.
func (s *SshHarden) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (s *SshHarden) Run() playbook.Result {
	cfg := s.GetConfig()
	nonRootUser := s.GetArg(ArgNonRootUser)
	if nonRootUser == "" {
		nonRootUser = "deploy" // Default non-root user name
	}

	log.Println("=== SSH Security Hardening ===")

	// Step 1: Backup
	log.Println("Step 1: Backing up current SSH configuration...")
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d)`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to backup SSH config", Error: err}
	}

	// Step 2: Verify non-root user
	log.Println("Step 2: Verifying non-root user exists...")
	cmd := fmt.Sprintf(`id %s >/dev/null 2>&1 && sudo -l -U %s >/dev/null 2>&1 && echo "OK" || echo "FAIL"`, nonRootUser, nonRootUser)
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil || !strings.Contains(output, "OK") {
		return playbook.Result{
			Changed: false,
			Message: "Non-root user not configured properly",
			Error:   fmt.Errorf("user '%s' doesn't exist or lacks sudo privileges", nonRootUser),
		}
	}

	// Apply security settings
	settings := []struct {
		name string
		cmd  string
	}{
		{"Disable root login", `sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config`},
		{"Disable password auth", `sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config`},
		{"Enable pubkey auth", `sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config`},
		{"Disable empty passwords", `sed -i 's/^#*PermitEmptyPasswords.*/PermitEmptyPasswords no/' /etc/ssh/sshd_config`},
		{"Set max auth tries", `grep -q "^MaxAuthTries" /etc/ssh/sshd_config && sed -i 's/^MaxAuthTries.*/MaxAuthTries 3/' /etc/ssh/sshd_config || echo "MaxAuthTries 3" >> /etc/ssh/sshd_config`},
		{"Disable X11 forwarding", `sed -i 's/^#*X11Forwarding.*/X11Forwarding no/' /etc/ssh/sshd_config`},
		{"Set client alive interval", `grep -q "^ClientAliveInterval" /etc/ssh/sshd_config && sed -i 's/^ClientAliveInterval.*/ClientAliveInterval 300/' /etc/ssh/sshd_config || echo "ClientAliveInterval 300" >> /etc/ssh/sshd_config`},
		{"Set client alive count", `grep -q "^ClientAliveCountMax" /etc/ssh/sshd_config && sed -i 's/^ClientAliveCountMax.*/ClientAliveCountMax 2/' /etc/ssh/sshd_config || echo "ClientAliveCountMax 2" >> /etc/ssh/sshd_config`},
	}

	for _, setting := range settings {
		log.Printf("Applying: %s...", setting.name)
		_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, setting.cmd)
	}

	// Validate configuration
	log.Println("Validating SSH configuration...")
	_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `sshd -t`)
	if err != nil {
		// Restore backup
		_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `cp /etc/ssh/sshd_config.backup.$(date +%Y%m%d) /etc/ssh/sshd_config`)
		return playbook.Result{
			Changed: false,
			Message: "SSH configuration validation failed, backup restored",
			Error:   err,
		}
	}

	// Restart SSH
	log.Println("Restarting SSH service...")
	_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `systemctl restart sshd`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart SSH", Error: err}
	}

	log.Println("=== SSH Hardening Complete ===")
	return playbook.Result{
		Changed: true,
		Message: "SSH security hardening applied successfully",
		Details: map[string]string{
			"backup":      "/etc/ssh/sshd_config.backup.<date>",
			"non-root-user": nonRootUser,
		},
	}
}

// NewSshHarden creates a new ssh-harden playbook.
func NewSshHarden() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSshHarden)
	pb.SetDescription("Apply security hardening to SSH server configuration")
	return &SshHarden{BasePlaybook: pb}
}
