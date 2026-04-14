package ork

import "github.com/dracory/ork/playbooks"

// Playbook ID constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook IDs.
// They are aliases to the constants in the playbooks package.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(ork.PlaybookPing)
const (
	// PlaybookPing checks SSH connectivity
	PlaybookPing = playbooks.IDPing

	// PlaybookAptUpdate refreshes the package database
	PlaybookAptUpdate = playbooks.IDAptUpdate

	// PlaybookAptUpgrade installs available updates
	PlaybookAptUpgrade = playbooks.IDAptUpgrade

	// PlaybookAptStatus shows available updates
	PlaybookAptStatus = playbooks.IDAptStatus

	// PlaybookReboot reboots the server
	PlaybookReboot = playbooks.IDReboot

	// PlaybookSwapCreate creates a swap file (requires "size" arg in GB)
	PlaybookSwapCreate = playbooks.IDSwapCreate

	// PlaybookSwapDelete removes the swap file
	PlaybookSwapDelete = playbooks.IDSwapDelete

	// PlaybookSwapStatus shows swap status
	PlaybookSwapStatus = playbooks.IDSwapStatus

	// PlaybookUserCreate creates a user with sudo (requires "username" arg)
	PlaybookUserCreate = playbooks.IDUserCreate

	// PlaybookUserDelete deletes a user (requires "username" arg)
	PlaybookUserDelete = playbooks.IDUserDelete

	// PlaybookUserList lists all non-system users
	PlaybookUserList = playbooks.IDUserList

	// PlaybookUserStatus shows user info (requires "username" arg)
	PlaybookUserStatus = playbooks.IDUserStatus
)
