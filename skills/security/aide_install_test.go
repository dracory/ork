package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestAideInstall_Run_DryRun verifies that dry-run mode correctly handles AIDE installation.
func TestAideInstall_Run_DryRun(t *testing.T) {
	pb := NewAideInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure AIDE"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAideInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAideInstall_Run_NotDryRun(t *testing.T) {
	pb := NewAideInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure AIDE" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAideInstall_NewAideInstall verifies that NewAideInstall creates a properly configured skill.
func TestAideInstall_NewAideInstall(t *testing.T) {
	pb := NewAideInstall()

	if pb.GetID() != "aide-install" {
		t.Errorf("Expected ID to be 'aide-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure AIDE file integrity monitoring"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestAideInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete AideInstall type.
func TestAideInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewAideInstall()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*AideInstall); !ok {
		t.Error("SetArgs should return *AideInstall, not just RunnableInterface")
	}
}

// TestAideInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete AideInstall type.
func TestAideInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewAideInstall()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*AideInstall); !ok {
		t.Error("SetArg should return *AideInstall, not just RunnableInterface")
	}
}

// TestAideInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete AideInstall type.
func TestAideInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewAideInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*AideInstall); !ok {
		t.Error("SetID should return *AideInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestAideInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete AideInstall type.
func TestAideInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewAideInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*AideInstall); !ok {
		t.Error("SetDescription should return *AideInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestAideInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete AideInstall type.
func TestAideInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewAideInstall()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*AideInstall); !ok {
		t.Error("SetTimeout should return *AideInstall, not just RunnableInterface")
	}
}

// TestAideInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestAideInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewAideInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*AideInstall); !ok {
		t.Error("Method chaining should preserve *AideInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
