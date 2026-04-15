package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestUserList_Run_DryRun verifies that dry-run mode correctly handles user list.
func TestUserList_Run_DryRun(t *testing.T) {
	pb := NewUserList()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// List is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would list all non-system users"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserList_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUserList_Run_NotDryRun(t *testing.T) {
	pb := NewUserList()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would list all non-system users" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// List is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestUserList_Check verifies that Check returns false for read-only operation.
func TestUserList_Check(t *testing.T) {
	pb := NewUserList()

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

// TestUserList_NewUserList verifies that NewUserList creates a properly configured skill.
func TestUserList_NewUserList(t *testing.T) {
	pb := NewUserList()

	if pb.GetID() != "user-list" {
		t.Errorf("Expected ID to be 'user-list', got '%s'", pb.GetID())
	}

	expectedDescription := "List all non-system users"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
