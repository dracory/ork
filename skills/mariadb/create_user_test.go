package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
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

// TestCreateUser_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete CreateUser type.
func TestCreateUser_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateUser()
	args := map[string]string{"username": "testuser", "password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*CreateUser); !ok {
		t.Error("SetArgs should return *CreateUser, not just RunnableInterface")
	}
}

// TestCreateUser_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete CreateUser type.
func TestCreateUser_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateUser()

	result := skill.SetArg("username", "testuser")

	if _, ok := result.(*CreateUser); !ok {
		t.Error("SetArg should return *CreateUser, not just RunnableInterface")
	}
}

// TestCreateUser_SetID_ReturnsConcreteType verifies that SetID returns the concrete CreateUser type.
func TestCreateUser_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateUser()

	result := skill.SetID("custom-id")

	if _, ok := result.(*CreateUser); !ok {
		t.Error("SetID should return *CreateUser, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestCreateUser_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete CreateUser type.
func TestCreateUser_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateUser()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*CreateUser); !ok {
		t.Error("SetDescription should return *CreateUser, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestCreateUser_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete CreateUser type.
func TestCreateUser_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewCreateUser()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*CreateUser); !ok {
		t.Error("SetTimeout should return *CreateUser, not just RunnableInterface")
	}
}

// TestCreateUser_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestCreateUser_MethodChaining_PreservesType(t *testing.T) {
	skill := NewCreateUser().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("username", "testuser").
		SetArgs(map[string]string{"password": "testpass"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*CreateUser); !ok {
		t.Error("Method chaining should preserve *CreateUser type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
