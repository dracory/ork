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

// UserList lists all non-system users on the system.
// This is a read-only skill that queries all user accounts with UID >= 1000.
//
// Usage:
//
//	go run . --playbook=user-list
//
// Execution Flow:
//  1. Parses /etc/passwd to find non-system users (UID >= 1000, < 65534)
//  2. Reports list of all regular user accounts
//
// Expected Output:
//   - Success: "Non-system users listed" with user list
//   - Failure: Error indicating failure to list users
//
// Result Details contains:
//   - users: List of usernames (one per line)
//
// Use Cases:
//   - Audit user accounts on system
//   - Inventory current users
//   - Verify user cleanup
//
// Idempotency:
//   - Always reports Changed=false since this is read-only
type UserList struct {
	*skills.BaseSkill
}

// Check always returns false since UserList is read-only.
// Per the skill interface convention, the bool return indicates whether
// the operation would modify system state. Since user-list only queries
// user information, this always returns false.
func (u *UserList) Check() (bool, error) {
	return false, nil
}

// Run displays all non-system users and returns detailed result.
// Changed is always false since this is a read-only operation.
//
// Result.Details contains:
//   - users: List of all non-system usernames (one per line)
func (u *UserList) Run() types.Result {
	cfg := u.GetNodeConfig()
	cmdListAll := types.Command{
		Command:     "awk -F: '$3 >= 1000 && $3 < 65534 {print $1}' /etc/passwd",
		Description: "List all non-system users",
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdListAll.Command)
		return types.Result{
			Changed: false,
			Message: "Would list all non-system users",
		}
	}

	// List all non-system users
	cfg.GetLoggerOrDefault().Info("listing all non-system users")

	output, err := ssh.Run(cfg, cmdListAll)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to list users",
			Error:   fmt.Errorf("failed to list users: %w", err),
		}
	}

	if output == "" {
		cfg.GetLoggerOrDefault().Info("no non-system users found")
		return types.Result{
			Changed: false,
			Message: "No non-system users found",
		}
	}

	cfg.GetLoggerOrDefault().Info("users found", "users", output)
	return types.Result{
		Changed: false,
		Message: "Non-system users listed",
		Details: map[string]string{"users": output},
	}
}

// NewUserList creates a new user-list skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDUserList identifier
//	and description "List all non-system users".
func NewUserList() types.SkillInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDUserList)
	pb.SetDescription("List all non-system users")
	return &UserList{BaseSkill: pb}
}
