package ping

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestPing_Run_DryRun verifies that dry-run mode correctly handles ping.
func TestPing_Run_DryRun(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Ping is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would ping: test.example.com"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPing_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPing_Run_NotDryRun(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would ping: test.example.com" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Ping is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestPing_Check verifies that Check returns false for read-only operation.
func TestPing_Check(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		Logger:  slog.Default(),
		SSHHost: "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	// Since there's no real SSH server, we expect an error
	if err == nil {
		t.Log("SSH connection succeeded in test environment")
	}

	// Ping is read-only, so should always return false
	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestPing_NewPing verifies that NewPing creates a properly configured skill.
func TestPing_NewPing(t *testing.T) {
	pb := NewPing()

	if pb.GetID() != "ping" {
		t.Errorf("Expected ID to be 'ping', got '%s'", pb.GetID())
	}

	expectedDescription := "Check SSH connectivity and show server uptime/load"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPing_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Ping type.
func TestPing_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPing()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Ping); !ok {
		t.Error("SetArgs should return *Ping, not just RunnableInterface")
	}
}

// TestPing_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Ping type.
func TestPing_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPing()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Ping); !ok {
		t.Error("SetArg should return *Ping, not just RunnableInterface")
	}
}

// TestPing_SetID_ReturnsConcreteType verifies that SetID returns the concrete Ping type.
func TestPing_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPing()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Ping); !ok {
		t.Error("SetID should return *Ping, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPing_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Ping type.
func TestPing_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPing()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Ping); !ok {
		t.Error("SetDescription should return *Ping, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPing_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Ping type.
func TestPing_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPing()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*Ping); !ok {
		t.Error("SetTimeout should return *Ping, not just RunnableInterface")
	}
}

// TestPing_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestPing_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPing().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*Ping); !ok {
		t.Error("Method chaining should preserve *Ping type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
