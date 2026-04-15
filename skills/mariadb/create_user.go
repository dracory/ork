package mariadb

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// CreateUser creates a new database user with configurable privileges.
// This skill creates users that can connect from specific hosts and grants
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
	*skills.BaseSkill
}

// Check determines if the user already exists.
func (m *CreateUser) Check() (bool, error) {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	username := m.GetArg(ArgUsername)
	host := m.GetArg(ArgHost)
	if host == "" {
		host = "%"
	}

	if rootPassword == "" || username == "" {
		return true, nil
	}

	cmdCheck := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT 1 FROM mysql.user WHERE user='%s' AND host='%s';"`, rootPassword, username, host), Description: "Check if user exists"}
	output, _ := ssh.Run(cfg, cmdCheck)
	return output == "", nil
}

// Run executes the skill and returns detailed result.
func (m *CreateUser) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	username := m.GetArg(ArgUsername)
	password := m.GetArg(ArgPassword)
	dbName := m.GetArg(ArgDbName)
	host := m.GetArg(ArgHost)
	if host == "" {
		host = "%"
	}

	if username == "" {
		return types.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username argument is required"),
		}
	}
	if password == "" {
		return types.Result{
			Changed: false,
			Message: "Password is required",
			Error:   fmt.Errorf("password argument is required"),
		}
	}
	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating database user", "username", username, "host", host)

	// Create user
	cmdCreate := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED BY '%s';"`,
		rootPassword, username, host, password), Description: "Create database user"}
	cmdFlush := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "FLUSH PRIVILEGES;"`, rootPassword), Description: "Flush privileges"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCreate.Command, "description", cmdCreate.Description)
		if dbName != "" {
			if dbName == "*" {
				cmdGrantAll := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "GRANT ALL PRIVILEGES ON *.* TO '%s'@'%s';"`,
					rootPassword, username, host), Description: "Grant all privileges"}
				cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGrantAll.Command, "description", cmdGrantAll.Description)
			} else {
				databases := strings.Split(dbName, ",")
				for _, db := range databases {
					db = strings.TrimSpace(db)
					cmdGrantDb := types.Command{Command: fmt.Sprintf("mysql -u root -p\"%s\" -e \"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%s';\"",
						rootPassword, db, username, host), Description: "Grant database privileges"}
					cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGrantDb.Command, "description", cmdGrantDb.Description)
				}
			}
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdFlush.Command, "description", cmdFlush.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create user '%s'@'%s'", username, host),
		}
	}

	output, err := ssh.Run(cfg, cmdCreate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create user",
			Error:   fmt.Errorf("failed to create user: %w\nOutput: %s", err, output),
		}
	}

	// Grant privileges
	grantedDBs := []string{}
	if dbName != "" {
		if dbName == "*" {
			cmdGrantAll := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "GRANT ALL PRIVILEGES ON *.* TO '%s'@'%s';"`,
				rootPassword, username, host), Description: "Grant all privileges"}
			_, _ = ssh.Run(cfg, cmdGrantAll)
			grantedDBs = append(grantedDBs, "*")
			cfg.GetLoggerOrDefault().Info("granted all privileges", "username", username, "host", host)
		} else {
			databases := strings.Split(dbName, ",")
			for _, db := range databases {
				db = strings.TrimSpace(db)
				cmdGrantDb := types.Command{Command: fmt.Sprintf("mysql -u root -p\"%s\" -e \"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%s';\"",
					rootPassword, db, username, host), Description: "Grant database privileges"}
				_, err = ssh.Run(cfg, cmdGrantDb)
				if err != nil {
					cfg.GetLoggerOrDefault().Warn("could not grant privileges", "database", db, "error", err)
				} else {
					grantedDBs = append(grantedDBs, db)
					cfg.GetLoggerOrDefault().Info("granted privileges", "database", db, "username", username, "host", host)
				}
			}
		}
	}

	// Flush privileges
	_, _ = ssh.Run(cfg, cmdFlush)

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s'@'%s' created successfully", username, host),
		Details: map[string]string{
			"granted_databases": strings.Join(grantedDBs, ","),
		},
	}
}

// NewCreateUser creates a new mariadb-create-user skill.
func NewCreateUser() types.SkillInterface {
	pb := skills.NewBaseSkill()
	pb.SetID(skills.IDMariadbCreateUser)
	pb.SetDescription("Create a new MariaDB user with configurable privileges")
	return &CreateUser{BaseSkill: pb}
}
