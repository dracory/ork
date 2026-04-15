package security

import (
	"fmt"
	"strconv"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Argument key constants for SSH port change.
const (
	// ArgPort specifies the new SSH port number (1024-65535)
	ArgPort = "port"
)

// SshChangePort changes the SSH port to reduce automated scanning.
// For publicly exposed servers, using a non-standard SSH port significantly
// reduces bot traffic and automated brute force attempts.
//
// Usage:
//
//	go run . --playbook=ssh-change-port --arg=port=2222
//
// Execution Flow:
//  1. Validates new port number (1024-65535)
//  2. Backs up current SSH configuration
//  3. Updates SSH port in sshd_config
//  4. Updates UFW firewall rules (if UFW is active)
//  5. Validates SSH configuration
//  6. Restarts SSH service
//
// Args:
//   - port: New SSH port number (required, 1024-65535)
//
// IMPORTANT:
//   - After running this, update your SSH client to use the new port
//   - Ensure firewall allows the new port before running
//   - Keep a backup SSH session open until verified
//
// Prerequisites:
//   - Root SSH access on current port
//   - Backup access method (console) in case of failure
//
// Related Playbooks:
//   - ssh-harden: Disable password auth, root login
//   - ufw-install: Configure firewall
type SshChangePort struct {
	*skills.BaseSkill
}

// Check determines if port change is needed.
func (s *SshChangePort) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (s *SshChangePort) Run() types.Result {
	cfg := s.GetNodeConfig()
	newPort := s.GetArg(ArgPort)

	if newPort == "" {
		return types.Result{
			Changed: false,
			Message: "Port parameter is required",
			Error:   fmt.Errorf("use --arg=port=<port_number>"),
		}
	}

	// Validate port
	portNum, err := strconv.Atoi(newPort)
	if err != nil || portNum < 1024 || portNum > 65535 {
		return types.Result{
			Changed: false,
			Message: "Invalid port number",
			Error:   fmt.Errorf("port must be between 1024 and 65535"),
		}
	}

	cfg.GetLoggerOrDefault().Info("changing SSH port", "port", newPort)

	// Define commands
	cmdBackup := types.Command{Command: `cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d_%H%M%S)`, Description: "Backup SSH config"}
	cmdCheckUfw := types.Command{Command: `ufw status | grep -q "Status: active" && echo "ACTIVE" || echo "INACTIVE"`, Description: "Check UFW status"}
	cmdAllowPort := types.Command{Command: fmt.Sprintf(`ufw allow %s/tcp comment 'SSH on custom port'`, newPort), Description: "Allow SSH on custom port in UFW"}
	cmdUpdatePort := types.Command{Command: fmt.Sprintf(`sed -i 's/^#*Port .*/Port %s/' /etc/ssh/sshd_config`, newPort), Description: "Update SSH port in config"}
	cmdValidate := types.Command{Command: `sshd -t`, Description: "Validate SSH config"}
	cmdRestore := types.Command{Command: `ls -t /etc/ssh/sshd_config.backup.* | head -1 | xargs -I {} cp {} /etc/ssh/sshd_config`, Description: "Restore SSH config backup"}
	cmdRestart := types.Command{Command: `systemctl restart sshd`, Description: "Restart SSH service"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCheckUfw.Command, "description", cmdCheckUfw.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAllowPort.Command, "description", cmdAllowPort.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdUpdatePort.Command, "description", cmdUpdatePort.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdValidate.Command, "description", cmdValidate.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would change SSH port to %s", newPort),
		}
	}

	// Backup
	cfg.GetLoggerOrDefault().Info("backing up SSH configuration")
	_, err = ssh.Run(cfg, cmdBackup)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to backup SSH config", Error: err}
	}

	// Update UFW if active
	ufwOutput, _ := ssh.Run(cfg, cmdCheckUfw)
	if ufwOutput == "ACTIVE" {
		_, _ = ssh.Run(cfg, cmdAllowPort)
	}

	// Update SSH port
	_, _ = ssh.Run(cfg, cmdUpdatePort)

	// Validate
	_, err = ssh.Run(cfg, cmdValidate)
	if err != nil {
		_, _ = ssh.Run(cfg, cmdRestore)
		return types.Result{Changed: false, Message: "SSH configuration validation failed, backup restored", Error: err}
	}

	// Restart SSH
	_, err = ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to restart SSH", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("SSH port change complete")
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("SSH port changed to %s", newPort),
		Details: map[string]string{
			"new_port": newPort,
		},
	}
}

// NewSshChangePort creates a new ssh-change-port skill.
func NewSshChangePort() types.SkillInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDSshChangePort)
	pb.SetDescription("Change the SSH port to reduce automated scanning")
	return &SshChangePort{BaseSkill: pb}
}
