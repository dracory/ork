package fail2ban

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestFail2banInstall_Run_DryRun verifies that dry-run mode correctly handles fail2ban installation.
func TestFail2banInstall_Run_DryRun(t *testing.T) {
	pb := NewFail2banInstall()

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

	expectedMessage := "Would install and enable fail2ban"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFail2banInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestFail2banInstall_Run_NotDryRun(t *testing.T) {
	pb := NewFail2banInstall()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and enable fail2ban" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestFail2banInstall_NewFail2banInstall verifies that NewFail2banInstall creates a properly configured playbook.
func TestFail2banInstall_NewFail2banInstall(t *testing.T) {
	pb := NewFail2banInstall()

	if pb.GetID() != "fail2ban-install" {
		t.Errorf("Expected ID to be 'fail2ban-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and enable fail2ban intrusion prevention system"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
