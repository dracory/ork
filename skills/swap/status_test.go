package swap

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestSwapStatus_Run_DryRun verifies that dry-run mode correctly handles swap status.
func TestSwapStatus_Run_DryRun(t *testing.T) {
	pb := NewSwapStatus()

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

	expectedMessage := "Would check swap status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSwapStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSwapStatus_Run_NotDryRun(t *testing.T) {
	pb := NewSwapStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check swap status" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestSwapStatus_Check verifies that Check returns false for read-only operation.
func TestSwapStatus_Check(t *testing.T) {
	pb := NewSwapStatus()

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

// TestSwapStatus_NewSwapStatus verifies that NewSwapStatus creates a properly configured skill.
func TestSwapStatus_NewSwapStatus(t *testing.T) {
	pb := NewSwapStatus()

	if pb.GetID() != "swap-status" {
		t.Errorf("Expected ID to be 'swap-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Show swap status and usage"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
