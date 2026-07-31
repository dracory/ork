package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUserAddToGroup_Run_DryRun verifies that dry-run mode correctly handles adding user to group.
func TestUserAddToGroup_Run_DryRun(t *testing.T) {
	pb := NewUserAddToGroup()
	pb.SetArg("username", "testuser")
	pb.SetArg("group", "www-data")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would add user 'testuser' to group 'www-data'"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUserAddToGroup_Run_DryRun_NoUsername verifies dry-run without username returns error.
func TestUserAddToGroup_Run_DryRun_NoUsername(t *testing.T) {
	pb := NewUserAddToGroup()
	pb.SetArg("group", "www-data")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for missing username")
	}

	if result.Message != "Username is required" {
		t.Errorf("Expected message 'Username is required', got '%s'", result.Message)
	}
}

// TestUserAddToGroup_Run_DryRun_NoGroup verifies dry-run without group returns error.
func TestUserAddToGroup_Run_DryRun_NoGroup(t *testing.T) {
	pb := NewUserAddToGroup()
	pb.SetArg("username", "testuser")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for missing group")
	}

	if result.Message != "Group is required" {
		t.Errorf("Expected message 'Group is required', got '%s'", result.Message)
	}
}

// TestUserAddToGroup_Run_NotDryRun verifies that non-dry-run mode doesn't return dry-run message.
func TestUserAddToGroup_Run_NotDryRun(t *testing.T) {
	pb := NewUserAddToGroup()
	pb.SetArg("username", "testuser")
	pb.SetArg("group", "www-data")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Message == "Would add user 'testuser' to group 'www-data'" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestUserAddToGroup_NewUserAddToGroup verifies that NewUserAddToGroup creates a properly configured skill.
func TestUserAddToGroup_NewUserAddToGroup(t *testing.T) {
	pb := NewUserAddToGroup()

	if pb.GetID() != "user-add-to-group" {
		t.Errorf("Expected ID to be 'user-add-to-group', got '%s'", pb.GetID())
	}

	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

// TestUserAddToGroup_SetUsername verifies that SetUsername sets the username arg.
func TestUserAddToGroup_SetUsername(t *testing.T) {
	skill := NewUserAddToGroup().SetUsername("alice")

	if skill.GetArg(ArgUsername) != "alice" {
		t.Errorf("Expected username 'alice', got '%s'", skill.GetArg(ArgUsername))
	}
}

// TestUserAddToGroup_SetGroup verifies that SetGroup sets the group arg.
func TestUserAddToGroup_SetGroup(t *testing.T) {
	skill := NewUserAddToGroup().SetGroup("docker")

	if skill.GetArg(ArgGroup) != "docker" {
		t.Errorf("Expected group 'docker', got '%s'", skill.GetArg(ArgGroup))
	}
}

// TestUserAddToGroup_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestUserAddToGroup_TypedSetters_Chaining(t *testing.T) {
	skill := NewUserAddToGroup().
		SetUsername("alice").
		SetGroup("docker")

	if skill.GetArg(ArgUsername) != "alice" {
		t.Errorf("Expected username 'alice', got '%s'", skill.GetArg(ArgUsername))
	}
	if skill.GetArg(ArgGroup) != "docker" {
		t.Errorf("Expected group 'docker', got '%s'", skill.GetArg(ArgGroup))
	}
}
