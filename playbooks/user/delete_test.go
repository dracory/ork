package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestUserDelete_Run_DryRun verifies that dry-run mode correctly handles user deletion.
func TestUserDelete_Run_DryRun(t *testing.T) {
	pb := NewUserDelete()
	pb.SetArg("username", "testuser")

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would delete user: testuser"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserDelete_Run_DryRun_NoUsername verifies dry-run without username returns error.
func TestUserDelete_Run_DryRun_NoUsername(t *testing.T) {
	pb := NewUserDelete()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
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

// TestUserDelete_Run_DryRun_RootUser verifies dry-run prevents root deletion.
func TestUserDelete_Run_DryRun_RootUser(t *testing.T) {
	pb := NewUserDelete()
	pb.SetArg("username", "root")

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for trying to delete root even in dry-run
	if result.Error == nil {
		t.Error("Expected error for attempting to delete root user")
	}

	if result.Message != "Cannot delete root user" {
		t.Errorf("Expected message 'Cannot delete root user', got '%s'", result.Message)
	}
}

// TestUserDelete_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUserDelete_Run_NotDryRun(t *testing.T) {
	pb := NewUserDelete()
	pb.SetArg("username", "testuser")

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would delete user: testuser" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestUserDelete_NewUserDelete verifies that NewUserDelete creates a properly configured playbook.
func TestUserDelete_NewUserDelete(t *testing.T) {
	pb := NewUserDelete()

	if pb.GetID() != "user-delete" {
		t.Errorf("Expected ID to be 'user-delete', got '%s'", pb.GetID())
	}

	expectedDescription := "Delete a user (username via args['username'])"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
