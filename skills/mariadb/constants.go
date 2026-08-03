// Package mariadb provides playbooks for managing MariaDB database servers.
// These playbooks handle installation, configuration, user management,
// database operations, security hardening, and backup operations.
package mariadb

// Argument key constants for use with GetArg.
const (
	// ArgRootPassword specifies the MariaDB root password
	ArgRootPassword = "root-password"

	// ArgDbName specifies the database name for create/delete operations
	ArgDbName = "db-name"

	// ArgUsername specifies the username for user operations
	ArgUsername = "username"

	// ArgPassword specifies the password for user operations
	ArgPassword = "password"

	// ArgHost specifies the host for user grants (default: localhost)
	ArgHost = "host"

	// ArgPrivileges specifies the privileges to grant (default: ALL PRIVILEGES)
	ArgPrivileges = "privileges"

	// ArgPort specifies the port for MariaDB server (default: 3306)
	ArgPort = "port"

	// ArgBackupPath specifies the backup file path
	ArgBackupPath = "backup-path"

	// ArgSslCertPath specifies the SSL certificate path
	ArgSslCertPath = "ssl-cert-path"

	// ArgSslKeyPath specifies the SSL key path
	ArgSslKeyPath = "ssl-key-path"

	// ArgSslCaPath specifies the SSL CA certificate path
	ArgSslCaPath = "ssl-ca-path"

	// ArgBackupDir is the directory to store backups
	ArgBackupDir = "backup-dir"

	// ArgDBName is the name of the database
	ArgDBName = "dbname"

	// ArgDataDir specifies the MariaDB data directory
	ArgDataDir = "data-dir"

	// ArgConfigPath specifies the MariaDB server configuration file path
	ArgConfigPath = "config-path"

	// ArgKeyFilePath specifies the encryption key file path
	ArgKeyFilePath = "keyfile-path"

	// ArgSchedule specifies the systemd OnCalendar expression for the backup
	// timer (default: *-*-* 02:00:00, i.e. daily at 02:00)
	ArgSchedule = "schedule"

	// ArgRetentionDays specifies how many days to keep backups before pruning
	// (default: 7)
	ArgRetentionDays = "retention-days"

	// ArgScriptPath specifies the path to the backup script
	// (default: /usr/local/bin/mariadb-backup.sh)
	ArgScriptPath = "script-path"

	// ArgServiceName specifies the systemd unit basename (without .service or
	// .timer suffix) for the backup units (default: mariadb-backup)
	ArgServiceName = "service-name"
)

// Default values.
const (
	// DefaultBackupDir is the default backup directory
	DefaultBackupDir = "/root/backups"

	// DefaultPort is the default MariaDB port
	DefaultPort = "3306"

	// DefaultHost is the default host for user grants
	DefaultHost = "localhost"

	// DefaultPrivileges is the default privileges for user grants
	DefaultPrivileges = "ALL PRIVILEGES"

	// DefaultDataDir is the default MariaDB data directory
	DefaultDataDir = "/var/lib/mysql"

	// DefaultConfigPath is the default MariaDB server config path
	DefaultConfigPath = "/etc/mysql/mariadb.conf.d/50-server.cnf"

	// DefaultKeyFilePath is the default encryption key file path
	DefaultKeyFilePath = "/var/lib/mysql-keyfile/keyfile.enc"

	// DefaultBackupScheduleDir is the default directory used by the
	// backup-schedule skill to store automated backups
	DefaultBackupScheduleDir = "/var/backups/mariadb"

	// DefaultBackupSchedule is the default systemd OnCalendar expression for
	// the backup timer (daily at 02:00)
	DefaultBackupSchedule = "*-*-* 02:00:00"

	// DefaultBackupRetentionDays is the default number of days to keep backups
	DefaultBackupRetentionDays = "7"

	// DefaultBackupScriptPath is the default path to the backup script
	DefaultBackupScriptPath = "/usr/local/bin/mariadb-backup.sh"

	// DefaultBackupServiceName is the default systemd unit basename (without
	// .service/.timer suffix) for the backup units
	DefaultBackupServiceName = "mariadb-backup"
)
