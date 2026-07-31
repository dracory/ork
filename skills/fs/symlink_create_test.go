package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestSymlinkCreate_Run_DryRun verifies that dry-run mode reports the would-create message.
func TestSymlinkCreate_Run_DryRun(t *testing.T) {
	pb := NewSymlinkCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgTarget, "/opt/node/bin/pm2")
	pb.SetArg(ArgLink, "/usr/local/bin/pm2")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would create symlink: /usr/local/bin/pm2 -> /opt/node/bin/pm2"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSymlinkCreate_Run_NoTarget verifies that missing ArgTarget returns an error.
func TestSymlinkCreate_Run_NoTarget(t *testing.T) {
	pb := NewSymlinkCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgLink, "/usr/local/bin/pm2")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no target specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no target is specified")
	}
}

// TestSymlinkCreate_Run_NoLink verifies that missing ArgLink returns an error.
func TestSymlinkCreate_Run_NoLink(t *testing.T) {
	pb := NewSymlinkCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgTarget, "/opt/node/bin/pm2")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no link specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no link is specified")
	}
}

// TestSymlinkCreate_Run_RelativeTarget verifies that a relative target returns an error.
func TestSymlinkCreate_Run_RelativeTarget(t *testing.T) {
	pb := NewSymlinkCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgTarget, "relative/path")
	pb.SetArg(ArgLink, "/usr/local/bin/pm2")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative target")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative target")
	}
}

// TestSymlinkCreate_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestSymlinkCreate_Run_NotDryRun(t *testing.T) {
	pb := NewSymlinkCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgTarget, "/opt/node/bin/pm2")
	pb.SetArg(ArgLink, "/usr/local/bin/pm2")

	result := pb.Run()

	if result.Message == "Would create symlink: /usr/local/bin/pm2 -> /opt/node/bin/pm2" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSymlinkCreate_NewSymlinkCreate verifies that NewSymlinkCreate creates a properly configured skill.
func TestSymlinkCreate_NewSymlinkCreate(t *testing.T) {
	pb := NewSymlinkCreate()

	if pb.GetID() != "fs-symlink-create" {
		t.Errorf("Expected ID to be 'fs-symlink-create', got '%s'", pb.GetID())
	}

	expectedDescription := "Create or update symbolic link (ln -sf)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestSymlinkCreate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete SymlinkCreate type.
func TestSymlinkCreate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSymlinkCreate()
	args := map[string]string{ArgTarget: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*SymlinkCreate); !ok {
		t.Error("SetArgs should return *SymlinkCreate, not just RunnableInterface")
	}
}

// TestSymlinkCreate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete SymlinkCreate type.
func TestSymlinkCreate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSymlinkCreate()

	result := skill.SetArg(ArgTarget, "/tmp/test")

	if _, ok := result.(*SymlinkCreate); !ok {
		t.Error("SetArg should return *SymlinkCreate, not just RunnableInterface")
	}
}

// TestSymlinkCreate_SetID_ReturnsConcreteType verifies that SetID returns the concrete SymlinkCreate type.
func TestSymlinkCreate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSymlinkCreate()

	result := skill.SetID("custom-id")

	if _, ok := result.(*SymlinkCreate); !ok {
		t.Error("SetID should return *SymlinkCreate, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSymlinkCreate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete SymlinkCreate type.
func TestSymlinkCreate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSymlinkCreate()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*SymlinkCreate); !ok {
		t.Error("SetDescription should return *SymlinkCreate, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSymlinkCreate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete SymlinkCreate type.
func TestSymlinkCreate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSymlinkCreate()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*SymlinkCreate); !ok {
		t.Error("SetTimeout should return *SymlinkCreate, not just RunnableInterface")
	}
}

// TestSymlinkCreate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSymlinkCreate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSymlinkCreate().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgTarget, "/tmp/test").
		SetArgs(map[string]string{ArgTarget: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*SymlinkCreate); !ok {
		t.Error("Method chaining should preserve *SymlinkCreate type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestSymlinkCreate_SetTarget verifies that SetTarget sets the target arg and returns *SymlinkCreate.
func TestSymlinkCreate_SetTarget(t *testing.T) {
	skill := NewSymlinkCreate()
	skill.SetTarget("/opt/node/bin/pm2")

	if skill.GetArg(ArgTarget) != "/opt/node/bin/pm2" {
		t.Errorf("Expected target '/opt/node/bin/pm2', got '%s'", skill.GetArg(ArgTarget))
	}
}

// TestSymlinkCreate_SetLink verifies that SetLink sets the link arg and returns *SymlinkCreate.
func TestSymlinkCreate_SetLink(t *testing.T) {
	skill := NewSymlinkCreate()
	skill.SetLink("/usr/local/bin/pm2")

	if skill.GetArg(ArgLink) != "/usr/local/bin/pm2" {
		t.Errorf("Expected link '/usr/local/bin/pm2', got '%s'", skill.GetArg(ArgLink))
	}
}

// TestSymlinkCreate_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestSymlinkCreate_TypedSetters_Chaining(t *testing.T) {
	skill := NewSymlinkCreate().
		SetTarget("/opt/node/bin/pm2").
		SetLink("/usr/local/bin/pm2")

	if skill.GetArg(ArgTarget) != "/opt/node/bin/pm2" {
		t.Errorf("Expected target '/opt/node/bin/pm2', got '%s'", skill.GetArg(ArgTarget))
	}
	if skill.GetArg(ArgLink) != "/usr/local/bin/pm2" {
		t.Errorf("Expected link '/usr/local/bin/pm2', got '%s'", skill.GetArg(ArgLink))
	}
}
