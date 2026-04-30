package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUfwStatus_Run_DryRun verifies that dry-run mode correctly handles status check.
func TestUfwStatus_Run_DryRun(t *testing.T) {
	pb := NewUfwStatus()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check UFW firewall status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUfwStatus_Run_NotDryRun(t *testing.T) {
	pb := NewUfwStatus()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check UFW firewall status" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestUfwStatus_Check verifies that Check always returns false since this is a read-only skill.
func TestUfwStatus_Check(t *testing.T) {
	pb := NewUfwStatus()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false for read-only playbook")
	}
}

// TestUfwStatus_NewUfwStatus verifies that NewUfwStatus creates a properly configured skill.
func TestUfwStatus_NewUfwStatus(t *testing.T) {
	pb := NewUfwStatus()

	if pb.GetID() != "ufw-status" {
		t.Errorf("Expected ID to be 'ufw-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Display UFW firewall configuration and status (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestUfwStatus_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete UfwStatus type.
func TestUfwStatus_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwStatus()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*UfwStatus); !ok {
		t.Error("SetArgs should return *UfwStatus, not just RunnableInterface")
	}
}

// TestUfwStatus_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete UfwStatus type.
func TestUfwStatus_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwStatus()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*UfwStatus); !ok {
		t.Error("SetArg should return *UfwStatus, not just RunnableInterface")
	}
}

// TestUfwStatus_SetID_ReturnsConcreteType verifies that SetID returns the concrete UfwStatus type.
func TestUfwStatus_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwStatus()

	result := skill.SetID("custom-id")

	if _, ok := result.(*UfwStatus); !ok {
		t.Error("SetID should return *UfwStatus, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestUfwStatus_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete UfwStatus type.
func TestUfwStatus_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwStatus()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*UfwStatus); !ok {
		t.Error("SetDescription should return *UfwStatus, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestUfwStatus_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete UfwStatus type.
func TestUfwStatus_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwStatus()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*UfwStatus); !ok {
		t.Error("SetTimeout should return *UfwStatus, not just RunnableInterface")
	}
}

// TestUfwStatus_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestUfwStatus_MethodChaining_PreservesType(t *testing.T) {
	skill := NewUfwStatus().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*UfwStatus); !ok {
		t.Error("Method chaining should preserve *UfwStatus type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
