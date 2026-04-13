package mariadb

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// CreateUser creates a new database user with configurable privileges.
// This playbook creates users that can connect from specific hosts and grants
// appropriate database access permissions following the principle of least privilege.
//
// Usage:
//
//	go run . --playbook=mariadb-create-user --arg=username=<name> --arg=password=<pass> [--arg=db-name=<db>] [--arg=host=<host>]
//
// Args:
//   - username: Database username to create (required)
//   - password: Password for the new user (required)
//   - db-name: Database name(s) to grant access to (optional, comma-separated for multiple)
//     Use "*" for all databases (superuser privileges - use with caution)
//   - host: Host pattern for user connections (default: '%' for any host)
//
// Host Patterns:
//   - '%' (default): Allow connections from any host
//   - 'localhost': Local connections only
//   - '192.168.1.%': Allow from any IP in subnet
//   - '10.0.0.5': Specific IP only
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-create-db: Create a database for this user
//   - mariadb-list-users: Verify user was created
type CreateUser struct {
	*playbook.BasePlaybook
}

// Check determines if the user already exists.
func (m *CreateUser) Check() (bool, error) {
	cfg := m.GetConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	username := m.GetArg(ArgUsername)
	host := m.GetArg(ArgHost)
	if host == "" {
		host = "%"
	}

	if rootPassword == "" || username == "" {
		return true, nil
	}

	cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT 1 FROM mysql.user WHERE user='%s' AND host='%s';"`, rootPassword, username, host)
	output, _ := ssh.Run(cfg, cmd)
	return output == "", nil
}

// Run executes the playbook and returns detailed result.
func (m *CreateUser) Run() playbook.Result {
	cfg := m.GetConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	username := m.GetArg(ArgUsername)
	password := m.GetArg(ArgPassword)
	dbName := m.GetArg(ArgDbName)
	host := m.GetArg(ArgHost)
	if host == "" {
		host = "%"
	}

	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username argument is required"),
		}
	}
	if password == "" {
		return playbook.Result{
			Changed: false,
			Message: "Password is required",
			Error:   fmt.Errorf("password argument is required"),
		}
	}
	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	log.Printf("Creating user: %s@%s", username, host)

	// Create user
	cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED BY '%s';"`,
		rootPassword, username, host, password)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create user",
			Error:   fmt.Errorf("failed to create user: %w\nOutput: %s", err, output),
		}
	}

	// Grant privileges
	grantedDBs := []string{}
	if dbName != "" {
		if dbName == "*" {
			cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "GRANT ALL PRIVILEGES ON *.* TO '%s'@'%s';"`,
				rootPassword, username, host)
			_, _ = ssh.Run(cfg, cmd)
			grantedDBs = append(grantedDBs, "*")
			log.Printf("Granted ALL PRIVILEGES on all databases to %s@%s", username, host)
		} else {
			databases := strings.Split(dbName, ",")
			for _, db := range databases {
				db = strings.TrimSpace(db)
				cmd = fmt.Sprintf("mysql -u root -p\"%s\" -e \"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%s';\"",
					rootPassword, db, username, host)
				_, err = ssh.Run(cfg, cmd)
				if err != nil {
					log.Printf("Warning: Could not grant privileges on %s: %v", db, err)
				} else {
					grantedDBs = append(grantedDBs, db)
					log.Printf("Granted privileges on '%s' to %s@%s", db, username, host)
				}
			}
		}
	}

	// Flush privileges
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "FLUSH PRIVILEGES;"`, rootPassword)
	_, _ = ssh.Run(cfg, cmd)

	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s'@'%s' created successfully", username, host),
		Details: map[string]string{
			"granted_databases": strings.Join(grantedDBs, ","),
		},
	}
}

// NewCreateUser creates a new mariadb-create-user playbook.
func NewCreateUser() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbCreateUser)
	pb.SetDescription("Create a new MariaDB user with configurable privileges")
	return &CreateUser{BasePlaybook: pb}
}
