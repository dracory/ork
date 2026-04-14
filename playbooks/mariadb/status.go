package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// Status displays comprehensive MariaDB server status information.
// This read-only playbook checks service status, network configuration, version,
// and tests database connectivity using the root credentials.
//
// Usage:
//
//	go run . --playbook=mariadb-status [--arg=root-password=<password>] [--arg=port=<port>]
//
// Args:
//   - root-password: MariaDB root password (optional, for connection test)
//   - port: MariaDB port (default: 3306)
//
// Prerequisites:
//   - MariaDB must be installed
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-install: Install MariaDB server
//   - mariadb-secure: Security hardening
type Status struct {
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (m *Status) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (m *Status) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	mariaDBPort := m.GetArg(ArgPort)
	if mariaDBPort == "" {
		mariaDBPort = DefaultPort
	}

	cfg.GetLoggerOrDefault().Info("checking MariaDB status")

	// Check service status
	serviceOutput, err := ssh.Run(cfg, "systemctl status mariadb --no-pager")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB is not running",
			Error:   fmt.Errorf("mariadb is not running: %w", err),
		}
	}
	cfg.GetLoggerOrDefault().Info("MariaDB service status", "output", serviceOutput)

	// Check if MariaDB is listening
	netOutput, _ := ssh.Run(cfg, fmt.Sprintf("ss -tlnp | grep :%s || netstat -tlnp | grep :%s || echo 'Port %s not found'", mariaDBPort, mariaDBPort, mariaDBPort))
	cfg.GetLoggerOrDefault().Info("MariaDB network status", "output", netOutput)

	// Check version
	versionOutput, _ := ssh.Run(cfg, "mysql --version")
	cfg.GetLoggerOrDefault().Info("MariaDB version", "output", versionOutput)

	// Check connection
	connectionStatus := "not tested"
	if rootPassword != "" {
		cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT 'MariaDB is working' as status;"`, rootPassword)
		connOutput, err := ssh.Run(cfg, cmd)
		if err != nil {
			connectionStatus = "failed"
			cfg.GetLoggerOrDefault().Warn("could not connect to MariaDB", "error", err)
		} else {
			connectionStatus = "successful"
			cfg.GetLoggerOrDefault().Info("MariaDB connection test", "output", connOutput)
		}
	}

	return playbook.Result{
		Changed: false,
		Message: "MariaDB status retrieved",
		Details: map[string]string{
			"service":    serviceOutput,
			"network":    netOutput,
			"version":    versionOutput,
			"connection": connectionStatus,
		},
	}
}

// NewStatus creates a new mariadb-status playbook.
func NewStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbStatus)
	pb.SetDescription("Display MariaDB server status information (read-only)")
	return &Status{BasePlaybook: pb}
}
