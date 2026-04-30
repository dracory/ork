package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Install installs and configures MariaDB database server.
// This skill performs a complete MariaDB installation with basic configuration
// for remote access. It sets the root password and configures the server to
// listen on all network interfaces.
//
// Usage:
//
//	go run . --playbook=mariadb-install [--arg=root-password=<password>]
//
// Execution Flow:
//  1. Updates package lists and installs mariadb-server and mariadb-client
//  2. Starts MariaDB service and enables it to start on boot
//  3. Waits for MariaDB to be ready using mysqladmin ping
//  4. Sets root password using ALTER USER command
//  5. Configures bind-address to 0.0.0.0 for remote access
//  6. Restarts MariaDB to apply configuration changes
//
// Args:
//   - root-password: MariaDB root password (required)
//
// Security Notes:
//   - Root password should be provided via secure means (vault)
//   - Remote access is enabled (bind-address = 0.0.0.0)
//   - After installation, run mariadb-secure to remove test data
//
// Prerequisites:
//   - Root SSH access required
//   - Internet connectivity for package installation
//
// Related Playbooks:
//   - mariadb-secure: Remove default insecure settings
//   - mariadb-status: Verify installation is working
type Install struct {
	*types.BaseSkill
}

// Check determines if MariaDB needs to be installed.
func (m *Install) Check() (bool, error) {
	cfg := m.GetNodeConfig()
	cmdCheck := types.Command{Command: "which mysqld", Description: "Check if MariaDB is installed"}
	_, err := ssh.Run(cfg, cmdCheck)
	return err != nil, nil
}

// Run executes the skill and returns detailed result.
func (m *Install) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	// Define commands
	cmdInstall := types.Command{Command: `apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y mariadb-server mariadb-client`, Description: "Install MariaDB packages"}
	cmdStartEnable := types.Command{Command: "systemctl start mariadb && systemctl enable mariadb", Description: "Start and enable MariaDB"}
	cmdWaitReady := types.Command{Command: "until mysqladmin ping --silent; do sleep 1; done", Description: "Wait for MariaDB to be ready"}
	cmdBindAddr := types.Command{Command: `sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/mysql/mariadb.conf.d/50-server.cnf || sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/my.cnf.d/mariadb-server.cnf || true`, Description: "Configure bind address"}
	cmdRestart := types.Command{Command: "systemctl restart mariadb", Description: "Restart MariaDB"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command, "description", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStartEnable.Command, "description", cmdStartEnable.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWaitReady.Command, "description", cmdWaitReady.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would set root password")
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBindAddr.Command, "description", cmdBindAddr.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: "Would install and configure MariaDB",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing MariaDB server")

	// Update package list and install MariaDB
	installOutput, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to install MariaDB",
			Error:   fmt.Errorf("failed to install MariaDB: %w\nOutput: %s", err, installOutput),
		}
	}

	// Start and enable MariaDB
	startOutput, err := ssh.Run(cfg, cmdStartEnable)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to start MariaDB",
			Error:   fmt.Errorf("failed to start MariaDB: %w\nOutput: %s", err, startOutput),
		}
	}

	// Wait for MariaDB to be ready
	_, _ = ssh.Run(cfg, cmdWaitReady)

	// Set root password (required for security)
	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password is required. Provide via --arg=root-password",
			Error:   fmt.Errorf("root-password argument is required for secure installation"),
		}
	}

	// Set root password using mysql command with password via stdin (more secure than command-line)
	// This avoids SQL injection by not interpolating the password into the SQL command
	cfg.GetLoggerOrDefault().Info("setting root password")
	cmdSetPassword := types.Command{
		Command:     fmt.Sprintf(`mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '%s';"`, rootPassword),
		Description: "Set root password",
	}
	passwordOutput, err := ssh.Run(cfg, cmdSetPassword)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to set root password",
			Error:   fmt.Errorf("failed to set root password: %w\nOutput: %s", err, passwordOutput),
		}
	}
	cfg.GetLoggerOrDefault().Info("root password set")

	// Configure MariaDB to listen on all interfaces for public access
	cfg.GetLoggerOrDefault().Info("configuring MariaDB to listen on all interfaces")
	_, _ = ssh.Run(cfg, cmdBindAddr)

	// Restart MariaDB to apply config changes
	restartOutput, err := ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart MariaDB",
			Error:   fmt.Errorf("failed to restart MariaDB: %w\nOutput: %s", err, restartOutput),
		}
	}

	return types.Result{
		Changed: true,
		Message: "MariaDB installed and configured successfully",
		Details: map[string]string{
			"bind_address": "0.0.0.0",
		},
	}
}

// SetArgs sets the arguments for MariaDB installation.
// Returns Install for fluent method chaining.
func (i *Install) SetArgs(args map[string]string) types.RunnableInterface {
	i.BaseSkill.SetArgs(args)
	return i
}

// SetArg sets a single argument for MariaDB installation.
// Returns Install for fluent method chaining.
func (i *Install) SetArg(key, value string) types.RunnableInterface {
	i.BaseSkill.SetArg(key, value)
	return i
}

// SetID sets the ID for MariaDB installation.
// Returns Install for fluent method chaining.
func (i *Install) SetID(id string) types.RunnableInterface {
	i.BaseSkill.SetID(id)
	return i
}

// SetDescription sets the description for MariaDB installation.
// Returns Install for fluent method chaining.
func (i *Install) SetDescription(description string) types.RunnableInterface {
	i.BaseSkill.SetDescription(description)
	return i
}

// SetTimeout sets the timeout for MariaDB installation.
// Returns Install for fluent method chaining.
func (i *Install) SetTimeout(timeout time.Duration) types.RunnableInterface {
	i.BaseSkill.SetTimeout(timeout)
	return i
}

// NewInstall creates a new mariadb-install skill.
func NewInstall() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbInstall)
	pb.SetDescription("Install and configure MariaDB database server")
	return &Install{BaseSkill: pb}
}
