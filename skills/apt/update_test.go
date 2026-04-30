package apt

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestAptUpdate_Run_DryRun verifies that dry-run mode correctly handles apt update.
func TestAptUpdate_Run_DryRun(t *testing.T) {
	pb := NewAptUpdate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would update package database: apt-get update -y"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAptUpdate_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAptUpdate_Run_NotDryRun(t *testing.T) {
	pb := NewAptUpdate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would update package database: apt-get update -y" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAptUpdate_Check verifies that Check always returns true.
func TestAptUpdate_Check(t *testing.T) {
	pb := NewAptUpdate()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true for apt update")
	}
}

// TestAptUpdate_NewAptUpdate verifies that NewAptUpdate creates a properly configured skill.
func TestAptUpdate_NewAptUpdate(t *testing.T) {
	pb := NewAptUpdate()

	if pb.GetID() != "apt-update" {
		t.Errorf("Expected ID to be 'apt-update', got '%s'", pb.GetID())
	}

	expectedDescription := "Refresh package database (apt-get update)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestAptUpdate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete AptUpdate type.
func TestAptUpdate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewAptUpdate()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*AptUpdate); !ok {
		t.Error("SetArgs should return *AptUpdate, not just RunnableInterface")
	}
}

// TestAptUpdate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete AptUpdate type.
func TestAptUpdate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewAptUpdate()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*AptUpdate); !ok {
		t.Error("SetArg should return *AptUpdate, not just RunnableInterface")
	}
}

// TestAptUpdate_SetID_ReturnsConcreteType verifies that SetID returns the concrete AptUpdate type.
func TestAptUpdate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewAptUpdate()

	result := skill.SetID("custom-id")

	if _, ok := result.(*AptUpdate); !ok {
		t.Error("SetID should return *AptUpdate, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestAptUpdate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete AptUpdate type.
func TestAptUpdate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewAptUpdate()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*AptUpdate); !ok {
		t.Error("SetDescription should return *AptUpdate, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestAptUpdate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete AptUpdate type.
func TestAptUpdate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewAptUpdate()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*AptUpdate); !ok {
		t.Error("SetTimeout should return *AptUpdate, not just RunnableInterface")
	}
}

// TestAptUpdate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestAptUpdate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewAptUpdate().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*AptUpdate); !ok {
		t.Error("Method chaining should preserve *AptUpdate type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
