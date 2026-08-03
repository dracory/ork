package apt

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/types"
)

// TestIsPkgInstalled_Run_DryRun verifies that dry-run mode correctly handles the check.
func TestIsPkgInstalled_Run_DryRun(t *testing.T) {
	pb := NewIsPkgInstalled().SetPackage("nginx")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Read-only operation, Changed should be false even in dry-run
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

// TestIsPkgInstalled_Run_NoPackage verifies that missing ArgPackage returns an error.
func TestIsPkgInstalled_Run_NoPackage(t *testing.T) {
	pb := NewIsPkgInstalled()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no package specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no package is specified")
	}
}

// TestIsPkgInstalled_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestIsPkgInstalled_Run_NotDryRun(t *testing.T) {
	pb := NewIsPkgInstalled().SetPackage("nginx")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode it will try SSH commands and likely fail without a real server.
	if result.Message == "Would check if package 'nginx' is installed" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestIsPkgInstalled_Check_NoPackage verifies that Check returns an error when no package is set.
func TestIsPkgInstalled_Check_NoPackage(t *testing.T) {
	pb := NewIsPkgInstalled()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	_, err := pb.Check()

	if err == nil {
		t.Error("Expected an error when no package is specified")
	}
}

// TestIsPkgInstalled_Check_DryRun verifies that Check returns false in dry-run mode.
func TestIsPkgInstalled_Check_DryRun(t *testing.T) {
	pb := NewIsPkgInstalled().SetPackage("nginx")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	installed, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check in dry-run mode, got: %v", err)
	}

	if installed {
		t.Error("Expected Check to return false in dry-run mode")
	}
}

// TestIsPkgInstalled_NewIsPkgInstalled verifies that NewIsPkgInstalled creates a properly configured skill.
func TestIsPkgInstalled_NewIsPkgInstalled(t *testing.T) {
	pb := NewIsPkgInstalled()

	if pb.GetID() != "apt-is-installed" {
		t.Errorf("Expected ID to be 'apt-is-installed', got '%s'", pb.GetID())
	}

	expectedDescription := "Check if a package is installed (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestIsPkgInstalled_Run_WithMock_Installed verifies detecting an installed package.
func TestIsPkgInstalled_Run_WithMock_Installed(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/",
		"nginx/jammy,now 1.18.0-0ubuntu1 amd64 [installed]",
	)

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/")
	test.AssertResultMessageContains(result, "package 'nginx' is installed")

	if result.Details["installed"] != "true" {
		t.Errorf("Expected installed='true', got '%s'", result.Details["installed"])
	}
	if result.Details["version"] != "1.18.0-0ubuntu1" {
		t.Errorf("Expected version='1.18.0-0ubuntu1', got '%s'", result.Details["version"])
	}
	if result.Details["architecture"] != "amd64" {
		t.Errorf("Expected architecture='amd64', got '%s'", result.Details["architecture"])
	}
	if result.Details["suite"] != "jammy,now" {
		t.Errorf("Expected suite='jammy,now', got '%s'", result.Details["suite"])
	}
}

// TestIsPkgInstalled_Run_WithMock_NotInstalled verifies detecting a missing package.
func TestIsPkgInstalled_Run_WithMock_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nonexistent'/",
		"",
	)

	pb := NewIsPkgInstalled().SetPackage("nonexistent")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "package 'nonexistent' is not installed")

	if result.Details["installed"] != "false" {
		t.Errorf("Expected installed='false', got '%s'", result.Details["installed"])
	}
}

// TestIsPkgInstalled_Run_WithMock_MultipleMatches verifies exact name matching when
// grep returns multiple lines (e.g. "nginx" matches "nginx-common" too).
func TestIsPkgInstalled_Run_WithMock_MultipleMatches(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/",
		"nginx-common/jammy,now 1.18.0-0ubuntu1 all [installed]\nnginx/jammy,now 1.18.0-0ubuntu1 amd64 [installed]",
	)

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertResultMessageContains(result, "package 'nginx' is installed")

	// Should pick the exact "nginx" line, not "nginx-common"
	if result.Details["version"] != "1.18.0-0ubuntu1" {
		t.Errorf("Expected version='1.18.0-0ubuntu1', got '%s'", result.Details["version"])
	}
	if result.Details["architecture"] != "amd64" {
		t.Errorf("Expected architecture='amd64' (exact match), got '%s'", result.Details["architecture"])
	}
}

// TestIsPkgInstalled_Run_WithMock_NoExactMatch verifies that a package whose
// name is a substring of another (e.g. "nginx" vs "nginx-common") is not
// falsely reported as installed when only the longer-named package exists.
func TestIsPkgInstalled_Run_WithMock_NoExactMatch(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/",
		"nginx-common/jammy,now 1.18.0-0ubuntu1 all [installed]",
	)

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

// TestIsPkgInstalled_Run_WithMockError verifies error handling.
func TestIsPkgInstalled_Run_WithMockError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/",
		fmt.Errorf("command not found"),
	)

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultError(result)
	test.AssertErrorContains(result.Error, "failed to check if package 'nginx' is installed")
}

// TestIsPkgInstalled_Check_WithMock_Installed verifies Check returns true when installed.
func TestIsPkgInstalled_Check_WithMock_Installed(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nginx'/",
		"nginx/jammy,now 1.18.0-0ubuntu1 amd64 [installed]",
	)

	pb := NewIsPkgInstalled().SetPackage("nginx")
	pb.SetNodeConfig(test.Config())
	installed, err := pb.Check()

	test.AssertNoError(err)
	if !installed {
		t.Error("Expected Check to return true when package is installed")
	}
}

// TestIsPkgInstalled_Check_WithMock_NotInstalled verifies Check returns false when missing.
func TestIsPkgInstalled_Check_WithMock_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand(
		"apt list --installed 2>/dev/null | tail -n +2 | grep -i -- ^'nonexistent'/",
		"",
	)

	pb := NewIsPkgInstalled().SetPackage("nonexistent")
	pb.SetNodeConfig(test.Config())
	installed, err := pb.Check()

	test.AssertNoError(err)
	if installed {
		t.Error("Expected Check to return false when package is not installed")
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

// TestIsPkgInstalled_SetPackage verifies that SetPackage sets the package arg.
func TestIsPkgInstalled_SetPackage(t *testing.T) {
	skill := NewIsPkgInstalled().SetPackage("curl")

	if skill.GetArg(ArgPackage) != "curl" {
		t.Errorf("Expected package 'curl', got '%s'", skill.GetArg(ArgPackage))
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

// TestParseAptListLine verifies parsing of apt list output lines.
func TestParseAptListLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected map[string]string
	}{
		{
			name: "standard line",
			line: "nginx/jammy,now 1.18.0-0ubuntu1 amd64 [installed]",
			expected: map[string]string{
				"version":      "1.18.0-0ubuntu1",
				"architecture": "amd64",
				"suite":        "jammy,now",
			},
		},
		{
			name: "architecture-independent package",
			line: "adduser/jammy,now 3.118 all [installed]",
			expected: map[string]string{
				"version":      "3.118",
				"architecture": "all",
				"suite":        "jammy,now",
			},
		},
		{
			name: "updates suite",
			line: "apt/jammy-updates,now 2.4.9 amd64 [installed]",
			expected: map[string]string{
				"version":      "2.4.9",
				"architecture": "amd64",
				"suite":        "jammy-updates,now",
			},
		},
		{
			name: "empty line",
			line: "",
			expected: map[string]string{
				"version":      "",
				"architecture": "",
				"suite":        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAptListLine(tt.line)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s='%s', got '%s'", k, v, result[k])
				}
			}
		})
	}
}
