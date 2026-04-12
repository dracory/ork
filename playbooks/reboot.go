package playbooks

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// Reboot reboots the remote server and optionally waits for it to come back.
type Reboot struct {
	// WaitForReconnect if true, will poll until server is back online
	WaitForReconnect bool
	// MaxWaitTime is the maximum time to wait for reconnection
	MaxWaitTime time.Duration

	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (r *Reboot) GetID() string {
	return playbook.IDReboot
}

// SetID sets the playbook identifier.
func (r *Reboot) SetID(id string) playbook.Playbook {
	return r
}

// GetDescription returns what this playbook does.
func (r *Reboot) GetDescription() string {
	return "Reboot the remote server"
}

// SetDescription sets the playbook description.
func (r *Reboot) SetDescription(description string) playbook.Playbook {
	return r
}

// GetConfig returns the current node configuration.
func (r *Reboot) GetConfig() config.Config {
	return r.cfg
}

// GetOptions returns the current playbook options.
func (r *Reboot) GetOptions() *playbook.PlaybookOptions {
	return r.opts
}

// SetConfig sets the node configuration for this playbook.
func (r *Reboot) SetConfig(cfg config.Config) playbook.Playbook {
	r.cfg = cfg
	return r
}

// SetOptions sets the playbook-specific options.
func (r *Reboot) SetOptions(opts *playbook.PlaybookOptions) playbook.Playbook {
	r.opts = opts
	return r
}

// Check always returns true for reboot since it's an explicit action.
// Reboot is always "needed" because the user explicitly requested it.
func (r *Reboot) Check() (bool, error) {
	return true, nil // Always reboot when requested
}

// Run executes the reboot and returns detailed result.
func (r *Reboot) Run() playbook.Result {
	log.Printf("Rebooting %s...", r.cfg.SSHHost)

	// Trigger reboot (non-blocking, command returns immediately)
	_, err := ssh.RunOnce(r.cfg.SSHHost, r.cfg.SSHPort, r.cfg.RootUser, r.cfg.SSHKey, "reboot")
	if err != nil {
		// reboot command often returns connection error since it kills the SSH session
		log.Printf("Reboot command sent (connection error expected): %v", err)
	}

	if !r.WaitForReconnect {
		log.Println("Reboot initiated. Not waiting for server to come back online.")
		return playbook.Result{
			Changed: true, // Reboot was initiated
			Message: fmt.Sprintf("Reboot initiated on %s", r.cfg.SSHHost),
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

		_, err := ssh.RunOnce(r.cfg.SSHHost, r.cfg.SSHPort, r.cfg.RootUser, r.cfg.SSHKey, "uptime")
		if err == nil {
			log.Println("Server is back online!")
			return playbook.Result{
				Changed: true,
				Message: fmt.Sprintf("Reboot completed on %s, server is back online", r.cfg.SSHHost),
				Details: map[string]string{
					"wait_for_reconnect": "true",
					"max_wait":           maxWait.String(),
				},
			}
		}
	}

	return playbook.Result{
		Changed: true, // Reboot was initiated even if we timed out waiting
		Message: fmt.Sprintf("Reboot initiated on %s, but timeout waiting for reconnect", r.cfg.SSHHost),
		Error:   fmt.Errorf("timeout waiting for server to come back online after %v", maxWait),
		Details: map[string]string{
			"wait_for_reconnect": "true",
			"max_wait":           maxWait.String(),
		},
	}
}

// NewReboot creates a new reboot playbook.
// By default, it does NOT wait for the server to reconnect.
func NewReboot() *Reboot {
	return &Reboot{
		WaitForReconnect: false,
		MaxWaitTime:      5 * time.Minute,
	}
}
