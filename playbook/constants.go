package playbook

// Playbook name constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook names.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(playbook.NamePing)
const (
	// NamePing checks SSH connectivity
	NamePing = "ping"

	// NameAptUpdate refreshes the package database
	NameAptUpdate = "apt-update"

	// NameAptUpgrade installs available updates
	NameAptUpgrade = "apt-upgrade"

	// NameAptStatus shows available updates
	NameAptStatus = "apt-status"

	// NameReboot reboots the server
	NameReboot = "reboot"

	// NameSwapCreate creates a swap file (requires "size" arg in GB)
	NameSwapCreate = "swap-create"

	// NameSwapDelete removes the swap file
	NameSwapDelete = "swap-delete"

	// NameSwapStatus shows swap status
	NameSwapStatus = "swap-status"

	// NameUserCreate creates a user with sudo (requires "username" arg)
	NameUserCreate = "user-create"

	// NameUserDelete deletes a user (requires "username" arg)
	NameUserDelete = "user-delete"

	// NameUserStatus shows user info (accepts optional "username" arg)
	NameUserStatus = "user-status"
)
