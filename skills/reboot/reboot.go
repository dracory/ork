// Package reboot provides a skill for rebooting remote servers.
// It supports both immediate reboot and wait-for-reconnect functionality
// to ensure the server comes back online after rebooting.
package reboot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// defaultMaxWait is the parsed form of DefaultMaxWait, computed once at init.
var defaultMaxWait = 5 * time.Minute

// defaultInitialWait is the parsed form of DefaultInitialWait, computed once at init.
var defaultInitialWait = 10 * time.Second

// defaultPollInterval is the parsed form of DefaultPollInterval, computed once at init.
var defaultPollInterval = 5 * time.Second

func init() {
	if d, err := time.ParseDuration(DefaultMaxWait); err == nil {
		defaultMaxWait = d
	}
	if d, err := time.ParseDuration(DefaultInitialWait); err == nil {
		defaultInitialWait = d
	}
	if d, err := time.ParseDuration(DefaultPollInterval); err == nil {
		defaultPollInterval = d
	}
}

// Reboot reboots the remote server and optionally waits for it to come back.
// This skill triggers a system reboot via the reboot command and can optionally
// poll the server until it responds again, confirming the reboot completed successfully.
//
// Usage:
//
//	node.Run(reboot.NewReboot())
//	// or with wait-for-reconnect:
//	node.Run(reboot.NewReboot().SetWaitForReconnect(true))
//
// Execution Flow (without wait):
//  1. Connects to remote server via SSH
//  2. Executes reboot command
//  3. Reports that reboot was initiated
//
// Execution Flow (with wait=true):
//  1. Connects to remote server via SSH
//  2. Executes reboot command
//  3. Waits the initial grace period for reboot to begin
//  4. Polls server every poll-interval until it responds to uptime command
//  5. Reports success when server is back online, or timeout if max-wait exceeded
//
// Expected Output:
//   - Success (no wait): "Reboot initiated on <host>"
//   - Success (with wait): "Reboot completed on <host>, server is back online"
//   - Timeout (with wait): Error indicating timeout waiting for reconnect
//
// Result Details:
//   - wait_for_reconnect: "true" or "false" indicating if wait was enabled
//   - max_wait: Duration string when wait is enabled (e.g. "5m0s")
//
// Use Cases:
//   - Apply kernel updates requiring reboot
//   - Recover from system issues
//   - Scheduled maintenance windows
//
// Safety Features:
//   - Connection errors after reboot command are expected and ignored
//   - Configurable maximum wait time prevents indefinite blocking
//   - Default max wait is 5 minutes if not specified
//   - max-wait is the total budget for the wait phase (initial grace period
//     plus polling); the skill will not block longer than max-wait
//
// Args:
//   - wait: "true"/"false" to enable wait-for-reconnect (default: false)
//   - max-wait: Go duration string for total reconnect wait (default: "5m",
//     only applies when wait is true)
//   - initial-wait: Go duration string for grace period before polling
//     (default: "10s", only applies when wait is true)
//   - poll-interval: Go duration string for delay between uptime probes
//     (default: "5s", only applies when wait is true)
//
// Note: By default, wait is false. The caller must explicitly enable waiting
// via SetWaitForReconnect(true) or the "wait" arg.
type Reboot struct {
	*types.BaseSkill
}

// Compile-time assertion that Reboot implements types.RunnableInterface.
var _ types.RunnableInterface = (*Reboot)(nil)

// Check always returns true for reboot since it's an explicit action.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since reboot is always explicitly
// requested by the user and always modifies system state, this always returns true.
//
// Note: Reboot is always "needed" because the user explicitly requested it.
// Unlike read-only skills (e.g. ping), reboot does not special-case dry-run
// here because Check reports intent to modify state, not execution preview.
func (r *Reboot) Check() (bool, error) {
	return true, nil // Always reboot when requested
}

// Run executes the reboot and returns detailed result.
// Changed is always true since reboot modifies the system state.
//
// When wait is enabled, this method will block until either:
//   - The server responds to SSH connections again (success)
//   - Max wait is exceeded (returns error with timeout message)
//
// The total blocking time is bounded by max-wait, which covers both the
// initial grace period and the polling phase.
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

	wait := r.getWaitForReconnect()
	maxWait := r.getMaxWaitTime()

	cfg.GetLoggerOrDefault().Info("rebooting server", "host", cfg.SSHHost)

	// Trigger reboot (non-blocking, command returns immediately)
	_, err := ssh.Run(cfg, cmdReboot)
	if err != nil {
		// reboot command often returns connection error since it kills the SSH session
		cfg.GetLoggerOrDefault().Info("reboot command sent", "host", cfg.SSHHost, "expected_error", err)
	}

	if !wait {
		cfg.GetLoggerOrDefault().Info("reboot initiated, not waiting", "host", cfg.SSHHost)
		return types.Result{
			Changed: true, // Reboot was initiated
			Message: fmt.Sprintf("Reboot initiated on %s", cfg.SSHHost),
			Details: map[string]string{
				"wait_for_reconnect": "false",
			},
		}
	}

	// Wait and poll for server to come back.
	// maxWait is the total budget for the whole wait phase (initial grace
	// period + polling). The skill will not block longer than maxWait.
	cfg.GetLoggerOrDefault().Info("waiting for server to come back online", "host", cfg.SSHHost)

	deadline := time.Now().Add(maxWait)

	// Initial grace period, capped to the remaining budget.
	initialWait := r.getInitialWait()
	if remaining := time.Until(deadline); initialWait > remaining {
		initialWait = remaining
	}
	if initialWait > 0 {
		time.Sleep(initialWait)
	}

	pollInterval := r.getPollInterval()
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		sleep := pollInterval
		if sleep > remaining {
			sleep = remaining
		}
		if sleep > 0 {
			time.Sleep(sleep)
		}

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

// getWaitForReconnect reads the "wait" arg and returns the parsed bool.
// Returns false when the arg is unset or invalid. Invalid values are logged.
func (r *Reboot) getWaitForReconnect() bool {
	v := r.GetArg(ArgWait)
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		r.GetNodeConfig().GetLoggerOrDefault().Warn("invalid wait arg, falling back to false", "arg", v, "err", err)
		return false
	}
	return b
}

// getMaxWaitTime reads the "max-wait" arg and returns the parsed duration.
// Falls back to defaultMaxWait when the arg is unset or invalid. Invalid
// values are logged.
func (r *Reboot) getMaxWaitTime() time.Duration {
	v := r.GetArg(ArgMaxWait)
	if v == "" {
		return defaultMaxWait
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		r.GetNodeConfig().GetLoggerOrDefault().Warn("invalid max-wait arg, falling back to default", "arg", v, "default", defaultMaxWait, "err", err)
		return defaultMaxWait
	}
	return d
}

// getInitialWait reads the "initial-wait" arg and returns the parsed duration.
// Falls back to defaultInitialWait when the arg is unset or invalid. Invalid
// values are logged.
func (r *Reboot) getInitialWait() time.Duration {
	v := r.GetArg(ArgInitialWait)
	if v == "" {
		return defaultInitialWait
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		r.GetNodeConfig().GetLoggerOrDefault().Warn("invalid initial-wait arg, falling back to default", "arg", v, "default", defaultInitialWait, "err", err)
		return defaultInitialWait
	}
	return d
}

// getPollInterval reads the "poll-interval" arg and returns the parsed duration.
// Falls back to defaultPollInterval when the arg is unset or invalid. Invalid
// values are logged.
func (r *Reboot) getPollInterval() time.Duration {
	v := r.GetArg(ArgPollInterval)
	if v == "" {
		return defaultPollInterval
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		r.GetNodeConfig().GetLoggerOrDefault().Warn("invalid poll-interval arg, falling back to default", "arg", v, "default", defaultPollInterval, "err", err)
		return defaultPollInterval
	}
	return d
}

// SetArgs sets the arguments for reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetWaitForReconnect enables or disables wait-for-reconnect behaviour.
// When true, the skill polls the server until it responds again after reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetWaitForReconnect(wait bool) *Reboot {
	r.BaseSkill.SetArg(ArgWait, strconv.FormatBool(wait))
	return r
}

// SetMaxWaitTime sets the maximum total time to wait for reconnection,
// covering both the initial grace period and the polling phase.
// Only applies when wait-for-reconnect is enabled.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetMaxWaitTime(d time.Duration) *Reboot {
	r.BaseSkill.SetArg(ArgMaxWait, d.String())
	return r
}

// SetInitialWait sets the grace period to wait after sending the reboot
// command before beginning to poll. Only applies when wait-for-reconnect
// is enabled.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetInitialWait(d time.Duration) *Reboot {
	r.BaseSkill.SetArg(ArgInitialWait, d.String())
	return r
}

// SetPollInterval sets the delay between successive uptime probes while
// waiting for the server to come back online. Only applies when
// wait-for-reconnect is enabled.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetPollInterval(d time.Duration) *Reboot {
	r.BaseSkill.SetArg(ArgPollInterval, d.String())
	return r
}

// SetID sets the ID for reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for reboot.
// Returns Reboot for fluent method chaining.
func (r *Reboot) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// WithNodeConfig sets the node config and returns Reboot for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (r *Reboot) WithNodeConfig(cfg types.NodeConfig) *Reboot {
	r.BaseSkill.SetNodeConfig(cfg)
	return r
}

// NewReboot creates a new reboot skill.
// By default, wait-for-reconnect is disabled (does not wait for server to come back).
//
// Returns:
//
//	A *Reboot configured with IDReboot identifier,
//	description "Reboot the remote server", and default max wait of 5 minutes.
//
// Configuration:
//
//	// Via fluent setters:
//	pb := reboot.NewReboot().SetWaitForReconnect(true).SetMaxWaitTime(10 * time.Minute)
//
//	// Via args (e.g. from RunByID / buildArgs):
//	pb := reboot.NewReboot().SetArg(reboot.ArgWait, "true").SetArg(reboot.ArgMaxWait, "10m")
//
// Note: max-wait only applies when wait is true.
func NewReboot() *Reboot {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDReboot)
	pb.SetDescription("Reboot the remote server")
	pb.SetArg(ArgWait, "false")
	pb.SetArg(ArgMaxWait, DefaultMaxWait)
	pb.SetArg(ArgInitialWait, DefaultInitialWait)
	pb.SetArg(ArgPollInterval, DefaultPollInterval)
	return &Reboot{
		BaseSkill: pb,
	}
}
