package apt

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestAptUpdate_Run_DryRun verifies that dry-run mode correctly handles apt update.
func TestAptUpdate_Run_DryRun(t *testing.T) {
	pb := NewAptUpdate()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would update package database: apt-get update -y"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAptUpdate_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAptUpdate_Run_NotDryRun(t *testing.T) {
	pb := NewAptUpdate()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would update package database: apt-get update -y" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAptUpdate_Check verifies that Check always returns true.
func TestAptUpdate_Check(t *testing.T) {
	pb := NewAptUpdate()

	cfg := config.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true for apt update")
	}
}

// TestAptUpdate_NewAptUpdate verifies that NewAptUpdate creates a properly configured skill.
func TestAptUpdate_NewAptUpdate(t *testing.T) {
	pb := NewAptUpdate()

	if pb.GetID() != "apt-update" {
		t.Errorf("Expected ID to be 'apt-update', got '%s'", pb.GetID())
	}

	expectedDescription := "Refresh package database (apt-get update)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
