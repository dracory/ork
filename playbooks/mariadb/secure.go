package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbooks"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Secure performs basic security hardening on a fresh MariaDB installation.
// This playbook removes insecure default settings that come with a standard MariaDB install,
// including anonymous users, test database, and remote root access.
//
// Usage:
//
//	go run . --playbook=mariadb-secure [--arg=root-password=<password>]
//
// Security Actions Performed:
//  1. Remove anonymous users (users with empty username)
//  2. Remove remote root access (restrict root to localhost only)
//  3. Remove test database and its privileges
//  4. Flush privileges to apply all changes
//
// Args:
//   - root-password: MariaDB root password (required)
//
// Prerequisites:
//   - MariaDB must be installed
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-install: Initial installation
//   - mariadb-create-user: Create restricted application users
type Secure struct {
	*playbooks.BasePlaybook
}

// Check always returns true since we always want to ensure security settings are applied.
func (m *Secure) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *Secure) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("securing MariaDB installation")

	// Define commands
	cmdAnon := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.user WHERE User='';"`, rootPassword), Description: "Remove anonymous users"}
	cmdRoot := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');"`, rootPassword), Description: "Restrict root remote access"}
	cmdTestDb := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "DROP DATABASE IF EXISTS test;"`, rootPassword), Description: "Remove test database"}
	cmdTestPriv := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%%';"`, rootPassword), Description: "Remove test database privileges"}
	cmdFlush := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "FLUSH PRIVILEGES;"`, rootPassword), Description: "Flush privileges"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAnon.Command, "description", cmdAnon.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRoot.Command, "description", cmdRoot.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdTestDb.Command, "description", cmdTestDb.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdTestPriv.Command, "description", cmdTestPriv.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdFlush.Command, "description", cmdFlush.Description)
		return types.Result{
			Changed: true,
			Message: "Would secure MariaDB installation",
		}
	}

	actions := []string{}

	// Remove anonymous users
	_, err := ssh.Run(cfg, cmdAnon)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove anonymous users", "error", err)
	} else {
		actions = append(actions, "removed_anonymous_users")
		cfg.GetLoggerOrDefault().Info("anonymous users removed")
	}

	// Remove remote root access (only localhost allowed for root)
	_, err = ssh.Run(cfg, cmdRoot)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not restrict root remote access", "error", err)
	} else {
		actions = append(actions, "restricted_root_access")
		cfg.GetLoggerOrDefault().Info("remote root access restricted")
	}

	// Remove test database
	_, err = ssh.Run(cfg, cmdTestDb)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove test database", "error", err)
	} else {
		actions = append(actions, "removed_test_database")
		cfg.GetLoggerOrDefault().Info("test database removed")
	}

	// Remove test database privileges
	_, _ = ssh.Run(cfg, cmdTestPriv)

	// Reload privileges
	_, err = ssh.Run(cfg, cmdFlush)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not flush privileges", "error", err)
	} else {
		actions = append(actions, "flushed_privileges")
		cfg.GetLoggerOrDefault().Info("privileges flushed")
	}

	return types.Result{
		Changed: len(actions) > 0,
		Message: "MariaDB security hardening completed",
		Details: map[string]string{
			"actions": fmt.Sprintf("%v", actions),
		},
	}
}

// NewSecure creates a new mariadb-secure playbook.
func NewSecure() types.PlaybookInterface {
	pb := playbooks.NewBasePlaybook()
	pb.SetID(playbooks.IDMariadbSecure)
	pb.SetDescription("Perform basic security hardening on MariaDB installation")
	return &Secure{BasePlaybook: pb}
}
