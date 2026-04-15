// Package user provides playbooks for managing Linux user accounts.
// It supports creating users with SSH key authentication, deleting users,
// and querying user status and group membership.
package user

import (
	"fmt"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UserStatus shows user information for a specific user.
// This is a read-only skill that queries user account details.
//
// Usage:
//
//	go run . --playbook=user-status --arg=username=<name>
//
// Arguments:
//   - username: Specific user to query (required)
//
// Execution Flow:
//  1. Runs id <username> to get user info
//  2. Runs groups <username> to get group membership
//  3. Reports user info and groups
//
// Expected Output:
//   - Success: "User info for '<username>'" with uid/gid info
//   - Failure: Error indicating user doesn't exist
//
// Result Details contains:
//   - info: Output from id command (uid, gid, groups)
//   - groups: Output from groups command
//
// Use Cases:
//   - Verify user creation/deletion
//   - Check group membership for sudo access
//   - Troubleshoot permission issues
//
// Idempotency:
//   - Always reports Changed=false since this is read-only
type UserStatus struct {
	*skills.BaseSkill
}

// Check always returns false since UserStatus is read-only.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since user-status only queries
// user information, this always returns false.
func (u *UserStatus) Check() (bool, error) {
	return false, nil
}

// Run displays user status and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// ArgUsername is required - returns details for that specific user.
//
// Result.Details contains:
//   - info: Output from id command (UID, GID, group membership)
//   - groups: Output from groups command
func (u *UserStatus) Run() types.Result {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)

	if username == "" {
		return types.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	// Define commands for specific user check
	cmdID := types.Command{
		Command:     fmt.Sprintf("id %s", username),
		Description: "Get user info",
	}
	cmdGroups := types.Command{
		Command:     fmt.Sprintf("groups %s", username),
		Description: "Get user groups",
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdID.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGroups.Command)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Would check user status for '%s'", username),
		}
	}

	// Check specific user
	cfg.GetLoggerOrDefault().Info("checking user", "username", username)

	output, err := ssh.Run(cfg, cmdID)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("User '%s' does not exist", username),
			Error:   fmt.Errorf("user '%s' not found", username),
		}
	}
	cfg.GetLoggerOrDefault().Info("user info", "output", output)

	// Check if user has sudo
	groupsOutput, err := ssh.Run(cfg, cmdGroups)
	if err == nil {
		cfg.GetLoggerOrDefault().Info("user groups", "groups", groupsOutput)
	}

	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("User info for '%s'", username),
		Details: map[string]string{"info": output, "groups": groupsOutput},
	}
}

// NewUserStatus creates a new user-status skill.
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
func NewUserStatus() types.RunnableInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDUserStatus)
	pb.SetDescription("Show user information")
	return &UserStatus{BaseSkill: pb}
}
