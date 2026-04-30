package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestKernelHarden_Run_DryRun verifies that dry-run mode correctly handles kernel hardening.
func TestKernelHarden_Run_DryRun(t *testing.T) {
	pb := NewKernelHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would harden kernel security parameters"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestKernelHarden_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestKernelHarden_Run_NotDryRun(t *testing.T) {
	pb := NewKernelHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would harden kernel security parameters" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestKernelHarden_NewKernelHarden verifies that NewKernelHarden creates a properly configured skill.
func TestKernelHarden_NewKernelHarden(t *testing.T) {
	pb := NewKernelHarden()

	if pb.GetID() != "kernel-harden" {
		t.Errorf("Expected ID to be 'kernel-harden', got '%s'", pb.GetID())
	}

	expectedDescription := "Apply security-focused kernel parameters via sysctl"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestKernelHarden_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete KernelHarden type.
func TestKernelHarden_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewKernelHarden()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*KernelHarden); !ok {
		t.Error("SetArgs should return *KernelHarden, not just RunnableInterface")
	}
}

// TestKernelHarden_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete KernelHarden type.
func TestKernelHarden_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewKernelHarden()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*KernelHarden); !ok {
		t.Error("SetArg should return *KernelHarden, not just RunnableInterface")
	}
}

// TestKernelHarden_SetID_ReturnsConcreteType verifies that SetID returns the concrete KernelHarden type.
func TestKernelHarden_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewKernelHarden()

	result := skill.SetID("custom-id")

	if _, ok := result.(*KernelHarden); !ok {
		t.Error("SetID should return *KernelHarden, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestKernelHarden_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete KernelHarden type.
func TestKernelHarden_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewKernelHarden()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*KernelHarden); !ok {
		t.Error("SetDescription should return *KernelHarden, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestKernelHarden_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete KernelHarden type.
func TestKernelHarden_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewKernelHarden()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*KernelHarden); !ok {
		t.Error("SetTimeout should return *KernelHarden, not just RunnableInterface")
	}
}

// TestKernelHarden_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestKernelHarden_MethodChaining_PreservesType(t *testing.T) {
	skill := NewKernelHarden().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*KernelHarden); !ok {
		t.Error("Method chaining should preserve *KernelHarden type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
