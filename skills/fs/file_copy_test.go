package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestFileCopy_Run_DryRun verifies that dry-run mode reports the would-copy message.
func TestFileCopy_Run_DryRun(t *testing.T) {
	pb := NewFileCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/etc/ssh/sshd_config")
	pb.SetArg(ArgDst, "/etc/ssh/sshd_config.bak")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would copy file: /etc/ssh/sshd_config -> /etc/ssh/sshd_config.bak"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFileCopy_Run_NoSrc verifies that missing ArgSrc returns an error.
func TestFileCopy_Run_NoSrc(t *testing.T) {
	pb := NewFileCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgDst, "/etc/ssh/sshd_config.bak")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no src specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no src is specified")
	}
}

// TestFileCopy_Run_NoDst verifies that missing ArgDst returns an error.
func TestFileCopy_Run_NoDst(t *testing.T) {
	pb := NewFileCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/etc/ssh/sshd_config")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no dst specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no dst is specified")
	}
}

// TestFileCopy_Run_RelativeSrc verifies that a relative src returns an error.
func TestFileCopy_Run_RelativeSrc(t *testing.T) {
	pb := NewFileCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "relative/path")
	pb.SetArg(ArgDst, "/etc/ssh/sshd_config.bak")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative src")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative src")
	}
}

// TestFileCopy_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestFileCopy_Run_NotDryRun(t *testing.T) {
	pb := NewFileCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/etc/ssh/sshd_config")
	pb.SetArg(ArgDst, "/etc/ssh/sshd_config.bak")

	result := pb.Run()

	if result.Message == "Would copy file: /etc/ssh/sshd_config -> /etc/ssh/sshd_config.bak" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestFileCopy_NewFileCopy verifies that NewFileCopy creates a properly configured skill.
func TestFileCopy_NewFileCopy(t *testing.T) {
	pb := NewFileCopy()

	if pb.GetID() != "fs-file-copy" {
		t.Errorf("Expected ID to be 'fs-file-copy', got '%s'", pb.GetID())
	}

	expectedDescription := "Copy file on remote server (cp)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFileCopy_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete FileCopy type.
func TestFileCopy_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCopy()
	args := map[string]string{ArgSrc: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*FileCopy); !ok {
		t.Error("SetArgs should return *FileCopy, not just RunnableInterface")
	}
}

// TestFileCopy_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete FileCopy type.
func TestFileCopy_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCopy()

	result := skill.SetArg(ArgSrc, "/tmp/test")

	if _, ok := result.(*FileCopy); !ok {
		t.Error("SetArg should return *FileCopy, not just RunnableInterface")
	}
}

// TestFileCopy_SetID_ReturnsConcreteType verifies that SetID returns the concrete FileCopy type.
func TestFileCopy_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCopy()

	result := skill.SetID("custom-id")

	if _, ok := result.(*FileCopy); !ok {
		t.Error("SetID should return *FileCopy, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFileCopy_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete FileCopy type.
func TestFileCopy_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCopy()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*FileCopy); !ok {
		t.Error("SetDescription should return *FileCopy, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFileCopy_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete FileCopy type.
func TestFileCopy_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFileCopy()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*FileCopy); !ok {
		t.Error("SetTimeout should return *FileCopy, not just RunnableInterface")
	}
}

// TestFileCopy_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFileCopy_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFileCopy().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgSrc, "/tmp/test").
		SetArgs(map[string]string{ArgSrc: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*FileCopy); !ok {
		t.Error("Method chaining should preserve *FileCopy type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestFileCopy_SetSrc verifies that SetSrc sets the src arg and returns *FileCopy.
func TestFileCopy_SetSrc(t *testing.T) {
	skill := NewFileCopy().SetSrc("/etc/ssh/sshd_config")

	if skill.GetArg(ArgSrc) != "/etc/ssh/sshd_config" {
		t.Errorf("Expected src '/etc/ssh/sshd_config', got '%s'", skill.GetArg(ArgSrc))
	}
}

// TestFileCopy_SetDst verifies that SetDst sets the dst arg and returns *FileCopy.
func TestFileCopy_SetDst(t *testing.T) {
	skill := NewFileCopy().SetDst("/etc/ssh/sshd_config.bak")

	if skill.GetArg(ArgDst) != "/etc/ssh/sshd_config.bak" {
		t.Errorf("Expected dst '/etc/ssh/sshd_config.bak', got '%s'", skill.GetArg(ArgDst))
	}
}

// TestFileCopy_SetForce verifies that SetForce sets the force arg and returns *FileCopy.
func TestFileCopy_SetForce(t *testing.T) {
	skill := NewFileCopy().SetForce(true)

	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}

	skill.SetForce(false)
	if skill.GetArg(ArgForce) != "false" {
		t.Errorf("Expected force 'false', got '%s'", skill.GetArg(ArgForce))
	}
}

// TestFileCopy_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestFileCopy_TypedSetters_Chaining(t *testing.T) {
	skill := NewFileCopy().
		SetSrc("/etc/ssh/sshd_config").
		SetDst("/etc/ssh/sshd_config.bak").
		SetForce(true)

	if skill.GetArg(ArgSrc) != "/etc/ssh/sshd_config" {
		t.Errorf("Expected src '/etc/ssh/sshd_config', got '%s'", skill.GetArg(ArgSrc))
	}
	if skill.GetArg(ArgDst) != "/etc/ssh/sshd_config.bak" {
		t.Errorf("Expected dst '/etc/ssh/sshd_config.bak', got '%s'", skill.GetArg(ArgDst))
	}
	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}
}
