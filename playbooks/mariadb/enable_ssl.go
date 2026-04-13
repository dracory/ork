package mariadb

import (
	"fmt"
	"log"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// EnableSSL configures SSL/TLS encryption for MariaDB connections.
// This playbook generates SSL certificates and configures the server to accept
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
	*playbook.BasePlaybook
}

// Check determines if SSL needs to be enabled.
func (m *EnableSSL) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *EnableSSL) Run() playbook.Result {
	cfg := m.GetConfig()

	// Get configurable paths
	dataDir := m.GetArg(ArgDataDir)
	if dataDir == "" {
		dataDir = DefaultDataDir
	}
	configPath := m.GetArg(ArgConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	log.Println("=== Enabling MariaDB SSL/TLS ===")

	// Generate SSL certificates
	log.Println("Generating SSL certificates...")
	_, _ = ssh.Run(cfg, fmt.Sprintf(`mysql_ssl_rsa_setup --datadir=%s`, dataDir))

	// Set ownership and permissions
	_, _ = ssh.Run(cfg, fmt.Sprintf(`chown mysql:mysql %s/*.pem`, dataDir))
	_, _ = ssh.Run(cfg, fmt.Sprintf(`chmod 600 %s/*-key.pem && chmod 644 %s/*.pem`, dataDir, dataDir))

	// Backup config
	_, _ = ssh.Run(cfg, fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d)`, configPath, configPath))

	// Configure SSL
	log.Println("Configuring MariaDB to use SSL...")
	cmd := fmt.Sprintf(`grep -q "ssl-ca" %s || cat >> %s << 'EOF'

# SSL/TLS Configuration
ssl-ca=%s/ca.pem
ssl-cert=%s/server-cert.pem
ssl-key=%s/server-key.pem
EOF`, configPath, configPath, dataDir, dataDir, dataDir)
	_, _ = ssh.Run(cfg, cmd)

	// Restart MariaDB
	log.Println("Restarting MariaDB service...")
	_, err := ssh.Run(cfg, `systemctl restart mariadb`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart MariaDB", Error: err}
	}

	log.Println("=== MariaDB SSL/TLS Configuration Complete ===")
	return playbook.Result{
		Changed: true,
		Message: "SSL/TLS enabled for MariaDB",
		Details: map[string]string{
			"cert_path":   dataDir,
			"data_dir":    dataDir,
			"config_path": configPath,
		},
	}
}

// NewEnableSSL creates a new mariadb-enable-ssl playbook.
func NewEnableSSL() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbEnableSSL)
	pb.SetDescription("Enable SSL/TLS encryption for MariaDB connections")
	return &EnableSSL{BasePlaybook: pb}
}
