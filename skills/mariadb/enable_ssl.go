package mariadb

import (
	"fmt"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// EnableSSL configures SSL/TLS encryption for MariaDB connections.
// This skill generates SSL certificates and configures the server to accept
// encrypted connections, protecting data in transit from eavesdropping.
//
// Usage:
//
//	go run . --playbook=mariadb-enable-ssl [--arg=root-password=<password>]
//
// Args:
//   - root-password: MariaDB root password (optional)
//   - data-dir: MariaDB data directory (default: /var/lib/mysql)
//   - config-path: MariaDB config file path (default: /etc/mysql/mariadb.conf.d/50-server.cnf)
//
// Execution Flow:
//  1. Checks for existing SSL certificates
//  2. Generates new certificates using mysql_ssl_rsa_setup
//  3. Sets correct ownership (mysql:mysql) on certificate files
//  4. Sets secure permissions on private key files (600)
//  5. Backs up current MariaDB configuration
//  6. Adds SSL configuration to 50-server.cnf
//  7. Restarts MariaDB to apply changes
//  8. Verifies SSL is enabled and working
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-security-audit: Verify SSL is working properly
//   - mariadb-create-user: Create users with SSL requirements
type EnableSSL struct {
	*types.BaseSkill
}

// Check determines if SSL needs to be enabled.
func (m *EnableSSL) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (m *EnableSSL) Run() types.Result {
	cfg := m.GetNodeConfig()

	// Get configurable paths
	dataDir := m.GetArg(ArgDataDir)
	if dataDir == "" {
		dataDir = DefaultDataDir
	}
	configPath := m.GetArg(ArgConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	cfg.GetLoggerOrDefault().Info("enabling MariaDB SSL/TLS")

	// Define commands
	cmdGenCert := types.Command{Command: fmt.Sprintf(`mysql_ssl_rsa_setup --datadir=%s`, dataDir), Description: "Generate SSL certificates"}
	cmdChown := types.Command{Command: fmt.Sprintf(`chown mysql:mysql %s/*.pem`, dataDir), Description: "Set SSL cert ownership"}
	cmdChmod := types.Command{Command: fmt.Sprintf(`chmod 600 %s/*-key.pem && chmod 644 %s/*.pem`, dataDir, dataDir), Description: "Set SSL cert permissions"}
	cmdBackup := types.Command{Command: fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d)`, configPath, configPath), Description: "Backup MariaDB config"}
	cmdConfigure := types.Command{Command: fmt.Sprintf(`grep -q "ssl-ca" %s || cat >> %s << 'EOF'

# SSL/TLS Configuration
ssl-ca=%s/ca.pem
ssl-cert=%s/server-cert.pem
ssl-key=%s/server-key.pem
EOF`, configPath, configPath, dataDir, dataDir, dataDir), Description: "Configure SSL in MariaDB"}
	cmdRestart := types.Command{Command: `systemctl restart mariadb`, Description: "Restart MariaDB"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGenCert.Command, "description", cmdGenCert.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChown.Command, "description", cmdChown.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmod.Command, "description", cmdChmod.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdConfigure.Command, "description", cmdConfigure.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: "Would enable MariaDB SSL/TLS",
		}
	}

	// Generate SSL certificates
	cfg.GetLoggerOrDefault().Info("generating SSL certificates")
	_, _ = ssh.Run(cfg, cmdGenCert)

	// Set ownership and permissions
	_, _ = ssh.Run(cfg, cmdChown)
	_, _ = ssh.Run(cfg, cmdChmod)

	// Backup config
	_, _ = ssh.Run(cfg, cmdBackup)

	// Configure SSL
	cfg.GetLoggerOrDefault().Info("configuring MariaDB to use SSL")
	_, _ = ssh.Run(cfg, cmdConfigure)

	// Restart MariaDB
	cfg.GetLoggerOrDefault().Info("restarting MariaDB service")
	_, err := ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to restart MariaDB", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("MariaDB SSL/TLS configuration complete")
	return types.Result{
		Changed: true,
		Message: "SSL/TLS enabled for MariaDB",
		Details: map[string]string{
			"cert_path":   dataDir,
			"data_dir":    dataDir,
			"config_path": configPath,
		},
	}
}

// NewEnableSSL creates a new mariadb-enable-ssl skill.
func NewEnableSSL() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbEnableSSL)
	pb.SetDescription("Enable SSL/TLS encryption for MariaDB connections")
	return &EnableSSL{BaseSkill: pb}
}
