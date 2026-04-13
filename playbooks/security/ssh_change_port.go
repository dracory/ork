package security

import (
	"fmt"
	"log"
	"strconv"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
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
	*playbook.BasePlaybook
}

// Check determines if port change is needed.
func (s *SshChangePort) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (s *SshChangePort) Run() playbook.Result {
	cfg := s.GetConfig()
	newPort := s.GetArg(ArgPort)

	if newPort == "" {
		return playbook.Result{
			Changed: false,
			Message: "Port parameter is required",
			Error:   fmt.Errorf("use --arg=port=<port_number>"),
		}
	}

	// Validate port
	portNum, err := strconv.Atoi(newPort)
	if err != nil || portNum < 1024 || portNum > 65535 {
		return playbook.Result{
			Changed: false,
			Message: "Invalid port number",
			Error:   fmt.Errorf("port must be between 1024 and 65535"),
		}
	}

	log.Println("=== Changing SSH Port ===")
	log.Printf("New SSH port: %s", newPort)

	// Backup
	log.Println("Backing up current SSH configuration...")
	_, err = ssh.Run(cfg, `cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d_%H%M%S)`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to backup SSH config", Error: err}
	}

	// Update UFW if active
	ufwOutput, _ := ssh.Run(cfg, `ufw status | grep -q "Status: active" && echo "ACTIVE" || echo "INACTIVE"`)
	if ufwOutput == "ACTIVE" {
		cmd := fmt.Sprintf(`ufw allow %s/tcp comment 'SSH on custom port'`, newPort)
		_, _ = ssh.Run(cfg, cmd)
	}

	// Update SSH port
	cmd := fmt.Sprintf(`sed -i 's/^#*Port .*/Port %s/' /etc/ssh/sshd_config`, newPort)
	_, _ = ssh.Run(cfg, cmd)

	// Validate
	_, err = ssh.Run(cfg, `sshd -t`)
	if err != nil {
		_, _ = ssh.Run(cfg, `ls -t /etc/ssh/sshd_config.backup.* | head -1 | xargs -I {} cp {} /etc/ssh/sshd_config`)
		return playbook.Result{Changed: false, Message: "SSH configuration validation failed, backup restored", Error: err}
	}

	// Restart SSH
	_, err = ssh.Run(cfg, `systemctl restart sshd`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart SSH", Error: err}
	}

	log.Println("=== SSH Port Change Complete ===")
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("SSH port changed to %s", newPort),
		Details: map[string]string{
			"new_port": newPort,
		},
	}
}

// NewSshChangePort creates a new ssh-change-port playbook.
func NewSshChangePort() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSshChangePort)
	pb.SetDescription("Change the SSH port to reduce automated scanning")
	return &SshChangePort{BasePlaybook: pb}
}
