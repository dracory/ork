package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestInstall_Run_DryRun verifies that dry-run mode correctly handles MariaDB installation.
func TestInstall_Run_DryRun(t *testing.T) {
	pb := NewInstall()

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

	expectedMessage := "Would install and configure MariaDB"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_Run_DryRun_WithPassword verifies dry-run with root password.
func TestInstall_Run_DryRun_WithPassword(t *testing.T) {
	pb := NewInstall()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure MariaDB"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestInstall_Run_NotDryRun(t *testing.T) {
	pb := NewInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure MariaDB" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestInstall_NewInstall verifies that NewInstall creates a properly configured skill.
func TestInstall_NewInstall(t *testing.T) {
	pb := NewInstall()

	if pb.GetID() != "mariadb-install" {
		t.Errorf("Expected ID to be 'mariadb-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure MariaDB database server"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Install type.
func TestInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	args := map[string]string{"root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Install); !ok {
		t.Error("SetArgs should return *Install, not just RunnableInterface")
	}
}

// TestInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Install type.
func TestInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()

	result := skill.SetArg("root-password", "testpass")

	if _, ok := result.(*Install); !ok {
		t.Error("SetArg should return *Install, not just RunnableInterface")
	}
}

// TestInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete Install type.
func TestInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Install); !ok {
		t.Error("SetID should return *Install, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Install type.
func TestInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Install); !ok {
		t.Error("SetDescription should return *Install, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Install type.
func TestInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Install); !ok {
		t.Error("SetTimeout should return *Install, not just RunnableInterface")
	}
}

// TestInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("root-password", "testpass").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Install); !ok {
		t.Error("Method chaining should preserve *Install type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
