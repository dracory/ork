package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestStatus_Run_DryRun verifies that dry-run mode correctly handles MariaDB status.
func TestStatus_Run_DryRun(t *testing.T) {
	pb := NewStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check MariaDB status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestStatus_Run_DryRun_WithPassword verifies dry-run with root password.
func TestStatus_Run_DryRun_WithPassword(t *testing.T) {
	pb := NewStatus()
	pb.SetArg("root-password", "testpass123")

	cfg := config.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Status is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check MariaDB status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestStatus_Run_NotDryRun(t *testing.T) {
	pb := NewStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check MariaDB status" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestStatus_Check verifies that Check returns false for read-only operation.
func TestStatus_Check(t *testing.T) {
	pb := NewStatus()

	cfg := config.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestStatus_NewStatus verifies that NewStatus creates a properly configured playbook.
func TestStatus_NewStatus(t *testing.T) {
	pb := NewStatus()

	if pb.GetID() != "mariadb-status" {
		t.Errorf("Expected ID to be 'mariadb-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Display MariaDB server status information (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
