package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestUserCreate_Run_DryRun verifies that dry-run mode correctly handles user creation.
func TestUserCreate_Run_DryRun(t *testing.T) {
	pb := NewUserCreate()
	pb.SetArg("username", "testuser")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create user: testuser"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserCreate_Run_DryRun_WithSSHKey verifies dry-run with SSH key.
func TestUserCreate_Run_DryRun_WithSSHKey(t *testing.T) {
	pb := NewUserCreate()
	pb.SetArg("username", "testuser")
	pb.SetArg("ssh-key", "ssh-rsa AAAAB3NzaC1yc2E...")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create user: testuser"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserCreate_Run_DryRun_NoUsername verifies dry-run without username returns error.
func TestUserCreate_Run_DryRun_NoUsername(t *testing.T) {
	pb := NewUserCreate()

	cfg := types.NodeConfig{
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

// TestUserCreate_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUserCreate_Run_NotDryRun(t *testing.T) {
	pb := NewUserCreate()
	pb.SetArg("username", "testuser")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create user: testuser" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestUserCreate_NewUserCreate verifies that NewUserCreate creates a properly configured skill.
func TestUserCreate_NewUserCreate(t *testing.T) {
	pb := NewUserCreate()

	if pb.GetID() != "user-create" {
		t.Errorf("Expected ID to be 'user-create', got '%s'", pb.GetID())
	}

	expectedDescription := "Create a new user with sudo access (username via args['username'])"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
