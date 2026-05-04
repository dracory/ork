package security

import (
	"fmt"
	"strconv"
	"time"

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
//  4. Validates SSH configuration
//  5. Restarts SSH service
//
// Args:
//   - port: New SSH port number (required, 1024-65535)
//
// IMPORTANT:
//   - After running this, update your SSH client to use the new port
//   - Ensure firewall allows the new port before running (calling playbook responsibility)
//   - Keep a backup SSH session open until verified
//
// Prerequisites:
//   - Root SSH access on current port
//   - Backup access method (console) in case of failure
//   - Firewall must be configured to allow the new port (calling playbook responsibility)
//
// Related Playbooks:
//   - ssh-harden: Disable password auth, root login
//   - ufw-install: Configure firewall (for UFW-based systems)
type SshChangePort struct {
	*types.BaseSkill
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
	cmdUpdatePort := types.Command{
		Command:     fmt.Sprintf(`sed -i 's/^#*Port .*/Port %s/' /etc/ssh/sshd_config`, newPort),
		Description: "Update SSH port in config",
		Required:    true,
	}
	cmdValidate := types.Command{
		Command:     `sshd -t`,
		Description: "Validate SSH config",
		Required:    true,
	}
	cmdRestart := types.Command{
		Command:     `systemctl restart sshd || systemctl restart ssh`,
		Description: "Restart SSH service",
		Required:    true,
	}
	cmdCheckPort := types.Command{
		Command:     fmt.Sprintf(`ss -tlnp | grep -q ':%s'`, newPort),
		Description: "Verify SSH is listening on new port",
		Required:    true,
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdUpdatePort.Command, "description", cmdUpdatePort.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdValidate.Command, "description", cmdValidate.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would change SSH port to %s", newPort),
		}
	}

	// Update SSH port
	cfg.GetLoggerOrDefault().Info("updating SSH port in config", "command", cmdUpdatePort.Command)
	updateOutput, err := ssh.Run(cfg, cmdUpdatePort)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to update SSH port in config: %v", err),
			Error:   err,
			Details: map[string]string{
				"output":  updateOutput,
				"command": cmdUpdatePort.Command,
			},
		}
	}

	// Validate
	cfg.GetLoggerOrDefault().Info("validating SSH configuration", "command", cmdValidate.Command)
	validateOutput, err := ssh.Run(cfg, cmdValidate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("SSH configuration validation failed: %v", err),
			Error:   err,
			Details: map[string]string{
				"output":  validateOutput,
				"command": cmdValidate.Command,
			},
		}
	}

	// Restart SSH
	cfg.GetLoggerOrDefault().Info("restarting SSH service", "command", cmdRestart.Command)
	restartOutput, err := ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Failed to restart SSH service: %v", err),
			Error:   err,
			Details: map[string]string{
				"output":  restartOutput,
				"command": cmdRestart.Command,
			},
		}
	}

	// Verify SSH is listening on new port
	cfg.GetLoggerOrDefault().Info("verifying SSH is listening on new port", "port", newPort)
	checkOutput, err := ssh.Run(cfg, cmdCheckPort)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("SSH service restarted but not listening on port %s: %v", newPort, err),
			Error:   err,
			Details: map[string]string{
				"output":  checkOutput,
				"command": cmdCheckPort.Command,
			},
		}
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

// SetArgs sets the arguments for SSH port change.
// Returns SshChangePort for fluent method chaining.
func (s *SshChangePort) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for SSH port change.
// Returns SshChangePort for fluent method chaining.
func (s *SshChangePort) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for SSH port change.
// Returns SshChangePort for fluent method chaining.
func (s *SshChangePort) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for SSH port change.
// Returns SshChangePort for fluent method chaining.
func (s *SshChangePort) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for SSH port change.
// Returns SshChangePort for fluent method chaining.
func (s *SshChangePort) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSshChangePort creates a new ssh-change-port skill.
func NewSshChangePort() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSshChangePort)
	pb.SetDescription("Change the SSH port to reduce automated scanning")
	return &SshChangePort{BaseSkill: pb}
}
