package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestKernelHarden_Run_DryRun verifies that dry-run mode correctly handles kernel hardening.
func TestKernelHarden_Run_DryRun(t *testing.T) {
	pb := NewKernelHarden()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would harden kernel security parameters"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestKernelHarden_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestKernelHarden_Run_NotDryRun(t *testing.T) {
	pb := NewKernelHarden()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would harden kernel security parameters" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestKernelHarden_NewKernelHarden verifies that NewKernelHarden creates a properly configured playbook.
func TestKernelHarden_NewKernelHarden(t *testing.T) {
	pb := NewKernelHarden()

	if pb.GetID() != "kernel-harden" {
		t.Errorf("Expected ID to be 'kernel-harden', got '%s'", pb.GetID())
	}

	expectedDescription := "Apply security-focused kernel parameters via sysctl"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
