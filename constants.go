package ork

import "github.com/dracory/ork/playbook"

// Playbook ID constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook IDs.
// They are aliases to the constants in the playbook package.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(ork.PlaybookPing)
const (
	// PlaybookPing checks SSH connectivity
	PlaybookPing = playbook.IDPing

	// PlaybookAptUpdate refreshes the package database
	PlaybookAptUpdate = playbook.IDAptUpdate

	// PlaybookAptUpgrade installs available updates
	PlaybookAptUpgrade = playbook.IDAptUpgrade

	// PlaybookAptStatus shows available updates
	PlaybookAptStatus = playbook.IDAptStatus

	// PlaybookReboot reboots the server
	PlaybookReboot = playbook.IDReboot

	// PlaybookSwapCreate creates a swap file (requires "size" arg in GB)
	PlaybookSwapCreate = playbook.IDSwapCreate

	// PlaybookSwapDelete removes the swap file
	PlaybookSwapDelete = playbook.IDSwapDelete

	// PlaybookSwapStatus shows swap status
	PlaybookSwapStatus = playbook.IDSwapStatus

	// PlaybookUserCreate creates a user with sudo (requires "username" arg)
	PlaybookUserCreate = playbook.IDUserCreate

	// PlaybookUserDelete deletes a user (requires "username" arg)
	PlaybookUserDelete = playbook.IDUserDelete

	// PlaybookUserStatus shows user info (accepts optional "username" arg)
	PlaybookUserStatus = playbook.IDUserStatus
)
