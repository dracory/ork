package mariadb

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// BackupEncrypt creates an encrypted backup of a MariaDB database.
// This playbook extends the standard backup process by encrypting the backup
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
	*playbook.BasePlaybook
}

// Check determines if backup can be created.
func (b *BackupEncrypt) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (b *BackupEncrypt) Run() playbook.Result {
	cfg := b.GetNodeConfig()
	rootPassword := cfg.GetArg(ArgRootPassword)
	dbName := cfg.GetArg(ArgDBName)

	if dbName == "" {
		return playbook.Result{
			Changed: false,
			Message: "Database name is required",
			Error:   fmt.Errorf("use --arg=dbname=<database_name>"),
		}
	}

	// Validate database name
	validDBName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validDBName.MatchString(dbName) {
		return playbook.Result{
			Changed: false,
			Message: "Invalid database name",
			Error:   fmt.Errorf("only alphanumeric characters, underscores, and hyphens allowed"),
		}
	}

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	backupDir := cfg.GetArgOr(ArgBackupDir, "/root/backups")
	timestamp := time.Now().Format("20060102_150405")

	cfg.GetLoggerOrDefault().Info("creating encrypted database backup", "database", dbName)
	_, _ = ssh.Run(cfg, fmt.Sprintf(`mkdir -p %s`, backupDir))
	_, _ = ssh.Run(cfg, `which openssl || DEBIAN_FRONTEND=noninteractive apt-get install -y openssl`)

	cfg.GetLoggerOrDefault().Info("creating encrypted backup")
	cmd := fmt.Sprintf(`(umask 077 && MYSQL_PWD='%s' mysqldump -u root --single-transaction --routines --triggers --events '%s' | gzip | openssl enc -aes-256-cbc -salt -pbkdf2 -pass env:MYSQL_PWD -out %s/%s_%s.sql.gz.enc)`,
		rootPassword, dbName, backupDir, dbName, timestamp)
	_, err := ssh.Run(cfg, cmd)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to create backup", Error: err}
	}

	cfg.GetLoggerOrDefault().Info("encrypted backup complete")
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("Encrypted backup created: %s/%s_%s.sql.gz.enc", backupDir, dbName, timestamp),
		Details: map[string]string{
			"backup_path": fmt.Sprintf("%s/%s_%s.sql.gz.enc", backupDir, dbName, timestamp),
		},
	}
}

// NewBackupEncrypt creates a new mariadb-backup-encrypt playbook.
func NewBackupEncrypt() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbBackupEncrypt)
	pb.SetDescription("Create an encrypted backup of a MariaDB database")
	return &BackupEncrypt{BasePlaybook: pb}
}
