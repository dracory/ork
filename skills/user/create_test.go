package user

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
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

// TestUserCreate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete UserCreate type.
func TestUserCreate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewUserCreate()
	args := map[string]string{"username": "testuser"}

	result := skill.SetArgs(args)

	// Verify the returned value is still a UserCreate (not just BaseSkill)
	if _, ok := result.(*UserCreate); !ok {
		t.Error("SetArgs should return *UserCreate, not just RunnableInterface")
	}

	// Verify the args were actually set
	if skill.GetArg("username") != "testuser" {
		t.Error("SetArgs should set the arguments")
	}
}

// TestUserCreate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete UserCreate type.
func TestUserCreate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewUserCreate()

	result := skill.SetArg("username", "testuser")

	// Verify the returned value is still a UserCreate (not just BaseSkill)
	if _, ok := result.(*UserCreate); !ok {
		t.Error("SetArg should return *UserCreate, not just RunnableInterface")
	}

	// Verify the arg was actually set
	if skill.GetArg("username") != "testuser" {
		t.Error("SetArg should set the argument")
	}
}

// TestUserCreate_SetID_ReturnsConcreteType verifies that SetID returns the concrete UserCreate type.
func TestUserCreate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewUserCreate()

	result := skill.SetID("custom-id")

	// Verify the returned value is still a UserCreate (not just BaseSkill)
	if _, ok := result.(*UserCreate); !ok {
		t.Error("SetID should return *UserCreate, not just RunnableInterface")
	}

	// Verify the ID was actually set
	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestUserCreate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete UserCreate type.
func TestUserCreate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewUserCreate()

	result := skill.SetDescription("custom description")

	// Verify the returned value is still a UserCreate (not just BaseSkill)
	if _, ok := result.(*UserCreate); !ok {
		t.Error("SetDescription should return *UserCreate, not just RunnableInterface")
	}

	// Verify the description was actually set
	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestUserCreate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete UserCreate type.
func TestUserCreate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewUserCreate()

	result := skill.SetTimeout(30 * 1000000000) // 30 seconds

	// Verify the returned value is still a UserCreate (not just BaseSkill)
	if _, ok := result.(*UserCreate); !ok {
		t.Error("SetTimeout should return *UserCreate, not just RunnableInterface")
	}
}

// TestUserCreate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestUserCreate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewUserCreate().
		SetID("custom-id").
		SetDescription("custom description").
		SetTimeout(30 * 1000000000).
		SetArgs(map[string]string{"username": "testuser", "ssh-key": "ssh-rsa test"})

	// Verify the final result is still a UserCreate (not just BaseSkill)
	if _, ok := skill.(*UserCreate); !ok {
		t.Error("Method chaining should preserve *UserCreate type")
	}

	// Verify all values were set correctly
	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}

	if skill.GetArg("username") != "testuser" {
		t.Error("Method chaining should set args")
	}

	if skill.GetArg("ssh-key") != "ssh-rsa test" {
		t.Error("Method chaining should set ssh-key arg")
	}
}
