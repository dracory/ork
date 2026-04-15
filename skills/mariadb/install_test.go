package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestInstall_Run_DryRun verifies that dry-run mode correctly handles MariaDB installation.
func TestInstall_Run_DryRun(t *testing.T) {
	pb := NewInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure MariaDB"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_Run_DryRun_WithPassword verifies dry-run with root password.
func TestInstall_Run_DryRun_WithPassword(t *testing.T) {
	pb := NewInstall()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure MariaDB"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestInstall_Run_NotDryRun(t *testing.T) {
	pb := NewInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure MariaDB" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestInstall_NewInstall verifies that NewInstall creates a properly configured skill.
func TestInstall_NewInstall(t *testing.T) {
	pb := NewInstall()

	if pb.GetID() != "mariadb-install" {
		t.Errorf("Expected ID to be 'mariadb-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure MariaDB database server"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
