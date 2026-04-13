package user

// Package user documentation is in status.go

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// UserDelete removes a user.
type UserDelete struct {
	*playbook.BasePlaybook
}

// Check determines if user exists and can be deleted.
// Returns true if user exists, false if user doesn't exist.
func (u *UserDelete) Check() (bool, error) {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
	return strings.Contains(output, username), nil
}

// Run removes a non-root system user.
// This playbook safely deletes user accounts while protecting critical system users
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
func (u *UserDelete) Run() playbook.Result {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)
	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	// Safety check: prevent deletion of root user
	if username == "root" {
		return playbook.Result{
			Changed: false,
			Message: "Cannot delete root user",
			Error:   fmt.Errorf("deletion of root user is not allowed for safety reasons"),
		}
	}

	log.Printf("Deleting user '%s'...", username)

	// Delete user and home directory (try -r first, then without)
	cmd := fmt.Sprintf("userdel -r %s 2>/dev/null || userdel %s", username, username)
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to delete user",
			Error:   fmt.Errorf("failed to delete user: %w\nOutput: %s", err, output),
		}
	}

	log.Printf("User '%s' deleted successfully", username)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' deleted", username),
	}
}

// NewUserDelete creates a new user-delete playbook.
func NewUserDelete() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUserDelete)
	pb.SetDescription("Delete a user (username via args['username'])")
	return &UserDelete{BasePlaybook: pb}
}
