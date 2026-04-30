package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestBackupEncrypt_Run_DryRun verifies that dry-run mode correctly handles encrypted MariaDB backup.
func TestBackupEncrypt_Run_DryRun(t *testing.T) {
	pb := NewBackupEncrypt()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			"dbname":        "testdb",
			"root-password": "testpass123",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create encrypted backup for database 'testdb'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestBackupEncrypt_Run_DryRun_NoDbName verifies dry-run without database name returns error.
func TestBackupEncrypt_Run_DryRun_NoDbName(t *testing.T) {
	pb := NewBackupEncrypt()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			"root-password": "testpass123",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing dbname even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing dbname")
	}

	if result.Message != "Database name is required" {
		t.Errorf("Expected message 'Database name is required', got '%s'", result.Message)
	}
}

// TestBackupEncrypt_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestBackupEncrypt_Run_NotDryRun(t *testing.T) {
	pb := NewBackupEncrypt()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args: map[string]string{
			"dbname":        "testdb",
			"root-password": "testpass123",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create encrypted backup for database 'testdb'" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestBackupEncrypt_NewBackupEncrypt verifies that NewBackupEncrypt creates a properly configured skill.
func TestBackupEncrypt_NewBackupEncrypt(t *testing.T) {
	pb := NewBackupEncrypt()

	if pb.GetID() != "mariadb-backup-encrypt" {
		t.Errorf("Expected ID to be 'mariadb-backup-encrypt', got '%s'", pb.GetID())
	}

	expectedDescription := "Create an encrypted backup of a MariaDB database"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestBackupEncrypt_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete BackupEncrypt type.
func TestBackupEncrypt_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupEncrypt()
	args := map[string]string{"dbname": "testdb", "root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*BackupEncrypt); !ok {
		t.Error("SetArgs should return *BackupEncrypt, not just RunnableInterface")
	}
}

// TestBackupEncrypt_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete BackupEncrypt type.
func TestBackupEncrypt_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupEncrypt()

	result := skill.SetArg("dbname", "testdb")

	if _, ok := result.(*BackupEncrypt); !ok {
		t.Error("SetArg should return *BackupEncrypt, not just RunnableInterface")
	}
}

// TestBackupEncrypt_SetID_ReturnsConcreteType verifies that SetID returns the concrete BackupEncrypt type.
func TestBackupEncrypt_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupEncrypt()

	result := skill.SetID("custom-id")

	if _, ok := result.(*BackupEncrypt); !ok {
		t.Error("SetID should return *BackupEncrypt, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestBackupEncrypt_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete BackupEncrypt type.
func TestBackupEncrypt_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupEncrypt()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*BackupEncrypt); !ok {
		t.Error("SetDescription should return *BackupEncrypt, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestBackupEncrypt_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete BackupEncrypt type.
func TestBackupEncrypt_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupEncrypt()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*BackupEncrypt); !ok {
		t.Error("SetTimeout should return *BackupEncrypt, not just RunnableInterface")
	}
}

// TestBackupEncrypt_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestBackupEncrypt_MethodChaining_PreservesType(t *testing.T) {
	skill := NewBackupEncrypt().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("dbname", "testdb").
		SetArgs(map[string]string{"root-password": "testpass"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*BackupEncrypt); !ok {
		t.Error("Method chaining should preserve *BackupEncrypt type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
