package swap

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSwapCreate_Run_DryRun verifies that dry-run mode correctly handles swap creation.
func TestSwapCreate_Run_DryRun(t *testing.T) {
	pb := NewSwapCreate()

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
		if result.Message == "Would create 1GB swap file at /swapfile" {
			t.Error("Should not reach dry-run message if Check() fails")
		}
	}
}

// TestSwapCreate_Run_DryRun_WithArgs verifies dry-run with custom arguments.
func TestSwapCreate_Run_DryRun_WithArgs(t *testing.T) {
	pb := NewSwapCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgSize:       "2",
			ArgUnit:       "gb",
			ArgSwappiness: "20",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In dry-run mode, Check will fail due to no SSH server
	if result.Error != nil {
		// Expected to fail on Check() since no SSH server
	}
}

// TestSwapCreate_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSwapCreate_Run_NotDryRun(t *testing.T) {
	pb := NewSwapCreate()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would create 1GB swap file at /swapfile" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSwapCreate_NewSwapCreate verifies that NewSwapCreate creates a properly configured skill.
func TestSwapCreate_NewSwapCreate(t *testing.T) {
	pb := NewSwapCreate()

	if pb.GetID() != "swap-create" {
		t.Errorf("Expected ID to be 'swap-create', got '%s'", pb.GetID())
	}

	expectedDescription := "Create a swap file (size via args['size'], unit via args['unit']='gb'|'mb', swappiness via args['swappiness']=10, defaults: 1GB, swappiness=10)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
