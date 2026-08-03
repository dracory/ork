package apt

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestPkgInstall_Run_DryRun verifies that dry-run mode correctly handles apt install.
func TestPkgInstall_Run_DryRun(t *testing.T) {
	pb := NewPkgInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPackages, "nodejs npm")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install packages: nodejs npm"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPkgInstall_Run_NoPackages verifies that missing ArgPackages returns an error.
func TestPkgInstall_Run_NoPackages(t *testing.T) {
	pb := NewPkgInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no packages specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no packages are specified")
	}
}

// TestPkgInstall_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestPkgInstall_Run_NotDryRun(t *testing.T) {
	pb := NewPkgInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPackages, "nodejs npm")

	result := pb.Run()

	// In non-dry-run mode it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install packages: nodejs npm" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestPkgInstall_Check_DryRun verifies that Check returns true in dry-run mode.
func TestPkgInstall_Check_DryRun(t *testing.T) {
	pb := NewPkgInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPackages, "nodejs npm")

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check in dry-run mode, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true in dry-run mode")
	}
}

// TestPkgInstall_Check_NoPackages verifies that Check returns an error when no packages are set.
func TestPkgInstall_Check_NoPackages(t *testing.T) {
	pb := NewPkgInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	_, err := pb.Check()

	if err == nil {
		t.Error("Expected an error when no packages are specified")
	}
}

// TestPkgInstall_NewPkgInstall verifies that NewPkgInstall creates a properly configured skill.
func TestPkgInstall_NewPkgInstall(t *testing.T) {
	pb := NewPkgInstall()

	if pb.GetID() != "apt-install" {
		t.Errorf("Expected ID to be 'apt-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install packages (apt-get install)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPkgInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete PkgInstall type.
func TestPkgInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgInstall()
	args := map[string]string{ArgPackages: "curl"}

	result := skill.SetArgs(args)

	if _, ok := result.(*PkgInstall); !ok {
		t.Error("SetArgs should return *PkgInstall, not just RunnableInterface")
	}
}

// TestPkgInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete PkgInstall type.
func TestPkgInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgInstall()

	result := skill.SetArg(ArgPackages, "curl")

	if _, ok := result.(*PkgInstall); !ok {
		t.Error("SetArg should return *PkgInstall, not just RunnableInterface")
	}
}

// TestPkgInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete PkgInstall type.
func TestPkgInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*PkgInstall); !ok {
		t.Error("SetID should return *PkgInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPkgInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete PkgInstall type.
func TestPkgInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*PkgInstall); !ok {
		t.Error("SetDescription should return *PkgInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPkgInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete PkgInstall type.
func TestPkgInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgInstall()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*PkgInstall); !ok {
		t.Error("SetTimeout should return *PkgInstall, not just RunnableInterface")
	}
}

// TestPkgInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPkgInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPkgInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPackages, "nodejs npm").
		SetArgs(map[string]string{ArgPackages: "curl"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*PkgInstall); !ok {
		t.Error("Method chaining should preserve *PkgInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestPkgInstall_SetPackages verifies that SetPackages sets the packages arg (variadic, joined with spaces).
func TestPkgInstall_SetPackages(t *testing.T) {
	skill := NewPkgInstall().SetPackages("nodejs", "npm")

	if skill.GetArg(ArgPackages) != "nodejs npm" {
		t.Errorf("Expected packages 'nodejs npm', got '%s'", skill.GetArg(ArgPackages))
	}
}

// TestPkgInstall_SetPackages_Single verifies that SetPackages works with a single package.
func TestPkgInstall_SetPackages_Single(t *testing.T) {
	skill := NewPkgInstall().SetPackages("curl")

	if skill.GetArg(ArgPackages) != "curl" {
		t.Errorf("Expected packages 'curl', got '%s'", skill.GetArg(ArgPackages))
	}
}

// TestPkgInstall_SetPackages_Chaining verifies that SetPackages chains with other setters.
func TestPkgInstall_SetPackages_Chaining(t *testing.T) {
	skill := NewPkgInstall().
		SetPackages("nodejs", "npm", "curl").
		SetID("custom-id").
		SetDescription("custom description")

	if skill.GetArg(ArgPackages) != "nodejs npm curl" {
		t.Errorf("Expected packages 'nodejs npm curl', got '%s'", skill.GetArg(ArgPackages))
	}
	if skill.GetID() != "custom-id" {
		t.Error("Chaining should set ID")
	}
	if skill.GetDescription() != "custom description" {
		t.Error("Chaining should set description")
	}
}
