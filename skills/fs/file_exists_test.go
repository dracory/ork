package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestFileExists_Run_DryRun verifies that dry-run mode reports the would-check message.
func TestFileExists_Run_DryRun(t *testing.T) {
	pb := NewFileExists()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/etc/hostname")

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

// TestFileExists_Run_NoPath verifies that missing ArgPath returns an error.
func TestFileExists_Run_NoPath(t *testing.T) {
	pb := NewFileExists()

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

// TestFileExists_Run_RelativePath verifies that a relative path returns an error.
func TestFileExists_Run_RelativePath(t *testing.T) {
	pb := NewFileExists()

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

// TestFileExists_Check_AlwaysFalse verifies that Check always returns false (read-only).
func TestFileExists_Check_AlwaysFalse(t *testing.T) {
	pb := NewFileExists()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/etc/hostname")

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false (read-only skill never needs changes)")
	}
}

// TestFileExists_NewFileExists verifies that NewFileExists creates a properly configured skill.
func TestFileExists_NewFileExists(t *testing.T) {
	pb := NewFileExists()

	if pb.GetID() != "fs-file-exists" {
		t.Errorf("Expected ID to be 'fs-file-exists', got '%s'", pb.GetID())
	}

	expectedDescription := "Check if file exists"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFileExists_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete FileExists type.
func TestFileExists_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFileExists()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*FileExists); !ok {
		t.Error("SetArgs should return *FileExists, not just RunnableInterface")
	}
}

// TestFileExists_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete FileExists type.
func TestFileExists_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFileExists()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*FileExists); !ok {
		t.Error("SetArg should return *FileExists, not just RunnableInterface")
	}
}

// TestFileExists_SetID_ReturnsConcreteType verifies that SetID returns the concrete FileExists type.
func TestFileExists_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFileExists()

	result := skill.SetID("custom-id")

	if _, ok := result.(*FileExists); !ok {
		t.Error("SetID should return *FileExists, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFileExists_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete FileExists type.
func TestFileExists_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFileExists()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*FileExists); !ok {
		t.Error("SetDescription should return *FileExists, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFileExists_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete FileExists type.
func TestFileExists_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFileExists()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*FileExists); !ok {
		t.Error("SetTimeout should return *FileExists, not just RunnableInterface")
	}
}

// TestFileExists_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFileExists_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFileExists().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*FileExists); !ok {
		t.Error("Method chaining should preserve *FileExists type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
