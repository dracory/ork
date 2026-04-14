package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestEnableEncryption_Run_DryRun verifies that dry-run mode correctly handles encryption enablement.
func TestEnableEncryption_Run_DryRun(t *testing.T) {
	pb := NewEnableEncryption()

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

	expectedMessage := "Would enable MariaDB encryption at rest"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestEnableEncryption_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestEnableEncryption_Run_NotDryRun(t *testing.T) {
	pb := NewEnableEncryption()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would enable MariaDB encryption at rest" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestEnableEncryption_NewEnableEncryption verifies that NewEnableEncryption creates a properly configured playbook.
func TestEnableEncryption_NewEnableEncryption(t *testing.T) {
	pb := NewEnableEncryption()

	if pb.GetID() != "mariadb-enable-encryption" {
		t.Errorf("Expected ID to be 'mariadb-enable-encryption', got '%s'", pb.GetID())
	}

	expectedDescription := "Enable data-at-rest encryption for MariaDB"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
