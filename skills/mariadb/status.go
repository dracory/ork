package mariadb

import (
	"fmt"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Status displays comprehensive MariaDB server status information.
// This read-only skill checks service status, network configuration, version,
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
	*types.BaseSkill
}

// Check always returns false since this is a read-only skill.
func (m *Status) Check() (bool, error) {
	return false, nil
}

// Run executes the skill and returns detailed result.
func (m *Status) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	mariaDBPort := m.GetArg(ArgPort)
	if mariaDBPort == "" {
		mariaDBPort = DefaultPort
	}

	// Check service status
	cmdService := types.Command{Command: "systemctl status mariadb --no-pager", Description: "Check MariaDB service status"}
	cmdNet := types.Command{Command: fmt.Sprintf("ss -tlnp | grep :%s || netstat -tlnp | grep :%s || echo 'Port %s not found'", mariaDBPort, mariaDBPort, mariaDBPort), Description: "Check MariaDB network status"}
	cmdVersion := types.Command{Command: "mysql --version", Description: "Check MariaDB version"}
	var cmdConn types.Command
	if rootPassword != "" {
		cmdConn = types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT 'MariaDB is working' as status;"`, rootPassword), Description: "Test MariaDB connection"}
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdService.Command, "description", cmdService.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdNet.Command, "description", cmdNet.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdVersion.Command, "description", cmdVersion.Description)
		if rootPassword != "" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdConn.Command, "description", cmdConn.Description)
		}
		return types.Result{
			Changed: false,
			Message: "Would check MariaDB status",
		}
	}

	cfg.GetLoggerOrDefault().Info("checking MariaDB status")

	serviceOutput, err := ssh.Run(cfg, cmdService)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "MariaDB is not running",
			Error:   fmt.Errorf("mariadb is not running: %w", err),
		}
	}
	cfg.GetLoggerOrDefault().Info("MariaDB service status", "output", serviceOutput)

	// Check if MariaDB is listening
	netOutput, _ := ssh.Run(cfg, cmdNet)
	cfg.GetLoggerOrDefault().Info("MariaDB network status", "output", netOutput)

	// Check version
	versionOutput, _ := ssh.Run(cfg, cmdVersion)
	cfg.GetLoggerOrDefault().Info("MariaDB version", "output", versionOutput)

	// Check connection
	connectionStatus := "not tested"
	if rootPassword != "" {
		connOutput, err := ssh.Run(cfg, cmdConn)
		if err != nil {
			connectionStatus = "failed"
			cfg.GetLoggerOrDefault().Warn("could not connect to MariaDB", "error", err)
		} else {
			connectionStatus = "successful"
			cfg.GetLoggerOrDefault().Info("MariaDB connection test", "output", connOutput)
		}
	}

	return types.Result{
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

// NewStatus creates a new mariadb-status skill.
func NewStatus() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbStatus)
	pb.SetDescription("Display MariaDB server status information (read-only)")
	return &Status{BaseSkill: pb}
}
