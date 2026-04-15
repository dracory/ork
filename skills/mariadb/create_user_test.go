package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestCreateUser_Run_DryRun verifies that dry-run mode correctly handles user creation.
func TestCreateUser_Run_DryRun(t *testing.T) {
	pb := NewCreateUser()
	pb.SetArg("username", "testuser")
	pb.SetArg("password", "testpass123")
	pb.SetArg("root-password", "rootpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create user 'testuser'@'%'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestCreateUser_Run_DryRun_WithDB verifies dry-run with database grant.
func TestCreateUser_Run_DryRun_WithDB(t *testing.T) {
	pb := NewCreateUser()
	pb.SetArg("username", "testuser")
	pb.SetArg("password", "testpass123")
	pb.SetArg("root-password", "rootpass123")
	pb.SetArg("db-name", "testdb")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create user 'testuser'@'%'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestCreateUser_Run_DryRun_NoUsername verifies dry-run without username returns error.
func TestCreateUser_Run_DryRun_NoUsername(t *testing.T) {
	pb := NewCreateUser()
	pb.SetArg("password", "testpass123")
	pb.SetArg("root-password", "rootpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing username even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing username")
	}

	if result.Message != "Username is required" {
		t.Errorf("Expected message 'Username is required', got '%s'", result.Message)
	}
}

// TestCreateUser_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestCreateUser_Run_NotDryRun(t *testing.T) {
	pb := NewCreateUser()
	pb.SetArg("username", "testuser")
	pb.SetArg("password", "testpass123")
	pb.SetArg("root-password", "rootpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create user 'testuser'@'%'" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestCreateUser_NewCreateUser verifies that NewCreateUser creates a properly configured skill.
func TestCreateUser_NewCreateUser(t *testing.T) {
	pb := NewCreateUser()

	if pb.GetID() != "mariadb-create-user" {
		t.Errorf("Expected ID to be 'mariadb-create-user', got '%s'", pb.GetID())
	}

	expectedDescription := "Create a new MariaDB user with configurable privileges"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
