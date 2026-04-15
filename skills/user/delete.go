package user

// Package user documentation is in status.go

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UserDelete removes a user.
type UserDelete struct {
	*skills.BaseSkill
}

// Check determines if user exists and can be deleted.
// Returns true if user exists, false if user doesn't exist.
func (u *UserDelete) Check() (bool, error) {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	cmdCheck := types.Command{Command: fmt.Sprintf("id %s", username), Description: "Check if user exists"}
	output, _ := ssh.Run(cfg, cmdCheck)
	return strings.Contains(output, username), nil
}

// Run removes a non-root system user.
// This skill safely deletes user accounts while protecting critical system users
// like root from accidental deletion.
//
// Usage:
//
//	go run . --playbook=user-delete [--arg=username=<name>]
//
// Parameters (passed via args):
//   - username (string, required): Name of the user to delete (cannot be 'root')
//
// Execution Flow:
//  1. Validates username is provided
//  2. Prevents deletion of root user (safety check)
//  3. Attempts deluser first (Debian/Ubuntu)
//  4. Falls back to userdel if deluser fails
//  5. Reports success or failure
//
// Safety Features:
//   - Explicitly prevents deletion of the root user
//   - Username must be provided (no default)
//   - Tries deluser first for Debian/Ubuntu compatibility
//
// Prerequisites:
//   - Root SSH access required
//   - User must exist on the system
//
// Args:
//   - username (string, required): Username to delete
func (u *UserDelete) Run() types.Result {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)
	if username == "" {
		return types.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	// Safety check: prevent deletion of root user
	if username == "root" {
		return types.Result{
			Changed: false,
			Message: "Cannot delete root user",
			Error:   fmt.Errorf("deletion of root user is not allowed for safety reasons"),
		}
	}

	cfg.GetLoggerOrDefault().Info("deleting user", "username", username)

	// Delete user and home directory (try -r first, then without)
	cmdDelete := types.Command{Command: fmt.Sprintf("userdel -r %s 2>/dev/null || userdel %s", username, username), Description: "Delete user"}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDelete.Command)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would delete user: %s", username),
		}
	}

	output, err := ssh.Run(cfg, cmdDelete)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to delete user",
			Error:   fmt.Errorf("failed to delete user: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("user deleted", "username", username)
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' deleted", username),
	}
}

// NewUserDelete creates a new user-delete skill.
func NewUserDelete() types.RunnableInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDUserDelete)
	pb.SetDescription("Delete a user (username via args['username'])")
	return &UserDelete{BaseSkill: pb}
}
