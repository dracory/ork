package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestUserStatus_Run_DryRun verifies that dry-run mode correctly handles user status.
func TestUserStatus_Run_DryRun(t *testing.T) {
	pb := NewUserStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would list all system users"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserStatus_Run_DryRun_WithUsername verifies dry-run with specific username.
func TestUserStatus_Run_DryRun_WithUsername(t *testing.T) {
	pb := NewUserStatus()
	pb.SetArg("username", "testuser")

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check user status for 'testuser'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUserStatus_Run_NotDryRun(t *testing.T) {
	pb := NewUserStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would list all system users" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestUserStatus_Check verifies that Check returns false for read-only operation.
func TestUserStatus_Check(t *testing.T) {
	pb := NewUserStatus()

	cfg := config.NodeConfig{
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

// TestUserStatus_NewUserStatus verifies that NewUserStatus creates a properly configured playbook.
func TestUserStatus_NewUserStatus(t *testing.T) {
	pb := NewUserStatus()

	if pb.GetID() != "user-status" {
		t.Errorf("Expected ID to be 'user-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Show user information"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
