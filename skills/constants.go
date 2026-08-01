package skills

// DebianNonInteractive prevents dpkg from prompting for user input during
// package operations (e.g. the "What do you want to do about modified
// configuration file?" prompt), which would hang the SSH session.
const DebianNonInteractive = "DEBIAN_FRONTEND=noninteractive"

// DpkgConfDef uses the package maintainer's default config when the user
// hasn't modified it locally. Matches Ansible's apt module default.
const DpkgConfDef = " -o Dpkg::Options::=\"--force-confdef\""

// DpkgConfOld keeps the locally modified config when the user has changed it.
// Matches Ansible's apt module default.
const DpkgConfOld = " -o Dpkg::Options::=\"--force-confold\""

// DpkgConfOptions combines DpkgConfDef and DpkgConfOld for convenience.
// This matches Ansible's apt module defaults for unattended package operations.
//
// Usage: append to any apt-get command string, e.g.:
//
//	cmdStr += skills.DpkgConfOptions
const DpkgConfOptions = DpkgConfDef + DpkgConfOld

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

	// IDAptInstall installs one or more packages (requires "packages" arg)
	IDAptInstall = "apt-install"

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

	// IDUserAddToGroup adds a user to a supplementary group (requires "username" and "group" args)
	IDUserAddToGroup = "user-add-to-group"

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

	// IDUfwAllow configures UFW to allow a port
	IDUfwAllow = "ufw-allow"

	// IDUfwDeny denies traffic on a port
	IDUfwDeny = "ufw-deny"

	// IDUfwDelete removes a UFW rule by number
	IDUfwDelete = "ufw-delete"

	// IDUfwEnable enables the UFW firewall
	IDUfwEnable = "ufw-enable"

	// IDUfwDisable disables the UFW firewall
	IDUfwDisable = "ufw-disable"

	// IDUfwReset resets UFW to factory defaults
	IDUfwReset = "ufw-reset"

	// IDUfwDefault sets UFW default policies
	IDUfwDefault = "ufw-default"

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

	// Filesystem skills (general-purpose primitives)
	// IDFSDirCreate creates a directory with ownership and permissions
	IDFSDirCreate = "fs-dir-create"

	// IDFSDirExists checks if a directory exists (read-only)
	IDFSDirExists = "fs-dir-exists"

	// IDFSDirDelete deletes a directory
	IDFSDirDelete = "fs-dir-delete"

	// IDFSFileCreate creates a file with content, ownership, and permissions
	IDFSFileCreate = "fs-file-create"

	// IDFSFileExists checks if a file exists (read-only)
	IDFSFileExists = "fs-file-exists"

	// IDFSFileDelete deletes a single file
	IDFSFileDelete = "fs-file-delete"

	// IDFSFileCopy copies a file on the remote server
	IDFSFileCopy = "fs-file-copy"

	// IDFSChangeOwner changes file/directory ownership (chown)
	IDFSChangeOwner = "fs-change-owner"

	// IDFSChangeMode changes file/directory permissions (chmod)
	IDFSChangeMode = "fs-change-mode"

	// IDFSSymlinkCreate creates or updates a symbolic link (ln -sf)
	IDFSSymlinkCreate = "fs-symlink-create"

	// IDFSRename renames/moves a file or directory (mv)
	IDFSRename = "fs-rename"

	// IDFSRemove removes a file or directory (rm)
	IDFSRemove = "fs-remove"

	// PHP skills
	// IDPhpInstall installs PHP with extensions and configures FPM
	IDPhpInstall = "php-install"

	// IDPhpUninstall removes PHP packages and FPM configuration
	IDPhpUninstall = "php-uninstall"

	// IDPhpInstallComposer installs Composer (PHP dependency manager)
	IDPhpInstallComposer = "php-install-composer"

	// IDPhpUninstallComposer removes Composer binary
	IDPhpUninstallComposer = "php-uninstall-composer"

	// IDPhpUpdateComposer updates Composer to the latest version
	IDPhpUpdateComposer = "php-update-composer"

	// Systemctl skills (systemd unit management)
	// IDSystemctlDaemonReload reloads the systemd manager configuration
	IDSystemctlDaemonReload = "systemctl-daemon-reload"

	// IDSystemctlRestart restarts a systemd unit (requires "service" arg)
	IDSystemctlRestart = "systemctl-restart"

	// IDSystemctlReload reloads a systemd unit, falling back to restart
	// (requires "service" arg)
	IDSystemctlReload = "systemctl-reload"

	// IDSystemctlStatus shows the status of a systemd unit (read-only,
	// requires "service" arg)
	IDSystemctlStatus = "systemctl-status"

	// IDSystemctlIsActive checks whether a systemd unit is active (read-only,
	// requires "service" arg)
	IDSystemctlIsActive = "systemctl-is-active"

	// IDSystemctlEnable enables a systemd unit, optionally starting it
	// (requires "service" arg, optional "start" arg)
	IDSystemctlEnable = "systemctl-enable"

	// IDSystemctlDisable disables a systemd unit, optionally stopping it
	// (requires "service" arg, optional "stop" arg)
	IDSystemctlDisable = "systemctl-disable"
)
