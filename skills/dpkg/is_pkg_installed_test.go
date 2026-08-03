package dpkg

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/types"
)

// TestIsPkgInstalled_Run_DryRun verifies that dry-run mode correctly handles dpkg-is-installed.
func TestIsPkgInstalled_Run_DryRun(t *testing.T) {
	pb := NewIsPkgInstalled().SetPackage("nginx")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would check if package 'nginx' is installed"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestIsPkgInstalled_Run_MissingPackage verifies error when no package is specified.
func TestIsPkgInstalled_Run_MissingPackage(t *testing.T) {
	pb := NewIsPkgInstalled()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
		Args:   map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error when no package is specified")
	}

	expectedMessage := "No package specified"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}
}

// TestIsPkgInstalled_Check_MissingPackage verifies Check returns error when no package is specified.
func TestIsPkgInstalled_Check_MissingPackage(t *testing.T) {
	pb := NewIsPkgInstalled()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
		Args:   map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	_, err := pb.Check()

	if err == nil {
		t.Error("Expected error from Check when no package is specified")
	}
}

// TestIsPkgInstalled_Run_WithMock_Installed verifies detecting an installed package.
func TestIsPkgInstalled_Run_WithMock_Installed(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("dpkg-query -W -f='${Status}' -- 'nginx' 2>/dev/null", "install ok installed")

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("dpkg-query -W -f='${Status}' -- 'nginx' 2>/dev/null")
	test.AssertResultMessageContains(result, "package 'nginx' is installed")

	if result.Details["installed"] != "true" {
		t.Errorf("Expected installed='true', got '%s'", result.Details["installed"])
	}
}

// TestIsPkgInstalled_Run_WithMock_NotInstalled verifies detecting a missing package.
func TestIsPkgInstalled_Run_WithMock_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("dpkg-query -W -f='${Status}' -- 'nginx' 2>/dev/null", fmt.Errorf("exit status 1"))

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "package 'nginx' is not installed")

	if result.Details["installed"] != "false" {
		t.Errorf("Expected installed='false', got '%s'", result.Details["installed"])
	}
}

// TestIsPkgInstalled_Check_WithMock_Installed verifies Check returns true for installed package.
func TestIsPkgInstalled_Check_WithMock_Installed(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("dpkg-query -W -f='${Status}' -- 'nginx' 2>/dev/null", "install ok installed")

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())

	isInstalled, err := pb.Check()

	test.AssertNoError(err)
	if !isInstalled {
		t.Error("Expected Check to return true for installed package")
	}
}

// TestIsPkgInstalled_Check_WithMock_NotInstalled verifies Check returns false for missing package.
func TestIsPkgInstalled_Check_WithMock_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("dpkg-query -W -f='${Status}' -- 'nginx' 2>/dev/null", fmt.Errorf("exit status 1"))

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())

	isInstalled, err := pb.Check()

	test.AssertNoError(err)
	if isInstalled {
		t.Error("Expected Check to return false for not-installed package")
	}
}

// TestIsPkgInstalled_NewIsPkgInstalled verifies that NewIsPkgInstalled creates a properly configured skill.
func TestIsPkgInstalled_NewIsPkgInstalled(t *testing.T) {
	pb := NewIsPkgInstalled()

	if pb.GetID() != "dpkg-is-installed" {
		t.Errorf("Expected ID to be 'dpkg-is-installed', got '%s'", pb.GetID())
	}

	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

// TestIsPkgInstalled_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete type.
func TestIsPkgInstalled_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewIsPkgInstalled()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*IsPkgInstalled); !ok {
		t.Error("SetArgs should return *IsPkgInstalled, not just RunnableInterface")
	}
}

// TestIsPkgInstalled_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete type.
func TestIsPkgInstalled_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewIsPkgInstalled()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*IsPkgInstalled); !ok {
		t.Error("SetArg should return *IsPkgInstalled, not just RunnableInterface")
	}
}

// TestIsPkgInstalled_SetID_ReturnsConcreteType verifies that SetID returns the concrete type.
func TestIsPkgInstalled_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewIsPkgInstalled()

	result := skill.SetID("custom-id")

	if _, ok := result.(*IsPkgInstalled); !ok {
		t.Error("SetID should return *IsPkgInstalled, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestIsPkgInstalled_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete type.
func TestIsPkgInstalled_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewIsPkgInstalled()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*IsPkgInstalled); !ok {
		t.Error("SetDescription should return *IsPkgInstalled, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestIsPkgInstalled_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete type.
func TestIsPkgInstalled_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewIsPkgInstalled()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*IsPkgInstalled); !ok {
		t.Error("SetTimeout should return *IsPkgInstalled, not just RunnableInterface")
	}
}

// TestIsPkgInstalled_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestIsPkgInstalled_MethodChaining_PreservesType(t *testing.T) {
	skill := NewIsPkgInstalled().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*IsPkgInstalled); !ok {
		t.Error("Method chaining should preserve *IsPkgInstalled type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
