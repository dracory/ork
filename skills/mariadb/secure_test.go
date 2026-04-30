package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSecure_Run_DryRun verifies that dry-run mode correctly handles MariaDB security hardening.
func TestSecure_Run_DryRun(t *testing.T) {
	pb := NewSecure()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would secure MariaDB installation"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSecure_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestSecure_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewSecure()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing root-password even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestSecure_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSecure_Run_NotDryRun(t *testing.T) {
	pb := NewSecure()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would secure MariaDB installation" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSecure_NewSecure verifies that NewSecure creates a properly configured skill.
func TestSecure_NewSecure(t *testing.T) {
	pb := NewSecure()

	if pb.GetID() != "mariadb-secure" {
		t.Errorf("Expected ID to be 'mariadb-secure', got '%s'", pb.GetID())
	}

	expectedDescription := "Perform basic security hardening on MariaDB installation"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestSecure_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Secure type.
func TestSecure_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSecure()
	args := map[string]string{"root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Secure); !ok {
		t.Error("SetArgs should return *Secure, not just RunnableInterface")
	}
}

// TestSecure_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Secure type.
func TestSecure_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSecure()

	result := skill.SetArg("root-password", "testpass")

	if _, ok := result.(*Secure); !ok {
		t.Error("SetArg should return *Secure, not just RunnableInterface")
	}
}

// TestSecure_SetID_ReturnsConcreteType verifies that SetID returns the concrete Secure type.
func TestSecure_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSecure()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Secure); !ok {
		t.Error("SetID should return *Secure, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSecure_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Secure type.
func TestSecure_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSecure()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Secure); !ok {
		t.Error("SetDescription should return *Secure, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSecure_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Secure type.
func TestSecure_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSecure()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Secure); !ok {
		t.Error("SetTimeout should return *Secure, not just RunnableInterface")
	}
}

// TestSecure_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSecure_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSecure().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("root-password", "testpass").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Secure); !ok {
		t.Error("Method chaining should preserve *Secure type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
