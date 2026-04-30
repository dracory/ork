package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestPurge_Run_DryRun verifies that dry-run mode correctly handles MariaDB purge.
func TestPurge_Run_DryRun(t *testing.T) {
	pb := NewPurge()

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

	expectedMessage := "Would purge MariaDB"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPurge_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPurge_Run_NotDryRun(t *testing.T) {
	pb := NewPurge()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would purge MariaDB" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestPurge_NewPurge verifies that NewPurge creates a properly configured skill.
func TestPurge_NewPurge(t *testing.T) {
	pb := NewPurge()

	if pb.GetID() != "mariadb-purge" {
		t.Errorf("Expected ID to be 'mariadb-purge', got '%s'", pb.GetID())
	}

	expectedDescription := "Remove MariaDB database server and all associated data"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPurge_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Purge type.
func TestPurge_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPurge()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Purge); !ok {
		t.Error("SetArgs should return *Purge, not just RunnableInterface")
	}
}

// TestPurge_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Purge type.
func TestPurge_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPurge()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Purge); !ok {
		t.Error("SetArg should return *Purge, not just RunnableInterface")
	}
}

// TestPurge_SetID_ReturnsConcreteType verifies that SetID returns the concrete Purge type.
func TestPurge_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPurge()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Purge); !ok {
		t.Error("SetID should return *Purge, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPurge_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Purge type.
func TestPurge_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPurge()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Purge); !ok {
		t.Error("SetDescription should return *Purge, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPurge_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Purge type.
func TestPurge_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPurge()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Purge); !ok {
		t.Error("SetTimeout should return *Purge, not just RunnableInterface")
	}
}

// TestPurge_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPurge_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPurge().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Purge); !ok {
		t.Error("Method chaining should preserve *Purge type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
