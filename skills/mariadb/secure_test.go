package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestSecure_Run_DryRun verifies that dry-run mode correctly handles MariaDB security hardening.
func TestSecure_Run_DryRun(t *testing.T) {
	pb := NewSecure()
	pb.SetArg("root-password", "testpass123")

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would secure MariaDB installation"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSecure_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestSecure_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewSecure()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing root-password even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestSecure_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSecure_Run_NotDryRun(t *testing.T) {
	pb := NewSecure()
	pb.SetArg("root-password", "testpass123")

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would secure MariaDB installation" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSecure_NewSecure verifies that NewSecure creates a properly configured skill.
func TestSecure_NewSecure(t *testing.T) {
	pb := NewSecure()

	if pb.GetID() != "mariadb-secure" {
		t.Errorf("Expected ID to be 'mariadb-secure', got '%s'", pb.GetID())
	}

	expectedDescription := "Perform basic security hardening on MariaDB installation"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
