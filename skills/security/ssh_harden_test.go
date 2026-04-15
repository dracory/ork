package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSshHarden_Run_DryRun verifies that dry-run mode correctly handles SSH hardening.
func TestSshHarden_Run_DryRun(t *testing.T) {
	pb := NewSshHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would harden SSH security configuration"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSshHarden_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSshHarden_Run_NotDryRun(t *testing.T) {
	pb := NewSshHarden()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would harden SSH security configuration" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSshHarden_NewSshHarden verifies that NewSshHarden creates a properly configured skill.
func TestSshHarden_NewSshHarden(t *testing.T) {
	pb := NewSshHarden()

	if pb.GetID() != "ssh-harden" {
		t.Errorf("Expected ID to be 'ssh-harden', got '%s'", pb.GetID())
	}

	expectedDescription := "Apply security hardening to SSH server configuration"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
