package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestEnableSSL_Run_DryRun verifies that dry-run mode correctly handles SSL enablement.
func TestEnableSSL_Run_DryRun(t *testing.T) {
	pb := NewEnableSSL()

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

	expectedMessage := "Would enable MariaDB SSL/TLS"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestEnableSSL_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestEnableSSL_Run_NotDryRun(t *testing.T) {
	pb := NewEnableSSL()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would enable MariaDB SSL/TLS" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestEnableSSL_NewEnableSSL verifies that NewEnableSSL creates a properly configured skill.
func TestEnableSSL_NewEnableSSL(t *testing.T) {
	pb := NewEnableSSL()

	if pb.GetID() != "mariadb-enable-ssl" {
		t.Errorf("Expected ID to be 'mariadb-enable-ssl', got '%s'", pb.GetID())
	}

	expectedDescription := "Enable SSL/TLS encryption for MariaDB connections"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
