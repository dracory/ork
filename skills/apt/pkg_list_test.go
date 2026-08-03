package apt

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/types"
)

// TestPkgList_Run_DryRun verifies that dry-run mode correctly handles apt list.
func TestPkgList_Run_DryRun(t *testing.T) {
	pb := NewPkgList()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// List is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would list installed packages"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPkgList_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPkgList_Run_NotDryRun(t *testing.T) {
	pb := NewPkgList()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would list installed packages" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// List is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestPkgList_Check verifies that Check returns false for read-only operation.
func TestPkgList_Check(t *testing.T) {
	pb := NewPkgList()

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

// TestPkgList_NewPkgList verifies that NewPkgList creates a properly configured skill.
func TestPkgList_NewPkgList(t *testing.T) {
	pb := NewPkgList()

	if pb.GetID() != "apt-list" {
		t.Errorf("Expected ID to be 'apt-list', got '%s'", pb.GetID())
	}

	expectedDescription := "List installed packages (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPkgList_Run_WithMock demonstrates using the mock SSH client for testing.
// This test verifies the actual command execution without requiring a real SSH server.
func TestPkgList_Run_WithMock(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt list --installed 2>/dev/null | tail -n +2", "nginx/stable,now 1.18.0-0ubuntu1 amd64 [installed]\ncurl/stable,now 7.68.0-1ubuntu2 amd64 [installed]")

	pb := NewPkgList()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("apt list --installed 2>/dev/null | tail -n +2")
	test.AssertResultMessageContains(result, "2 installed packages")
}

// TestPkgList_Run_WithMockNoPackages demonstrates testing when no packages match.
func TestPkgList_Run_WithMockNoPackages(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt list --installed 2>/dev/null | tail -n +2", "")

	pb := NewPkgList()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "No installed packages matched")
}

// TestPkgList_Run_WithMockFiltered verifies filtering by a single package name.
func TestPkgList_Run_WithMockFiltered(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/", "nginx/stable,now 1.18.0-0ubuntu1 amd64 [installed]")

	pb := NewPkgList().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/")
	test.AssertResultMessageContains(result, "1 installed packages")
}

// TestPkgList_Run_WithMockError demonstrates testing error scenarios.
func TestPkgList_Run_WithMockError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("apt list --installed 2>/dev/null | tail -n +2", fmt.Errorf("command not found"))

	pb := NewPkgList()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultError(result)
	test.AssertErrorContains(result.Error, "failed to list installed packages")
}

// TestPkgList_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete PkgList type.
func TestPkgList_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgList()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*PkgList); !ok {
		t.Error("SetArgs should return *PkgList, not just RunnableInterface")
	}
}

// TestPkgList_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete PkgList type.
func TestPkgList_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgList()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*PkgList); !ok {
		t.Error("SetArg should return *PkgList, not just RunnableInterface")
	}
}

// TestPkgList_SetID_ReturnsConcreteType verifies that SetID returns the concrete PkgList type.
func TestPkgList_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgList()

	result := skill.SetID("custom-id")

	if _, ok := result.(*PkgList); !ok {
		t.Error("SetID should return *PkgList, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPkgList_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete PkgList type.
func TestPkgList_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgList()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*PkgList); !ok {
		t.Error("SetDescription should return *PkgList, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPkgList_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete PkgList type.
func TestPkgList_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgList()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*PkgList); !ok {
		t.Error("SetTimeout should return *PkgList, not just RunnableInterface")
	}
}

// TestPkgList_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPkgList_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPkgList().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*PkgList); !ok {
		t.Error("Method chaining should preserve *PkgList type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
