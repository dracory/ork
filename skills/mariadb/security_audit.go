package mariadb

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SecurityAudit performs a comprehensive security audit of MariaDB configuration.
// This read-only skill checks for common security issues and misconfigurations,
// providing recommendations for hardening the database server.
//
// Usage:
//
//	node.Run(mariadb.NewSecurityAudit().SetRootPassword("<password>"))
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
	*types.BaseSkill
}

// Compile-time assertion that SecurityAudit implements types.RunnableInterface.
var _ types.RunnableInterface = (*SecurityAudit)(nil)

// Check always returns false since this is a read-only skill.
func (m *SecurityAudit) Check() (bool, error) {
	return false, nil
}

// Run executes the skill and returns detailed result.
func (m *SecurityAudit) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	// Define commands â€” use MYSQL_PWD env var to avoid shell injection via -p argument
	shellEscapedPwd := mariadbEscapeShellQuote(rootPassword)
	cmdAnon := types.Command{Command: fmt.Sprintf(`MYSQL_PWD='%s' mysql -u root -e "SELECT User, Host FROM mysql.user WHERE User='';"`, shellEscapedPwd), Description: "Check for anonymous users"}
	cmdTestDb := types.Command{Command: fmt.Sprintf(`MYSQL_PWD='%s' mysql -u root -e "SHOW DATABASES LIKE 'test';"`, shellEscapedPwd), Description: "Check for test database"}
	cmdSsl := types.Command{Command: fmt.Sprintf(`MYSQL_PWD='%s' mysql -u root -e "SHOW VARIABLES LIKE 'have_ssl';"`, shellEscapedPwd), Description: "Check SSL status"}
	cmdWildcard := types.Command{Command: fmt.Sprintf(`MYSQL_PWD='%s' mysql -u root -e "SELECT User, Host FROM mysql.user WHERE Host='%%';"`, shellEscapedPwd), Description: "Check wildcard hosts"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAnon.Command, "description", cmdAnon.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdTestDb.Command, "description", cmdTestDb.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdSsl.Command, "description", cmdSsl.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWildcard.Command, "description", cmdWildcard.Description)
		return types.Result{
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
	return types.Result{
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

// SetArgs sets the arguments for MariaDB security audit.
// Returns SecurityAudit for fluent method chaining.
func (s *SecurityAudit) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetRootPassword sets the MariaDB root password and returns SecurityAudit for chaining.
func (s *SecurityAudit) SetRootPassword(password string) *SecurityAudit {
	s.BaseSkill.SetArg(ArgRootPassword, password)
	return s
}

// SetArg sets a single argument for MariaDB security audit.
// Returns SecurityAudit for fluent method chaining.
func (s *SecurityAudit) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for MariaDB security audit.
// Returns SecurityAudit for fluent method chaining.
func (s *SecurityAudit) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for MariaDB security audit.
// Returns SecurityAudit for fluent method chaining.
func (s *SecurityAudit) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for MariaDB security audit.
// Returns SecurityAudit for fluent method chaining.
func (s *SecurityAudit) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSecurityAudit creates a new mariadb-security-audit skill.
func NewSecurityAudit() *SecurityAudit {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbSecurityAudit)
	pb.SetDescription("Perform a comprehensive security audit of MariaDB (read-only)")
	return &SecurityAudit{BaseSkill: pb}
}
