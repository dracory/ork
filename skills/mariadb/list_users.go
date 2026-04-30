package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ListUsers displays all database user accounts and their allowed hosts.
// This read-only skill queries the mysql.user table to show authentication
// information for all users configured in the MariaDB server.
//
// Usage:
//
//	go run . --playbook=mariadb-list-users [--arg=root-password=<password>]
//
// Args:
//   - root-password: MariaDB root password (required)
//
// Output Format:
//   - User column: Username for authentication
//   - Host column: Allowed connection sources
//
// Common Host Patterns:
//   - localhost: Local socket connections only
//   - 127.0.0.1: Local TCP connections
//   - ::1: Local IPv6 connections
//   - %: Any host (wildcard)
//   - 192.168.1.%: IP subnet pattern
//   - specific.ip.address: Single IP only
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-create-user: Create a new user
//   - mariadb-secure: Remove insecure default users
type ListUsers struct {
	*types.BaseSkill
}

// Check always returns false since this is a read-only skill.
func (m *ListUsers) Check() (bool, error) {
	return false, nil
}

// Run executes the skill and returns detailed result.
func (m *ListUsers) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cmdList := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user;"`, rootPassword), Description: "List all database users"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdList.Command, "description", cmdList.Description)
		return types.Result{
			Changed: false,
			Message: "Would list all database users",
		}
	}

	cfg.GetLoggerOrDefault().Info("listing all database users")

	output, err := ssh.Run(cfg, cmdList)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to list users",
			Error:   fmt.Errorf("failed to list users: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("database users", "output", output)
	return types.Result{
		Changed: false,
		Message: "User list retrieved",
		Details: map[string]string{
			"users": output,
		},
	}
}

// SetArgs sets the arguments for listing MariaDB users.
// Returns ListUsers for fluent method chaining.
func (l *ListUsers) SetArgs(args map[string]string) types.RunnableInterface {
	l.BaseSkill.SetArgs(args)
	return l
}

// SetArg sets a single argument for listing MariaDB users.
// Returns ListUsers for fluent method chaining.
func (l *ListUsers) SetArg(key, value string) types.RunnableInterface {
	l.BaseSkill.SetArg(key, value)
	return l
}

// SetID sets the ID for listing MariaDB users.
// Returns ListUsers for fluent method chaining.
func (l *ListUsers) SetID(id string) types.RunnableInterface {
	l.BaseSkill.SetID(id)
	return l
}

// SetDescription sets the description for listing MariaDB users.
// Returns ListUsers for fluent method chaining.
func (l *ListUsers) SetDescription(description string) types.RunnableInterface {
	l.BaseSkill.SetDescription(description)
	return l
}

// SetTimeout sets the timeout for listing MariaDB users.
// Returns ListUsers for fluent method chaining.
func (l *ListUsers) SetTimeout(timeout time.Duration) types.RunnableInterface {
	l.BaseSkill.SetTimeout(timeout)
	return l
}

// NewListUsers creates a new mariadb-list-users skill.
func NewListUsers() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbListUsers)
	pb.SetDescription("Display all database user accounts and their allowed hosts (read-only)")
	return &ListUsers{BaseSkill: pb}
}
