package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestBackup_Run_DryRun verifies that dry-run mode correctly handles MariaDB backup.
func TestBackup_Run_DryRun(t *testing.T) {
	pb := NewBackup()
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

	expectedMessage := "Would create backup for database 'testdb'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestBackup_Run_DryRun_NoDbName verifies dry-run without database name returns error.
func TestBackup_Run_DryRun_NoDbName(t *testing.T) {
	pb := NewBackup()
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

// TestBackup_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestBackup_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewBackup()
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

// TestBackup_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestBackup_Run_NotDryRun(t *testing.T) {
	pb := NewBackup()
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
	if result.Message == "Would create backup for database 'testdb'" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestBackup_NewBackup verifies that NewBackup creates a properly configured skill.
func TestBackup_NewBackup(t *testing.T) {
	pb := NewBackup()

	if pb.GetID() != "mariadb-backup" {
		t.Errorf("Expected ID to be 'mariadb-backup', got '%s'", pb.GetID())
	}

	expectedDescription := "Create a compressed SQL dump of a MariaDB database"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestBackup_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Backup type.
func TestBackup_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewBackup()
	args := map[string]string{"db-name": "testdb", "root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Backup); !ok {
		t.Error("SetArgs should return *Backup, not just RunnableInterface")
	}
}

// TestBackup_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Backup type.
func TestBackup_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewBackup()

	result := skill.SetArg("db-name", "testdb")

	if _, ok := result.(*Backup); !ok {
		t.Error("SetArg should return *Backup, not just RunnableInterface")
	}
}

// TestBackup_SetID_ReturnsConcreteType verifies that SetID returns the concrete Backup type.
func TestBackup_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewBackup()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Backup); !ok {
		t.Error("SetID should return *Backup, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestBackup_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Backup type.
func TestBackup_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewBackup()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Backup); !ok {
		t.Error("SetDescription should return *Backup, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestBackup_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Backup type.
func TestBackup_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewBackup()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Backup); !ok {
		t.Error("SetTimeout should return *Backup, not just RunnableInterface")
	}
}

// TestBackup_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestBackup_MethodChaining_PreservesType(t *testing.T) {
	skill := NewBackup().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("db-name", "testdb").
		SetArgs(map[string]string{"root-password": "testpass"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Backup); !ok {
		t.Error("Method chaining should preserve *Backup type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
