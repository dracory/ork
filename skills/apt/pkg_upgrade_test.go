package apt

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/dracory/ork/types"
)

// TestPkgUpgrade_Run_DryRun verifies that dry-run mode correctly handles apt upgrade.
func TestPkgUpgrade_Run_DryRun(t *testing.T) {
	pb := NewPkgUpgrade()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In dry-run mode, Check() returns (true, nil) without SSH, so Run() reaches its dry-run guard
	if result.Error != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", result.Error)
	}

	if !strings.HasPrefix(result.Message, "Would upgrade packages:") {
		t.Errorf("Expected dry-run message prefix 'Would upgrade packages:', got: %s", result.Message)
	}

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}
}

// TestPkgUpgrade_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPkgUpgrade_Run_NotDryRun(t *testing.T) {
	pb := NewPkgUpgrade()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if strings.HasPrefix(result.Message, "Would upgrade packages:") {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestPkgUpgrade_NewPkgUpgrade verifies that NewPkgUpgrade creates a properly configured skill.
func TestPkgUpgrade_NewPkgUpgrade(t *testing.T) {
	pb := NewPkgUpgrade()

	if pb.GetID() != "apt-upgrade" {
		t.Errorf("Expected ID to be 'apt-upgrade', got '%s'", pb.GetID())
	}

	expectedDescription := "Install available package updates (apt-get upgrade)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPkgUpgrade_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete PkgUpgrade type.
func TestPkgUpgrade_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgUpgrade()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*PkgUpgrade); !ok {
		t.Error("SetArgs should return *PkgUpgrade, not just RunnableInterface")
	}
}

// TestPkgUpgrade_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete PkgUpgrade type.
func TestPkgUpgrade_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgUpgrade()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*PkgUpgrade); !ok {
		t.Error("SetArg should return *PkgUpgrade, not just RunnableInterface")
	}
}

// TestPkgUpgrade_SetID_ReturnsConcreteType verifies that SetID returns the concrete PkgUpgrade type.
func TestPkgUpgrade_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgUpgrade()

	result := skill.SetID("custom-id")

	if _, ok := result.(*PkgUpgrade); !ok {
		t.Error("SetID should return *PkgUpgrade, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPkgUpgrade_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete PkgUpgrade type.
func TestPkgUpgrade_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgUpgrade()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*PkgUpgrade); !ok {
		t.Error("SetDescription should return *PkgUpgrade, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPkgUpgrade_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete PkgUpgrade type.
func TestPkgUpgrade_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPkgUpgrade()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*PkgUpgrade); !ok {
		t.Error("SetTimeout should return *PkgUpgrade, not just RunnableInterface")
	}
}

// TestPkgUpgrade_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPkgUpgrade_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPkgUpgrade().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*PkgUpgrade); !ok {
		t.Error("Method chaining should preserve *PkgUpgrade type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
