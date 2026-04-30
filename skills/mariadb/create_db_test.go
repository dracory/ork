package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestCreateDB_Run_DryRun verifies that dry-run mode correctly handles database creation.
func TestCreateDB_Run_DryRun(t *testing.T) {
	pb := NewCreateDB()
	pb.SetArg("db-name", "testdb")
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create database 'testdb'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestCreateDB_Run_DryRun_NoDbName verifies dry-run without database name returns error.
func TestCreateDB_Run_DryRun_NoDbName(t *testing.T) {
	pb := NewCreateDB()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing db-name even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing db-name")
	}

	if result.Message != "Database name is required" {
		t.Errorf("Expected message 'Database name is required', got '%s'", result.Message)
	}
}

// TestCreateDB_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestCreateDB_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewCreateDB()
	pb.SetArg("db-name", "testdb")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing root-password even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestCreateDB_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestCreateDB_Run_NotDryRun(t *testing.T) {
	pb := NewCreateDB()
	pb.SetArg("db-name", "testdb")
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create database 'testdb'" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestCreateDB_NewCreateDB verifies that NewCreateDB creates a properly configured skill.
func TestCreateDB_NewCreateDB(t *testing.T) {
	pb := NewCreateDB()

	if pb.GetID() != "mariadb-create-db" {
		t.Errorf("Expected ID to be 'mariadb-create-db', got '%s'", pb.GetID())
	}

	expectedDescription := "Create a new MariaDB database with UTF-8 encoding"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestCreateDB_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete CreateDB type.
func TestCreateDB_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateDB()
	args := map[string]string{"db-name": "testdb", "root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*CreateDB); !ok {
		t.Error("SetArgs should return *CreateDB, not just RunnableInterface")
	}
}

// TestCreateDB_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete CreateDB type.
func TestCreateDB_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateDB()

	result := skill.SetArg("db-name", "testdb")

	if _, ok := result.(*CreateDB); !ok {
		t.Error("SetArg should return *CreateDB, not just RunnableInterface")
	}
}

// TestCreateDB_SetID_ReturnsConcreteType verifies that SetID returns the concrete CreateDB type.
func TestCreateDB_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateDB()

	result := skill.SetID("custom-id")

	if _, ok := result.(*CreateDB); !ok {
		t.Error("SetID should return *CreateDB, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestCreateDB_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete CreateDB type.
func TestCreateDB_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateDB()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*CreateDB); !ok {
		t.Error("SetDescription should return *CreateDB, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestCreateDB_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete CreateDB type.
func TestCreateDB_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateDB()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*CreateDB); !ok {
		t.Error("SetTimeout should return *CreateDB, not just RunnableInterface")
	}
}

// TestCreateDB_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestCreateDB_MethodChaining_PreservesType(t *testing.T) {
	skill := NewCreateDB().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("db-name", "testdb").
		SetArgs(map[string]string{"root-password": "testpass"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*CreateDB); !ok {
		t.Error("Method chaining should preserve *CreateDB type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
