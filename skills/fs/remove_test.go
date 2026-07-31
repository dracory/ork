package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestRemove_Run_DryRun verifies that dry-run mode reports the would-remove message.
func TestRemove_Run_DryRun(t *testing.T) {
	pb := NewRemove()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/old-data")
	pb.SetArg(ArgRecursive, "true")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would remove: /tmp/old-data"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestRemove_Run_NoPath verifies that missing ArgPath returns an error.
func TestRemove_Run_NoPath(t *testing.T) {
	pb := NewRemove()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no path specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no path is specified")
	}
}

// TestRemove_Run_RelativePath verifies that a relative path returns an error.
func TestRemove_Run_RelativePath(t *testing.T) {
	pb := NewRemove()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "relative/path")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative path")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative path")
	}
}

// TestRemove_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestRemove_Run_NotDryRun(t *testing.T) {
	pb := NewRemove()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/old-data")

	result := pb.Run()

	if result.Message == "Would remove: /tmp/old-data" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestRemove_NewRemove verifies that NewRemove creates a properly configured skill.
func TestRemove_NewRemove(t *testing.T) {
	pb := NewRemove()

	if pb.GetID() != "fs-remove" {
		t.Errorf("Expected ID to be 'fs-remove', got '%s'", pb.GetID())
	}

	expectedDescription := "Remove file or directory (rm)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestRemove_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Remove type.
func TestRemove_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewRemove()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Remove); !ok {
		t.Error("SetArgs should return *Remove, not just RunnableInterface")
	}
}

// TestRemove_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Remove type.
func TestRemove_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewRemove()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*Remove); !ok {
		t.Error("SetArg should return *Remove, not just RunnableInterface")
	}
}

// TestRemove_SetID_ReturnsConcreteType verifies that SetID returns the concrete Remove type.
func TestRemove_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewRemove()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Remove); !ok {
		t.Error("SetID should return *Remove, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestRemove_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Remove type.
func TestRemove_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewRemove()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Remove); !ok {
		t.Error("SetDescription should return *Remove, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestRemove_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Remove type.
func TestRemove_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewRemove()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*Remove); !ok {
		t.Error("SetTimeout should return *Remove, not just RunnableInterface")
	}
}

// TestRemove_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestRemove_MethodChaining_PreservesType(t *testing.T) {
	skill := NewRemove().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Remove); !ok {
		t.Error("Method chaining should preserve *Remove type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestRemove_SetPath verifies that SetPath sets the path arg and returns *Remove.
func TestRemove_SetPath(t *testing.T) {
	skill := NewRemove().SetPath("/tmp/old-data")

	if skill.GetArg(ArgPath) != "/tmp/old-data" {
		t.Errorf("Expected path '/tmp/old-data', got '%s'", skill.GetArg(ArgPath))
	}
}

// TestRemove_SetRecursive verifies that SetRecursive sets the recursive arg and returns *Remove.
func TestRemove_SetRecursive(t *testing.T) {
	skill := NewRemove().SetRecursive(true)

	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}

	skill.SetRecursive(false)
	if skill.GetArg(ArgRecursive) != "false" {
		t.Errorf("Expected recursive 'false', got '%s'", skill.GetArg(ArgRecursive))
	}
}

// TestRemove_SetForce verifies that SetForce sets the force arg and returns *Remove.
func TestRemove_SetForce(t *testing.T) {
	skill := NewRemove().SetForce(true)

	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}

	skill.SetForce(false)
	if skill.GetArg(ArgForce) != "false" {
		t.Errorf("Expected force 'false', got '%s'", skill.GetArg(ArgForce))
	}
}

// TestRemove_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestRemove_TypedSetters_Chaining(t *testing.T) {
	skill := NewRemove().
		SetPath("/tmp/old-data").
		SetRecursive(true).
		SetForce(true)

	if skill.GetArg(ArgPath) != "/tmp/old-data" {
		t.Errorf("Expected path '/tmp/old-data', got '%s'", skill.GetArg(ArgPath))
	}
	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}
	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}
}
