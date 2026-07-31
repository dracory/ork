package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestFileDelete_Run_DryRun verifies that dry-run mode reports the would-delete message.
func TestFileDelete_Run_DryRun(t *testing.T) {
	pb := NewFileDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/test.log")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would delete file: /tmp/test.log"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFileDelete_Run_NoPath verifies that missing ArgPath returns an error.
func TestFileDelete_Run_NoPath(t *testing.T) {
	pb := NewFileDelete()

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

// TestFileDelete_Run_RelativePath verifies that a relative path returns an error.
func TestFileDelete_Run_RelativePath(t *testing.T) {
	pb := NewFileDelete()

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

// TestFileDelete_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestFileDelete_Run_NotDryRun(t *testing.T) {
	pb := NewFileDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/tmp/test.log")

	result := pb.Run()

	if result.Message == "Would delete file: /tmp/test.log" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestFileDelete_NewFileDelete verifies that NewFileDelete creates a properly configured skill.
func TestFileDelete_NewFileDelete(t *testing.T) {
	pb := NewFileDelete()

	if pb.GetID() != "fs-file-delete" {
		t.Errorf("Expected ID to be 'fs-file-delete', got '%s'", pb.GetID())
	}

	expectedDescription := "Delete a single file"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFileDelete_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete FileDelete type.
func TestFileDelete_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFileDelete()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*FileDelete); !ok {
		t.Error("SetArgs should return *FileDelete, not just RunnableInterface")
	}
}

// TestFileDelete_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete FileDelete type.
func TestFileDelete_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFileDelete()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*FileDelete); !ok {
		t.Error("SetArg should return *FileDelete, not just RunnableInterface")
	}
}

// TestFileDelete_SetID_ReturnsConcreteType verifies that SetID returns the concrete FileDelete type.
func TestFileDelete_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFileDelete()

	result := skill.SetID("custom-id")

	if _, ok := result.(*FileDelete); !ok {
		t.Error("SetID should return *FileDelete, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFileDelete_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete FileDelete type.
func TestFileDelete_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFileDelete()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*FileDelete); !ok {
		t.Error("SetDescription should return *FileDelete, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFileDelete_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete FileDelete type.
func TestFileDelete_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFileDelete()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*FileDelete); !ok {
		t.Error("SetTimeout should return *FileDelete, not just RunnableInterface")
	}
}

// TestFileDelete_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFileDelete_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFileDelete().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*FileDelete); !ok {
		t.Error("Method chaining should preserve *FileDelete type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
