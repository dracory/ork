package ufw

import (
	"fmt"
	"log"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// UfwStatus displays the current UFW firewall configuration and state.
// This read-only playbook checks whether UFW is installed, enabled, and
// displays all active rules. Use this to verify firewall configuration.
//
// Usage:
//
//	go run . --playbook=ufw-status
//
// Execution Flow:
//  1. Runs 'ufw status verbose' to get detailed status
//  2. Displays firewall state (active/inactive)
//  3. Shows default policies
//  4. Lists all configured rules with numbers
//
// Output Information:
//   - Status: active or inactive
//   - Default Policy: incoming/outgoing/routed
//   - Rules: numbered list with action, direction, and target
//   - Logging: current logging level
//
// Understanding the Output:
//   - Status active: Firewall is enforcing rules
//   - To Action: From -> Destination direction
//   - Anywhere: Applies to all IP addresses
//   - Numbers: Use with 'ufw delete <number>' to remove rules
//
// Prerequisites:
//   - UFW must be installed (use ufw-install playbook)
//   - Root SSH access required
//
// Related Playbooks:
//   - ufw-install: Install UFW firewall
//   - ufw-allow: Allow additional ports
type UfwStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (u *UfwStatus) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (u *UfwStatus) Run() playbook.Result {
	cfg := u.GetConfig()

	log.Println("Checking UFW status...")

	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "ufw status verbose")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to check UFW status",
			Error:   fmt.Errorf("failed to check UFW status: %w", err),
		}
	}

	log.Printf("UFW Status:\n%s", output)
	return playbook.Result{
		Changed: false,
		Message: "UFW status retrieved",
		Details: map[string]string{
			"status": output,
		},
	}
}

// NewUfwStatus creates a new ufw-status playbook.
func NewUfwStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUfwStatus)
	pb.SetDescription("Display UFW firewall configuration and status (read-only)")
	return &UfwStatus{BasePlaybook: pb}
}
