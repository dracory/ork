package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestEnableSSL_Run_DryRun verifies that dry-run mode correctly handles SSL enablement.
func TestEnableSSL_Run_DryRun(t *testing.T) {
	pb := NewEnableSSL()

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

	expectedMessage := "Would enable MariaDB SSL/TLS"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestEnableSSL_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestEnableSSL_Run_NotDryRun(t *testing.T) {
	pb := NewEnableSSL()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would enable MariaDB SSL/TLS" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestEnableSSL_NewEnableSSL verifies that NewEnableSSL creates a properly configured skill.
func TestEnableSSL_NewEnableSSL(t *testing.T) {
	pb := NewEnableSSL()

	if pb.GetID() != "mariadb-enable-ssl" {
		t.Errorf("Expected ID to be 'mariadb-enable-ssl', got '%s'", pb.GetID())
	}

	expectedDescription := "Enable SSL/TLS encryption for MariaDB connections"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestEnableSSL_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete EnableSSL type.
func TestEnableSSL_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableSSL()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*EnableSSL); !ok {
		t.Error("SetArgs should return *EnableSSL, not just RunnableInterface")
	}
}

// TestEnableSSL_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete EnableSSL type.
func TestEnableSSL_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableSSL()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*EnableSSL); !ok {
		t.Error("SetArg should return *EnableSSL, not just RunnableInterface")
	}
}

// TestEnableSSL_SetID_ReturnsConcreteType verifies that SetID returns the concrete EnableSSL type.
func TestEnableSSL_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableSSL()

	result := skill.SetID("custom-id")

	if _, ok := result.(*EnableSSL); !ok {
		t.Error("SetID should return *EnableSSL, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestEnableSSL_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete EnableSSL type.
func TestEnableSSL_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableSSL()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*EnableSSL); !ok {
		t.Error("SetDescription should return *EnableSSL, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestEnableSSL_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete EnableSSL type.
func TestEnableSSL_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewEnableSSL()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*EnableSSL); !ok {
		t.Error("SetTimeout should return *EnableSSL, not just RunnableInterface")
	}
}

// TestEnableSSL_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestEnableSSL_MethodChaining_PreservesType(t *testing.T) {
	skill := NewEnableSSL().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*EnableSSL); !ok {
		t.Error("Method chaining should preserve *EnableSSL type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
