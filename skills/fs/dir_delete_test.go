package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestDirDelete_Run_DryRun verifies that dry-run mode reports the would-delete message.
func TestDirDelete_Run_DryRun(t *testing.T) {
	pb := NewDirDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/old-build")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would delete directory: /tmp/old-build"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestDirDelete_Run_NoPath verifies that missing ArgPath returns an error.
func TestDirDelete_Run_NoPath(t *testing.T) {
	pb := NewDirDelete()

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

// TestDirDelete_Run_RelativePath verifies that a relative path returns an error.
func TestDirDelete_Run_RelativePath(t *testing.T) {
	pb := NewDirDelete()

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

// TestDirDelete_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestDirDelete_Run_NotDryRun(t *testing.T) {
	pb := NewDirDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/old-build")

	result := pb.Run()

	if result.Message == "Would delete directory: /tmp/old-build" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestDirDelete_NewDirDelete verifies that NewDirDelete creates a properly configured skill.
func TestDirDelete_NewDirDelete(t *testing.T) {
	pb := NewDirDelete()

	if pb.GetID() != "fs-dir-delete" {
		t.Errorf("Expected ID to be 'fs-dir-delete', got '%s'", pb.GetID())
	}

	expectedDescription := "Delete a directory"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestDirDelete_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete DirDelete type.
func TestDirDelete_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewDirDelete()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*DirDelete); !ok {
		t.Error("SetArgs should return *DirDelete, not just RunnableInterface")
	}
}

// TestDirDelete_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete DirDelete type.
func TestDirDelete_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewDirDelete()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*DirDelete); !ok {
		t.Error("SetArg should return *DirDelete, not just RunnableInterface")
	}
}

// TestDirDelete_SetID_ReturnsConcreteType verifies that SetID returns the concrete DirDelete type.
func TestDirDelete_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewDirDelete()

	result := skill.SetID("custom-id")

	if _, ok := result.(*DirDelete); !ok {
		t.Error("SetID should return *DirDelete, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestDirDelete_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete DirDelete type.
func TestDirDelete_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewDirDelete()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*DirDelete); !ok {
		t.Error("SetDescription should return *DirDelete, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestDirDelete_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete DirDelete type.
func TestDirDelete_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewDirDelete()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*DirDelete); !ok {
		t.Error("SetTimeout should return *DirDelete, not just RunnableInterface")
	}
}

// TestDirDelete_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestDirDelete_MethodChaining_PreservesType(t *testing.T) {
	skill := NewDirDelete().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*DirDelete); !ok {
		t.Error("Method chaining should preserve *DirDelete type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
