package fail2ban

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestFail2banInstall_Run_DryRun verifies that dry-run mode correctly handles fail2ban installation.
func TestFail2banInstall_Run_DryRun(t *testing.T) {
	pb := NewFail2banInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and enable fail2ban"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestFail2banInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestFail2banInstall_Run_NotDryRun(t *testing.T) {
	pb := NewFail2banInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and enable fail2ban" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestFail2banInstall_NewFail2banInstall verifies that NewFail2banInstall creates a properly configured skill.
func TestFail2banInstall_NewFail2banInstall(t *testing.T) {
	pb := NewFail2banInstall()

	if pb.GetID() != "fail2ban-install" {
		t.Errorf("Expected ID to be 'fail2ban-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and enable fail2ban intrusion prevention system"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestFail2banInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Fail2banInstall type.
func TestFail2banInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banInstall()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Fail2banInstall); !ok {
		t.Error("SetArgs should return *Fail2banInstall, not just RunnableInterface")
	}
}

// TestFail2banInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Fail2banInstall type.
func TestFail2banInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banInstall()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Fail2banInstall); !ok {
		t.Error("SetArg should return *Fail2banInstall, not just RunnableInterface")
	}
}

// TestFail2banInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete Fail2banInstall type.
func TestFail2banInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Fail2banInstall); !ok {
		t.Error("SetID should return *Fail2banInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestFail2banInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Fail2banInstall type.
func TestFail2banInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Fail2banInstall); !ok {
		t.Error("SetDescription should return *Fail2banInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestFail2banInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Fail2banInstall type.
func TestFail2banInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewFail2banInstall()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Fail2banInstall); !ok {
		t.Error("SetTimeout should return *Fail2banInstall, not just RunnableInterface")
	}
}

// TestFail2banInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestFail2banInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewFail2banInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Fail2banInstall); !ok {
		t.Error("Method chaining should preserve *Fail2banInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
