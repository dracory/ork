package fail2ban

import (
	"fmt"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// NewFail2banInstall creates a new fail2ban-install skill.
func NewFail2banInstall() types.RunnableInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDFail2banInstall)
	pb.SetDescription("Install and enable fail2ban intrusion prevention system")
	return &Fail2banInstall{BaseSkill: pb}
}

// Fail2banInstall installs and enables the fail2ban intrusion prevention system.
// Fail2ban monitors log files for suspicious activity (like brute-force login attempts)
// and automatically bans IPs that show malicious patterns.
//
// Usage:
//
//	go run . --playbook=fail2ban-install
//
// Execution Flow:
//  1. Updates package lists via apt-get update
//  2. Installs fail2ban package
//  3. Enables fail2ban to start on boot
//  4. Starts the fail2ban service
//
// Default Behavior:
//   - Uses default fail2ban configuration (/etc/fail2ban/jail.conf)
//   - Monitors SSH (/var/log/auth.log) for failed login attempts
//   - Default ban time: 10 minutes
//   - Default max retries: 5 attempts
//
// Monitored Services (by default):
//   - SSH (sshd jail) - primary protection against brute-force
//
// Security Benefits:
//   - Automatically blocks IPs with failed login attempts
//   - Reduces server load from brute-force attacks
//   - Provides audit trail of banned IPs
//
// Prerequisites:
//   - Root SSH access required
//   - Internet connectivity for package installation
//
// Post-Installation:
//   - Check status with: fail2ban-status
//   - View banned IPs: fail2ban-client status sshd
//
// Related Playbooks:
//   - fail2ban-status: Check service and jail status
type Fail2banInstall struct {
	*skills.BaseSkill
}

// Check determines if fail2ban needs to be installed.
func (f *Fail2banInstall) Check() (bool, error) {
	cfg := f.GetNodeConfig()
	cmdCheck := types.Command{
		Command:     "which fail2ban-server",
		Description: "Check if fail2ban is installed",
	}

	_, err := ssh.Run(cfg, cmdCheck)
	return err != nil, nil
}

// Run executes the skill and returns detailed result.
func (f *Fail2banInstall) Run() types.Result {
	cfg := f.GetNodeConfig()

	cmdInstall := types.Command{
		Command:     "apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y fail2ban",
		Description: "Install fail2ban",
	}
	cmdEnable := types.Command{
		Command:     "systemctl enable fail2ban && systemctl start fail2ban",
		Description: "Enable and start fail2ban",
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command",
			"cmd:", cmdInstall.Command,
			"description:", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command",
			"cmd:", cmdEnable.Command,
			"description:", cmdEnable.Description)

		return types.Result{
			Changed: true,
			Message: "Would install and enable fail2ban",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing fail2ban")

	output, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to install fail2ban",
			Error:   fmt.Errorf("failed to install fail2ban: %w\nOutput: %s", err, output),
		}
	}

	// Enable and start fail2ban
	output, err = ssh.Run(cfg, cmdEnable)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable/start fail2ban",
			Error:   fmt.Errorf("failed to enable/start fail2ban: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("fail2ban installed")
	return types.Result{
		Changed: true,
		Message: "Fail2ban installed and enabled",
	}
}
