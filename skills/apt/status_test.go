package apt

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/internal/skilltest"
)

// TestAptStatus_Run_DryRun verifies that dry-run mode correctly handles apt status.
func TestAptStatus_Run_DryRun(t *testing.T) {
	pb := NewAptStatus()

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

	expectedMessage := "Would check for available package updates"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAptStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAptStatus_Run_NotDryRun(t *testing.T) {
	pb := NewAptStatus()

	cfg := config.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check for available package updates" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestAptStatus_Check verifies that Check returns false for read-only operation.
func TestAptStatus_Check(t *testing.T) {
	pb := NewAptStatus()

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

// TestAptStatus_NewAptStatus verifies that NewAptStatus creates a properly configured skill.
func TestAptStatus_NewAptStatus(t *testing.T) {
	pb := NewAptStatus()

	if pb.GetID() != "apt-status" {
		t.Errorf("Expected ID to be 'apt-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Show available package updates (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestAptStatus_Run_WithMock demonstrates using the mock SSH client for testing.
// This test verifies the actual command execution without requiring a real SSH server.
func TestAptStatus_Run_WithMock(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt-get update -qq", "")
	test.ExpectCommand("apt list --upgradable 2>/dev/null | tail -n +2", "nginx/stable 1.18.0-0ubuntu1 amd64 [upgradable from 1.17.0-0ubuntu1]")

	pb := NewAptStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("apt-get update -qq")
	test.AssertCommandRun("apt list --upgradable 2>/dev/null | tail -n +2")
	test.AssertResultMessageContains(result, "1 packages available for upgrade")
}

// TestAptStatus_Run_WithMockNoUpdates demonstrates testing when no updates are available.
func TestAptStatus_Run_WithMockNoUpdates(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt-get update -qq", "")
	test.ExpectCommand("apt list --upgradable 2>/dev/null | tail -n +2", "")

	pb := NewAptStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "All packages are up to date")
}

// TestAptStatus_Run_WithMockError demonstrates testing error scenarios.
func TestAptStatus_Run_WithMockError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("apt-get update -qq", fmt.Errorf("failed to lock apt directory"))

	pb := NewAptStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultError(result)
	test.AssertErrorContains(result.Error, "failed to update package lists")
}
