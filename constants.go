package ork

import "github.com/dracory/ork/skills"

// Skill ID constants for use with RunSkill.
// These constants provide compile-time safety and IDE autocomplete for skill IDs.
// They are aliases to the constants in the skills package.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunSkill(ork.SkillPing)
const (
	// SkillPing checks SSH connectivity
	SkillPing = skills.IDPing

	// SkillAptUpdate refreshes the package database
	SkillAptUpdate = skills.IDAptUpdate

	// SkillAptUpgrade installs available updates
	SkillAptUpgrade = skills.IDAptUpgrade

	// SkillAptStatus shows available updates
	SkillAptStatus = skills.IDAptStatus

	// SkillReboot reboots the server
	SkillReboot = skills.IDReboot

	// SkillSwapCreate creates a swap file (requires "size" arg in GB)
	SkillSwapCreate = skills.IDSwapCreate

	// SkillSwapDelete removes the swap file
	SkillSwapDelete = skills.IDSwapDelete

	// SkillSwapStatus shows swap status
	SkillSwapStatus = skills.IDSwapStatus

	// SkillUserCreate creates a user with sudo (requires "username" arg)
	SkillUserCreate = skills.IDUserCreate

	// SkillUserDelete deletes a user (requires "username" arg)
	SkillUserDelete = skills.IDUserDelete

	// SkillUserList lists all non-system users
	SkillUserList = skills.IDUserList

	// SkillUserStatus shows user info (requires "username" arg)
	SkillUserStatus = skills.IDUserStatus
)
