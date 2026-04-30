package fail2ban

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestFail2banStatus_Run_DryRun verifies that dry-run mode correctly handles fail2ban status.
func TestFail2banStatus_Run_DryRun(t *testing.T) {
	pb := NewFail2banStatus()

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

	expectedMessage := "Would check fail2ban status"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFail2banStatus_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestFail2banStatus_Run_NotDryRun(t *testing.T) {
	pb := NewFail2banStatus()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would check fail2ban status" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Status is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestFail2banStatus_Check verifies that Check returns false for read-only operation.
func TestFail2banStatus_Check(t *testing.T) {
	pb := NewFail2banStatus()

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

// TestFail2banStatus_NewFail2banStatus verifies that NewFail2banStatus creates a properly configured skill.
func TestFail2banStatus_NewFail2banStatus(t *testing.T) {
	pb := NewFail2banStatus()

	if pb.GetID() != "fail2ban-status" {
		t.Errorf("Expected ID to be 'fail2ban-status', got '%s'", pb.GetID())
	}

	expectedDescription := "Display fail2ban service status and SSH jail information (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFail2banStatus_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Fail2banStatus type.
func TestFail2banStatus_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banStatus()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Fail2banStatus); !ok {
		t.Error("SetArgs should return *Fail2banStatus, not just RunnableInterface")
	}
}

// TestFail2banStatus_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Fail2banStatus type.
func TestFail2banStatus_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banStatus()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Fail2banStatus); !ok {
		t.Error("SetArg should return *Fail2banStatus, not just RunnableInterface")
	}
}

// TestFail2banStatus_SetID_ReturnsConcreteType verifies that SetID returns the concrete Fail2banStatus type.
func TestFail2banStatus_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banStatus()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Fail2banStatus); !ok {
		t.Error("SetID should return *Fail2banStatus, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFail2banStatus_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Fail2banStatus type.
func TestFail2banStatus_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banStatus()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Fail2banStatus); !ok {
		t.Error("SetDescription should return *Fail2banStatus, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFail2banStatus_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Fail2banStatus type.
func TestFail2banStatus_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banStatus()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Fail2banStatus); !ok {
		t.Error("SetTimeout should return *Fail2banStatus, not just RunnableInterface")
	}
}

// TestFail2banStatus_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFail2banStatus_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFail2banStatus().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Fail2banStatus); !ok {
		t.Error("Method chaining should preserve *Fail2banStatus type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
