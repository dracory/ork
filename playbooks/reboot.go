package playbooks

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// NewReboot creates a new reboot playbook.
// By default, it does NOT wait for the server to reconnect.
func NewReboot() *reboot {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDReboot)
	pb.SetDescription("Reboot the remote server")
	return &reboot{
		BasePlaybook:     pb,
		WaitForReconnect: false,
		MaxWaitTime:      5 * time.Minute,
	}
}

// Reboot reboots the remote server and optionally waits for it to come back.
type reboot struct {
	*playbook.BasePlaybook
	// WaitForReconnect if true, will poll until server is back online
	WaitForReconnect bool
	// MaxWaitTime is the maximum time to wait for reconnection
	MaxWaitTime time.Duration
}

// Check always returns true for reboot since it's an explicit action.
// Reboot is always "needed" because the user explicitly requested it.
func (r *reboot) Check() (bool, error) {
	return true, nil // Always reboot when requested
}

// Run executes the reboot and returns detailed result.
func (r *reboot) Run() playbook.Result {
	cfg := r.GetConfig()
	log.Printf("Rebooting %s...", cfg.SSHHost)

	// Trigger reboot (non-blocking, command returns immediately)
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "reboot")
	if err != nil {
		// reboot command often returns connection error since it kills the SSH session
		log.Printf("Reboot command sent (connection error expected): %v", err)
	}

	if !r.WaitForReconnect {
		log.Println("Reboot initiated. Not waiting for server to come back online.")
		return playbook.Result{
			Changed: true, // Reboot was initiated
			Message: fmt.Sprintf("Reboot initiated on %s", cfg.SSHHost),
			Details: map[string]string{
				"wait_for_reconnect": "false",
			},
		}
	}

	// Wait and poll for server to come back
	maxWait := r.MaxWaitTime
	if maxWait == 0 {
		maxWait = 5 * time.Minute
	}

	log.Println("Waiting for server to come back online...")
	time.Sleep(10 * time.Second) // Give it time to actually start rebooting

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)

		_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "uptime")
		if err == nil {
			log.Println("Server is back online!")
			return playbook.Result{
				Changed: true,
				Message: fmt.Sprintf("Reboot completed on %s, server is back online", cfg.SSHHost),
				Details: map[string]string{
					"wait_for_reconnect": "true",
					"max_wait":           maxWait.String(),
				},
			}
		}
	}

	return playbook.Result{
		Changed: true, // Reboot was initiated even if we timed out waiting
		Message: fmt.Sprintf("Reboot initiated on %s, but timeout waiting for reconnect", cfg.SSHHost),
		Error:   fmt.Errorf("timeout waiting for server to come back online after %v", maxWait),
		Details: map[string]string{
			"wait_for_reconnect": "true",
			"max_wait":           maxWait.String(),
		},
	}
}
