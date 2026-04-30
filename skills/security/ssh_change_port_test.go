package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSshChangePort_Run_DryRun verifies that dry-run mode correctly handles SSH port change.
func TestSshChangePort_Run_DryRun(t *testing.T) {
	pb := NewSshChangePort()
	pb.SetArg("port", "2222")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would change SSH port to 2222"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSshChangePort_Run_DryRun_NoPort verifies dry-run without port returns error.
func TestSshChangePort_Run_DryRun_NoPort(t *testing.T) {
	pb := NewSshChangePort()

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

// TestSshChangePort_Run_DryRun_InvalidPort verifies dry-run with invalid port returns error.
func TestSshChangePort_Run_DryRun_InvalidPort(t *testing.T) {
	pb := NewSshChangePort()
	pb.SetArg("port", "22")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for invalid port (< 1024) even in dry-run
	if result.Error == nil {
		t.Error("Expected error for invalid port")
	}

	if result.Message != "Invalid port number" {
		t.Errorf("Expected message 'Invalid port number', got '%s'", result.Message)
	}
}

// TestSshChangePort_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSshChangePort_Run_NotDryRun(t *testing.T) {
	pb := NewSshChangePort()
	pb.SetArg("port", "2222")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would change SSH port to 2222" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSshChangePort_NewSshChangePort verifies that NewSshChangePort creates a properly configured skill.
func TestSshChangePort_NewSshChangePort(t *testing.T) {
	pb := NewSshChangePort()

	if pb.GetID() != "ssh-change-port" {
		t.Errorf("Expected ID to be 'ssh-change-port', got '%s'", pb.GetID())
	}

	expectedDescription := "Change the SSH port to reduce automated scanning"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestSshChangePort_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete SshChangePort type.
func TestSshChangePort_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSshChangePort()
	args := map[string]string{"port": "2222"}

	result := skill.SetArgs(args)

	if _, ok := result.(*SshChangePort); !ok {
		t.Error("SetArgs should return *SshChangePort, not just RunnableInterface")
	}
}

// TestSshChangePort_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete SshChangePort type.
func TestSshChangePort_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSshChangePort()

	result := skill.SetArg("port", "2222")

	if _, ok := result.(*SshChangePort); !ok {
		t.Error("SetArg should return *SshChangePort, not just RunnableInterface")
	}
}

// TestSshChangePort_SetID_ReturnsConcreteType verifies that SetID returns the concrete SshChangePort type.
func TestSshChangePort_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSshChangePort()

	result := skill.SetID("custom-id")

	if _, ok := result.(*SshChangePort); !ok {
		t.Error("SetID should return *SshChangePort, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSshChangePort_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete SshChangePort type.
func TestSshChangePort_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSshChangePort()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*SshChangePort); !ok {
		t.Error("SetDescription should return *SshChangePort, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSshChangePort_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete SshChangePort type.
func TestSshChangePort_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSshChangePort()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*SshChangePort); !ok {
		t.Error("SetTimeout should return *SshChangePort, not just RunnableInterface")
	}
}

// TestSshChangePort_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSshChangePort_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSshChangePort().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("port", "2222").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*SshChangePort); !ok {
		t.Error("Method chaining should preserve *SshChangePort type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
