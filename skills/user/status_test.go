package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUserStatus_Run_DryRun verifies that dry-run mode correctly handles user status.
func TestUserStatus_Run_DryRun(t *testing.T) {
	pb := NewUserStatus()
	pb.SetArg("username", "testuser")

	cfg := types.NodeConfig{
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

// TestUserStatus_Run_RequiresUsername verifies that username is required.
func TestUserStatus_Run_RequiresUsername(t *testing.T) {
	pb := NewUserStatus()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error when username is not provided")
	}

	expectedMessage := "Username is required"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}
}

// TestUserStatus_Check verifies that Check returns false for read-only operation.
func TestUserStatus_Check(t *testing.T) {
	pb := NewUserStatus()

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

// TestUserStatus_NewUserStatus verifies that NewUserStatus creates a properly configured skill.
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
