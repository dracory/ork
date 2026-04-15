package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestAuditdInstall_Run_DryRun verifies that dry-run mode correctly handles auditd installation.
func TestAuditdInstall_Run_DryRun(t *testing.T) {
	pb := NewAuditdInstall()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure auditd"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAuditdInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAuditdInstall_Run_NotDryRun(t *testing.T) {
	pb := NewAuditdInstall()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure auditd" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAuditdInstall_NewAuditdInstall verifies that NewAuditdInstall creates a properly configured skill.
func TestAuditdInstall_NewAuditdInstall(t *testing.T) {
	pb := NewAuditdInstall()

	if pb.GetID() != "auditd-install" {
		t.Errorf("Expected ID to be 'auditd-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure the Linux Audit Framework"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
