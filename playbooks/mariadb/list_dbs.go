package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ListDBs displays all databases in the MariaDB server.
// This read-only playbook executes SHOW DATABASES and displays the results,
// including system databases (mysql, information_schema, performance_schema, sys).
//
// Usage:
//
//	go run . --playbook=mariadb-list-dbs [--arg=root-password=<password>]
//
// Args:
//   - root-password: MariaDB root password (required)
//
// Output:
//   - List of all database names
//   - Includes system databases:
//   - mysql: System tables (users, permissions, etc.)
//   - information_schema: Metadata about database objects
//   - performance_schema: Performance monitoring data
//   - sys: Simplified views of performance schema
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-create-db: Create a new database
//   - mariadb-backup: Backup an existing database
type ListDBs struct {
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (m *ListDBs) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (m *ListDBs) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("listing all databases")

	cmdList := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SHOW DATABASES;"`, rootPassword), Description: "List all databases"}
	output, err := ssh.Run(cfg, cmdList)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list databases",
			Error:   fmt.Errorf("failed to list databases: %w", err),
		}
	}

	cfg.GetLoggerOrDefault().Info("databases", "output", output)
	return playbook.Result{
		Changed: false,
		Message: "Database list retrieved",
		Details: map[string]string{
			"databases": output,
		},
	}
}

// NewListDBs creates a new mariadb-list-dbs playbook.
func NewListDBs() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbListDBs)
	pb.SetDescription("Display all databases in the MariaDB server (read-only)")
	return &ListDBs{BasePlaybook: pb}
}
