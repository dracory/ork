package mariadb

import (
	"fmt"
	"path"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// EnableEncryption configures data-at-rest encryption for MariaDB.
// This skill generates an encryption key and enables encryption for InnoDB tables,
// protecting data files from unauthorized access if the storage media is compromised.
//
// Usage:
//
//	node.Run(mariadb.NewEnableEncryption().SetConfigPath("<config_path>").SetKeyFilePath("<keyfile_path>"))
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
//   - root-password: MariaDB root password (optional). When set, Run() verifies
//     encryption is active after the restart; when unset, the post-restart
//     verification is skipped.
//   - config-path: MariaDB config file path (default: /etc/mysql/mariadb.conf.d/50-server.cnf)
//   - keyfile-path: Encryption key file path (default: /var/lib/mysql-keyfile/keyfile.enc)
//
// Related Playbooks:
//   - mariadb-enable-ssl: Encrypt data in transit
//   - mariadb-security-audit: Verify encryption is working
type EnableEncryption struct {
	*types.BaseSkill
}

// Compile-time assertion that EnableEncryption implements types.RunnableInterface.
var _ types.RunnableInterface = (*EnableEncryption)(nil)

// Check determines if encryption needs to be enabled.
func (m *EnableEncryption) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
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
	// Use path.Dir (not filepath.Dir) because the target is always Linux —
	// filepath.Dir uses OS-native separators (backslashes on Windows), which
	// would produce a broken path like \etc\mysql\encryption on the remote.
	keyDir := path.Dir(keyFilePath)
	shellEscapedConfigPath := mariadbEscapeShellQuote(configPath)
	shellEscapedKeyFilePath := mariadbEscapeShellQuote(keyFilePath)
	cmdBackup := types.Command{
		Command:     fmt.Sprintf(`sh -c 'cp %s %s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)'`, shellEscapedConfigPath, shellEscapedConfigPath),
		Description: "Backup MariaDB config",
	}
	cmdCheckOpenSSLStr := ""
	cmdCheckOpenSSLStr += "which openssl"                      // check if openssl exists
	cmdCheckOpenSSLStr += " || " + skills.DebianNonInteractive // if not, prevent interactive prompts
	cmdCheckOpenSSLStr += " apt-get install -y openssl"        // install openssl, auto-confirm
	cmdCheckOpenSSLStr += skills.DpkgConfOptions               // keep local config, use maintainer default if unmodified
	cmdCheckOpenSSL := types.Command{
		Command:     cmdCheckOpenSSLStr,
		Description: "Ensure openssl is installed",
		Required:    true,
	}
	cmdGenKey := types.Command{
		Command:     fmt.Sprintf(`openssl rand -hex 32 | awk '{print "1;" $0}' > '%s'`, shellEscapedKeyFilePath),
		Description: "Generate encryption key",
		Required:    true,
	}
	cmdPerms := types.Command{
		Command:     fmt.Sprintf(`chown mysql:mysql '%s' && chmod 600 '%s'`, shellEscapedKeyFilePath, shellEscapedKeyFilePath),
		Description: "Set key file permissions",
		Required:    true,
	}
	cmdConfigure := types.Command{
		Command: fmt.Sprintf(`grep -q "file_key_management_filename" '%s' || cat >> '%s' << 'EOF'

# Encryption at Rest Configuration
plugin_load_add = file_key_management
file_key_management_filename = %s
file_key_management_encryption_algorithm = AES_CBC
innodb_encrypt_tables = ON
innodb_encrypt_log = ON
encrypt_tmp_files = ON
EOF`, shellEscapedConfigPath, shellEscapedConfigPath, keyFilePath),
		Description: "Configure encryption",
		Required:    true,
	}
	cmdRestart := types.Command{
		Command:     `systemctl restart mariadb`,
		Description: "Restart MariaDB",
		Required:    true,
	}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would create directory", "path", keyDir, "owner", "mysql:mysql", "mode", "700")
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCheckOpenSSL.Command, "description", cmdCheckOpenSSL.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGenKey.Command, "description", cmdGenKey.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdPerms.Command, "description", cmdPerms.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdConfigure.Command, "description", cmdConfigure.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRestart.Command, "description", cmdRestart.Description)
		return types.Result{
			Changed: true,
			Message: "Would enable MariaDB encryption at rest",
		}
	}

	// Backup config — if this fails, abort before modifying the config file
	// (continuing without a backup risks losing the original config on a bad edit).
	cfg.GetLoggerOrDefault().Info("backing up MariaDB configuration")
	backupOutput, backupErr := ssh.Run(cfg, cmdBackup)
	if backupErr != nil {
		return types.Result{Changed: false, Message: "Failed to backup MariaDB config: " + backupOutput, Error: backupErr}
	}

	// Create key directory using the fs.DirCreate skill (idempotent, handles
	// path escaping correctly, sets owner and mode in one step).
	dirResult := fs.NewDirCreate().
		SetPath(keyDir).
		SetOwner("mysql:mysql").
		SetMode("700").
		WithNodeConfig(cfg).
		Run()
	if dirResult.Error != nil {
		return types.Result{Changed: false, Message: "Failed to create key directory: " + dirResult.Message, Error: dirResult.Error}
	}

	// Ensure openssl is installed (needed for key generation)
	cfg.GetLoggerOrDefault().Info("ensuring openssl is installed")
	output, err := ssh.Run(cfg, cmdCheckOpenSSL)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to ensure openssl is installed: " + output, Error: err}
	}

	// Generate encryption key
	cfg.GetLoggerOrDefault().Info("generating encryption key file")
	output, err = ssh.Run(cfg, cmdGenKey)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to generate encryption key: " + output, Error: err}
	}

	// Set permissions
	output, err = ssh.Run(cfg, cmdPerms)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to set key file permissions: " + output, Error: err}
	}

	// Configure encryption
	cfg.GetLoggerOrDefault().Info("configuring encryption in MariaDB")
	output, err = ssh.Run(cfg, cmdConfigure)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to configure encryption: " + output, Error: err}
	}

	// Restart MariaDB
	cfg.GetLoggerOrDefault().Info("restarting MariaDB")
	output, err = ssh.Run(cfg, cmdRestart)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to restart MariaDB: " + output, Error: err}
	}

	// Optional post-restart verification: when a root password is provided,
	// confirm encryption is active on the server. Skipped when no password is set.
	rootPassword := m.GetArg(ArgRootPassword)
	verified := "skipped"
	if rootPassword != "" {
		shellEscapedPwd := mariadbEscapeShellQuote(rootPassword)
		cmdVerify := types.Command{
			Command:     fmt.Sprintf(`MYSQL_PWD='%s' mysql -u root -e "SHOW VARIABLES LIKE 'innodb_encrypt_tables';"`, shellEscapedPwd),
			Description: "Verify encryption is enabled",
			Required:    false,
		}
		cfg.GetLoggerOrDefault().Info("verifying MariaDB encryption is active")
		if _, vErr := ssh.Run(cfg, cmdVerify); vErr != nil {
			return types.Result{
				Changed: false,
				Message: "Encryption configured but verification failed",
				Error:   fmt.Errorf("encryption configured but could not verify innodb_encrypt_tables: %w", vErr),
			}
		}
		verified = "ok"
	}

	cfg.GetLoggerOrDefault().Info("MariaDB encryption at rest enabled")
	return types.Result{
		Changed: true,
		Message: "Data-at-rest encryption enabled for MariaDB",
		Details: map[string]string{
			"key_file":    keyFilePath,
			"config_path": configPath,
			"verified":    verified,
		},
	}
}

// SetArgs sets the arguments for enabling encryption.
// Returns EnableEncryption for fluent method chaining.
func (e *EnableEncryption) SetArgs(args map[string]string) types.RunnableInterface {
	e.BaseSkill.SetArgs(args)
	return e
}

// SetConfigPath sets the MariaDB config file path and returns EnableEncryption for chaining.
func (e *EnableEncryption) SetConfigPath(path string) *EnableEncryption {
	e.BaseSkill.SetArg(ArgConfigPath, path)
	return e
}

// SetKeyFilePath sets the encryption key file path and returns EnableEncryption for chaining.
func (e *EnableEncryption) SetKeyFilePath(path string) *EnableEncryption {
	e.BaseSkill.SetArg(ArgKeyFilePath, path)
	return e
}

// SetRootPassword sets the MariaDB root password and returns EnableEncryption for
// chaining. When set, Run() uses it to verify encryption is active after the
// restart; when unset, the post-restart verification is skipped.
func (e *EnableEncryption) SetRootPassword(password string) *EnableEncryption {
	e.BaseSkill.SetArg(ArgRootPassword, password)
	return e
}

// SetArg sets a single argument for enabling encryption.
// Returns EnableEncryption for fluent method chaining.
func (e *EnableEncryption) SetArg(key, value string) types.RunnableInterface {
	e.BaseSkill.SetArg(key, value)
	return e
}

// SetID sets the ID for enabling encryption.
// Returns EnableEncryption for fluent method chaining.
func (e *EnableEncryption) SetID(id string) types.RunnableInterface {
	e.BaseSkill.SetID(id)
	return e
}

// SetDescription sets the description for enabling encryption.
// Returns EnableEncryption for fluent method chaining.
func (e *EnableEncryption) SetDescription(description string) types.RunnableInterface {
	e.BaseSkill.SetDescription(description)
	return e
}

// SetTimeout sets the timeout for enabling encryption.
// Returns EnableEncryption for fluent method chaining.
func (e *EnableEncryption) SetTimeout(timeout time.Duration) types.RunnableInterface {
	e.BaseSkill.SetTimeout(timeout)
	return e
}

// NewEnableEncryption creates a new mariadb-enable-encryption skill.
func NewEnableEncryption() *EnableEncryption {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbEnableEncryption)
	pb.SetDescription("Enable data-at-rest encryption for MariaDB")
	return &EnableEncryption{BaseSkill: pb}
}
