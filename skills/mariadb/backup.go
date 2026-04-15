package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Backup creates a compressed SQL dump of a MariaDB database.
// This skill generates timestamped backups using mysqldump with transaction
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
	*types.BaseSkill
}

// Check always returns true since we always want to create a fresh backup.
func (m *Backup) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (m *Backup) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	dbName := m.GetArg(ArgDbName)

	if dbName == "" {
		return types.Result{
			Changed: false,
			Message: "Database name is required",
			Error:   fmt.Errorf("db-name argument is required"),
		}
	}

	if rootPassword == "" {
		return types.Result{
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

	backupPath := fmt.Sprintf("%s/%s", backupDir, backupFile)

	// Define commands
	cmdMkdir := types.Command{Command: fmt.Sprintf("mkdir -p %s", backupDir), Description: "Create backup directory"}
	cmdDump := types.Command{Command: fmt.Sprintf(`mysqldump -u root -p"%s" --single-transaction --routines --triggers "%s" > "%s"`,
		rootPassword, dbName, backupPath), Description: "Create database backup"}
	cmdCompress := types.Command{Command: fmt.Sprintf("gzip -f %s", backupPath), Description: "Compress backup"}
	cmdChecksum := types.Command{Command: fmt.Sprintf("sha256sum %s.gz > %s.gz.sha256", backupPath, backupPath), Description: "Generate backup checksum"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdir.Command, "description", cmdMkdir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDump.Command, "description", cmdDump.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCompress.Command, "description", cmdCompress.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChecksum.Command, "description", cmdChecksum.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create backup for database '%s'", dbName),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating database backup", "database", dbName)

	// Create backup directory
	_, _ = ssh.Run(cfg, cmdMkdir)

	// Create backup
	output, err := ssh.Run(cfg, cmdDump)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create backup",
			Error:   fmt.Errorf("failed to create backup: %w\nOutput: %s", err, output),
		}
	}

	// Compress backup
	_, _ = ssh.Run(cfg, cmdCompress)

	// Generate checksum
	_, _ = ssh.Run(cfg, cmdChecksum)

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Backup created: %s.gz", backupPath),
		Details: map[string]string{
			"backup_path": fmt.Sprintf("%s.gz", backupPath),
			"checksum":    fmt.Sprintf("%s.gz.sha256", backupPath),
		},
	}
}

// NewBackup creates a new mariadb-backup skill.
func NewBackup() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbBackup)
	pb.SetDescription("Create a compressed SQL dump of a MariaDB database")
	return &Backup{BaseSkill: pb}
}
