package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestRename_Run_DryRun verifies that dry-run mode reports the would-rename message.
func TestRename_Run_DryRun(t *testing.T) {
	pb := NewRename()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/tmp/config.tmp")
	pb.SetArg(ArgDst, "/etc/myapp/config")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would rename: /tmp/config.tmp -> /etc/myapp/config"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestRename_Run_NoSrc verifies that missing ArgSrc returns an error.
func TestRename_Run_NoSrc(t *testing.T) {
	pb := NewRename()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgDst, "/etc/myapp/config")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no src specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no src is specified")
	}
}

// TestRename_Run_NoDst verifies that missing ArgDst returns an error.
func TestRename_Run_NoDst(t *testing.T) {
	pb := NewRename()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/tmp/config.tmp")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no dst specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no dst is specified")
	}
}

// TestRename_Run_RelativeSrc verifies that a relative src returns an error.
func TestRename_Run_RelativeSrc(t *testing.T) {
	pb := NewRename()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "relative/path")
	pb.SetArg(ArgDst, "/etc/myapp/config")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative src")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative src")
	}
}

// TestRename_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestRename_Run_NotDryRun(t *testing.T) {
	pb := NewRename()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/tmp/config.tmp")
	pb.SetArg(ArgDst, "/etc/myapp/config")

	result := pb.Run()

	if result.Message == "Would rename: /tmp/config.tmp -> /etc/myapp/config" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestRename_NewRename verifies that NewRename creates a properly configured skill.
func TestRename_NewRename(t *testing.T) {
	pb := NewRename()

	if pb.GetID() != "fs-rename" {
		t.Errorf("Expected ID to be 'fs-rename', got '%s'", pb.GetID())
	}

	expectedDescription := "Rename/move file or directory (mv)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestRename_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Rename type.
func TestRename_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewRename()
	args := map[string]string{ArgSrc: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Rename); !ok {
		t.Error("SetArgs should return *Rename, not just RunnableInterface")
	}
}

// TestRename_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Rename type.
func TestRename_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewRename()

	result := skill.SetArg(ArgSrc, "/tmp/test")

	if _, ok := result.(*Rename); !ok {
		t.Error("SetArg should return *Rename, not just RunnableInterface")
	}
}

// TestRename_SetID_ReturnsConcreteType verifies that SetID returns the concrete Rename type.
func TestRename_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewRename()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Rename); !ok {
		t.Error("SetID should return *Rename, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestRename_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Rename type.
func TestRename_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewRename()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Rename); !ok {
		t.Error("SetDescription should return *Rename, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestRename_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Rename type.
func TestRename_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewRename()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*Rename); !ok {
		t.Error("SetTimeout should return *Rename, not just RunnableInterface")
	}
}

// TestRename_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestRename_MethodChaining_PreservesType(t *testing.T) {
	skill := NewRename().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgSrc, "/tmp/test").
		SetArgs(map[string]string{ArgSrc: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Rename); !ok {
		t.Error("Method chaining should preserve *Rename type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
