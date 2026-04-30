package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestStatus_Run_DryRun verifies that dry-run mode correctly handles MariaDB status.
func TestStatus_Run_DryRun(t *testing.T) {
	pb := NewStatus()

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

	expectedMessage := "Would check MariaDB status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestStatus_Run_DryRun_WithPassword verifies dry-run with root password.
func TestStatus_Run_DryRun_WithPassword(t *testing.T) {
	pb := NewStatus()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check MariaDB status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestStatus_Run_NotDryRun(t *testing.T) {
	pb := NewStatus()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check MariaDB status" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestStatus_Check verifies that Check returns false for read-only operation.
func TestStatus_Check(t *testing.T) {
	pb := NewStatus()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestStatus_NewStatus verifies that NewStatus creates a properly configured skill.
func TestStatus_NewStatus(t *testing.T) {
	pb := NewStatus()

	if pb.GetID() != "mariadb-status" {
		t.Errorf("Expected ID to be 'mariadb-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Display MariaDB server status information (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestStatus_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Status type.
func TestStatus_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	args := map[string]string{"root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Status); !ok {
		t.Error("SetArgs should return *Status, not just RunnableInterface")
	}
}

// TestStatus_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Status type.
func TestStatus_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()

	result := skill.SetArg("root-password", "testpass")

	if _, ok := result.(*Status); !ok {
		t.Error("SetArg should return *Status, not just RunnableInterface")
	}
}

// TestStatus_SetID_ReturnsConcreteType verifies that SetID returns the concrete Status type.
func TestStatus_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Status); !ok {
		t.Error("SetID should return *Status, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestStatus_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Status type.
func TestStatus_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Status); !ok {
		t.Error("SetDescription should return *Status, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestStatus_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Status type.
func TestStatus_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Status); !ok {
		t.Error("SetTimeout should return *Status, not just RunnableInterface")
	}
}

// TestStatus_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestStatus_MethodChaining_PreservesType(t *testing.T) {
	skill := NewStatus().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("root-password", "testpass").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Status); !ok {
		t.Error("Method chaining should preserve *Status type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
