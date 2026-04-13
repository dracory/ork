package playbook

// Playbook ID constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook IDs.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(playbook.IDPing)
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

	// IDUserStatus shows user info (accepts optional "username" arg)
	IDUserStatus = "user-status"

	// IDFail2banInstall installs fail2ban intrusion prevention
	IDFail2banInstall = "fail2ban-install"

	// IDFail2banStatus shows fail2ban service and jail status
	IDFail2banStatus = "fail2ban-status"
)
