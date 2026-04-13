package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
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
	*playbook.BasePlaybook
}

// Check always returns true since we always want to ensure security settings are applied.
func (m *Secure) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (m *Secure) Run() playbook.Result {
	cfg := m.GetConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("securing MariaDB installation")

	actions := []string{}

	// Remove anonymous users
	cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.user WHERE User='';"`, rootPassword)
	_, err := ssh.Run(cfg, cmd)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove anonymous users", "error", err)
	} else {
		actions = append(actions, "removed_anonymous_users")
		cfg.GetLoggerOrDefault().Info("anonymous users removed")
	}

	// Remove remote root access (only localhost allowed for root)
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');"`, rootPassword)
	_, err = ssh.Run(cfg, cmd)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not restrict root remote access", "error", err)
	} else {
		actions = append(actions, "restricted_root_access")
		cfg.GetLoggerOrDefault().Info("remote root access restricted")
	}

	// Remove test database
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "DROP DATABASE IF EXISTS test;"`, rootPassword)
	_, err = ssh.Run(cfg, cmd)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not remove test database", "error", err)
	} else {
		actions = append(actions, "removed_test_database")
		cfg.GetLoggerOrDefault().Info("test database removed")
	}

	// Remove test database privileges
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%%';"`, rootPassword)
	_, _ = ssh.Run(cfg, cmd)

	// Reload privileges
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "FLUSH PRIVILEGES;"`, rootPassword)
	_, err = ssh.Run(cfg, cmd)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("could not flush privileges", "error", err)
	} else {
		actions = append(actions, "flushed_privileges")
		cfg.GetLoggerOrDefault().Info("privileges flushed")
	}

	return playbook.Result{
		Changed: len(actions) > 0,
		Message: "MariaDB security hardening completed",
		Details: map[string]string{
			"actions": fmt.Sprintf("%v", actions),
		},
	}
}

// NewSecure creates a new mariadb-secure playbook.
func NewSecure() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbSecure)
	pb.SetDescription("Perform basic security hardening on MariaDB installation")
	return &Secure{BasePlaybook: pb}
}
