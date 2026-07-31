package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestDirCreate_Run_DryRun verifies that dry-run mode reports the would-create message.
func TestDirCreate_Run_DryRun(t *testing.T) {
	pb := NewDirCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create directory: /var/www/myapp"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestDirCreate_Run_NoPath verifies that missing ArgPath returns an error.
func TestDirCreate_Run_NoPath(t *testing.T) {
	pb := NewDirCreate()

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

// TestDirCreate_Run_RelativePath verifies that a relative path returns an error.
func TestDirCreate_Run_RelativePath(t *testing.T) {
	pb := NewDirCreate()

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

// TestDirCreate_Run_InvalidMode verifies that an invalid mode returns an error.
func TestDirCreate_Run_InvalidMode(t *testing.T) {
	pb := NewDirCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "999")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for invalid mode")
	}

	if result.Error == nil {
		t.Error("Expected an error for invalid mode '999'")
	}
}

// TestDirCreate_Run_InvalidOwner verifies that an invalid owner returns an error.
func TestDirCreate_Run_InvalidOwner(t *testing.T) {
	pb := NewDirCreate()

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

// TestDirCreate_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestDirCreate_Run_NotDryRun(t *testing.T) {
	pb := NewDirCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")

	result := pb.Run()

	// In non-dry-run mode it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create directory: /var/www/myapp" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestDirCreate_NewDirCreate verifies that NewDirCreate creates a properly configured skill.
func TestDirCreate_NewDirCreate(t *testing.T) {
	pb := NewDirCreate()

	if pb.GetID() != "fs-dir-create" {
		t.Errorf("Expected ID to be 'fs-dir-create', got '%s'", pb.GetID())
	}

	expectedDescription := "Create directory with ownership and permissions"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestDirCreate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete DirCreate type.
func TestDirCreate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCreate()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*DirCreate); !ok {
		t.Error("SetArgs should return *DirCreate, not just RunnableInterface")
	}
}

// TestDirCreate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete DirCreate type.
func TestDirCreate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCreate()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*DirCreate); !ok {
		t.Error("SetArg should return *DirCreate, not just RunnableInterface")
	}
}

// TestDirCreate_SetID_ReturnsConcreteType verifies that SetID returns the concrete DirCreate type.
func TestDirCreate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCreate()

	result := skill.SetID("custom-id")

	if _, ok := result.(*DirCreate); !ok {
		t.Error("SetID should return *DirCreate, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestDirCreate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete DirCreate type.
func TestDirCreate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCreate()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*DirCreate); !ok {
		t.Error("SetDescription should return *DirCreate, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestDirCreate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete DirCreate type.
func TestDirCreate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCreate()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*DirCreate); !ok {
		t.Error("SetTimeout should return *DirCreate, not just RunnableInterface")
	}
}

// TestDirCreate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestDirCreate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewDirCreate().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*DirCreate); !ok {
		t.Error("Method chaining should preserve *DirCreate type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestDirCreate_SetPath verifies that SetPath sets the path arg and returns *DirCreate.
func TestDirCreate_SetPath(t *testing.T) {
	skill := NewDirCreate()
	skill.SetPath("/var/www/myapp")

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
}

// TestDirCreate_SetOwner verifies that SetOwner sets the owner arg and returns *DirCreate.
func TestDirCreate_SetOwner(t *testing.T) {
	skill := NewDirCreate()
	skill.SetOwner("www-data:www-data")

	if skill.GetArg(ArgOwner) != "www-data:www-data" {
		t.Errorf("Expected owner 'www-data:www-data', got '%s'", skill.GetArg(ArgOwner))
	}
}

// TestDirCreate_SetMode verifies that SetMode sets the mode arg and returns *DirCreate.
func TestDirCreate_SetMode(t *testing.T) {
	skill := NewDirCreate()
	skill.SetMode("755")

	if skill.GetArg(ArgMode) != "755" {
		t.Errorf("Expected mode '755', got '%s'", skill.GetArg(ArgMode))
	}
}

// TestDirCreate_SetParents verifies that SetParents sets the parents arg as a string bool and returns *DirCreate.
func TestDirCreate_SetParents(t *testing.T) {
	skill := NewDirCreate()
	skill.SetParents(true)

	if skill.GetArg(ArgParents) != "true" {
		t.Errorf("Expected parents 'true', got '%s'", skill.GetArg(ArgParents))
	}

	skill.SetParents(false)
	if skill.GetArg(ArgParents) != "false" {
		t.Errorf("Expected parents 'false', got '%s'", skill.GetArg(ArgParents))
	}
}

// TestDirCreate_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestDirCreate_TypedSetters_Chaining(t *testing.T) {
	skill := NewDirCreate().
		SetPath("/var/www/myapp").
		SetOwner("www-data:www-data").
		SetMode("755").
		SetParents(true)

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
	if skill.GetArg(ArgOwner) != "www-data:www-data" {
		t.Errorf("Expected owner 'www-data:www-data', got '%s'", skill.GetArg(ArgOwner))
	}
	if skill.GetArg(ArgMode) != "755" {
		t.Errorf("Expected mode '755', got '%s'", skill.GetArg(ArgMode))
	}
	if skill.GetArg(ArgParents) != "true" {
		t.Errorf("Expected parents 'true', got '%s'", skill.GetArg(ArgParents))
	}
}
