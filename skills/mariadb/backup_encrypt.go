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
//	go run . --playbook=mariadb-backup-encrypt --arg=dbname=<database_name> [--arg=dir=/path/to/backups]
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
	cmdMkdir := types.Command{Command: fmt.Sprintf(`mkdir -p %s`, backupDir), Description: "Create backup directory"}
	cmdCheckOpenSSL := types.Command{Command: `which openssl || DEBIAN_FRONTEND=noninteractive apt-get install -y openssl`, Description: "Ensure openssl is installed"}
	cmdBackup := types.Command{Command: fmt.Sprintf(`(umask 077 && MYSQL_PWD='%s' mysqldump -u root --single-transaction --routines --triggers --events '%s' | gzip | openssl enc -aes-256-cbc -salt -pbkdf2 -pass env:MYSQL_PWD -out %s/%s_%s.sql.gz.enc)`,
		rootPassword, dbName, backupDir, dbName, timestamp), Description: "Create encrypted backup"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdir.Command, "description", cmdMkdir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCheckOpenSSL.Command, "description", cmdCheckOpenSSL.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdBackup.Command, "description", cmdBackup.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create encrypted backup for database '%s'", dbName),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating encrypted database backup", "database", dbName)
	_, _ = ssh.Run(cfg, cmdMkdir)
	_, _ = ssh.Run(cfg, cmdCheckOpenSSL)

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

// NewBackupEncrypt creates a new mariadb-backup-encrypt skill.
func NewBackupEncrypt() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbBackupEncrypt)
	pb.SetDescription("Create an encrypted backup of a MariaDB database")
	return &BackupEncrypt{BaseSkill: pb}
}
