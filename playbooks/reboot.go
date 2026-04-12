package playbooks

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
)

// Reboot reboots the remote server and optionally waits for it to come back.
type Reboot struct {
	// WaitForReconnect if true, will poll until server is back online
	WaitForReconnect bool
	// MaxWaitTime is the maximum time to wait for reconnection
	MaxWaitTime time.Duration
}

// Name returns the playbook identifier.
func (r *Reboot) Name() string {
	return "reboot"
}

// Description returns what this playbook does.
func (r *Reboot) Description() string {
	return "Reboot the remote server"
}

// Run executes the reboot.
func (r *Reboot) Run(cfg config.Config) error {
	log.Printf("Rebooting %s...", cfg.SSHHost)

	// Trigger reboot (non-blocking, command returns immediately)
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "reboot")
	if err != nil {
		// reboot command often returns connection error since it kills the SSH session
		log.Printf("Reboot command sent (connection error expected): %v", err)
	}

	if !r.WaitForReconnect {
		log.Println("Reboot initiated. Not waiting for server to come back online.")
		return nil
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
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for server to come back online after %v", maxWait)
}

// NewReboot creates a new reboot playbook.
// By default, it does NOT wait for the server to reconnect.
func NewReboot() *Reboot {
	return &Reboot{
		WaitForReconnect: false,
		MaxWaitTime:      5 * time.Minute,
	}
}
