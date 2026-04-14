// Package reboot provides a playbook for rebooting remote servers.
// It supports both immediate reboot and wait-for-reconnect functionality
// to ensure the server comes back online after rebooting.
package reboot

import (
	"fmt"
	"time"

	"github.com/dracory/ork/playbooks"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Reboot reboots the remote server and optionally waits for it to come back.
// This playbook triggers a system reboot via the reboot command and can optionally
// poll the server until it responds again, confirming the reboot completed successfully.
//
// Usage:
//
//	go run . --playbook=reboot
//
// Execution Flow (without wait):
//  1. Connects to remote server via SSH
//  2. Executes reboot command
//  3. Reports that reboot was initiated
//
// Execution Flow (with WaitForReconnect=true):
//  1. Connects to remote server via SSH
//  2. Executes reboot command
//  3. Waits 10 seconds for reboot to begin
//  4. Polls server every 5 seconds until it responds to uptime command
//  5. Reports success when server is back online, or timeout if max wait exceeded
//
// Expected Output:
//   - Success (no wait): "Reboot initiated on <host>"
//   - Success (with wait): "Reboot completed on <host>, server is back online"
//   - Timeout (with wait): Error indicating timeout waiting for reconnect
//
// Result Details:
//   - wait_for_reconnect: "true" or "false" indicating if wait was enabled
//   - max_wait: Duration string when wait is enabled (e.g., "5m0s")
//
// Use Cases:
//   - Apply kernel updates requiring reboot
//   - Recover from system issues
//   - Scheduled maintenance windows
//
// Safety Features:
//   - Connection errors after reboot command are expected and ignored
//   - Configurable maximum wait time prevents indefinite blocking
//   - Default MaxWaitTime is 5 minutes if not specified
//
// Note: By default, WaitForReconnect is false. The caller must explicitly
// enable waiting by setting WaitForReconnect=true on the returned instance.
type Reboot struct {
	*playbooks.BasePlaybook
	// WaitForReconnect if true, will poll until server is back online
	WaitForReconnect bool
	// MaxWaitTime is the maximum time to wait for reconnection (default: 5m)
	MaxWaitTime time.Duration
}

// Check always returns true for reboot since it's an explicit action.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since reboot is always explicitly
// requested by the user and always modifies system state, this always returns true.
//
// Note: Reboot is always "needed" because the user explicitly requested it.
func (r *Reboot) Check() (bool, error) {
	return true, nil // Always reboot when requested
}

// Run executes the reboot and returns detailed result.
// Changed is always true since reboot modifies the system state.
//
// When WaitForReconnect is true, this method will block until either:
//   - The server responds to SSH connections again (success)
//   - MaxWaitTime is exceeded (returns error with timeout message)
//
// Result.Details contains:
//   - wait_for_reconnect: "true" or "false"
//   - max_wait: Maximum wait duration string (when wait is enabled)
func (r *Reboot) Run() types.Result {
	cfg := r.GetNodeConfig()
	cmdReboot := types.Command{Command: "reboot", Description: "Reboot server"}

	// Check for dry-run mode - display actual command
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdReboot.Command, "host", cfg.SSHHost)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would reboot %s", cfg.SSHHost),
		}
	}

	cfg.GetLoggerOrDefault().Info("rebooting server", "host", cfg.SSHHost)

	// Trigger reboot (non-blocking, command returns immediately)
	_, err := ssh.Run(cfg, cmdReboot)
	if err != nil {
		// reboot command often returns connection error since it kills the SSH session
		cfg.GetLoggerOrDefault().Info("reboot command sent", "host", cfg.SSHHost, "expected_error", err)
	}

	if !r.WaitForReconnect {
		cfg.GetLoggerOrDefault().Info("reboot initiated, not waiting", "host", cfg.SSHHost)
		return types.Result{
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

	cfg.GetLoggerOrDefault().Info("waiting for server to come back online", "host", cfg.SSHHost)
	time.Sleep(10 * time.Second) // Give it time to actually start rebooting

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)

		cmdUptime := types.Command{Command: "uptime", Description: "Check if server is back online"}
		_, err := ssh.Run(cfg, cmdUptime)
		if err == nil {
			cfg.GetLoggerOrDefault().Info("server is back online", "host", cfg.SSHHost)
			return types.Result{
				Changed: true,
				Message: fmt.Sprintf("Reboot completed on %s, server is back online", cfg.SSHHost),
				Details: map[string]string{
					"wait_for_reconnect": "true",
					"max_wait":           maxWait.String(),
				},
			}
		}
	}

	return types.Result{
		Changed: true, // Reboot was initiated even if we timed out waiting
		Message: fmt.Sprintf("Reboot initiated on %s, but timeout waiting for reconnect", cfg.SSHHost),
		Error:   fmt.Errorf("timeout waiting for server to come back online after %v", maxWait),
		Details: map[string]string{
			"wait_for_reconnect": "true",
			"max_wait":           maxWait.String(),
		},
	}
}

// NewReboot creates a new reboot playbook.
// By default, WaitForReconnect is false (does not wait for server to come back).
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDReboot identifier,
//	description "Reboot the remote server", and default MaxWaitTime of 5 minutes.
//
// Configuration:
//
//	  The returned *Reboot can be type-asserted to configure WaitForReconnect:
//
//		pb := NewReboot().(*Reboot)
//		pb.WaitForReconnect = true
//		pb.MaxWaitTime = 10 * time.Minute
//
// Note: MaxWaitTime only applies when WaitForReconnect is true.
func NewReboot() types.PlaybookInterface {
	pb := playbooks.NewBasePlaybook()
	pb.SetID(playbooks.IDReboot)
	pb.SetDescription("Reboot the remote server")
	return &Reboot{
		BasePlaybook:     pb,
		WaitForReconnect: false,
		MaxWaitTime:      5 * time.Minute,
	}
}
