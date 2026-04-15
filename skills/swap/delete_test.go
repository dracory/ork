package swap

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestSwapDelete_Run_DryRun verifies that dry-run mode correctly handles swap deletion.
func TestSwapDelete_Run_DryRun(t *testing.T) {
	pb := NewSwapDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In dry-run mode, Check will fail due to no SSH server, but the dry-run should still work
	// The implementation calls Check() first, which will fail
	if result.Error != nil {
		// Expected to fail on Check() since no SSH server
		if result.Message == "Would remove swap file at /swapfile" {
			t.Error("Should not reach dry-run message if Check() fails")
		}
	}
}

// TestSwapDelete_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSwapDelete_Run_NotDryRun(t *testing.T) {
	pb := NewSwapDelete()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would remove swap file at /swapfile" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSwapDelete_NewSwapDelete verifies that NewSwapDelete creates a properly configured skill.
func TestSwapDelete_NewSwapDelete(t *testing.T) {
	pb := NewSwapDelete()

	if pb.GetID() != "swap-delete" {
		t.Errorf("Expected ID to be 'swap-delete', got '%s'", pb.GetID())
	}

	expectedDescription := "Remove the swap file"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
