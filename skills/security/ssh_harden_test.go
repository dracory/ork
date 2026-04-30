package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSshHarden_Run_DryRun verifies that dry-run mode correctly handles SSH hardening.
func TestSshHarden_Run_DryRun(t *testing.T) {
	pb := NewSshHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would harden SSH security configuration"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSshHarden_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSshHarden_Run_NotDryRun(t *testing.T) {
	pb := NewSshHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would harden SSH security configuration" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSshHarden_NewSshHarden verifies that NewSshHarden creates a properly configured skill.
func TestSshHarden_NewSshHarden(t *testing.T) {
	pb := NewSshHarden()

	if pb.GetID() != "ssh-harden" {
		t.Errorf("Expected ID to be 'ssh-harden', got '%s'", pb.GetID())
	}

	expectedDescription := "Apply security hardening to SSH server configuration"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestSshHarden_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete SshHarden type.
func TestSshHarden_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSshHarden()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*SshHarden); !ok {
		t.Error("SetArgs should return *SshHarden, not just RunnableInterface")
	}
}

// TestSshHarden_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete SshHarden type.
func TestSshHarden_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSshHarden()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*SshHarden); !ok {
		t.Error("SetArg should return *SshHarden, not just RunnableInterface")
	}
}

// TestSshHarden_SetID_ReturnsConcreteType verifies that SetID returns the concrete SshHarden type.
func TestSshHarden_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSshHarden()

	result := skill.SetID("custom-id")

	if _, ok := result.(*SshHarden); !ok {
		t.Error("SetID should return *SshHarden, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSshHarden_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete SshHarden type.
func TestSshHarden_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSshHarden()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*SshHarden); !ok {
		t.Error("SetDescription should return *SshHarden, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSshHarden_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete SshHarden type.
func TestSshHarden_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSshHarden()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*SshHarden); !ok {
		t.Error("SetTimeout should return *SshHarden, not just RunnableInterface")
	}
}

// TestSshHarden_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSshHarden_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSshHarden().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*SshHarden); !ok {
		t.Error("Method chaining should preserve *SshHarden type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
