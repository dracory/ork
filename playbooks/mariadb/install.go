package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Install installs and configures MariaDB database server.
// This playbook performs a complete MariaDB installation with basic configuration
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
	*playbook.BasePlaybook
}

// Check determines if MariaDB needs to be installed.
func (m *Install) Check() (bool, error) {
	cfg := m.GetNodeConfig()
	_, err := ssh.Run(cfg, types.Command{Command: "which mysqld", Description: "Check if MariaDB is installed"})
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (m *Install) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	cfg.GetLoggerOrDefault().Info("installing MariaDB server")

	// Update package list and install MariaDB
	cmd := `apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y mariadb-server mariadb-client`
	output, err := ssh.Run(cfg, types.Command{Command: cmd, Description: "Install MariaDB packages"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to install MariaDB",
			Error:   fmt.Errorf("failed to install MariaDB: %w\nOutput: %s", err, output),
		}
	}

	// Start and enable MariaDB
	cmd = "systemctl start mariadb && systemctl enable mariadb"
	output, err = ssh.Run(cfg, types.Command{Command: cmd, Description: "Start and enable MariaDB"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to start MariaDB",
			Error:   fmt.Errorf("failed to start MariaDB: %w\nOutput: %s", err, output),
		}
	}

	// Wait for MariaDB to be ready
	cmd = "until mysqladmin ping --silent; do sleep 1; done"
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Wait for MariaDB to be ready"})

	// Set root password if provided
	if rootPassword != "" {
		cmd = fmt.Sprintf(`mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '%s';"`, rootPassword)
		_, err = ssh.Run(cfg, types.Command{Command: cmd, Description: "Set root password"})
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("could not set root password", "error", err)
		} else {
			cfg.GetLoggerOrDefault().Info("root password set")
		}
	}

	// Configure MariaDB to listen on all interfaces for public access
	cfg.GetLoggerOrDefault().Info("configuring MariaDB to listen on all interfaces")
	cmd = `sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/mysql/mariadb.conf.d/50-server.cnf || sed -i 's/^bind-address.*/bind-address = 0.0.0.0/' /etc/my.cnf.d/mariadb-server.cnf || true`
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Configure bind address"})

	// Restart MariaDB to apply config changes
	cmd = "systemctl restart mariadb"
	output, err = ssh.Run(cfg, types.Command{Command: cmd, Description: "Restart MariaDB"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to restart MariaDB",
			Error:   fmt.Errorf("failed to restart MariaDB: %w\nOutput: %s", err, output),
		}
	}

	return playbook.Result{
		Changed: true,
		Message: "MariaDB installed and configured successfully",
		Details: map[string]string{
			"bind_address": "0.0.0.0",
		},
	}
}

// NewInstall creates a new mariadb-install playbook.
func NewInstall() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbInstall)
	pb.SetDescription("Install and configure MariaDB database server")
	return &Install{BasePlaybook: pb}
}
