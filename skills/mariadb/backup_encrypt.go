package mariadb

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// BackupEncrypt creates an encrypted backup of a MariaDB database.
// This skill extends the standard backup process by encrypting the backup
// with AES-256-CBC using PBKDF2 key derivation, protecting sensitive data at rest.
//
// Usage:
//
//	node.Run(mariadb.NewBackupEncrypt().SetArg("dbname", "<database_name>").SetArg("dir", "/path/to/backups"))
//
// Args:
//   - dbname: Name of the database to backup (required)
//   - dir: Directory to store encrypted backup (default: /root/backups)
//   - root-password: MariaDB root password (optional, uses env if not provided)
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-backup: Standard (non-encrypted) backup
type BackupEncrypt struct {
	*types.BaseSkill
}

// Compile-time assertion that BackupEncrypt implements types.RunnableInterface.
var _ types.RunnableInterface = (*BackupEncrypt)(nil)

// Check determines if backup can be created.
func (b *BackupEncrypt) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (b *BackupEncrypt) Run() types.Result {
	cfg := b.GetNodeConfig()
	rootPassword := cfg.GetArg(ArgRootPassword)
	dbName := cfg.GetArg(ArgDBName)

	if dbName == "" {
		return types.Result{
			Changed: false,
			Message: "Database name is required",
			Error:   fmt.Errorf("use --arg=dbname=<database_name>"),
		}
	}

	// Validate database name
	validDBName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validDBName.MatchString(dbName) {
		return types.Result{
			Changed: false,
			Message: "Invalid database name",
			Error:   fmt.Errorf("only alphanumeric characters, underscores, and hyphens allowed"),
		}
	}

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	backupDir := cfg.GetArgOr(ArgBackupDir, "/root/backups")
	timestamp := time.Now().Format("20060102_150405")

	// Define commands
	shellEscapedBackupDir := mariadbEscapeShellQuote(backupDir)
	cmdMkdir := types.Command{Command: fmt.Sprintf(`mkdir -p '%s'`, shellEscapedBackupDir), Description: "Create backup directory"}
	cmdCheckOpenSSLStr := ""
	cmdCheckOpenSSLStr += "which openssl"                      // check if openssl exists
	cmdCheckOpenSSLStr += " || " + skills.DebianNonInteractive // if not, prevent interactive prompts
	cmdCheckOpenSSLStr += " apt-get install -y openssl"        // install openssl, auto-confirm
	cmdCheckOpenSSLStr += skills.DpkgConfOptions               // keep local config, use maintainer default if unmodified

	cmdCheckOpenSSLCmd := types.Command{Command: cmdCheckOpenSSLStr, Description: "Ensure openssl is installed"}
	shellEscapedPassword := mariadbEscapeShellQuote(rootPassword)
	cmdBackup := types.Command{Command: fmt.Sprintf(`(umask 077 && MYSQL_PWD='%s' mysqldump -u root --single-transaction --routines --triggers --events '%s' | gzip | openssl enc -aes-256-cbc -salt -pbkdf2 -pass env:MYSQL_PWD -out '%s'/%s_%s.sql.gz.enc)`,
		shellEscapedPassword, dbName, shellEscapedBackupDir, dbName, timestamp), Description: "Create encrypted backup"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdir.Command, "description", cmdMkdir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCheckOpenSSLCmd.Command, "description", cmdCheckOpenSSLCmd.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create encrypted backup for database '%s'", dbName),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating encrypted database backup", "database", dbName)
	_, _ = ssh.Run(cfg, cmdMkdir)
	_, _ = ssh.Run(cfg, cmdCheckOpenSSLCmd)

	cfg.GetLoggerOrDefault().Info("creating encrypted backup")
	_, err := ssh.Run(cfg, cmdBackup)
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to create backup", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("encrypted backup complete")
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Encrypted backup created: %s/%s_%s.sql.gz.enc", backupDir, dbName, timestamp),
		Details: map[string]string{
			"backup_path": fmt.Sprintf("%s/%s_%s.sql.gz.enc", backupDir, dbName, timestamp),
		},
	}
}

// SetArgs sets the arguments for encrypted MariaDB backup.
// Returns BackupEncrypt for fluent method chaining.
func (b *BackupEncrypt) SetArgs(args map[string]string) types.RunnableInterface {
	b.BaseSkill.SetArgs(args)
	return b
}

// SetArg sets a single argument for encrypted MariaDB backup.
// Returns BackupEncrypt for fluent method chaining.
func (b *BackupEncrypt) SetArg(key, value string) types.RunnableInterface {
	b.BaseSkill.SetArg(key, value)
	return b
}

// SetID sets the ID for encrypted MariaDB backup.
// Returns BackupEncrypt for fluent method chaining.
func (b *BackupEncrypt) SetID(id string) types.RunnableInterface {
	b.BaseSkill.SetID(id)
	return b
}

// SetDescription sets the description for encrypted MariaDB backup.
// Returns BackupEncrypt for fluent method chaining.
func (b *BackupEncrypt) SetDescription(description string) types.RunnableInterface {
	b.BaseSkill.SetDescription(description)
	return b
}

// SetTimeout sets the timeout for encrypted MariaDB backup.
// Returns BackupEncrypt for fluent method chaining.
func (b *BackupEncrypt) SetTimeout(timeout time.Duration) types.RunnableInterface {
	b.BaseSkill.SetTimeout(timeout)
	return b
}

// NewBackupEncrypt creates a new mariadb-backup-encrypt skill.
func NewBackupEncrypt() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbBackupEncrypt)
	pb.SetDescription("Create an encrypted backup of a MariaDB database")
	return &BackupEncrypt{BaseSkill: pb}
}
