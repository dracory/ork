package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestChangeOwner_Run_DryRun verifies that dry-run mode reports the would-change message.
func TestChangeOwner_Run_DryRun(t *testing.T) {
	pb := NewChangeOwner()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgOwner, "www-data:www-data")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would change owner to www-data:www-data on /var/www/myapp"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestChangeOwner_Run_NoPath verifies that missing ArgPath returns an error.
func TestChangeOwner_Run_NoPath(t *testing.T) {
	pb := NewChangeOwner()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgOwner, "www-data:www-data")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no path specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no path is specified")
	}
}

// TestChangeOwner_Run_NoOwner verifies that missing ArgOwner returns an error.
func TestChangeOwner_Run_NoOwner(t *testing.T) {
	pb := NewChangeOwner()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no owner specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no owner is specified")
	}
}

// TestChangeOwner_Run_InvalidOwner verifies that an invalid owner returns an error.
func TestChangeOwner_Run_InvalidOwner(t *testing.T) {
	pb := NewChangeOwner()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgOwner, "bad owner!") // space is not allowed

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for invalid owner")
	}

	if result.Error == nil {
		t.Error("Expected an error for invalid owner")
	}
}

// TestChangeOwner_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestChangeOwner_Run_NotDryRun(t *testing.T) {
	pb := NewChangeOwner()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgOwner, "www-data:www-data")

	result := pb.Run()

	if result.Message == "Would change owner to www-data:www-data on /var/www/myapp" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestChangeOwner_NewChangeOwner verifies that NewChangeOwner creates a properly configured skill.
func TestChangeOwner_NewChangeOwner(t *testing.T) {
	pb := NewChangeOwner()

	if pb.GetID() != "fs-change-owner" {
		t.Errorf("Expected ID to be 'fs-change-owner', got '%s'", pb.GetID())
	}

	expectedDescription := "Change file/directory ownership (chown)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestChangeOwner_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete ChangeOwner type.
func TestChangeOwner_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeOwner()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*ChangeOwner); !ok {
		t.Error("SetArgs should return *ChangeOwner, not just RunnableInterface")
	}
}

// TestChangeOwner_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete ChangeOwner type.
func TestChangeOwner_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeOwner()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*ChangeOwner); !ok {
		t.Error("SetArg should return *ChangeOwner, not just RunnableInterface")
	}
}

// TestChangeOwner_SetID_ReturnsConcreteType verifies that SetID returns the concrete ChangeOwner type.
func TestChangeOwner_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeOwner()

	result := skill.SetID("custom-id")

	if _, ok := result.(*ChangeOwner); !ok {
		t.Error("SetID should return *ChangeOwner, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestChangeOwner_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete ChangeOwner type.
func TestChangeOwner_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeOwner()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*ChangeOwner); !ok {
		t.Error("SetDescription should return *ChangeOwner, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestChangeOwner_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete ChangeOwner type.
func TestChangeOwner_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeOwner()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*ChangeOwner); !ok {
		t.Error("SetTimeout should return *ChangeOwner, not just RunnableInterface")
	}
}

// TestChangeOwner_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestChangeOwner_MethodChaining_PreservesType(t *testing.T) {
	skill := NewChangeOwner().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*ChangeOwner); !ok {
		t.Error("Method chaining should preserve *ChangeOwner type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestChangeOwner_SetPath verifies that SetPath sets the path arg and returns *ChangeOwner.
func TestChangeOwner_SetPath(t *testing.T) {
	skill := NewChangeOwner()
	skill.SetPath("/var/www/myapp")

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
}

// TestChangeOwner_SetOwner verifies that SetOwner sets the owner arg and returns *ChangeOwner.
func TestChangeOwner_SetOwner(t *testing.T) {
	skill := NewChangeOwner()
	skill.SetOwner("www-data:www-data")

	if skill.GetArg(ArgOwner) != "www-data:www-data" {
		t.Errorf("Expected owner 'www-data:www-data', got '%s'", skill.GetArg(ArgOwner))
	}
}

// TestChangeOwner_SetRecursive verifies that SetRecursive sets the recursive arg as a string bool and returns *ChangeOwner.
func TestChangeOwner_SetRecursive(t *testing.T) {
	skill := NewChangeOwner()
	skill.SetRecursive(true)

	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}

	skill.SetRecursive(false)
	if skill.GetArg(ArgRecursive) != "false" {
		t.Errorf("Expected recursive 'false', got '%s'", skill.GetArg(ArgRecursive))
	}
}

// TestChangeOwner_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestChangeOwner_TypedSetters_Chaining(t *testing.T) {
	skill := NewChangeOwner().
		SetPath("/var/www/myapp").
		SetOwner("www-data:www-data").
		SetRecursive(true)

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
	if skill.GetArg(ArgOwner) != "www-data:www-data" {
		t.Errorf("Expected owner 'www-data:www-data', got '%s'", skill.GetArg(ArgOwner))
	}
	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}
}
