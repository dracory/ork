package reboot

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestReboot_Run_DryRun verifies that dry-run mode correctly handles reboot.
func TestReboot_Run_DryRun(t *testing.T) {
	pb := NewReboot().(*Reboot)

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would reboot test.example.com"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestReboot_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestReboot_Run_NotDryRun(t *testing.T) {
	pb := NewReboot().(*Reboot)

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would reboot test.example.com" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Reboot should always report Changed=true
	if !result.Changed {
		t.Error("Expected Changed to be true for reboot operation")
	}
}

// TestReboot_Check verifies that Check always returns true.
func TestReboot_Check(t *testing.T) {
	pb := NewReboot().(*Reboot)

	cfg := types.NodeConfig{
		Logger:  slog.Default(),
		SSHHost: "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true for reboot operation")
	}
}

// TestReboot_NewReboot verifies that NewReboot creates a properly configured skill.
func TestReboot_NewReboot(t *testing.T) {
	pb := NewReboot().(*Reboot)

	if pb.GetID() != "reboot" {
		t.Errorf("Expected ID to be 'reboot', got '%s'", pb.GetID())
	}

	expectedDescription := "Reboot the remote server"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}

	// Verify default values
	if pb.WaitForReconnect {
		t.Error("Expected WaitForReconnect to be false by default")
	}

	if pb.MaxWaitTime != 5*time.Minute {
		t.Errorf("Expected MaxWaitTime to be 5 minutes, got %v", pb.MaxWaitTime)
	}
}

// TestReboot_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Reboot type.
func TestReboot_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetArgs should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Reboot type.
func TestReboot_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetArg should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_SetID_ReturnsConcreteType verifies that SetID returns the concrete Reboot type.
func TestReboot_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetID should return *Reboot, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestReboot_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Reboot type.
func TestReboot_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetDescription should return *Reboot, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestReboot_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Reboot type.
func TestReboot_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetTimeout should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestReboot_MethodChaining_PreservesType(t *testing.T) {
	skill := NewReboot().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Reboot); !ok {
		t.Error("Method chaining should preserve *Reboot type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
