package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestListUsers_Run_DryRun verifies that dry-run mode correctly handles user listing.
func TestListUsers_Run_DryRun(t *testing.T) {
	pb := NewListUsers()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// List is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would list all database users"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestListUsers_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestListUsers_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewListUsers()

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

// TestListUsers_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestListUsers_Run_NotDryRun(t *testing.T) {
	pb := NewListUsers()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would list all database users" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// List is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestListUsers_Check verifies that Check returns false for read-only operation.
func TestListUsers_Check(t *testing.T) {
	pb := NewListUsers()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestListUsers_NewListUsers verifies that NewListUsers creates a properly configured skill.
func TestListUsers_NewListUsers(t *testing.T) {
	pb := NewListUsers()

	if pb.GetID() != "mariadb-list-users" {
		t.Errorf("Expected ID to be 'mariadb-list-users', got '%s'", pb.GetID())
	}

	expectedDescription := "Display all database user accounts and their allowed hosts (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
