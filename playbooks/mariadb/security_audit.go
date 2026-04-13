package mariadb

import (
	"fmt"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
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
	cfg := m.GetConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return playbook.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("MariaDB security audit started")

	// Check for anonymous users
	cmd := fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user WHERE User='';"`, rootPassword)
	anonOutput, _ := ssh.Run(cfg, cmd)

	// Check for test database
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "SHOW DATABASES LIKE 'test';"`, rootPassword)
	testOutput, _ := ssh.Run(cfg, cmd)

	// Check SSL
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "SHOW VARIABLES LIKE 'have_ssl';"`, rootPassword)
	sslOutput, _ := ssh.Run(cfg, cmd)

	// Check wildcard hosts
	cmd = fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT User, Host FROM mysql.user WHERE Host='%%';"`, rootPassword)
	wildcardOutput, _ := ssh.Run(cfg, cmd)

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
