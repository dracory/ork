package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestFileCreate_Run_DryRun verifies that dry-run mode reports the would-create message.
func TestFileCreate_Run_DryRun(t *testing.T) {
	pb := NewFileCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp/config.json")
	pb.SetArg(ArgContent, `{"key": "value"}`)
	pb.SetArg(ArgMode, "644")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create file: /var/www/myapp/config.json"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFileCreate_Run_NoPath verifies that missing ArgPath returns an error.
func TestFileCreate_Run_NoPath(t *testing.T) {
	pb := NewFileCreate()

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

// TestFileCreate_Run_RelativePath verifies that a relative path returns an error.
func TestFileCreate_Run_RelativePath(t *testing.T) {
	pb := NewFileCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "relative/path")
	pb.SetArg(ArgContent, "test")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative path")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative path")
	}
}

// TestFileCreate_Run_InvalidMode verifies that an invalid mode returns an error.
func TestFileCreate_Run_InvalidMode(t *testing.T) {
	pb := NewFileCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp/config.json")
	pb.SetArg(ArgContent, "test")
	pb.SetArg(ArgMode, "999") // 9 is not a valid octal digit

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for invalid mode")
	}

	if result.Error == nil {
		t.Error("Expected an error for invalid mode '999'")
	}
}

// TestFileCreate_Run_InvalidOwner verifies that an invalid owner returns an error.
func TestFileCreate_Run_InvalidOwner(t *testing.T) {
	pb := NewFileCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp/config.json")
	pb.SetArg(ArgContent, "test")
	pb.SetArg(ArgOwner, "bad owner!") // space is not allowed

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for invalid owner")
	}

	if result.Error == nil {
		t.Error("Expected an error for invalid owner")
	}
}

// TestFileCreate_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestFileCreate_Run_NotDryRun(t *testing.T) {
	pb := NewFileCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp/config.json")
	pb.SetArg(ArgContent, "test")

	result := pb.Run()

	if result.Message == "Would create file: /var/www/myapp/config.json" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestFileCreate_NewFileCreate verifies that NewFileCreate creates a properly configured skill.
func TestFileCreate_NewFileCreate(t *testing.T) {
	pb := NewFileCreate()

	if pb.GetID() != "fs-file-create" {
		t.Errorf("Expected ID to be 'fs-file-create', got '%s'", pb.GetID())
	}

	expectedDescription := "Create file with content, ownership, and permissions"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFileCreate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete FileCreate type.
func TestFileCreate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCreate()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*FileCreate); !ok {
		t.Error("SetArgs should return *FileCreate, not just RunnableInterface")
	}
}

// TestFileCreate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete FileCreate type.
func TestFileCreate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCreate()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*FileCreate); !ok {
		t.Error("SetArg should return *FileCreate, not just RunnableInterface")
	}
}

// TestFileCreate_SetID_ReturnsConcreteType verifies that SetID returns the concrete FileCreate type.
func TestFileCreate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCreate()

	result := skill.SetID("custom-id")

	if _, ok := result.(*FileCreate); !ok {
		t.Error("SetID should return *FileCreate, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFileCreate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete FileCreate type.
func TestFileCreate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCreate()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*FileCreate); !ok {
		t.Error("SetDescription should return *FileCreate, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFileCreate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete FileCreate type.
func TestFileCreate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCreate()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*FileCreate); !ok {
		t.Error("SetTimeout should return *FileCreate, not just RunnableInterface")
	}
}

// TestFileCreate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFileCreate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFileCreate().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*FileCreate); !ok {
		t.Error("Method chaining should preserve *FileCreate type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
