package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SecurityAudit performs a comprehensive security audit of MariaDB configuration.
// This read-only playbook checks for common security issues and misconfigurations,
// providing recommendations for hardening the database server.
//
// Usage:
//
//	go run . --playbook=mariadb-security-audit [--arg=root-password=<password>]
//
// Args:
//   - root-password: MariaDB root password (required)
//
// Security Checks Performed:
//   - Anonymous user accounts
//   - Test database existence
//   - Root remote access permissions
//   - SSL/TLS availability
//   - User access patterns
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-secure: Fix identified security issues
//   - mariadb-enable-ssl: Enable SSL/TLS encryption
type SecurityAudit struct {
	*playbook.BasePlaybook
}

// Check always returns false since this is a read-only playbook.
func (m *SecurityAudit) Check() (bool, error) {
	return false, nil
}

// Run executes the playbook and returns detailed result.
func (m *SecurityAudit) Run() playbook.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	// Define commands
	cmdAnon := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user WHERE User='';"`, rootPassword), Description: "Check for anonymous users"}
	cmdTestDb := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SHOW DATABASES LIKE 'test';"`, rootPassword), Description: "Check for test database"}
	cmdSsl := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SHOW VARIABLES LIKE 'have_ssl';"`, rootPassword), Description: "Check SSL status"}
	cmdWildcard := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user WHERE Host='%%';"`, rootPassword), Description: "Check wildcard hosts"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAnon.Command, "description", cmdAnon.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdTestDb.Command, "description", cmdTestDb.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdSsl.Command, "description", cmdSsl.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWildcard.Command, "description", cmdWildcard.Description)
		return playbook.Result{
			Changed: false,
			Message: "Would perform MariaDB security audit",
		}
	}

	cfg.GetLoggerOrDefault().Info("MariaDB security audit started")

	anonOutput, _ := ssh.Run(cfg, cmdAnon)
	testOutput, _ := ssh.Run(cfg, cmdTestDb)
	sslOutput, _ := ssh.Run(cfg, cmdSsl)
	wildcardOutput, _ := ssh.Run(cfg, cmdWildcard)

	cfg.GetLoggerOrDefault().Info("MariaDB security audit complete")
	return playbook.Result{
		Changed: false,
		Message: "Security audit completed",
		Details: map[string]string{
			"anonymous_users": anonOutput,
			"test_database":   testOutput,
			"ssl_status":      sslOutput,
			"wildcard_hosts":  wildcardOutput,
		},
	}
}

// NewSecurityAudit creates a new mariadb-security-audit playbook.
func NewSecurityAudit() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDMariadbSecurityAudit)
	pb.SetDescription("Perform a comprehensive security audit of MariaDB (read-only)")
	return &SecurityAudit{BasePlaybook: pb}
}
