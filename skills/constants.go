package skills

// Skill ID constants for use with RunSkill.
// These constants provide compile-time safety and IDE autocomplete for skill IDs.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.Run(skills.IDPing)
const (
	// IDPing checks SSH connectivity
	IDPing = "ping"

	// IDAptUpdate refreshes the package database
	IDAptUpdate = "apt-update"

	// IDAptUpgrade installs available updates
	IDAptUpgrade = "apt-upgrade"

	// IDAptStatus shows available updates
	IDAptStatus = "apt-status"

	// IDReboot reboots the server
	IDReboot = "reboot"

	// IDSwapCreate creates a swap file (requires "size" arg in GB)
	IDSwapCreate = "swap-create"

	// IDSwapDelete removes the swap file
	IDSwapDelete = "swap-delete"

	// IDSwapStatus shows swap status
	IDSwapStatus = "swap-status"

	// IDUserCreate creates a user with sudo (requires "username" arg)
	IDUserCreate = "user-create"

	// IDUserDelete deletes a user (requires "username" arg)
	IDUserDelete = "user-delete"

	// IDUserList lists all non-system users
	IDUserList = "user-list"

	// IDUserStatus shows user info (requires "username" arg)
	IDUserStatus = "user-status"

	// IDFail2banInstall installs fail2ban intrusion prevention
	IDFail2banInstall = "fail2ban-install"

	// IDFail2banStatus shows fail2ban service and jail status
	IDFail2banStatus = "fail2ban-status"

	// IDUfwInstall installs and configures UFW firewall
	IDUfwInstall = "ufw-install"

	// IDUfwStatus checks UFW firewall status
	IDUfwStatus = "ufw-status"

	// IDUfwAllowMariaDB configures UFW for MariaDB access
	IDUfwAllowMariaDB = "ufw-allow-mariadb"

	// IDSshHarden applies security hardening to SSH server configuration
	IDSshHarden = "ssh-harden"

	// IDKernelHarden applies security-focused kernel parameters
	IDKernelHarden = "kernel-harden"

	// IDAideInstall installs and configures AIDE file integrity monitoring
	IDAideInstall = "aide-install"

	// IDAuditdInstall installs and configures the Linux Audit Framework
	IDAuditdInstall = "auditd-install"

	// IDSshChangePort changes the SSH port to reduce automated scanning
	IDSshChangePort = "ssh-change-port"

	// MariaDB skills
	// IDMariadbInstall installs and configures MariaDB database server
	IDMariadbInstall = "mariadb-install"

	// IDMariadbSecure performs security hardening on MariaDB
	IDMariadbSecure = "mariadb-secure"

	// IDMariadbCreateDB creates a new MariaDB database
	IDMariadbCreateDB = "mariadb-create-db"

	// IDMariadbCreateUser creates a new MariaDB user
	IDMariadbCreateUser = "mariadb-create-user"

	// IDMariadbStatus displays MariaDB server status
	IDMariadbStatus = "mariadb-status"

	// IDMariadbListDBs displays all databases
	IDMariadbListDBs = "mariadb-list-dbs"

	// IDMariadbListUsers displays all users
	IDMariadbListUsers = "mariadb-list-users"

	// IDMariadbBackup creates a compressed SQL dump
	IDMariadbBackup = "mariadb-backup"

	// IDMariadbSecurityAudit performs security audit
	IDMariadbSecurityAudit = "mariadb-security-audit"

	// IDMariadbChangePort changes MariaDB port
	IDMariadbChangePort = "mariadb-change-port"

	// IDMariadbEnableSSL enables SSL/TLS encryption
	IDMariadbEnableSSL = "mariadb-enable-ssl"

	// IDMariadbEnableEncryption enables data-at-rest encryption
	IDMariadbEnableEncryption = "mariadb-enable-encryption"

	// IDMariadbBackupEncrypt creates an encrypted backup
	IDMariadbBackupEncrypt = "mariadb-backup-encrypt"
)
