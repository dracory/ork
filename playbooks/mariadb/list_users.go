package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ListUsers displays all database user accounts and their allowed hosts.
// This read-only playbook queries the mysql.user table to show authentication
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
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (m *ListUsers) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (m *ListUsers) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("listing all database users")

	cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user;"`, rootPassword)
	output, err := ssh.Run(cfg, types.Command{Command: cmd, Description: "List all database users"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list users",
			Error:   fmt.Errorf("failed to list users: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("database users", "output", output)
	return playbook.Result{
		Changed: false,
		Message: "User list retrieved",
		Details: map[string]string{
			"users": output,
		},
	}
}

// NewListUsers creates a new mariadb-list-users playbook.
func NewListUsers() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbListUsers)
	pb.SetDescription("Display all database user accounts and their allowed hosts (read-only)")
	return &ListUsers{BasePlaybook: pb}
}
