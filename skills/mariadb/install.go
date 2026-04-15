package mariadb

import (
	"fmt"

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
//   - root-password: MariaDB root password (optional but recommended)
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
	*skills.BaseSkill
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
	var cmdSetPassword types.Command
	if rootPassword != "" {
		cmdSetPassword = types.Command{Command: fmt.Sprintf(`mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '%s';"`, rootPassword), Description: "Set root password"}
	}
	cmdBindAddr := types.Command{Command: `sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/mysql/mariadb.conf.d/50-server.cnf || sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/my.cnf.d/mariadb-server.cnf || true`, Description: "Configure bind address"}
	cmdRestart := types.Command{Command: "systemctl restart mariadb", Description: "Restart MariaDB"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall.Command, "description", cmdInstall.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdStartEnable.Command, "description", cmdStartEnable.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWaitReady.Command, "description", cmdWaitReady.Description)
		if rootPassword != "" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdSetPassword.Command, "description", cmdSetPassword.Description)
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBindAddr.Command, "description", cmdBindAddr.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: "Would install and configure MariaDB",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing MariaDB server")

	// Update package list and install MariaDB
	output, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to install MariaDB",
			Error:   fmt.Errorf("failed to install MariaDB: %w\nOutput: %s", err, output),
		}
	}

	// Start and enable MariaDB
	output, err = ssh.Run(cfg, cmdStartEnable)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to start MariaDB",
			Error:   fmt.Errorf("failed to start MariaDB: %w\nOutput: %s", err, output),
		}
	}

	// Wait for MariaDB to be ready
	_, _ = ssh.Run(cfg, cmdWaitReady)

	// Set root password if provided
	if rootPassword != "" {
		_, err = ssh.Run(cfg, cmdSetPassword)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("could not set root password", "error", err)
		} else {
			cfg.GetLoggerOrDefault().Info("root password set")
		}
	}

	// Configure MariaDB to listen on all interfaces for public access
	cfg.GetLoggerOrDefault().Info("configuring MariaDB to listen on all interfaces")
	_, _ = ssh.Run(cfg, cmdBindAddr)

	// Restart MariaDB to apply config changes
	output, err = ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart MariaDB",
			Error:   fmt.Errorf("failed to restart MariaDB: %w\nOutput: %s", err, output),
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

// NewInstall creates a new mariadb-install skill.
func NewInstall() types.RunnableInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDMariadbInstall)
	pb.SetDescription("Install and configure MariaDB database server")
	return &Install{BaseSkill: pb}
}
