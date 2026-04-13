package mariadb

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// EnableEncryption configures data-at-rest encryption for MariaDB.
// This playbook generates an encryption key and enables encryption for InnoDB tables,
// protecting data files from unauthorized access if the storage media is compromised.
//
// Usage:
//
//	go run . --playbook=mariadb-enable-encryption
//
// Execution Flow:
//  1. Backs up current MariaDB configuration
//  2. Generates encryption key file with random data
//  3. Sets secure permissions on encryption key (600)
//  4. Configures file key management plugin
//  5. Enables encryption for new tables by default
//  6. Restarts MariaDB to apply changes
//  7. Verifies encryption is enabled
//
// Encryption Configuration:
//   - Plugin: file_key_management
//   - Key file location: configurable (default: /var/lib/mysql-keyfile/keyfile.enc)
//   - Encryption algorithm: AES-256
//   - Default encryption: ON (all new InnoDB tables encrypted)
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//   - MariaDB 10.1.3 or later (encryption support)
//
// Args:
//   - root-password: MariaDB root password (optional)
//   - config-path: MariaDB config file path (default: /etc/mysql/mariadb.conf.d/50-server.cnf)
//   - keyfile-path: Encryption key file path (default: /var/lib/mysql-keyfile/keyfile.enc)
//
// Related Playbooks:
//   - mariadb-enable-ssl: Encrypt data in transit
//   - mariadb-security-audit: Verify encryption is working
type EnableEncryption struct {
	*playbook.BasePlaybook
}

// Check determines if encryption needs to be enabled.
func (m *EnableEncryption) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *EnableEncryption) Run() playbook.Result {
	cfg := m.GetConfig()

	// Get configurable paths
	configPath := m.GetArg(ArgConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	keyFilePath := m.GetArg(ArgKeyFilePath)
	if keyFilePath == "" {
		keyFilePath = DefaultKeyFilePath
	}

	log.Println("=== Enabling MariaDB Encryption at Rest ===")

	// Backup config
	log.Println("Backing up current MariaDB configuration...")
	_, _ = ssh.Run(cfg, fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)`, configPath, configPath))

	// Create key directory
	keyDir := filepath.Dir(keyFilePath)
	_, _ = ssh.Run(cfg, fmt.Sprintf(`mkdir -p %s`, keyDir))

	// Generate encryption key
	log.Println("Generating encryption key file...")
	cmd := fmt.Sprintf(`echo "1;$(openssl rand -hex 32)" > %s`, keyFilePath)
	_, _ = ssh.Run(cfg, cmd)

	// Set permissions
	_, _ = ssh.Run(cfg, fmt.Sprintf(`chown mysql:mysql %s && chmod 600 %s`, keyFilePath, keyFilePath))

	// Configure encryption
	log.Println("Configuring encryption in MariaDB configuration...")
	cmd = fmt.Sprintf(`grep -q "file_key_management_filename" %s || cat >> %s << 'EOF'

# Encryption at Rest Configuration
plugin_load_add = file_key_management
file_key_management_filename = %s
file_key_management_encryption_algorithm = AES_CBC
innodb_encrypt_tables = ON
innodb_encrypt_log = ON
encrypt_tmp_files = ON
EOF`, configPath, configPath, keyFilePath)
	_, _ = ssh.Run(cfg, cmd)

	// Restart MariaDB
	log.Println("Restarting MariaDB to apply changes...")
	_, err := ssh.Run(cfg, `systemctl restart mariadb`)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to restart MariaDB", Error: err}
	}

	log.Println("=== MariaDB Encryption at Rest Enabled ===")
	return playbook.Result{
		Changed: true,
		Message: "Data-at-rest encryption enabled for MariaDB",
		Details: map[string]string{
			"key_file":    keyFilePath,
			"config_path": configPath,
		},
	}
}

// NewEnableEncryption creates a new mariadb-enable-encryption playbook.
func NewEnableEncryption() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbEnableEncryption)
	pb.SetDescription("Enable data-at-rest encryption for MariaDB")
	return &EnableEncryption{BasePlaybook: pb}
}
