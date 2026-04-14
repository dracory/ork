package mariadb

import (
	"fmt"
	"strconv"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ChangePort changes the MariaDB server port from the default 3306.
// This playbook updates the MariaDB configuration and firewall rules
// to use a custom port for the database server.
//
// Usage:
//
//	go run . --playbook=mariadb-change-port --arg=port=<new_port> [--arg=root-password=<password>]
//
// Args:
//   - port: New port number (1024-65535, not 3306) (required)
//   - root-password: MariaDB root password (optional)
//   - config-path: MariaDB config file path (default: /etc/mysql/mariadb.conf.d/50-server.cnf)
//
// IMPORTANT:
//   - After changing the port, update your application configurations
//   - Ensure firewall allows the new port before changing
//   - Keep existing connections open until verified
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-status: Verify server is running after port change
//   - ufw-install: Configure firewall for new port
type ChangePort struct {
	*playbook.BasePlaybook
}

// Check always returns true to apply the port change.
func (m *ChangePort) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *ChangePort) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	newPort := m.GetArg(ArgPort)
	configPath := m.GetArg(ArgConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	if newPort == "" {
		return playbook.Result{
			Changed: false,
			Message: "Port parameter is required",
			Error:   fmt.Errorf("use --arg=port=<port_number>"),
		}
	}

	// Validate port
	portNum, err := strconv.Atoi(newPort)
	if err != nil || portNum < 1024 || portNum > 65535 || portNum == 3306 {
		return playbook.Result{
			Changed: false,
			Message: "Invalid port number",
			Error:   fmt.Errorf("port must be 1024-65535, not 3306"),
		}
	}

	cfg.GetLoggerOrDefault().Info("changing MariaDB port", "port", newPort)

	// Backup
	cfg.GetLoggerOrDefault().Info("backing up MariaDB configuration")
	_, err = ssh.Run(cfg, types.Command{Command: fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)`, configPath, configPath), Description: "Backup MariaDB config"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to backup config", Error: err}
	}

	// Update UFW if active
	ufwOutput, _ := ssh.Run(cfg, types.Command{Command: `ufw status | grep -q "Status: active" && echo "ACTIVE" || echo "INACTIVE"`, Description: "Check UFW status"})
	if ufwOutput == "ACTIVE" {
		cmd := fmt.Sprintf(`ufw allow %s/tcp comment 'MariaDB on custom port'`, newPort)
		_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Allow MariaDB custom port in UFW"})
	}

	// Update MariaDB port
	cmd := fmt.Sprintf(`sed -i 's/^#*port[[:space:]]*=.*/port = %s/' %s`, newPort, configPath)
	_, _ = ssh.Run(cfg, types.Command{Command: cmd, Description: "Update MariaDB port in config"})

	// Restart MariaDB
	cfg.GetLoggerOrDefault().Info("restarting MariaDB service")
	_, err = ssh.Run(cfg, types.Command{Command: `systemctl restart mariadb`, Description: "Restart MariaDB service"})
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart MariaDB", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("MariaDB port change complete")
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("MariaDB port changed to %s", newPort),
		Details: map[string]string{
			"new_port":    newPort,
			"config_path": configPath,
		},
	}
}

// NewChangePort creates a new mariadb-change-port playbook.
func NewChangePort() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbChangePort)
	pb.SetDescription("Change the MariaDB server port from default 3306")
	return &ChangePort{BasePlaybook: pb}
}
