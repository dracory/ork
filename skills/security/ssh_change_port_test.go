package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestSshChangePort_Run_DryRun verifies that dry-run mode correctly handles SSH port change.
func TestSshChangePort_Run_DryRun(t *testing.T) {
	pb := NewSshChangePort()
	pb.SetArg("port", "2222")

	cfg := config.NodeConfig{
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

	cfg := config.NodeConfig{
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

	cfg := config.NodeConfig{
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

	cfg := config.NodeConfig{
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
