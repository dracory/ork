package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Purge removes MariaDB database server and all associated data.
// This skill performs a complete removal of MariaDB including:
// - Stopping the MariaDB service
// - Removing MariaDB packages
// - Removing configuration files
// - Removing data directories
//
// Usage:
//
//	go run . --playbook=mariadb-purge
//
// Execution Flow:
//  1. Stops MariaDB service if running
//  2. Removes MariaDB packages (purge)
//  3. Removes MariaDB configuration files
//  4. Removes MariaDB data directories
//  5. Cleans up dependencies
//
// Security Notes:
//   - This operation is destructive and cannot be undone
//   - All databases and data will be permanently deleted
//   - Ensure you have backups before running this skill
//
// Prerequisites:
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-install: Reinstall MariaDB after purge
//   - mariadb-status: Check if MariaDB is still installed
type Purge struct {
	*types.BaseSkill
}

// Check determines if MariaDB needs to be purged.
func (m *Purge) Check() (bool, error) {
	cfg := m.GetNodeConfig()
	cmdCheck := types.Command{Command: "which mysqld", Description: "Check if MariaDB is installed"}
	_, err := ssh.Run(cfg, cmdCheck)
	return err == nil, nil // Return true if MariaDB is installed
}

// Run executes the skill and returns detailed result.
func (m *Purge) Run() types.Result {
	cfg := m.GetNodeConfig()

	// Define commands
	cmdStop := types.Command{Command: "systemctl stop mariadb", Description: "Stop MariaDB service"}
	cmdPurge := types.Command{Command: "apt-get purge -y mariadb-server mariadb-client mariadb-common", Description: "Remove MariaDB packages"}
	cmdAutoremove := types.Command{Command: "apt-get autoremove -y", Description: "Remove unused dependencies"}
	cmdRemoveConfig := types.Command{Command: "rm -rf /etc/mysql", Description: "Remove MariaDB configuration"}
	cmdRemoveData := types.Command{Command: "rm -rf /var/lib/mysql", Description: "Remove MariaDB data directory"}
	cmdRemoveLog := types.Command{Command: "rm -rf /var/log/mysql", Description: "Remove MariaDB logs"}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStop.Command, "description", cmdStop.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdPurge.Command, "description", cmdPurge.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAutoremove.Command, "description", cmdAutoremove.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRemoveConfig.Command, "description", cmdRemoveConfig.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRemoveData.Command, "description", cmdRemoveData.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRemoveLog.Command, "description", cmdRemoveLog.Description)
		return types.Result{
			Changed: true,
			Message: "Would purge MariaDB",
		}
	}

	// Track failures for reporting
	var failures []string

	// Stop MariaDB service (non-critical - service may not be running)
	cfg.GetLoggerOrDefault().Info("stopping MariaDB service")
	_, err := ssh.Run(cfg, cmdStop)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not stop MariaDB service (may not be running)", "error", err)
	}

	// Remove MariaDB packages (critical)
	cfg.GetLoggerOrDefault().Info("removing MariaDB packages")
	output, err := ssh.Run(cfg, cmdPurge)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to remove MariaDB packages",
			Error:   err,
		}
	}

	// Remove dependencies (non-critical)
	cfg.GetLoggerOrDefault().Info("removing unused dependencies")
	_, err = ssh.Run(cfg, cmdAutoremove)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove dependencies", "error", err)
		failures = append(failures, "autoremove")
	}

	// Remove configuration files (non-critical - may not exist)
	cfg.GetLoggerOrDefault().Info("removing MariaDB configuration files")
	_, err = ssh.Run(cfg, cmdRemoveConfig)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove configuration files (may not exist)", "error", err)
		failures = append(failures, "config")
	}

	// Remove data directory (critical)
	cfg.GetLoggerOrDefault().Info("removing MariaDB data directory")
	_, err = ssh.Run(cfg, cmdRemoveData)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove data directory", "error", err)
		failures = append(failures, "data directory")
	}

	// Remove log files (non-critical - may not exist)
	cfg.GetLoggerOrDefault().Info("removing MariaDB log files")
	_, err = ssh.Run(cfg, cmdRemoveLog)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove log files (may not exist)", "error", err)
		failures = append(failures, "logs")
	}

	// Build result message
	message := "MariaDB purged successfully"
	if len(failures) > 0 {
		message = fmt.Sprintf("MariaDB purged (with partial failures: %v)", failures)
	}

	return types.Result{
		Changed: true,
		Message: message,
		Details: map[string]string{
			"output":   output,
			"failures": fmt.Sprintf("%v", failures),
		},
	}
}

// SetArgs sets the arguments for MariaDB purge.
// Returns Purge for fluent method chaining.
func (p *Purge) SetArgs(args map[string]string) types.RunnableInterface {
	p.BaseSkill.SetArgs(args)
	return p
}

// SetArg sets a single argument for MariaDB purge.
// Returns Purge for fluent method chaining.
func (p *Purge) SetArg(key, value string) types.RunnableInterface {
	p.BaseSkill.SetArg(key, value)
	return p
}

// SetID sets the ID for MariaDB purge.
// Returns Purge for fluent method chaining.
func (p *Purge) SetID(id string) types.RunnableInterface {
	p.BaseSkill.SetID(id)
	return p
}

// SetDescription sets the description for MariaDB purge.
// Returns Purge for fluent method chaining.
func (p *Purge) SetDescription(description string) types.RunnableInterface {
	p.BaseSkill.SetDescription(description)
	return p
}

// SetTimeout sets the timeout for MariaDB purge.
// Returns Purge for fluent method chaining.
func (p *Purge) SetTimeout(timeout time.Duration) types.RunnableInterface {
	p.BaseSkill.SetTimeout(timeout)
	return p
}

// NewPurge creates a new MariaDB purge skill.
func NewPurge() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID("mariadb-purge")
	pb.SetDescription("Remove MariaDB database server and all associated data")
	return &Purge{BaseSkill: pb}
}
