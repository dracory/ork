package ork

import "github.com/dracory/ork/playbook"

// Playbook name constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook names.
// They are aliases to the constants in the playbook package.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(ork.PlaybookPing)
const (
	// PlaybookPing checks SSH connectivity
	PlaybookPing = playbook.NamePing

	// PlaybookAptUpdate refreshes the package database
	PlaybookAptUpdate = playbook.NameAptUpdate

	// PlaybookAptUpgrade installs available updates
	PlaybookAptUpgrade = playbook.NameAptUpgrade

	// PlaybookAptStatus shows available updates
	PlaybookAptStatus = playbook.NameAptStatus

	// PlaybookReboot reboots the server
	PlaybookReboot = playbook.NameReboot

	// PlaybookSwapCreate creates a swap file (requires "size" arg in GB)
	PlaybookSwapCreate = playbook.NameSwapCreate

	// PlaybookSwapDelete removes the swap file
	PlaybookSwapDelete = playbook.NameSwapDelete

	// PlaybookSwapStatus shows swap status
	PlaybookSwapStatus = playbook.NameSwapStatus

	// PlaybookUserCreate creates a user with sudo (requires "username" arg)
	PlaybookUserCreate = playbook.NameUserCreate

	// PlaybookUserDelete deletes a user (requires "username" arg)
	PlaybookUserDelete = playbook.NameUserDelete

	// PlaybookUserStatus shows user info (accepts optional "username" arg)
	PlaybookUserStatus = playbook.NameUserStatus
)
