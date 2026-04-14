// Package user provides playbooks for managing Linux user accounts.
// It supports creating users with SSH key authentication, deleting users,
// and querying user status and group membership.
package user

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UserStatus shows user information.
// This is a read-only playbook that queries user account details.
// It can show information for a specific user (via username argument)
// or list all non-system users on the system.
//
// Usage (specific user):
//
//	go run . --playbook=user-status --arg=username=<name>
//
// Usage (all users):
//
//	go run . --playbook=user-status
//
// Arguments:
//   - username: Specific user to query (optional, omit to list all users)
//
// Execution Flow (with username):
//  1. Runs id <username> to get user info
//  2. Runs groups <username> to get group membership
//  3. Reports user info and groups
//
// Execution Flow (without username):
//  1. Parses /etc/passwd to find non-system users (UID >= 1000, < 65534)
//  2. Reports list of all regular user accounts
//
// Expected Output:
//   - Success (specific user): "User info for '<username>'" with uid/gid info
//   - Success (all users): "Non-system users listed" with user list
//   - Failure (user not found): Error indicating user doesn't exist
//
// Result Details (specific user):
//   - info: Output from id command (uid, gid, groups)
//   - groups: Output from groups command
//
// Result Details (all users):
//   - users: List of usernames (one per line)
//
// Use Cases:
//   - Verify user creation/deletion
//   - Check group membership for sudo access
//   - Audit user accounts on system
//   - Troubleshoot permission issues
//
// Idempotency:
//   - Always reports Changed=false since this is read-only
type UserStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since UserStatus is read-only.
// Per the playbook interface convention, the bool return indicates whether
// the operation would modify system state. Since user-status only queries
// user information, this always returns false.
func (u *UserStatus) Check() (bool, error) {
	return false, nil
}

// Run displays user status and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// If ArgUsername is provided, returns details for that specific user.
// If ArgUsername is empty, lists all non-system users (UID >= 1000).
//
// Result.Details (specific user) contains:
//   - info: Output from id command (UID, GID, group membership)
//   - groups: Output from groups command
//
// Result.Details (all users) contains:
//   - users: List of all non-system usernames (one per line)
func (u *UserStatus) Run() playbook.Result {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)
	cmdListAll := types.Command{Command: "awk -F: '$3 >= 1000 && $3 < 65534 {print $1}' /etc/passwd", Description: "List all non-system users"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		if username != "" {
			cmdID := types.Command{Command: fmt.Sprintf("id %s", username), Description: "Get user info"}
			cmdGroups := types.Command{Command: fmt.Sprintf("groups %s", username), Description: "Get user groups"}
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdID.Command)
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGroups.Command)
			return playbook.Result{
				Changed: false,
				Message: fmt.Sprintf("Would check user status for '%s'", username),
			}
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdListAll.Command)
		return playbook.Result{
			Changed: false,
			Message: "Would list all system users",
		}
	}

	if username != "" {
		// Check specific user
		cfg.GetLoggerOrDefault().Info("checking user", "username", username)

		cmdID := types.Command{Command: fmt.Sprintf("id %s", username), Description: "Get user info"}
		output, err := ssh.Run(cfg, cmdID)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: fmt.Sprintf("User '%s' does not exist", username),
				Error:   fmt.Errorf("user '%s' not found", username),
			}
		}
		cfg.GetLoggerOrDefault().Info("user info", "output", output)

		// Check if user has sudo
		cmdGroups := types.Command{Command: fmt.Sprintf("groups %s", username), Description: "Get user groups"}
		groupsOutput, err := ssh.Run(cfg, cmdGroups)
		if err == nil {
			cfg.GetLoggerOrDefault().Info("user groups", "groups", groupsOutput)
		}

		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("User info for '%s'", username),
			Details: map[string]string{"info": output, "groups": groupsOutput},
		}
	}

	// List all non-system users
	cfg.GetLoggerOrDefault().Info("listing all system users")

	output, err := ssh.Run(cfg, cmdListAll)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list users",
			Error:   fmt.Errorf("failed to list users: %w", err),
		}
	}

	if output == "" {
		cfg.GetLoggerOrDefault().Info("no non-system users found")
		return playbook.Result{
			Changed: false,
			Message: "No non-system users found",
		}
	}

	cfg.GetLoggerOrDefault().Info("users found", "users", output)
	return playbook.Result{
		Changed: false,
		Message: "Non-system users listed",
		Details: map[string]string{"users": output},
	}
}

// NewUserStatus creates a new user-status playbook.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDUserStatus identifier
//	and description "Show user information".
//
// Usage Note:
//
//	Pass ArgUsername via --arg=username=<name> to query a specific user.
//	Omit the username argument to list all non-system users.
func NewUserStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUserStatus)
	pb.SetDescription("Show user information")
	return &UserStatus{BasePlaybook: pb}
}
