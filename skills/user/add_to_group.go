package user

// Package user documentation is in status.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UserAddToGroup adds a user to a supplementary group.
// This is useful for granting access to shared resources like
// the www-data group for PHP-FPM socket access.
type UserAddToGroup struct {
	*types.BaseSkill
}

// Check determines if the user needs to be added to the group.
// Returns true if the user is not already a member of the group.
func (u *UserAddToGroup) Check() (bool, error) {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)
	group := u.GetArg(ArgGroup)
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	if group == "" {
		return false, fmt.Errorf("group is required (pass via --arg=group=value)")
	}
	cmdCheck := types.Command{Command: fmt.Sprintf("groups %s", skills.ShellEscapeArg(username)), Description: "Check user group membership"}
	output, _ := ssh.Run(cfg, cmdCheck)
	for field := range strings.FieldsSeq(output) {
		if field == group {
			return false, nil // already in group
		}
	}
	return true, nil // not in group
}

// Run adds a user to a supplementary group using usermod -aG.
// This operation is idempotent - if the user is already in the group,
// usermod -aG silently succeeds.
//
// Usage:
//
//	go run . --playbook=user-add-to-group --arg=username=<name> --arg=group=<name>
//
// Parameters (passed via args):
//   - username (string, required): Name of the user to add to the group
//   - group (string, required): Name of the group to add the user to
//
// Execution Flow:
//  1. Validates username and group parameters
//  2. Runs usermod -aG <group> <username>
//  3. Reports success or failure
//
// Idempotency:
//   - usermod -aG is safe to re-run; it won't duplicate group membership
//
// Prerequisites:
//   - Root SSH access required for usermod
//   - User must exist on the system
//   - Group must exist on the system
//
// Args:
//   - username (string, required): Username to add to group
//   - group (string, required): Group name to add user to
func (u *UserAddToGroup) Run() types.Result {
	cfg := u.GetNodeConfig()
	username := u.GetArg(ArgUsername)
	group := u.GetArg(ArgGroup)

	if username == "" {
		return types.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}
	if group == "" {
		return types.Result{
			Changed: false,
			Message: "Group is required",
			Error:   fmt.Errorf("group is required (pass via --arg=group=value)"),
		}
	}

	cfg.GetLoggerOrDefault().Info("adding user to group", "username", username, "group", group)

	cmdAdd := types.Command{
		Command:     fmt.Sprintf("usermod -aG %s %s", skills.ShellEscapeArg(group), skills.ShellEscapeArg(username)),
		Description: "Add user to supplementary group",
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAdd.Command)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would add user '%s' to group '%s'", username, group),
		}
	}

	output, err := ssh.Run(cfg, cmdAdd)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to add user to group",
			Error:   fmt.Errorf("failed to add user '%s' to group '%s': %w\nOutput: %s", username, group, err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("user added to group", "username", username, "group", group)
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' added to group '%s'", username, group),
	}
}

// SetArgs sets the arguments for adding user to group.
// Returns UserAddToGroup for fluent method chaining.
func (u *UserAddToGroup) SetArgs(args map[string]string) types.RunnableInterface {
	u.BaseSkill.SetArgs(args)
	return u
}

// SetArg sets a single argument for adding user to group.
// Returns UserAddToGroup for fluent method chaining.
func (u *UserAddToGroup) SetArg(key, value string) types.RunnableInterface {
	u.BaseSkill.SetArg(key, value)
	return u
}

// SetID sets the ID for adding user to group.
// Returns UserAddToGroup for fluent method chaining.
func (u *UserAddToGroup) SetID(id string) types.RunnableInterface {
	u.BaseSkill.SetID(id)
	return u
}

// SetDescription sets the description for adding user to group.
// Returns UserAddToGroup for fluent method chaining.
func (u *UserAddToGroup) SetDescription(description string) types.RunnableInterface {
	u.BaseSkill.SetDescription(description)
	return u
}

// SetTimeout sets the timeout for adding user to group.
// Returns UserAddToGroup for fluent method chaining.
func (u *UserAddToGroup) SetTimeout(timeout time.Duration) types.RunnableInterface {
	u.BaseSkill.SetTimeout(timeout)
	return u
}

// NewUserAddToGroup creates a new user-add-to-group skill.
func NewUserAddToGroup() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUserAddToGroup)
	pb.SetDescription("Add a user to a supplementary group (username via args['username'], group via args['group'])")
	return &UserAddToGroup{BaseSkill: pb}
}
