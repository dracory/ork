package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestChangePort_Run_DryRun verifies that dry-run mode correctly handles MariaDB port change.
func TestChangePort_Run_DryRun(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3307")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would change MariaDB port to 3307"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestChangePort_Run_DryRun_NoPort verifies dry-run without port returns error.
func TestChangePort_Run_DryRun_NoPort(t *testing.T) {
	pb := NewChangePort()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing port even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing port")
	}

	if result.Message != "Port parameter is required" {
		t.Errorf("Expected message 'Port parameter is required', got '%s'", result.Message)
	}
}

// TestChangePort_Run_DryRun_InvalidPort verifies dry-run with invalid port returns error.
func TestChangePort_Run_DryRun_InvalidPort(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3306")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for invalid port (3306 is default) even in dry-run
	if result.Error == nil {
		t.Error("Expected error for invalid port")
	}

	if result.Message != "Invalid port number" {
		t.Errorf("Expected message 'Invalid port number', got '%s'", result.Message)
	}
}

// TestChangePort_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestChangePort_Run_NotDryRun(t *testing.T) {
	pb := NewChangePort()
	pb.SetArg("port", "3307")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would change MariaDB port to 3307" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestChangePort_NewChangePort verifies that NewChangePort creates a properly configured skill.
func TestChangePort_NewChangePort(t *testing.T) {
	pb := NewChangePort()

	if pb.GetID() != "mariadb-change-port" {
		t.Errorf("Expected ID to be 'mariadb-change-port', got '%s'", pb.GetID())
	}

	expectedDescription := "Change the MariaDB server port from default 3306"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
