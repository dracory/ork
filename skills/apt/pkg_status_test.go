package apt

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/types"
)

// TestPkgStatus_Run_DryRun verifies that dry-run mode correctly handles apt status.
func TestPkgStatus_Run_DryRun(t *testing.T) {
	pb := NewPkgStatus()

	cfg := types.NodeConfig{
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

// TestPkgStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPkgStatus_Run_NotDryRun(t *testing.T) {
	pb := NewPkgStatus()

	cfg := types.NodeConfig{
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

// TestPkgStatus_Check verifies that Check returns false for read-only operation.
func TestPkgStatus_Check(t *testing.T) {
	pb := NewPkgStatus()

	cfg := types.NodeConfig{
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

// TestPkgStatus_NewPkgStatus verifies that NewPkgStatus creates a properly configured skill.
func TestPkgStatus_NewPkgStatus(t *testing.T) {
	pb := NewPkgStatus()

	if pb.GetID() != "apt-status" {
		t.Errorf("Expected ID to be 'apt-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Show available package updates (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPkgStatus_Run_WithMock demonstrates using the mock SSH client for testing.
// This test verifies the actual command execution without requiring a real SSH server.
func TestPkgStatus_Run_WithMock(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt list --upgradable 2>/dev/null | tail -n +2", "nginx/stable 1.18.0-0ubuntu1 amd64 [upgradable from 1.17.0-0ubuntu1]")

	pb := NewPkgStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("apt list --upgradable 2>/dev/null | tail -n +2")
	test.AssertResultMessageContains(result, "1 packages available for upgrade")
}

// TestPkgStatus_Run_WithMockNoUpdates demonstrates testing when no updates are available.
func TestPkgStatus_Run_WithMockNoUpdates(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt list --upgradable 2>/dev/null | tail -n +2", "")

	pb := NewPkgStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "All packages are up to date")
}

// TestPkgStatus_Run_WithMockError demonstrates testing error scenarios.
func TestPkgStatus_Run_WithMockError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("apt list --upgradable 2>/dev/null | tail -n +2", fmt.Errorf("failed to list packages"))

	pb := NewPkgStatus()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultError(result)
	test.AssertErrorContains(result.Error, "failed to list upgradable packages")
}

// TestPkgStatus_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete PkgStatus type.
func TestPkgStatus_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgStatus()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*PkgStatus); !ok {
		t.Error("SetArgs should return *PkgStatus, not just RunnableInterface")
	}
}

// TestPkgStatus_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete PkgStatus type.
func TestPkgStatus_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgStatus()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*PkgStatus); !ok {
		t.Error("SetArg should return *PkgStatus, not just RunnableInterface")
	}
}

// TestPkgStatus_SetID_ReturnsConcreteType verifies that SetID returns the concrete PkgStatus type.
func TestPkgStatus_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgStatus()

	result := skill.SetID("custom-id")

	if _, ok := result.(*PkgStatus); !ok {
		t.Error("SetID should return *PkgStatus, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPkgStatus_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete PkgStatus type.
func TestPkgStatus_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgStatus()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*PkgStatus); !ok {
		t.Error("SetDescription should return *PkgStatus, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPkgStatus_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete PkgStatus type.
func TestPkgStatus_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgStatus()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*PkgStatus); !ok {
		t.Error("SetTimeout should return *PkgStatus, not just RunnableInterface")
	}
}

// TestPkgStatus_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPkgStatus_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPkgStatus().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*PkgStatus); !ok {
		t.Error("Method chaining should preserve *PkgStatus type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
