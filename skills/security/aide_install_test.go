package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestAideInstall_Run_DryRun verifies that dry-run mode correctly handles AIDE installation.
func TestAideInstall_Run_DryRun(t *testing.T) {
	pb := NewAideInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure AIDE"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAideInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAideInstall_Run_NotDryRun(t *testing.T) {
	pb := NewAideInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure AIDE" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAideInstall_NewAideInstall verifies that NewAideInstall creates a properly configured skill.
func TestAideInstall_NewAideInstall(t *testing.T) {
	pb := NewAideInstall()

	if pb.GetID() != "aide-install" {
		t.Errorf("Expected ID to be 'aide-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure AIDE file integrity monitoring"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
