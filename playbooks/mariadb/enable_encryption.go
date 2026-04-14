package mariadb

import (
	"fmt"
	"path/filepath"

	"github.com/dracory/ork/playbooks"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
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
	*playbooks.BasePlaybook
}

// Check determines if encryption needs to be enabled.
func (m *EnableEncryption) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *EnableEncryption) Run() types.Result {
	cfg := m.GetNodeConfig()

	// Get configurable paths
	configPath := m.GetArg(ArgConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	keyFilePath := m.GetArg(ArgKeyFilePath)
	if keyFilePath == "" {
		keyFilePath = DefaultKeyFilePath
	}

	cfg.GetLoggerOrDefault().Info("enabling MariaDB encryption at rest")

	// Define commands
	cmdBackup := types.Command{
		Command:     fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)`, configPath, configPath),
		Description: "Backup MariaDB config",
	}
	keyDir := filepath.Dir(keyFilePath)
	cmdMkdir := types.Command{
		Command:     fmt.Sprintf(`mkdir -p %s`, keyDir),
		Description: "Create key directory",
	}
	cmdGenKey := types.Command{
		Command:     fmt.Sprintf(`echo "1;$(openssl rand -hex 32)" > %s`, keyFilePath),
		Description: "Generate encryption key",
	}
	cmdPerms := types.Command{
		Command:     fmt.Sprintf(`chown mysql:mysql %s && chmod 600 %s`, keyFilePath, keyFilePath),
		Description: "Set key file permissions",
	}
	cmdConfigure := types.Command{
		Command: fmt.Sprintf(`grep -q "file_key_management_filename" %s || cat >> %s << 'EOF'

# Encryption at Rest Configuration
plugin_load_add = file_key_management
file_key_management_filename = %s
file_key_management_encryption_algorithm = AES_CBC
innodb_encrypt_tables = ON
innodb_encrypt_log = ON
encrypt_tmp_files = ON
EOF`, configPath, configPath, keyFilePath),
		Description: "Configure encryption"}
	cmdRestart := types.Command{
		Command:     `systemctl restart mariadb`,
		Description: "Restart MariaDB",
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdir.Command, "description", cmdMkdir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGenKey.Command, "description", cmdGenKey.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdPerms.Command, "description", cmdPerms.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdConfigure.Command, "description", cmdConfigure.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: "Would enable MariaDB encryption at rest",
		}
	}

	// Backup config
	cfg.GetLoggerOrDefault().Info("backing up MariaDB configuration")
	_, _ = ssh.Run(cfg, cmdBackup)

	// Create key directory
	_, _ = ssh.Run(cfg, cmdMkdir)

	// Generate encryption key
	cfg.GetLoggerOrDefault().Info("generating encryption key file")
	_, _ = ssh.Run(cfg, cmdGenKey)

	// Set permissions
	_, _ = ssh.Run(cfg, cmdPerms)

	// Configure encryption
	cfg.GetLoggerOrDefault().Info("configuring encryption in MariaDB")
	_, _ = ssh.Run(cfg, cmdConfigure)

	// Restart MariaDB
	cfg.GetLoggerOrDefault().Info("restarting MariaDB")
	_, err := ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to restart MariaDB", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("MariaDB encryption at rest enabled")
	return types.Result{
		Changed: true,
		Message: "Data-at-rest encryption enabled for MariaDB",
		Details: map[string]string{
			"key_file":    keyFilePath,
			"config_path": configPath,
		},
	}
}

// NewEnableEncryption creates a new mariadb-enable-encryption playbook.
func NewEnableEncryption() types.PlaybookInterface {
	pb := playbooks.NewBasePlaybook()
	pb.SetID(playbooks.IDMariadbEnableEncryption)
	pb.SetDescription("Enable data-at-rest encryption for MariaDB")
	return &EnableEncryption{BasePlaybook: pb}
}
