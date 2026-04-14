package security

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
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
//   - ssh-config-path: SSH configuration file path (default: /etc/ssh/sshd_config)
//   - max-auth-tries: Maximum authentication attempts (default: 3)
//   - client-alive-interval: Client alive interval in seconds (default: 300)
//   - client-alive-count-max: Client alive count max (default: 2)
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
	cfg := s.GetNodeConfig()
	nonRootUser := s.GetArg(ArgNonRootUser)
	if nonRootUser == "" {
		nonRootUser = DefaultNonRootUser
	}

	sshConfigPath := s.GetArg(ArgSSHConfigPath)
	if sshConfigPath == "" {
		sshConfigPath = DefaultSSHConfigPath
	}

	maxAuthTries := s.GetArg(ArgMaxAuthTries)
	if maxAuthTries == "" {
		maxAuthTries = DefaultMaxAuthTries
	}

	clientAliveInterval := s.GetArg(ArgClientAliveInterval)
	if clientAliveInterval == "" {
		clientAliveInterval = DefaultClientAliveInterval
	}

	clientAliveCountMax := s.GetArg(ArgClientAliveCountMax)
	if clientAliveCountMax == "" {
		clientAliveCountMax = DefaultClientAliveCountMax
	}

	cfg.GetLoggerOrDefault().Info("SSH security hardening started")

	// Define commands
	cmdBackup := fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d)`, sshConfigPath, sshConfigPath)
	cmdVerifyUser := fmt.Sprintf(`id %s >/dev/null 2>&1 && sudo -l -U %s >/dev/null 2>&1 && echo "OK" || echo "FAIL"`, nonRootUser, nonRootUser)
	cmdValidate := fmt.Sprintf(`sshd -t -f %s`, sshConfigPath)
	cmdRestore := fmt.Sprintf(`cp %s.backup.$(date +%%Y%%m%%d) %s`, sshConfigPath, sshConfigPath)
	cmdRestart := "systemctl restart sshd"

	// Apply security settings
	settings := []struct {
		name string
		cmd  string
	}{
		{"Disable root login", fmt.Sprintf(`sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' %s`, sshConfigPath)},
		{"Disable password auth", fmt.Sprintf(`sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' %s`, sshConfigPath)},
		{"Enable pubkey auth", fmt.Sprintf(`sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' %s`, sshConfigPath)},
		{"Disable empty passwords", fmt.Sprintf(`sed -i 's/^#*PermitEmptyPasswords.*/PermitEmptyPasswords no/' %s`, sshConfigPath)},
		{"Set max auth tries", fmt.Sprintf(`grep -q "^MaxAuthTries" %s && sed -i 's/^MaxAuthTries.*/MaxAuthTries %s/' %s || echo "MaxAuthTries %s" >> %s`, sshConfigPath, maxAuthTries, sshConfigPath, maxAuthTries, sshConfigPath)},
		{"Disable X11 forwarding", fmt.Sprintf(`sed -i 's/^#*X11Forwarding.*/X11Forwarding no/' %s`, sshConfigPath)},
		{"Set client alive interval", fmt.Sprintf(`grep -q "^ClientAliveInterval" %s && sed -i 's/^ClientAliveInterval.*/ClientAliveInterval %s/' %s || echo "ClientAliveInterval %s" >> %s`, sshConfigPath, clientAliveInterval, sshConfigPath, clientAliveInterval, sshConfigPath)},
		{"Set client alive count", fmt.Sprintf(`grep -q "^ClientAliveCountMax" %s && sed -i 's/^ClientAliveCountMax.*/ClientAliveCountMax %s/' %s || echo "ClientAliveCountMax %s" >> %s`, sshConfigPath, clientAliveCountMax, sshConfigPath, clientAliveCountMax, sshConfigPath)},
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdVerifyUser)
		for _, setting := range settings {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "setting", setting.name, "cmd", setting.cmd)
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdValidate)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart)
		return playbook.Result{
			Changed: true,
			Message: "Would harden SSH security configuration",
		}
	}

	// Step 1: Backup
	cfg.GetLoggerOrDefault().Info("backing up SSH configuration")
	_, err := ssh.Run(cfg, types.Command{Command: cmdBackup, Description: "Backup SSH config"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to backup SSH config", Error: err}
	}

	// Step 2: Verify non-root user
	cfg.GetLoggerOrDefault().Info("verifying non-root user exists")
	output, err := ssh.Run(cfg, types.Command{Command: cmdVerifyUser, Description: "Verify non-root user exists"})
	_ = output
	if err != nil || !strings.Contains(output, "OK") {
		return playbook.Result{
			Changed: false,
			Message: "Non-root user not configured properly",
			Error:   fmt.Errorf("user '%s' doesn't exist or lacks sudo privileges", nonRootUser),
		}
	}

	for _, setting := range settings {
		cfg.GetLoggerOrDefault().Info("applying SSH setting", "setting", setting.name)
		_, _ = ssh.Run(cfg, types.Command{Command: setting.cmd, Description: "Apply SSH setting: " + setting.name})
	}

	// Validate configuration
	cfg.GetLoggerOrDefault().Info("validating SSH configuration")
	_, err = ssh.Run(cfg, types.Command{Command: cmdValidate, Description: "Validate SSH config"})
	if err != nil {
		// Restore backup
		_, _ = ssh.Run(cfg, types.Command{Command: cmdRestore, Description: "Restore SSH config backup"})
		return playbook.Result{
			Changed: false,
			Message: "SSH configuration validation failed, backup restored",
			Error:   err,
		}
	}

	// Restart SSH
	cfg.GetLoggerOrDefault().Info("restarting SSH service")
	_, err = ssh.Run(cfg, types.Command{Command: cmdRestart, Description: "Restart SSH service"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart SSH", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("SSH hardening complete")
	return playbook.Result{
		Changed: true,
		Message: "SSH security hardening applied successfully",
		Details: map[string]string{
			"backup":                fmt.Sprintf("%s.backup.<date>", sshConfigPath),
			"non-root-user":         nonRootUser,
			"ssh-config-path":       sshConfigPath,
			"max-auth-tries":        maxAuthTries,
			"client-alive-interval": clientAliveInterval,
			"client-alive-count":    clientAliveCountMax,
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
