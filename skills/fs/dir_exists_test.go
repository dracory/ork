package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestDirExists_Run_DryRun verifies that dry-run mode reports the would-check message.
func TestDirExists_Run_DryRun(t *testing.T) {
	pb := NewDirExists()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false (read-only skill)")
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}

	if result.Details["exists"] != "unknown" {
		t.Errorf("Expected Details['exists'] to be 'unknown' in dry-run, got '%s'", result.Details["exists"])
	}
}

// TestDirExists_Run_NoPath verifies that missing ArgPath returns an error.
func TestDirExists_Run_NoPath(t *testing.T) {
	pb := NewDirExists()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false (read-only skill)")
	}

	if result.Error == nil {
		t.Error("Expected an error when no path is specified")
	}
}

// TestDirExists_Run_RelativePath verifies that a relative path returns an error.
func TestDirExists_Run_RelativePath(t *testing.T) {
	pb := NewDirExists()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "relative/path")

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected an error for relative path")
	}
}

// TestDirExists_Check_AlwaysFalse verifies that Check always returns false (read-only).
func TestDirExists_Check_AlwaysFalse(t *testing.T) {
	pb := NewDirExists()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www")

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false (read-only skill never needs changes)")
	}
}

// TestDirExists_NewDirExists verifies that NewDirExists creates a properly configured skill.
func TestDirExists_NewDirExists(t *testing.T) {
	pb := NewDirExists()

	if pb.GetID() != "fs-dir-exists" {
		t.Errorf("Expected ID to be 'fs-dir-exists', got '%s'", pb.GetID())
	}

	expectedDescription := "Check if directory exists"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestDirExists_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete DirExists type.
func TestDirExists_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewDirExists()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*DirExists); !ok {
		t.Error("SetArgs should return *DirExists, not just RunnableInterface")
	}
}

// TestDirExists_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete DirExists type.
func TestDirExists_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewDirExists()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*DirExists); !ok {
		t.Error("SetArg should return *DirExists, not just RunnableInterface")
	}
}

// TestDirExists_SetID_ReturnsConcreteType verifies that SetID returns the concrete DirExists type.
func TestDirExists_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewDirExists()

	result := skill.SetID("custom-id")

	if _, ok := result.(*DirExists); !ok {
		t.Error("SetID should return *DirExists, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestDirExists_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete DirExists type.
func TestDirExists_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewDirExists()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*DirExists); !ok {
		t.Error("SetDescription should return *DirExists, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestDirExists_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete DirExists type.
func TestDirExists_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewDirExists()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*DirExists); !ok {
		t.Error("SetTimeout should return *DirExists, not just RunnableInterface")
	}
}

// TestDirExists_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestDirExists_MethodChaining_PreservesType(t *testing.T) {
	skill := NewDirExists().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*DirExists); !ok {
		t.Error("Method chaining should preserve *DirExists type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestDirExists_SetPath verifies that SetPath sets the path arg and returns *DirExists.
func TestDirExists_SetPath(t *testing.T) {
	skill := NewDirExists().SetPath("/var/www")

	if skill.GetArg(ArgPath) != "/var/www" {
		t.Errorf("Expected path '/var/www', got '%s'", skill.GetArg(ArgPath))
	}
}
