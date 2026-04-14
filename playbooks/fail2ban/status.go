package fail2ban

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// Fail2banStatus displays the fail2ban service status and SSH jail information.
// This read-only playbook shows whether fail2ban is running and which IPs are currently
// banned for suspicious activity.
//
// Usage:
//
//	go run . --playbook=fail2ban-status
//
// Execution Flow:
//  1. Checks fail2ban service status using systemctl
//  2. Displays service state, uptime, and process information
//  3. Queries the SSH jail for currently banned IPs
//
// Output Information:
//
//	Service Status:
//	  - Active/Inactive state
//	  - Process ID and main PID
//	  - Memory usage
//	  - Recent log entries
//
//	SSH Jail Status:
//	  - List of currently banned IP addresses
//	  - Number of failed attempts per IP
//	  - Time remaining on bans
//
// Understanding Bans:
//   - Banned IPs are blocked at the firewall level
//   - Default ban duration: 10 minutes (configurable)
//   - IPs are automatically unbanned after the ban time expires
//   - Persistent attackers may be banned repeatedly
//
// Common Indicators:
//   - Many banned IPs: Indicates active brute-force attacks
//   - Few/no bans: Either no attacks or fail2ban not working
//   - Service inactive: fail2ban is not running
//
// Prerequisites:
//   - fail2ban must be installed
//   - Root SSH access required
//
// Related Playbooks:
//   - fail2ban-install: Install fail2ban
type Fail2banStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (f *Fail2banStatus) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (f *Fail2banStatus) Run() playbook.Result {
	cfg := f.GetNodeConfig()
	cmdStatus := "systemctl status fail2ban --no-pager"
	cmdJail := "fail2ban-client status sshd 2>/dev/null || echo 'No SSH jail configured'"

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStatus)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdJail)
		return playbook.Result{
			Changed: false,
			Message: "Would check fail2ban status",
		}
	}

	cfg.GetLoggerOrDefault().Info("checking fail2ban status")
	output, err := ssh.Run(cfg, cmdStatus)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Fail2ban is not running",
			Error:   fmt.Errorf("fail2ban is not running: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("fail2ban status", "output", output)
	jailOutput, _ := ssh.Run(cfg, cmdJail)
	cfg.GetLoggerOrDefault().Info("ssh jail status", "output", jailOutput)

	return playbook.Result{
		Changed: false,
		Message: "Fail2ban status retrieved",
		Details: map[string]string{
			"status": output,
			"jail":   jailOutput,
		},
	}
}

// NewFail2banStatus creates a new fail2ban-status playbook.
func NewFail2banStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDFail2banStatus)
	pb.SetDescription("Display fail2ban service status and SSH jail information (read-only)")
	return &Fail2banStatus{BasePlaybook: pb}
}
