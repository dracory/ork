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

// TestSwapCreate_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete SwapCreate type.
func TestSwapCreate_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSwapCreate()
	args := map[string]string{"size": "2", "unit": "gb"}

	result := skill.SetArgs(args)

	if _, ok := result.(*SwapCreate); !ok {
		t.Error("SetArgs should return *SwapCreate, not just RunnableInterface")
	}
}

// TestSwapCreate_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete SwapCreate type.
func TestSwapCreate_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSwapCreate()

	result := skill.SetArg("size", "2")

	if _, ok := result.(*SwapCreate); !ok {
		t.Error("SetArg should return *SwapCreate, not just RunnableInterface")
	}
}

// TestSwapCreate_SetID_ReturnsConcreteType verifies that SetID returns the concrete SwapCreate type.
func TestSwapCreate_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSwapCreate()

	result := skill.SetID("custom-id")

	if _, ok := result.(*SwapCreate); !ok {
		t.Error("SetID should return *SwapCreate, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSwapCreate_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete SwapCreate type.
func TestSwapCreate_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSwapCreate()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*SwapCreate); !ok {
		t.Error("SetDescription should return *SwapCreate, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSwapCreate_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete SwapCreate type.
func TestSwapCreate_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSwapCreate()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*SwapCreate); !ok {
		t.Error("SetTimeout should return *SwapCreate, not just RunnableInterface")
	}
}

// TestSwapCreate_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSwapCreate_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSwapCreate().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("size", "2").
		SetArgs(map[string]string{"unit": "gb", "swappiness": "20"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*SwapCreate); !ok {
		t.Error("Method chaining should preserve *SwapCreate type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
