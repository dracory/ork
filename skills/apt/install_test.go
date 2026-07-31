package apt

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestAptInstall_Run_DryRun verifies that dry-run mode correctly handles apt install.
func TestAptInstall_Run_DryRun(t *testing.T) {
	pb := NewAptInstall()

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

// TestAptInstall_Run_NoPackages verifies that missing ArgPackages returns an error.
func TestAptInstall_Run_NoPackages(t *testing.T) {
	pb := NewAptInstall()

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

// TestAptInstall_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestAptInstall_Run_NotDryRun(t *testing.T) {
	pb := NewAptInstall()

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

// TestAptInstall_Check_DryRun verifies that Check returns true in dry-run mode.
func TestAptInstall_Check_DryRun(t *testing.T) {
	pb := NewAptInstall()

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

// TestAptInstall_Check_NoPackages verifies that Check returns an error when no packages are set.
func TestAptInstall_Check_NoPackages(t *testing.T) {
	pb := NewAptInstall()

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

// TestAptInstall_NewAptInstall verifies that NewAptInstall creates a properly configured skill.
func TestAptInstall_NewAptInstall(t *testing.T) {
	pb := NewAptInstall()

	if pb.GetID() != "apt-install" {
		t.Errorf("Expected ID to be 'apt-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install packages (apt-get install)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestAptInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete AptInstall type.
func TestAptInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewAptInstall()
	args := map[string]string{ArgPackages: "curl"}

	result := skill.SetArgs(args)

	if _, ok := result.(*AptInstall); !ok {
		t.Error("SetArgs should return *AptInstall, not just RunnableInterface")
	}
}

// TestAptInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete AptInstall type.
func TestAptInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewAptInstall()

	result := skill.SetArg(ArgPackages, "curl")

	if _, ok := result.(*AptInstall); !ok {
		t.Error("SetArg should return *AptInstall, not just RunnableInterface")
	}
}

// TestAptInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete AptInstall type.
func TestAptInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewAptInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*AptInstall); !ok {
		t.Error("SetID should return *AptInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestAptInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete AptInstall type.
func TestAptInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewAptInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*AptInstall); !ok {
		t.Error("SetDescription should return *AptInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestAptInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete AptInstall type.
func TestAptInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewAptInstall()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*AptInstall); !ok {
		t.Error("SetTimeout should return *AptInstall, not just RunnableInterface")
	}
}

// TestAptInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestAptInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewAptInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPackages, "nodejs npm").
		SetArgs(map[string]string{ArgPackages: "curl"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*AptInstall); !ok {
		t.Error("Method chaining should preserve *AptInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestAptInstall_SetPackages verifies that SetPackages sets the packages arg (variadic, joined with spaces).
func TestAptInstall_SetPackages(t *testing.T) {
	skill := NewAptInstall().SetPackages("nodejs", "npm")

	if skill.GetArg(ArgPackages) != "nodejs npm" {
		t.Errorf("Expected packages 'nodejs npm', got '%s'", skill.GetArg(ArgPackages))
	}
}

// TestAptInstall_SetPackages_Single verifies that SetPackages works with a single package.
func TestAptInstall_SetPackages_Single(t *testing.T) {
	skill := NewAptInstall().SetPackages("curl")

	if skill.GetArg(ArgPackages) != "curl" {
		t.Errorf("Expected packages 'curl', got '%s'", skill.GetArg(ArgPackages))
	}
}

// TestAptInstall_SetPackages_Chaining verifies that SetPackages chains with other setters.
func TestAptInstall_SetPackages_Chaining(t *testing.T) {
	skill := NewAptInstall().
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
