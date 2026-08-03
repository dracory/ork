package mariadb

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// BackupSchedule installs a daily automated MariaDB backup script and a
// systemd timer to run it. This ensures backups happen automatically without
// manual intervention, and stores them in a dedicated directory. It also
// configures /root/.my.cnf so mysqldump can authenticate without passing
// credentials on the command line.
//
// Usage:
//
//	node.Run(mariadb.NewBackupSchedule().SetRootPassword("<password>"))
//
// Execution Flow:
//  1. Creates the backup directory with restrictive permissions (0700, root:root)
//  2. Writes /root/.my.cnf so mysqldump authenticates without CLI flags (mode 0600)
//  3. Writes the backup script to /usr/local/bin/mariadb-backup.sh (mode 0700)
//  4. Writes the systemd oneshot service unit (sandboxed)
//  5. Writes the systemd timer unit (daily at 02:00 by default, Persistent=true)
//  6. Reloads systemd, then enables and starts the timer
//
// Args:
//   - root-password: MariaDB root password (required)
//   - port: MariaDB port written into /root/.my.cnf (default: 3306)
//   - backup-dir: Directory to store backups (default: /var/backups/mariadb)
//   - schedule: systemd OnCalendar expression (default: *-*-* 02:00:00)
//   - retention-days: Days to keep backups before pruning (default: 7)
//   - script-path: Path to the backup script (default: /usr/local/bin/mariadb-backup.sh)
//   - service-name: systemd unit basename without suffix (default: mariadb-backup)
//
// Security Notes:
//   - /root/.my.cnf is mode 0600, root:root — only root can read it
//   - The backup script is mode 0700, root:root
//   - The systemd service is sandboxed (ProtectSystem=strict, PrivateTmp, etc.)
//   - Backups are written mode 0600
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//   - systemd-based host
//
// Related Playbooks:
//   - mariadb-backup: Perform a one-shot backup
//   - mariadb-install: Install MariaDB server
type BackupSchedule struct {
	*types.BaseSkill
}

// Compile-time assertion that BackupSchedule implements types.RunnableInterface.
var _ types.RunnableInterface = (*BackupSchedule)(nil)

// Check always returns true since we always want to ensure the schedule is
// installed and up to date. The Run method is idempotent: overwriting the
// unit files and re-running systemctl enable/start are no-ops when unchanged.
func (m *BackupSchedule) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (m *BackupSchedule) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	port := m.GetArg(ArgPort)
	if port == "" {
		port = DefaultPort
	}

	backupDir := m.GetArg(ArgBackupDir)
	if backupDir == "" {
		backupDir = DefaultBackupScheduleDir
	}

	schedule := m.GetArg(ArgSchedule)
	if schedule == "" {
		schedule = DefaultBackupSchedule
	}

	retentionDays := m.GetArg(ArgRetentionDays)
	if retentionDays == "" {
		retentionDays = DefaultBackupRetentionDays
	}

	scriptPath := m.GetArg(ArgScriptPath)
	if scriptPath == "" {
		scriptPath = DefaultBackupScriptPath
	}

	serviceName := m.GetArg(ArgServiceName)
	if serviceName == "" {
		serviceName = DefaultBackupServiceName
	}

	// Validate args to prevent config/file injection
	if !regexp.MustCompile(`^\d+$`).MatchString(port) {
		return types.Result{
			Changed: false,
			Message: "Invalid port: must be numeric",
			Error:   fmt.Errorf("port must be a positive integer, got '%s'", port),
		}
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(retentionDays) {
		return types.Result{
			Changed: false,
			Message: "Invalid retention-days: must be numeric",
			Error:   fmt.Errorf("retention-days must be a positive integer, got '%s'", retentionDays),
		}
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(serviceName) {
		return types.Result{
			Changed: false,
			Message: "Invalid service-name: only alphanumeric, hyphens, and underscores allowed",
			Error:   fmt.Errorf("service-name must match ^[a-zA-Z0-9_-]+$, got '%s'", serviceName),
		}
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9 _:,*-]+$`).MatchString(schedule) {
		return types.Result{
			Changed: false,
			Message: "Invalid schedule: contains illegal characters",
			Error:   fmt.Errorf("schedule must match ^[a-zA-Z0-9 _:,*-]+$, got '%s'", schedule),
		}
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_/.-]+$`).MatchString(backupDir) {
		return types.Result{
			Changed: false,
			Message: "Invalid backup-dir: contains illegal characters",
			Error:   fmt.Errorf("backup-dir must match ^[a-zA-Z0-9_/.-]+$, got '%s'", backupDir),
		}
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_/.-]+$`).MatchString(scriptPath) {
		return types.Result{
			Changed: false,
			Message: "Invalid script-path: contains illegal characters",
			Error:   fmt.Errorf("script-path must match ^[a-zA-Z0-9_/.-]+$, got '%s'", scriptPath),
		}
	}

	serviceUnitPath := "/etc/systemd/system/" + serviceName + ".service"
	timerUnitPath := "/etc/systemd/system/" + serviceName + ".timer"
	timerUnit := serviceName + ".timer"

	// Build the /root/.my.cnf content. The password is written verbatim
	// because the heredoc uses a single-quoted delimiter ('EOF') which
	// prevents all shell expansion — no escaping is needed or correct.
	myCnfContent := fmt.Sprintf(`[client]
user=root
password=%s
host=127.0.0.1
port=%s
`, rootPassword, port)

	// Build the backup script content. The backup dir, retention, and script
	// path are interpolated here; the script itself reads /root/.my.cnf for
	// credentials at runtime.
	backupScript := fmt.Sprintf(`#!/bin/bash
# MariaDB automated backup script
# Managed by ork mariadb-backup-schedule skill — do not edit by hand

set -euo pipefail

BACKUP_DIR="%s"
DATE=$(date +%%Y%%m%%d_%%H%%M%%S)
BACKUP_FILE="${BACKUP_DIR}/mariadb_all_${DATE}.sql.gz"
RETENTION_DAYS=%s

# Dump all databases and compress.
# Credentials are read from /root/.my.cnf (configured by this skill).
# --defaults-extra-file ensures the .my.cnf is used even if mysqldump
# has other default config.
mysqldump --defaults-extra-file=/root/.my.cnf --all-databases --single-transaction --routines --triggers --events | gzip > "${BACKUP_FILE}"

# Set restrictive permissions
chmod 600 "${BACKUP_FILE}"

# Remove backups older than retention period
find "${BACKUP_DIR}" -name "mariadb_all_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

echo "Backup completed: ${BACKUP_FILE}"
`, backupDir, retentionDays)

	// Build the systemd service unit. ReadWritePaths must include the backup
	// dir and /root (for /root/.my.cnf reads — ProtectSystem=strict otherwise
	// blocks it).
	serviceUnit := fmt.Sprintf(`[Unit]
Description=MariaDB automated backup
After=mariadb.service
Requires=mariadb.service

[Service]
Type=oneshot
ExecStart=%s
User=root
Group=root
# Sandboxing
ProtectSystem=strict
PrivateTmp=true
PrivateDevices=true
NoNewPrivileges=true
ReadWritePaths=%s /root
`, scriptPath, backupDir)

	// Build the systemd timer unit.
	timerUnitContent := fmt.Sprintf(`[Unit]
Description=Daily MariaDB backup timer

[Timer]
OnCalendar=%s
Persistent=true

[Install]
WantedBy=timers.target
`, schedule)

	// Define commands. We use single-quoted heredocs ('EOF') so the body is
	// written verbatim — no shell expansion of $ or backticks in the content.
	// The backup script body uses literal $VAR references that must survive
	// the write intact, so it MUST use a quoted delimiter.
	cmdMkdirBackupDir := types.Command{
		Command:     fmt.Sprintf("mkdir -p '%s' && chmod 700 '%s' && chown root:root '%s'", mariadbEscapeShellQuote(backupDir), mariadbEscapeShellQuote(backupDir), mariadbEscapeShellQuote(backupDir)),
		Description: "Create backup directory",
	}
	cmdWriteMyCnf := types.Command{
		Command:     fmt.Sprintf("umask 077 && cat > /root/.my.cnf <<'EOF'\n%sEOF\nchmod 600 /root/.my.cnf && chown root:root /root/.my.cnf", myCnfContent),
		Description: "Write /root/.my.cnf for backup authentication",
		Sensitive:   true,
	}
	cmdWriteScript := types.Command{
		Command:     fmt.Sprintf("cat > '%s' <<'EOF'\n%sEOF\nchmod 700 '%s' && chown root:root '%s'", mariadbEscapeShellQuote(scriptPath), backupScript, mariadbEscapeShellQuote(scriptPath), mariadbEscapeShellQuote(scriptPath)),
		Description: "Write backup script",
	}
	cmdWriteService := types.Command{
		Command:     fmt.Sprintf("cat > '%s' <<'EOF'\n%sEOF\nchmod 644 '%s' && chown root:root '%s'", mariadbEscapeShellQuote(serviceUnitPath), serviceUnit, mariadbEscapeShellQuote(serviceUnitPath), mariadbEscapeShellQuote(serviceUnitPath)),
		Description: "Write systemd service unit",
	}
	cmdWriteTimer := types.Command{
		Command:     fmt.Sprintf("cat > '%s' <<'EOF'\n%sEOF\nchmod 644 '%s' && chown root:root '%s'", mariadbEscapeShellQuote(timerUnitPath), timerUnitContent, mariadbEscapeShellQuote(timerUnitPath), mariadbEscapeShellQuote(timerUnitPath)),
		Description: "Write systemd timer unit",
	}
	cmdDaemonReload := types.Command{
		Command:     "systemctl daemon-reload",
		Description: "Reload systemd to pick up new unit files",
	}
	cmdEnableStart := types.Command{
		Command:     fmt.Sprintf("systemctl enable '%s' && systemctl start '%s'", mariadbEscapeShellQuote(timerUnit), mariadbEscapeShellQuote(timerUnit)),
		Description: "Enable and start the backup timer",
	}

	// Dry-run mode: log what would run, return without executing.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdirBackupDir.Command, "description", cmdMkdirBackupDir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", "[redacted]", "description", cmdWriteMyCnf.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWriteScript.Command, "description", cmdWriteScript.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWriteService.Command, "description", cmdWriteService.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWriteTimer.Command, "description", cmdWriteTimer.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDaemonReload.Command, "description", cmdDaemonReload.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdEnableStart.Command, "description", cmdEnableStart.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would install MariaDB backup schedule (daily '%s', %s-day retention, %s)", schedule, retentionDays, backupDir),
		}
	}

	cfg.GetLoggerOrDefault().Info("installing MariaDB backup schedule", "schedule", schedule, "backup_dir", backupDir)

	// 1. Create backup directory
	if out, err := ssh.Run(cfg, cmdMkdirBackupDir); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create backup directory",
			Error:   fmt.Errorf("failed to create backup directory: %w\nOutput: %s", err, out),
		}
	}

	// 2. Write /root/.my.cnf
	if out, err := ssh.Run(cfg, cmdWriteMyCnf); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write /root/.my.cnf for backup authentication",
			Error:   fmt.Errorf("failed to write /root/.my.cnf: %w\nOutput: %s", err, out),
		}
	}

	// 3. Write the backup script
	if out, err := ssh.Run(cfg, cmdWriteScript); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write backup script",
			Error:   fmt.Errorf("failed to write backup script: %w\nOutput: %s", err, out),
		}
	}

	// 4. Write the systemd service unit
	if out, err := ssh.Run(cfg, cmdWriteService); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write backup service unit",
			Error:   fmt.Errorf("failed to write backup service unit: %w\nOutput: %s", err, out),
		}
	}

	// 5. Write the systemd timer unit
	if out, err := ssh.Run(cfg, cmdWriteTimer); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write backup timer unit",
			Error:   fmt.Errorf("failed to write backup timer unit: %w\nOutput: %s", err, out),
		}
	}

	// 6. Reload systemd to pick up the new unit files
	if out, err := ssh.Run(cfg, cmdDaemonReload); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to reload systemd for backup timer",
			Error:   fmt.Errorf("failed to reload systemd: %w\nOutput: %s", err, out),
		}
	}

	// 7. Enable and start the timer (idempotent: no-ops if already enabled/active)
	if out, err := ssh.Run(cfg, cmdEnableStart); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable MariaDB backup timer",
			Error:   fmt.Errorf("failed to enable/start backup timer: %w\nOutput: %s", err, out),
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("MariaDB automated backup scheduled (daily '%s', %s-day retention, %s, /root/.my.cnf configured)", schedule, retentionDays, backupDir),
		Details: map[string]string{
			"backup_dir":     backupDir,
			"schedule":       schedule,
			"retention_days": retentionDays,
			"script_path":    scriptPath,
			"timer_unit":     timerUnit,
		},
	}
}

// SetArgs sets the arguments for this skill and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetArgs(args map[string]string) types.RunnableInterface {
	m.BaseSkill.SetArgs(args)
	return m
}

// SetRootPassword sets the MariaDB root password and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetRootPassword(password string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgRootPassword, password)
	return m
}

// SetPort sets the MariaDB port written into /root/.my.cnf and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetPort(port string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgPort, port)
	return m
}

// SetBackupDir sets the backup directory and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetBackupDir(dir string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgBackupDir, dir)
	return m
}

// SetSchedule sets the systemd OnCalendar expression (e.g. "*-*-* 02:00:00")
// and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetSchedule(schedule string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgSchedule, schedule)
	return m
}

// SetRetentionDays sets the backup retention period in days and returns
// BackupSchedule for chaining.
func (m *BackupSchedule) SetRetentionDays(days string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgRetentionDays, days)
	return m
}

// SetScriptPath sets the path to the backup script and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetScriptPath(path string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgScriptPath, path)
	return m
}

// SetServiceName sets the systemd unit basename (without .service/.timer suffix)
// and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetServiceName(name string) *BackupSchedule {
	m.BaseSkill.SetArg(ArgServiceName, name)
	return m
}

// SetArg sets a single argument and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetArg(key, value string) types.RunnableInterface {
	m.BaseSkill.SetArg(key, value)
	return m
}

// SetID sets the ID for this skill and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetID(id string) types.RunnableInterface {
	m.BaseSkill.SetID(id)
	return m
}

// SetDescription sets the description for this skill and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetDescription(description string) types.RunnableInterface {
	m.BaseSkill.SetDescription(description)
	return m
}

// SetTimeout sets the timeout for this skill and returns BackupSchedule for chaining.
func (m *BackupSchedule) SetTimeout(timeout time.Duration) types.RunnableInterface {
	m.BaseSkill.SetTimeout(timeout)
	return m
}

// NewBackupSchedule creates a new mariadb-backup-schedule skill.
func NewBackupSchedule() *BackupSchedule {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbBackupSchedule)
	pb.SetDescription("Install automated daily MariaDB backup with systemd timer")
	return &BackupSchedule{BaseSkill: pb}
}
