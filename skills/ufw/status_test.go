package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
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
