package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestAuditdInstall_Run_DryRun verifies that dry-run mode correctly handles auditd installation.
func TestAuditdInstall_Run_DryRun(t *testing.T) {
	pb := NewAuditdInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure auditd"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAuditdInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAuditdInstall_Run_NotDryRun(t *testing.T) {
	pb := NewAuditdInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure auditd" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAuditdInstall_NewAuditdInstall verifies that NewAuditdInstall creates a properly configured skill.
func TestAuditdInstall_NewAuditdInstall(t *testing.T) {
	pb := NewAuditdInstall()

	if pb.GetID() != "auditd-install" {
		t.Errorf("Expected ID to be 'auditd-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure the Linux Audit Framework"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestAuditdInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete AuditdInstall type.
func TestAuditdInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewAuditdInstall()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*AuditdInstall); !ok {
		t.Error("SetArgs should return *AuditdInstall, not just RunnableInterface")
	}
}

// TestAuditdInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete AuditdInstall type.
func TestAuditdInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewAuditdInstall()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*AuditdInstall); !ok {
		t.Error("SetArg should return *AuditdInstall, not just RunnableInterface")
	}
}

// TestAuditdInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete AuditdInstall type.
func TestAuditdInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewAuditdInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*AuditdInstall); !ok {
		t.Error("SetID should return *AuditdInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestAuditdInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete AuditdInstall type.
func TestAuditdInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewAuditdInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*AuditdInstall); !ok {
		t.Error("SetDescription should return *AuditdInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestAuditdInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete AuditdInstall type.
func TestAuditdInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewAuditdInstall()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*AuditdInstall); !ok {
		t.Error("SetTimeout should return *AuditdInstall, not just RunnableInterface")
	}
}

// TestAuditdInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestAuditdInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewAuditdInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*AuditdInstall); !ok {
		t.Error("Method chaining should preserve *AuditdInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
