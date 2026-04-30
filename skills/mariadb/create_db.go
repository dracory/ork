package mariadb

import (
	"fmt"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// CreateDB creates a new MariaDB database with UTF-8 encoding.
// The database is created with utf8mb4 character set and utf8mb4_unicode_ci
// collation for full Unicode support including emojis and special characters.
//
// Usage:
//
//	go run . --playbook=mariadb-create-db --arg=db-name=<database_name> [--arg=root-password=<password>]
//
// Args:
//   - db-name: Name of the database to create (required)
//   - root-password: MariaDB root password (required if not using socket auth)
//
// Database Configuration:
//   - Character Set: utf8mb4 (supports all Unicode characters including emojis)
//   - Collation: utf8mb4_unicode_ci (Unicode sorting and comparison)
//   - Creation: IF NOT EXISTS (idempotent - safe to run multiple times)
//
// Prerequisites:
//   - MariaDB must be installed and running
//   - Root SSH access required
//
// Related Playbooks:
//   - mariadb-create-user: Create a user and grant access to this database
//   - mariadb-list-dbs: Verify database was created
type CreateDB struct {
	*types.BaseSkill
}

// Check determines if the database already exists.
func (m *CreateDB) Check() (bool, error) {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	dbName := m.GetArg(ArgDbName)

	if rootPassword == "" || dbName == "" {
		return true, nil
	}

	cmdCheck := types.Command{Command: fmt.Sprintf(`mysql -u root -p"%s" -e "SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s';"`, rootPassword, dbName), Description: "Check if database exists"}
	output, _ := ssh.Run(cfg, cmdCheck)
	return output == "", nil
}

// Run executes the skill and returns detailed result.
func (m *CreateDB) Run() types.Result {
	cfg := m.GetNodeConfig()
	rootPassword := m.GetArg(ArgRootPassword)
	dbName := m.GetArg(ArgDbName)

	if dbName == "" {
		return types.Result{
			Changed: false,
			Message: "Database name is required",
			Error:   fmt.Errorf("db-name argument is required"),
		}
	}

	if rootPassword == "" {
		return types.Result{
			Changed: false,
			Message: "MariaDB root password not provided",
			Error:   fmt.Errorf("root-password is required"),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating database", "database", dbName)

	cmdCreate := types.Command{Command: fmt.Sprintf("mysql -u root -p\"%s\" -e \"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\"", rootPassword, dbName), Description: "Create database"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCreate.Command, "description", cmdCreate.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create database '%s'", dbName),
		}
	}

	output, err := ssh.Run(cfg, cmdCreate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create database",
			Error:   fmt.Errorf("failed to create database: %w\nOutput: %s", err, output),
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Database '%s' created successfully", dbName),
	}
}

// SetArgs sets the arguments for creating MariaDB database.
// Returns CreateDB for fluent method chaining.
func (c *CreateDB) SetArgs(args map[string]string) types.RunnableInterface {
	c.BaseSkill.SetArgs(args)
	return c
}

// NewCreateDB creates a new mariadb-create-db skill.
func NewCreateDB() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDMariadbCreateDB)
	pb.SetDescription("Create a new MariaDB database with UTF-8 encoding")
	return &CreateDB{BaseSkill: pb}
}
