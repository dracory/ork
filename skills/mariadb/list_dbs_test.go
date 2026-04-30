package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestListDBs_Run_DryRun verifies that dry-run mode correctly handles database listing.
func TestListDBs_Run_DryRun(t *testing.T) {
	pb := NewListDBs()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// List is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would list all databases"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestListDBs_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestListDBs_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewListDBs()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing root-password even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestListDBs_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestListDBs_Run_NotDryRun(t *testing.T) {
	pb := NewListDBs()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would list all databases" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// List is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestListDBs_Check verifies that Check returns false for read-only operation.
func TestListDBs_Check(t *testing.T) {
	pb := NewListDBs()

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

// TestListDBs_NewListDBs verifies that NewListDBs creates a properly configured skill.
func TestListDBs_NewListDBs(t *testing.T) {
	pb := NewListDBs()

	if pb.GetID() != "mariadb-list-dbs" {
		t.Errorf("Expected ID to be 'mariadb-list-dbs', got '%s'", pb.GetID())
	}

	expectedDescription := "Display all databases in the MariaDB server (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestListDBs_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete ListDBs type.
func TestListDBs_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewListDBs()
	args := map[string]string{"root-password": "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*ListDBs); !ok {
		t.Error("SetArgs should return *ListDBs, not just RunnableInterface")
	}
}

// TestListDBs_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete ListDBs type.
func TestListDBs_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewListDBs()

	result := skill.SetArg("root-password", "testpass")

	if _, ok := result.(*ListDBs); !ok {
		t.Error("SetArg should return *ListDBs, not just RunnableInterface")
	}
}

// TestListDBs_SetID_ReturnsConcreteType verifies that SetID returns the concrete ListDBs type.
func TestListDBs_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewListDBs()

	result := skill.SetID("custom-id")

	if _, ok := result.(*ListDBs); !ok {
		t.Error("SetID should return *ListDBs, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestListDBs_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete ListDBs type.
func TestListDBs_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewListDBs()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*ListDBs); !ok {
		t.Error("SetDescription should return *ListDBs, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestListDBs_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete ListDBs type.
func TestListDBs_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewListDBs()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*ListDBs); !ok {
		t.Error("SetTimeout should return *ListDBs, not just RunnableInterface")
	}
}

// TestListDBs_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestListDBs_MethodChaining_PreservesType(t *testing.T) {
	skill := NewListDBs().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("root-password", "testpass").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*ListDBs); !ok {
		t.Error("Method chaining should preserve *ListDBs type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
