package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// Backup creates a compressed SQL dump of a MariaDB database.
// This playbook generates timestamped backups using mysqldump with transaction
// consistency, then compresses the output with gzip to save disk space.
//
// Usage:
//
//	go run . --playbook=mariadb-backup --arg=db-name=<database_name> [--arg=backup-dir=/path/to/backups]
//
// Args:
//   - db-name: Name of the database to backup (required)
//   - root-password: MariaDB root password (required)
//   - backup-dir: Directory to store backup file (default: /root/backups)
//
// Backup Options Used:
//   - --single-transaction: Consistent backup without locking (InnoDB only)
//   - --routines: Include stored procedures and functions
//   - --triggers: Include table triggers
//
// Output Format:
//   - Filename: {dbname}_{timestamp}.sql.gz
//   - Location: Specified directory (default: /root/backups)
//   - Compression: gzip (typically 70-90% size reduction)
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-list-dbs: List available databases
type Backup struct {
	*playbook.BasePlaybook
}

// Check always returns true since we always want to create a fresh backup.
func (m *Backup) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *Backup) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	dbName := m.GetArg(ArgDbName)

	if dbName == "" {
		return playbook.Result{
			Changed: false,
			Message: "Database name is required",
			Error:   fmt.Errorf("db-name argument is required"),
		}
	}

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	backupDir := m.GetArg(ArgBackupDir)
	if backupDir == "" {
		backupDir = "/root/backups"
	}
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("%s_%s.sql", dbName, timestamp)

	cfg.GetLoggerOrDefault().Info("creating database backup", "database", dbName)

	// Create backup directory
	cmd := fmt.Sprintf("mkdir -p %s", backupDir)
	_, _ = ssh.Run(cfg, cmd)

	// Create backup
	backupPath := fmt.Sprintf("%s/%s", backupDir, backupFile)
	cmd = fmt.Sprintf(`mysqldump -u root -p"%s" --single-transaction --routines --triggers "%s" > "%s"`,
		rootPassword, dbName, backupPath)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create backup",
			Error:   fmt.Errorf("failed to create backup: %w\nOutput: %s", err, output),
		}
	}

	// Compress backup
	cmd = fmt.Sprintf("gzip -f %s", backupPath)
	_, _ = ssh.Run(cfg, cmd)

	// Generate checksum
	cmd = fmt.Sprintf("sha256sum %s.gz > %s.gz.sha256", backupPath, backupPath)
	_, _ = ssh.Run(cfg, cmd)

	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("Backup created: %s.gz", backupPath),
		Details: map[string]string{
			"backup_path": fmt.Sprintf("%s.gz", backupPath),
			"checksum":    fmt.Sprintf("%s.gz.sha256", backupPath),
		},
	}
}

// NewBackup creates a new mariadb-backup playbook.
func NewBackup() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbBackup)
	pb.SetDescription("Create a compressed SQL dump of a MariaDB database")
	return &Backup{BasePlaybook: pb}
}
