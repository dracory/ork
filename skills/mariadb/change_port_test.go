package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestChangePort_Run_DryRun verifies that dry-run mode correctly handles MariaDB port change.
func TestChangePort_Run_DryRun(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3307")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would change MariaDB port to 3307"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestChangePort_Run_DryRun_NoPort verifies dry-run without port returns error.
func TestChangePort_Run_DryRun_NoPort(t *testing.T) {
	pb := NewChangePort()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing port even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing port")
	}

	if result.Message != "Port parameter is required" {
		t.Errorf("Expected message 'Port parameter is required', got '%s'", result.Message)
	}
}

// TestChangePort_Run_DryRun_InvalidPort verifies dry-run with invalid port returns error.
func TestChangePort_Run_DryRun_InvalidPort(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3306")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for invalid port (3306 is default) even in dry-run
	if result.Error == nil {
		t.Error("Expected error for invalid port")
	}

	if result.Message != "Invalid port number" {
		t.Errorf("Expected message 'Invalid port number', got '%s'", result.Message)
	}
}

// TestChangePort_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestChangePort_Run_NotDryRun(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3307")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would change MariaDB port to 3307" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestChangePort_NewChangePort verifies that NewChangePort creates a properly configured skill.
func TestChangePort_NewChangePort(t *testing.T) {
	pb := NewChangePort()

	if pb.GetID() != "mariadb-change-port" {
		t.Errorf("Expected ID to be 'mariadb-change-port', got '%s'", pb.GetID())
	}

	expectedDescription := "Change the MariaDB server port from default 3306"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestChangePort_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete ChangePort type.
func TestChangePort_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewChangePort()
	args := map[string]string{"port": "3307"}

	result := skill.SetArgs(args)

	if _, ok := result.(*ChangePort); !ok {
		t.Error("SetArgs should return *ChangePort, not just RunnableInterface")
	}
}

// TestChangePort_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete ChangePort type.
func TestChangePort_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewChangePort()

	result := skill.SetArg("port", "3307")

	if _, ok := result.(*ChangePort); !ok {
		t.Error("SetArg should return *ChangePort, not just RunnableInterface")
	}
}

// TestChangePort_SetID_ReturnsConcreteType verifies that SetID returns the concrete ChangePort type.
func TestChangePort_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewChangePort()

	result := skill.SetID("custom-id")

	if _, ok := result.(*ChangePort); !ok {
		t.Error("SetID should return *ChangePort, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestChangePort_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete ChangePort type.
func TestChangePort_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewChangePort()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*ChangePort); !ok {
		t.Error("SetDescription should return *ChangePort, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestChangePort_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete ChangePort type.
func TestChangePort_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewChangePort()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*ChangePort); !ok {
		t.Error("SetTimeout should return *ChangePort, not just RunnableInterface")
	}
}

// TestChangePort_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestChangePort_MethodChaining_PreservesType(t *testing.T) {
	skill := NewChangePort().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("port", "3307").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*ChangePort); !ok {
		t.Error("Method chaining should preserve *ChangePort type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
