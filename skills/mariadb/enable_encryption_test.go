package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestEnableEncryption_Run_DryRun verifies that dry-run mode correctly handles encryption enablement.
func TestEnableEncryption_Run_DryRun(t *testing.T) {
	pb := NewEnableEncryption()

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

	expectedMessage := "Would enable MariaDB encryption at rest"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestEnableEncryption_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestEnableEncryption_Run_NotDryRun(t *testing.T) {
	pb := NewEnableEncryption()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would enable MariaDB encryption at rest" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestEnableEncryption_NewEnableEncryption verifies that NewEnableEncryption creates a properly configured skill.
func TestEnableEncryption_NewEnableEncryption(t *testing.T) {
	pb := NewEnableEncryption()

	if pb.GetID() != "mariadb-enable-encryption" {
		t.Errorf("Expected ID to be 'mariadb-enable-encryption', got '%s'", pb.GetID())
	}

	expectedDescription := "Enable data-at-rest encryption for MariaDB"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestEnableEncryption_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete EnableEncryption type.
func TestEnableEncryption_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableEncryption()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*EnableEncryption); !ok {
		t.Error("SetArgs should return *EnableEncryption, not just RunnableInterface")
	}
}

// TestEnableEncryption_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete EnableEncryption type.
func TestEnableEncryption_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableEncryption()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*EnableEncryption); !ok {
		t.Error("SetArg should return *EnableEncryption, not just RunnableInterface")
	}
}

// TestEnableEncryption_SetID_ReturnsConcreteType verifies that SetID returns the concrete EnableEncryption type.
func TestEnableEncryption_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableEncryption()

	result := skill.SetID("custom-id")

	if _, ok := result.(*EnableEncryption); !ok {
		t.Error("SetID should return *EnableEncryption, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestEnableEncryption_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete EnableEncryption type.
func TestEnableEncryption_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableEncryption()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*EnableEncryption); !ok {
		t.Error("SetDescription should return *EnableEncryption, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestEnableEncryption_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete EnableEncryption type.
func TestEnableEncryption_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableEncryption()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*EnableEncryption); !ok {
		t.Error("SetTimeout should return *EnableEncryption, not just RunnableInterface")
	}
}

// TestEnableEncryption_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestEnableEncryption_MethodChaining_PreservesType(t *testing.T) {
	skill := NewEnableEncryption().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*EnableEncryption); !ok {
		t.Error("Method chaining should preserve *EnableEncryption type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
